package account

import (
	"testing"
)

// =============================================================================
// TopicTechnology.String Tests
// =============================================================================

func TestTopicTechnology_String(t *testing.T) {
	tests := []struct {
		tech     TopicTechnology
		expected string
	}{
		{TopicTechnologyHedera, "hedera"},
		{TopicTechnologyKafka, "kafka"},
		{TopicTechnologyCustom, "custom"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if got := tc.tech.String(); got != tc.expected {
				t.Errorf("String() = %v, want %v", got, tc.expected)
			}
		})
	}

	t.Run("unknown technology returns itself", func(t *testing.T) {
		unknown := TopicTechnology("unknown")
		if got := unknown.String(); got != "unknown" {
			t.Errorf("String() = %v, want 'unknown'", got)
		}
	})
}

// =============================================================================
// TopicTechnology.IsValid Tests
// =============================================================================

func TestTopicTechnology_IsValid(t *testing.T) {
	t.Run("Hedera is valid", func(t *testing.T) {
		if !TopicTechnologyHedera.IsValid() {
			t.Error("Hedera should be valid")
		}
	})

	t.Run("Kafka is valid", func(t *testing.T) {
		if !TopicTechnologyKafka.IsValid() {
			t.Error("Kafka should be valid")
		}
	})

	t.Run("Custom is valid", func(t *testing.T) {
		if !TopicTechnologyCustom.IsValid() {
			t.Error("Custom should be valid")
		}
	})

	t.Run("empty string is invalid", func(t *testing.T) {
		if TopicTechnology("").IsValid() {
			t.Error("empty string should be invalid")
		}
	})

	t.Run("unknown technology is invalid", func(t *testing.T) {
		if TopicTechnology("unknown").IsValid() {
			t.Error("unknown technology should be invalid")
		}
	})

	t.Run("case sensitive - HEDERA is invalid", func(t *testing.T) {
		if TopicTechnology("HEDERA").IsValid() {
			t.Error("HEDERA (uppercase) should be invalid")
		}
	})
}
