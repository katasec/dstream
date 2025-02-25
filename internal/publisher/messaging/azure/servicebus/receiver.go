package servicebus

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// Receiver wraps azservicebus.Receiver for easier use
type Receiver struct {
	client   *azservicebus.Client
	receiver *azservicebus.Receiver
}

// NewReceiver creates a new Service Bus receiver
func NewReceiver(connectionString, queueName string) (*Receiver, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create receiver: %w", err)
	}

	return &Receiver{
		client:   client,
		receiver: receiver,
	}, nil
}

// ReceiveMessages receives messages from the queue
func (r *Receiver) ReceiveMessages(ctx context.Context, maxMessages int, options *azservicebus.ReceiveMessagesOptions) ([]*azservicebus.ReceivedMessage, error) {
	return r.receiver.ReceiveMessages(ctx, maxMessages, options)
}

// CompleteMessage marks a message as complete
func (r *Receiver) CompleteMessage(ctx context.Context, message *azservicebus.ReceivedMessage) error {
	return r.receiver.CompleteMessage(ctx, message, nil)
}

// Close closes the receiver
func (r *Receiver) Close(ctx context.Context) error {
	if err := r.receiver.Close(ctx); err != nil {
		return err
	}
	return r.client.Close(ctx)
}
