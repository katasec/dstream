package servicebus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/internal/types"
)

var log = logging.GetLogger()

// ServiceBusChangeDataTransport implements types.ServiceBusChangeDataTransport, transports change data messages to Azure Service Bus
type ServiceBusChangeDataTransport struct {
	connectionString string
	client           *azservicebus.Client
	sender           *azservicebus.Sender
	entityName       string
	isQueue          bool
	batchSize        int
	batchQueue       chan map[string]interface{}
}





// NewServiceBusChangeDataTransport creates a new ServiceBusChangeDataTransport with the provided connection string and entity name
func NewServiceBusChangeDataTransport(connectionString, entityName string, isQueue bool) (*ServiceBusChangeDataTransport, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Debug("Failed to create Service Bus client")
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	sender, err := client.NewSender(entityName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	p := &ServiceBusChangeDataTransport{
		connectionString: connectionString,
		client:           client,
		sender:           sender,
		entityName:       entityName,
		isQueue:          isQueue,
		batchSize:        100,
		batchQueue:       make(chan map[string]interface{}, 1000),
	}

	// Start processing messages in batches
	go p.processMessages()

	return p, nil
}

// Create creates a new transport for a specific destination
func (p *ServiceBusChangeDataTransport) Create(destination string) (types.ChangeDataTransport, error) {
	return NewServiceBusChangeDataTransport(p.connectionString, destination, p.isQueue)
}

// PublishBatch publishes a batch of messages to a topic or queue
func (p *ServiceBusChangeDataTransport) PublishBatch(ctx context.Context, messages []interface{}) error {
	if len(messages) == 0 {
		return nil // Nothing to publish
	}
	
	// Create a message batch
	messageBatch, err := p.sender.NewMessageBatch(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create message batch: %w", err)
	}
	
	// Convert messages to Service Bus messages and add them to the batch
	for _, message := range messages {
		var sbMessage *azservicebus.Message
		
		switch msg := message.(type) {
		case *azservicebus.ReceivedMessage:
			// Create a new message with the same body and properties
			sbMessage = &azservicebus.Message{
				Body:                 msg.Body,
				ApplicationProperties: msg.ApplicationProperties,
				ContentType:          msg.ContentType,
				CorrelationID:        msg.CorrelationID,
				MessageID:            &msg.MessageID,
				Subject:              msg.Subject,
				To:                   msg.To,
			}
		case []byte:
			// Create a message with the byte array body
			sbMessage = &azservicebus.Message{
				Body: msg,
			}
		default:
			return fmt.Errorf("unsupported message type: %T", message)
		}
		
		// Try to add the message to the batch
		if err := messageBatch.AddMessage(sbMessage, nil); err != nil {
			// If the batch is full, send it and create a new one
			if err == azservicebus.ErrMessageTooLarge {
				// If this is the first message and it's too large, it will never fit
				if messageBatch.NumMessages() == 0 {
					return fmt.Errorf("message too large for any batch: %w", err)
				}
				
				// Send the current batch
				log.Info("Batch full, sending current batch", "messageCount", messageBatch.NumMessages())
				if err := p.sender.SendMessageBatch(ctx, messageBatch, nil); err != nil {
					return fmt.Errorf("failed to send message batch: %w", err)
				}
				
				// Create a new batch
				messageBatch, err = p.sender.NewMessageBatch(ctx, nil)
				if err != nil {
					return fmt.Errorf("failed to create new message batch: %w", err)
				}
				
				// Try to add the message to the new batch
				if err := messageBatch.AddMessage(sbMessage, nil); err != nil {
					return fmt.Errorf("message too large for a new batch: %w", err)
				}
			} else {
				return fmt.Errorf("failed to add message to batch: %w", err)
			}
		}
	}
	
	// Send the final batch if it contains any messages
	if messageBatch.NumMessages() > 0 {
		log.Info("Sending final batch", "messageCount", messageBatch.NumMessages())
		if err := p.sender.SendMessageBatch(ctx, messageBatch, nil); err != nil {
			return fmt.Errorf("failed to send final message batch: %w", err)
		}
	}
	
	return nil
}

// PublishServiceBusMessage publishes a Service Bus message by wrapping it in a batch
// This is kept for backward compatibility with the ServiceBusChangeDataTransport interface
func (p *ServiceBusChangeDataTransport) PublishServiceBusMessage(ctx context.Context, message *azservicebus.ReceivedMessage) error {
	// Use the batch API for all publishing
	return p.PublishBatch(ctx, []interface{}{message})
}

// EnsureDestinationExists ensures that a topic or queue exists, creating it if necessary
func (p *ServiceBusChangeDataTransport) EnsureDestinationExists(entityName string) error {
	// Create an admin client
	adminClient, err := admin.NewClientFromConnectionString(p.connectionString, nil)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}

	// Create the topic if it doesn't exist
	if !p.isQueue {
		if err := CreateTopicIfNotExists(adminClient, entityName); err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}
	}

	return nil
}

// Close closes the transport and releases any resources
func (p *ServiceBusChangeDataTransport) Close() error {
	if p.sender != nil {
		return p.sender.Close(context.Background())
	}
	return nil
}

// processMessages reads from batchQueue and sends messages in batches
func (p *ServiceBusChangeDataTransport) processMessages() {
	var batch []interface{}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg := <-p.batchQueue:
			// Convert map to JSON bytes for the batch
			data, err := json.Marshal(msg)
			if err != nil {
				log.Error("Failed to marshal message", "error", err)
				continue
			}
			
			// Add the JSON bytes to the batch
			batch = append(batch, data)
			if len(batch) >= p.batchSize {
				p.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// sendBatch sends a batch of messages to the Service Bus topic
func (p *ServiceBusChangeDataTransport) sendBatch(batch []interface{}) {
	ctx := context.Background()
	
	// Use our new PublishBatch method which handles all the batch logic
	if err := p.PublishBatch(ctx, batch); err != nil {
		log.Error("Failed to publish batch", "error", err, "batchSize", len(batch))
	} else {
		log.Info("Successfully published batch", "batchSize", len(batch))
	}
}
