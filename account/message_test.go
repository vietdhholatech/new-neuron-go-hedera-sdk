package account

import (
	"testing"
	"time"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// TopicMessage.IsZero Tests
// =============================================================================

func TestTopicMessage_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var msg TopicMessage
		if !msg.IsZero() {
			t.Error("zero value should be zero")
		}
	})

	t.Run("message with payload is not zero", func(t *testing.T) {
		msg := TopicMessage{
			Payload: []byte("hello"),
		}
		if msg.IsZero() {
			t.Error("message with payload should not be zero")
		}
	})

	t.Run("message with timestamp is not zero", func(t *testing.T) {
		msg := TopicMessage{
			Timestamp: time.Now(),
		}
		if msg.IsZero() {
			t.Error("message with timestamp should not be zero")
		}
	})

	t.Run("message with sequence number is not zero", func(t *testing.T) {
		msg := TopicMessage{
			SequenceNumber: 1,
		}
		if msg.IsZero() {
			t.Error("message with sequence number should not be zero")
		}
	})

	t.Run("message with only topic ID is still zero", func(t *testing.T) {
		msg := TopicMessage{
			TopicID: "0.0.12345",
		}
		// TopicID alone doesn't make it non-zero based on IsZero implementation
		if !msg.IsZero() {
			t.Error("message with only topic ID should be zero")
		}
	})
}

// =============================================================================
// TopicMessage.HasSender Tests
// =============================================================================

func TestTopicMessage_HasSender(t *testing.T) {
	t.Run("message without sender returns false", func(t *testing.T) {
		var msg TopicMessage
		if msg.HasSender() {
			t.Error("message without sender should return false")
		}
	})

	t.Run("message with sender returns true", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		msg := TopicMessage{
			SenderPublicKey: privKey.PublicKey(),
		}
		if !msg.HasSender() {
			t.Error("message with sender should return true")
		}
	})
}

// =============================================================================
// TopicMessage Metadata Accessor Tests
// =============================================================================

func TestTopicMessage_GetMetadataString(t *testing.T) {
	t.Run("returns string value", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": "value",
			},
		}
		if got := msg.GetMetadataString("key"); got != "value" {
			t.Errorf("GetMetadataString() = %v, want 'value'", got)
		}
	})

	t.Run("returns empty for missing key", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{},
		}
		if got := msg.GetMetadataString("missing"); got != "" {
			t.Errorf("GetMetadataString() = %v, want ''", got)
		}
	})

	t.Run("returns empty for non-string value", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": 123,
			},
		}
		if got := msg.GetMetadataString("key"); got != "" {
			t.Errorf("GetMetadataString() = %v, want ''", got)
		}
	})

	t.Run("returns empty for nil metadata", func(t *testing.T) {
		msg := TopicMessage{}
		if got := msg.GetMetadataString("key"); got != "" {
			t.Errorf("GetMetadataString() = %v, want ''", got)
		}
	})
}

func TestTopicMessage_GetMetadataInt64(t *testing.T) {
	t.Run("returns int64 value", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": int64(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("converts int to int64", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": int(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("converts int32 to int64", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": int32(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("converts uint64 to int64", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": uint64(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("converts uint to int64", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": uint(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("converts uint32 to int64", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": uint32(42),
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 42 {
			t.Errorf("GetMetadataInt64() = %v, want 42", got)
		}
	})

	t.Run("returns 0 for missing key", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{},
		}
		if got := msg.GetMetadataInt64("missing"); got != 0 {
			t.Errorf("GetMetadataInt64() = %v, want 0", got)
		}
	})

	t.Run("returns 0 for non-numeric value", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": "string",
			},
		}
		if got := msg.GetMetadataInt64("key"); got != 0 {
			t.Errorf("GetMetadataInt64() = %v, want 0", got)
		}
	})

	t.Run("returns 0 for nil metadata", func(t *testing.T) {
		msg := TopicMessage{}
		if got := msg.GetMetadataInt64("key"); got != 0 {
			t.Errorf("GetMetadataInt64() = %v, want 0", got)
		}
	})
}

func TestTopicMessage_GetMetadataBytes(t *testing.T) {
	t.Run("returns bytes value", func(t *testing.T) {
		expected := []byte("hello")
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": expected,
			},
		}
		got := msg.GetMetadataBytes("key")
		if string(got) != string(expected) {
			t.Errorf("GetMetadataBytes() = %v, want %v", got, expected)
		}
	})

	t.Run("returns nil for missing key", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{},
		}
		if got := msg.GetMetadataBytes("missing"); got != nil {
			t.Errorf("GetMetadataBytes() = %v, want nil", got)
		}
	})

	t.Run("returns nil for non-bytes value", func(t *testing.T) {
		msg := TopicMessage{
			Metadata: map[string]any{
				"key": "string",
			},
		}
		if got := msg.GetMetadataBytes("key"); got != nil {
			t.Errorf("GetMetadataBytes() = %v, want nil", got)
		}
	})

	t.Run("returns nil for nil metadata", func(t *testing.T) {
		msg := TopicMessage{}
		if got := msg.GetMetadataBytes("key"); got != nil {
			t.Errorf("GetMetadataBytes() = %v, want nil", got)
		}
	})
}

