package cdc

// ChangeType represents the type of change detected in a table
type ChangeType string

const (
	// Insert represents a new row being added
	Insert ChangeType = "insert"
	// Update represents a row being modified
	Update ChangeType = "update"
	// Delete represents a row being removed
	Delete ChangeType = "delete"
)

// ChangeEvent represents a change detected in a table
type ChangeEvent struct {
	TableName  string                 `json:"table_name"`
	ChangeType ChangeType            `json:"change_type"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  string                 `json:"timestamp"`
	LSN       string                 `json:"lsn"`
}

// ChangePublisher is an interface for publishing CDC change messages
type ChangePublisher interface {
	// PublishChange publishes a change event to a queue or topic
	PublishChange(data map[string]interface{}) error

	// Close releases any resources used by the publisher
	Close() error
}
