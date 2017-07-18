package pubsub

// TopicID uniquely identifies a pubsub topic
type TopicID string

// Topic allows stores events that function can subsribe to.
type Topic struct {
	ID TopicID `json:"topicId"`
}