// =============================================================================
// TopicMessageBuilder Tests
// =============================================================================

func TestNewTopicMessageBuilder(t *testing.T) {
	t.Run("creates builder with initialized metadata", func(t *testing.T) {
		builder := NewTopicMessageBuilder()
		if builder == nil {
			t.Fatal("expected non-nil builder")
		}
		msg := builder.Build()
		if msg.Metadata == nil {
			t.Error("metadata should be initialized")
		}
	})
}

func TestTopicMessageBuilder_WithPayload(t *testing.T) {
	t.Run("sets payload", func(t *testing.T) {
		payload := []byte("test payload")
		msg := NewTopicMessageBuilder().
			WithPayload(payload).
			Build()
		if string(msg.Payload) != string(payload) {
			t.Errorf("Payload = %v, want %v", msg.Payload, payload)
		}
	})
}

func TestTopicMessageBuilder_WithTimestamp(t *testing.T) {
	t.Run("sets timestamp", func(t *testing.T) {
		ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		msg := NewTopicMessageBuilder().
			WithTimestamp(ts).
			Build()
		if !msg.Timestamp.Equal(ts) {
			t.Errorf("Timestamp = %v, want %v", msg.Timestamp, ts)
		}
	})
}

func TestTopicMessageBuilder_WithSequenceNumber(t *testing.T) {
	t.Run("sets sequence number", func(t *testing.T) {
		msg := NewTopicMessageBuilder().
			WithSequenceNumber(12345).
			Build()
		if msg.SequenceNumber != 12345 {
			t.Errorf("SequenceNumber = %v, want 12345", msg.SequenceNumber)
		}
	})
}

func TestTopicMessageBuilder_WithSender(t *testing.T) {
	t.Run("sets sender public key", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		msg := NewTopicMessageBuilder().
			WithSender(pubKey).
			Build()
		if !msg.SenderPublicKey.Equal(pubKey) {
			t.Error("SenderPublicKey not set correctly")
		}
	})
}

func TestTopicMessageBuilder_WithTopicID(t *testing.T) {
	t.Run("sets topic ID", func(t *testing.T) {
		msg := NewTopicMessageBuilder().
			WithTopicID("0.0.12345").
			Build()
		if msg.TopicID != "0.0.12345" {
			t.Errorf("TopicID = %v, want '0.0.12345'", msg.TopicID)
		}
	})
}

func TestTopicMessageBuilder_WithTechnology(t *testing.T) {
	t.Run("sets technology", func(t *testing.T) {
		msg := NewTopicMessageBuilder().
			WithTechnology(TopicTechnologyHedera).
			Build()
		if msg.Technology != TopicTechnologyHedera {
			t.Errorf("Technology = %v, want Hedera", msg.Technology)
		}
	})
}

func TestTopicMessageBuilder_WithMetadata(t *testing.T) {
	t.Run("adds metadata key-value", func(t *testing.T) {
		msg := NewTopicMessageBuilder().
			WithMetadata("key1", "value1").
			WithMetadata("key2", 42).
			Build()
		if msg.Metadata["key1"] != "value1" {
			t.Error("metadata key1 not set")
		}
		if msg.Metadata["key2"] != 42 {
			t.Error("metadata key2 not set")
		}
	})
}

func TestTopicMessageBuilder_Fluent(t *testing.T) {
	t.Run("supports fluent chaining", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		ts := time.Now()
		payload := []byte("test")

		msg := NewTopicMessageBuilder().
			WithPayload(payload).
			WithTimestamp(ts).
			WithSequenceNumber(100).
			WithSender(privKey.PublicKey()).
			WithTopicID("0.0.12345").
			WithTechnology(TopicTechnologyHedera).
			WithMetadata("txid", "0.0.1@123").
			Build()

		if string(msg.Payload) != "test" {
			t.Error("Payload not set")
		}
		if !msg.Timestamp.Equal(ts) {
			t.Error("Timestamp not set")
		}
		if msg.SequenceNumber != 100 {
			t.Error("SequenceNumber not set")
		}
		if !msg.HasSender() {
			t.Error("Sender not set")
		}
		if msg.TopicID != "0.0.12345" {
			t.Error("TopicID not set")
		}
		if msg.Technology != TopicTechnologyHedera {
			t.Error("Technology not set")
		}
		if msg.Metadata["txid"] != "0.0.1@123" {
			t.Error("Metadata not set")
		}
	})
}

func TestTopicMessageBuilder_Build(t *testing.T) {
	t.Run("returns copy not reference", func(t *testing.T) {
		builder := NewTopicMessageBuilder().
			WithPayload([]byte("original"))

		msg1 := builder.Build()
		msg2 := builder.WithPayload([]byte("modified")).Build()

		// msg1 should still have original (it's a copy)
		if string(msg1.Payload) != "original" {
			t.Error("Build() should return a copy")
		}
		if string(msg2.Payload) != "modified" {
			t.Error("second build should have modified value")
		}
	})
}
