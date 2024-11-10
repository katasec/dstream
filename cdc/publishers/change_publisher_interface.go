package publishers

// ChangePublisher is an interface for publishing CDC change messages
type ChangePublisher interface {
	PublishChange(data map[string]interface{})
}
