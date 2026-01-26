package account

// TopicTechnology identifies the underlying technology for a topic.
// This enables technology-agnostic topic handling in NeuronAccount.
type TopicTechnology string

// Supported topic technologies.
const (
	// TopicTechnologyHedera represents Hedera Consensus Service topics.
	TopicTechnologyHedera TopicTechnology = "hedera"

	// TopicTechnologyKafka represents Apache Kafka topics.
	TopicTechnologyKafka TopicTechnology = "kafka"

	// TopicTechnologyCustom represents a custom/third-party topic implementation.
	TopicTechnologyCustom TopicTechnology = "custom"
)

// String returns the string representation of the topic technology.
func (t TopicTechnology) String() string {
	return string(t)
}

// IsValid checks if the topic technology is a known type.
func (t TopicTechnology) IsValid() bool {
	switch t {
	case TopicTechnologyHedera, TopicTechnologyKafka, TopicTechnologyCustom:
		return true
	default:
		return false
	}
}
