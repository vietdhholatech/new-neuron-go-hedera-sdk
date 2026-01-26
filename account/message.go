package account

import (
	"time"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// TopicMessage represents a message received from or sent to a topic.
// It is technology-agnostic and contains the common fields needed across
// all topic implementations (Hedera, Kafka, etc.).
type TopicMessage struct {
	// Payload is the raw message content.
	// The encoding/format is determined by the application layer.
	Payload []byte

	// Timestamp is when the message was produced/received.
	// For Hedera, this is the consensus timestamp.
	// For Kafka, this is the message timestamp.
	Timestamp time.Time

	// SequenceNumber is the ordinal position of this message in the topic.
	// For Hedera, this is the topic sequence number.
	// For Kafka, this is the partition offset.
	// Zero value indicates the sequence number is not available.
	SequenceNumber uint64

	// SenderPublicKey identifies the message sender, if available.
	// This may be zero-value if the sender is unknown or not authenticated.
	// For Hedera, this comes from the transaction submit key.
	SenderPublicKey keylib.NeuronPublicKey

	// TopicID identifies which topic this message belongs to.
	TopicID string

	// Technology indicates which backend produced this message.
	Technology TopicTechnology

	// Metadata contains technology-specific additional fields.
	// Examples:
	//   - Hedera: "transactionId", "runningHash"
	//   - Kafka: "partition", "key", "headers"
	Metadata map[string]any
}

// IsZero returns true if this is a zero-value message.
func (m TopicMessage) IsZero() bool {
	return len(m.Payload) == 0 && m.Timestamp.IsZero() && m.SequenceNumber == 0
}

// HasSender returns true if the message has a known sender.
func (m TopicMessage) HasSender() bool {
	return !m.SenderPublicKey.IsZero()
}

// GetMetadataString retrieves a string value from metadata.
// Returns empty string if the key doesn't exist or isn't a string.
func (m TopicMessage) GetMetadataString(key string) string {
	if m.Metadata == nil {
		return ""
	}
	if v, ok := m.Metadata[key].(string); ok {
		return v
	}
	return ""
}

// GetMetadataInt64 retrieves an int64 value from metadata.
// Returns 0 if the key doesn't exist or isn't numeric.
func (m TopicMessage) GetMetadataInt64(key string) int64 {
	if m.Metadata == nil {
		return 0
	}
	switch v := m.Metadata[key].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint64:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	default:
		return 0
	}
}

// GetMetadataBytes retrieves a []byte value from metadata.
// Returns nil if the key doesn't exist or isn't []byte.
func (m TopicMessage) GetMetadataBytes(key string) []byte {
	if m.Metadata == nil {
		return nil
	}
	if v, ok := m.Metadata[key].([]byte); ok {
		return v
	}
	return nil
}

// TopicMessageBuilder provides a fluent API for constructing TopicMessage instances.
type TopicMessageBuilder struct {
	msg TopicMessage
}

// NewTopicMessageBuilder creates a new builder for TopicMessage.
func NewTopicMessageBuilder() *TopicMessageBuilder {
	return &TopicMessageBuilder{
		msg: TopicMessage{
			Metadata: make(map[string]any),
		},
	}
}

// WithPayload sets the message payload.
func (b *TopicMessageBuilder) WithPayload(payload []byte) *TopicMessageBuilder {
	b.msg.Payload = payload
	return b
}

// WithTimestamp sets the message timestamp.
func (b *TopicMessageBuilder) WithTimestamp(ts time.Time) *TopicMessageBuilder {
	b.msg.Timestamp = ts
	return b
}

// WithSequenceNumber sets the message sequence number.
func (b *TopicMessageBuilder) WithSequenceNumber(seq uint64) *TopicMessageBuilder {
	b.msg.SequenceNumber = seq
	return b
}

// WithSender sets the sender's public key.
func (b *TopicMessageBuilder) WithSender(pubKey keylib.NeuronPublicKey) *TopicMessageBuilder {
	b.msg.SenderPublicKey = pubKey
	return b
}

// WithTopicID sets the topic identifier.
func (b *TopicMessageBuilder) WithTopicID(id string) *TopicMessageBuilder {
	b.msg.TopicID = id
	return b
}

// WithTechnology sets the topic technology.
func (b *TopicMessageBuilder) WithTechnology(tech TopicTechnology) *TopicMessageBuilder {
	b.msg.Technology = tech
	return b
}

// WithMetadata adds a metadata key-value pair.
func (b *TopicMessageBuilder) WithMetadata(key string, value any) *TopicMessageBuilder {
	if b.msg.Metadata == nil {
		b.msg.Metadata = make(map[string]any)
	}
	b.msg.Metadata[key] = value
	return b
}

// Build returns the constructed TopicMessage.
func (b *TopicMessageBuilder) Build() TopicMessage {
	return b.msg
}
