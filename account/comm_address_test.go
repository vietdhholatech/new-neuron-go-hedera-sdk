package account

import (
	"errors"
	"strings"
	"testing"
)

// =============================================================================
// CommAddressKind.String Tests
// =============================================================================

func TestCommAddressKind_String(t *testing.T) {
	tests := []struct {
		kind     CommAddressKind
		expected string
	}{
		{CommAddressKind(HederaTopicKind), "hedera-topic"},
		{CommAddressKind(KafkaTopicKind), "kafka-topic"},
		{CommAddressKind(CustomTopicKind), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// CommAddressKind.IsValid Tests
// =============================================================================

func TestCommAddressKind_IsValid(t *testing.T) {
	t.Run("HederaTopic is valid", func(t *testing.T) {
		if !CommAddressKind(HederaTopicKind).IsValid() {
			t.Error("HederaTopic should be valid")
		}
	})

	t.Run("KafkaTopic is valid", func(t *testing.T) {
		if !CommAddressKind(KafkaTopicKind).IsValid() {
			t.Error("KafkaTopic should be valid")
		}
	})

	t.Run("Custom is valid", func(t *testing.T) {
		if !CommAddressKind(CustomTopicKind).IsValid() {
			t.Error("Custom should be valid")
		}
	})

	t.Run("empty string is invalid", func(t *testing.T) {
		var emptyKind CommAddressKind
		if emptyKind.IsValid() {
			t.Error("empty kind should be invalid")
		}
	})

	t.Run("unknown kind is invalid", func(t *testing.T) {
		unknownKind := CommAddressKind("unknown")
		if unknownKind.IsValid() {
			t.Error("unknown kind should be invalid")
		}
	})
}

// =============================================================================
// CommAddressKind.TopicTechnologyForKind Tests
// =============================================================================

func TestCommAddressKind_TopicTechnologyForKind(t *testing.T) {
	t.Run("HederaTopic returns TopicTechnologyHedera", func(t *testing.T) {
		got := CommAddressKind(HederaTopicKind).TopicTechnologyForKind()
		if got != TopicTechnologyHedera {
			t.Errorf("TopicTechnologyForKind() = %v, want %v", got, TopicTechnologyHedera)
		}
	})

	t.Run("KafkaTopic returns TopicTechnologyKafka", func(t *testing.T) {
		got := CommAddressKind(KafkaTopicKind).TopicTechnologyForKind()
		if got != TopicTechnologyKafka {
			t.Errorf("TopicTechnologyForKind() = %v, want %v", got, TopicTechnologyKafka)
		}
	})

	t.Run("Custom returns TopicTechnologyCustom", func(t *testing.T) {
		got := CommAddressKind(CustomTopicKind).TopicTechnologyForKind()
		if got != TopicTechnologyCustom {
			t.Errorf("TopicTechnologyForKind() = %v, want %v", got, TopicTechnologyCustom)
		}
	})
}

// =============================================================================
// NewHederaTopicAddress Tests
// =============================================================================

func TestNewHederaTopicAddress_Comprehensive(t *testing.T) {
	validCases := []struct {
		name    string
		locator string
	}{
		{"basic format 0.0.12345", "0.0.12345"},
		{"large topic number", "0.0.999999999"},
		{"non-zero shard", "1.0.12345"},
		{"non-zero realm", "0.1.12345"},
		{"all non-zero", "1.2.12345"},
		{"zero topic", "0.0.0"},
		{"all zeros", "0.0.0"},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := NewHederaTopicAddress(tc.locator)
			if err != nil {
				t.Errorf("NewHederaTopicAddress(%s) returned error: %v", tc.locator, err)
				return
			}
			if addr.Kind() != CommAddressKind(HederaTopicKind) {
				t.Errorf("Kind() = %v, want %v", addr.Kind(), CommAddressKind(HederaTopicKind))
			}
			if addr.Locator() != tc.locator {
				t.Errorf("Locator() = %v, want %v", addr.Locator(), tc.locator)
			}
		})
	}

	invalidCases := []struct {
		name    string
		locator string
	}{
		{"empty string", ""},
		{"single number", "12345"},
		{"two parts", "0.0"},
		{"four parts", "0.0.1.2"},
		{"non-numeric first", "a.0.12345"},
		{"non-numeric second", "0.b.12345"},
		{"non-numeric third", "0.0.abc"},
		{"with spaces", " 0.0.123"},
		{"trailing spaces", "0.0.123 "},
		{"negative number", "-1.0.12345"},
		{"decimal number", "0.0.123.45"},
		{"special characters", "0.0.12345!"},
	}

	for _, tc := range invalidCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			_, err := NewHederaTopicAddress(tc.locator)
			if err == nil {
				t.Errorf("NewHederaTopicAddress(%s) should return error", tc.locator)
			}
		})
	}
}

// =============================================================================
// NewKafkaTopicAddress Tests
// =============================================================================

func TestNewKafkaTopicAddress_Comprehensive(t *testing.T) {
	validCases := []struct {
		name    string
		locator string
	}{
		{"alphanumeric", "mytopic"},
		{"with dots", "my.topic.name"},
		{"with underscores", "my_topic"},
		{"with hyphens", "my-topic"},
		{"mixed characters", "my.topic_name-1"},
		{"single character", "a"},
		{"numbers only", "12345"},
		{"exactly 249 chars", strings.Repeat("a", 249)},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := NewKafkaTopicAddress(tc.locator)
			if err != nil {
				t.Errorf("NewKafkaTopicAddress(%s) returned error: %v", tc.locator, err)
				return
			}
			if addr.Kind() != CommAddressKind(KafkaTopicKind) {
				t.Errorf("Kind() = %v, want %v", addr.Kind(), CommAddressKind(KafkaTopicKind))
			}
			if addr.Locator() != tc.locator {
				t.Errorf("Locator() = %v, want %v", addr.Locator(), tc.locator)
			}
		})
	}

	invalidCases := []struct {
		name    string
		locator string
	}{
		{"empty string", ""},
		{"exceeds 249 chars", strings.Repeat("a", 250)},
		{"with spaces", "my topic"},
		{"special char @", "topic@name"},
		{"special char #", "topic#name"},
		{"special char $", "topic$name"},
		{"with slash", "cluster/topic"},
		{"with colon", "topic:name"},
	}

	for _, tc := range invalidCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			_, err := NewKafkaTopicAddress(tc.locator)
			if err == nil {
				t.Errorf("NewKafkaTopicAddress(%s) should return error", tc.locator)
			}
		})
	}
}

// =============================================================================
// NewCustomAddress Tests
// =============================================================================

func TestNewCustomAddress(t *testing.T) {
	t.Run("accepts custom kind with locator", func(t *testing.T) {
		addr, err := NewCustomAddress("my-custom-kind", "my-locator")
		if err != nil {
			t.Errorf("NewCustomAddress returned error: %v", err)
			return
		}
		if addr.Kind() != "my-custom-kind" {
			t.Errorf("Kind() = %v, want 'my-custom-kind'", addr.Kind())
		}
		if addr.Locator() != "my-locator" {
			t.Errorf("Locator() = %v, want 'my-locator'", addr.Locator())
		}
	})

	t.Run("rejects empty locator", func(t *testing.T) {
		_, err := NewCustomAddress("my-custom-kind", "")
		if err == nil {
			t.Error("NewCustomAddress with empty locator should return error")
		}
	})
}

// =============================================================================
// ParseCommAddress Tests
// =============================================================================

func TestParseCommAddress_Comprehensive(t *testing.T) {
	validCases := []struct {
		name    string
		input   string
		kind    CommAddressKind
		locator string
	}{
		{"hedera topic", "hedera-topic:0.0.12345", CommAddressKind(HederaTopicKind), "0.0.12345"},
		{"kafka topic", "kafka-topic:my-topic", CommAddressKind(KafkaTopicKind), "my-topic"},
		{"custom address", "custom:any-locator", CommAddressKind(CustomTopicKind), "any-locator"},
		{"unknown kind as custom", "unknown-kind:some-value", CommAddressKind("unknown-kind"), "some-value"},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := ParseCommAddress(tc.input)
			if err != nil {
				t.Errorf("ParseCommAddress(%s) returned error: %v", tc.input, err)
				return
			}
			if addr.Kind() != tc.kind {
				t.Errorf("Kind() = %v, want %v", addr.Kind(), tc.kind)
			}
			if addr.Locator() != tc.locator {
				t.Errorf("Locator() = %v, want %v", addr.Locator(), tc.locator)
			}
		})
	}

	invalidCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"no colon separator", "hedera-topic0.0.12345"},
		{"empty locator", "hedera-topic:"},
		{"only colon", ":"},
		{"invalid hedera format after parse", "hedera-topic:invalid"},
	}

	for _, tc := range invalidCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			_, err := ParseCommAddress(tc.input)
			if err == nil {
				t.Errorf("ParseCommAddress(%s) should return error", tc.input)
			}
		})
	}
}

func TestMustParseCommAddress(t *testing.T) {
	t.Run("parses valid address", func(t *testing.T) {
		addr := MustParseCommAddress("hedera-topic:0.0.12345")
		if addr.Locator() != "0.0.12345" {
			t.Errorf("Locator() = %v, want '0.0.12345'", addr.Locator())
		}
	})

	t.Run("panics on invalid address", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseCommAddress should panic on invalid address")
			}
		}()
		MustParseCommAddress("invalid")
	})
}

// =============================================================================
// CommAddress Method Tests
// =============================================================================

func TestCommAddress_Kind(t *testing.T) {
	t.Run("returns correct kind for Hedera", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if addr.Kind() != CommAddressKind(HederaTopicKind) {
			t.Errorf("Kind() = %v, want %v", addr.Kind(), CommAddressKind(HederaTopicKind))
		}
	})

	t.Run("returns correct kind for Kafka", func(t *testing.T) {
		addr, _ := NewKafkaTopicAddress("my-topic")
		if addr.Kind() != CommAddressKind(KafkaTopicKind) {
			t.Errorf("Kind() = %v, want %v", addr.Kind(), CommAddressKind(KafkaTopicKind))
		}
	})

	t.Run("returns empty for zero value", func(t *testing.T) {
		var addr CommAddress
		if addr.Kind() != "" {
			t.Errorf("Kind() for zero value = %v, want ''", addr.Kind())
		}
	})
}

func TestCommAddress_Locator(t *testing.T) {
	t.Run("returns correct locator for Hedera", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if addr.Locator() != "0.0.12345" {
			t.Errorf("Locator() = %v, want '0.0.12345'", addr.Locator())
		}
	})

	t.Run("returns correct locator for Kafka", func(t *testing.T) {
		addr, _ := NewKafkaTopicAddress("my-topic")
		if addr.Locator() != "my-topic" {
			t.Errorf("Locator() = %v, want 'my-topic'", addr.Locator())
		}
	})

	t.Run("returns empty for zero value", func(t *testing.T) {
		var addr CommAddress
		if addr.Locator() != "" {
			t.Errorf("Locator() for zero value = %v, want ''", addr.Locator())
		}
	})
}

func TestCommAddress_Technology(t *testing.T) {
	t.Run("returns Hedera technology", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if addr.Technology() != TopicTechnologyHedera {
			t.Errorf("Technology() = %v, want %v", addr.Technology(), TopicTechnologyHedera)
		}
	})

	t.Run("returns Kafka technology", func(t *testing.T) {
		addr, _ := NewKafkaTopicAddress("my-topic")
		if addr.Technology() != TopicTechnologyKafka {
			t.Errorf("Technology() = %v, want %v", addr.Technology(), TopicTechnologyKafka)
		}
	})
}

func TestCommAddress_String(t *testing.T) {
	t.Run("formats hedera-topic:locator", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		expected := "hedera-topic:0.0.12345"
		if addr.String() != expected {
			t.Errorf("String() = %v, want %v", addr.String(), expected)
		}
	})

	t.Run("formats kafka-topic:locator", func(t *testing.T) {
		addr, _ := NewKafkaTopicAddress("my-topic")
		expected := "kafka-topic:my-topic"
		if addr.String() != expected {
			t.Errorf("String() = %v, want %v", addr.String(), expected)
		}
	})

	t.Run("zero value returns empty string", func(t *testing.T) {
		var addr CommAddress
		if addr.String() != "" {
			t.Errorf("String() for zero value = %v, want ''", addr.String())
		}
	})
}

func TestCommAddress_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var addr CommAddress
		if !addr.IsZero() {
			t.Error("zero value should be zero")
		}
	})

	t.Run("parsed address is not zero", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if addr.IsZero() {
			t.Error("parsed address should not be zero")
		}
	})

	t.Run("only kind empty is not zero", func(t *testing.T) {
		// This shouldn't be possible through normal constructors
		addr := CommAddress{kind: "", locator: "something"}
		if addr.IsZero() {
			t.Error("address with locator should not be zero")
		}
	})
}

func TestCommAddress_Validate(t *testing.T) {
	t.Run("valid Hedera address passes", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if err := addr.Validate(); err != nil {
			t.Errorf("Validate() returned error: %v", err)
		}
	})

	t.Run("valid Kafka address passes", func(t *testing.T) {
		addr, _ := NewKafkaTopicAddress("my-topic")
		if err := addr.Validate(); err != nil {
			t.Errorf("Validate() returned error: %v", err)
		}
	})

	t.Run("zero value fails with ZeroValue kind", func(t *testing.T) {
		var addr CommAddress
		err := addr.Validate()
		if err == nil {
			t.Error("zero value Validate() should fail")
			return
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
		if ae.Kind != ErrKindZeroValue {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
		}
	})
}

func TestCommAddress_Equal(t *testing.T) {
	t.Run("same addresses are equal", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.12345")
		addr2, _ := NewHederaTopicAddress("0.0.12345")
		if !addr1.Equal(addr2) {
			t.Error("same addresses should be equal")
		}
	})

	t.Run("different kinds are not equal", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.12345")
		addr2, _ := NewKafkaTopicAddress("my-topic")
		if addr1.Equal(addr2) {
			t.Error("different kinds should not be equal")
		}
	})

	t.Run("different locators are not equal", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.12345")
		addr2, _ := NewHederaTopicAddress("0.0.67890")
		if addr1.Equal(addr2) {
			t.Error("different locators should not be equal")
		}
	})

	t.Run("zero values are equal", func(t *testing.T) {
		var addr1, addr2 CommAddress
		if !addr1.Equal(addr2) {
			t.Error("zero values should be equal")
		}
	})
}

// =============================================================================
// CommAddress Round-Trip Tests
// =============================================================================

func TestCommAddress_RoundTrip(t *testing.T) {
	t.Run("round trip Hedera address", func(t *testing.T) {
		original, _ := NewHederaTopicAddress("0.0.12345")
		str := original.String()
		restored, err := ParseCommAddress(str)
		if err != nil {
			t.Fatalf("ParseCommAddress failed: %v", err)
		}
		if !original.Equal(restored) {
			t.Error("round trip failed")
		}
	})

	t.Run("round trip Kafka address", func(t *testing.T) {
		original, _ := NewKafkaTopicAddress("my-topic")
		str := original.String()
		restored, err := ParseCommAddress(str)
		if err != nil {
			t.Fatalf("ParseCommAddress failed: %v", err)
		}
		if !original.Equal(restored) {
			t.Error("round trip failed")
		}
	})
}

// =============================================================================
// CommAddressSet Tests
// =============================================================================

func TestNewCommAddressSet(t *testing.T) {
	t.Run("creates set from addresses", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.1")
		addr2, _ := NewHederaTopicAddress("0.0.2")
		set := NewCommAddressSet(addr1, addr2)
		if set.Len() != 2 {
			t.Errorf("Len() = %d, want 2", set.Len())
		}
	})

	t.Run("filters out zero values", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.1")
		var zeroAddr CommAddress
		set := NewCommAddressSet(addr1, zeroAddr)
		if set.Len() != 1 {
			t.Errorf("Len() = %d, want 1 (zero value filtered)", set.Len())
		}
	})

	t.Run("empty input creates empty set", func(t *testing.T) {
		set := NewCommAddressSet()
		if !set.IsEmpty() {
			t.Error("set should be empty")
		}
	})
}

func TestCommAddressSet_Methods(t *testing.T) {
	addr1, _ := NewHederaTopicAddress("0.0.1")
	addr2, _ := NewHederaTopicAddress("0.0.2")
	set := NewCommAddressSet(addr1, addr2)

	t.Run("Addresses returns copy", func(t *testing.T) {
		addrs := set.Addresses()
		if len(addrs) != 2 {
			t.Errorf("len(Addresses()) = %d, want 2", len(addrs))
		}
		// Verify it's a copy by modifying
		addrs[0] = CommAddress{}
		if set.First().IsZero() {
			t.Error("original should not be modified")
		}
	})

	t.Run("IsEmpty on non-empty set", func(t *testing.T) {
		if set.IsEmpty() {
			t.Error("non-empty set should not be empty")
		}
	})

	t.Run("Len returns correct count", func(t *testing.T) {
		if set.Len() != 2 {
			t.Errorf("Len() = %d, want 2", set.Len())
		}
	})

	t.Run("First returns first address", func(t *testing.T) {
		if !set.First().Equal(addr1) {
			t.Error("First() should return first address")
		}
	})

	t.Run("First on empty returns zero", func(t *testing.T) {
		emptySet := NewCommAddressSet()
		if !emptySet.First().IsZero() {
			t.Error("First() on empty set should return zero")
		}
	})

	t.Run("IsEmpty on empty set", func(t *testing.T) {
		emptySet := NewCommAddressSet()
		if !emptySet.IsEmpty() {
			t.Error("empty set should be empty")
		}
	})
}

// =============================================================================
// ValidateHederaTopicLocator Tests
// =============================================================================

func TestValidateHederaTopicLocator(t *testing.T) {
	t.Run("empty string fails", func(t *testing.T) {
		err := ValidateHederaTopicLocator("")
		if err == nil {
			t.Error("empty string should fail validation")
		}
	})

	t.Run("valid format passes", func(t *testing.T) {
		err := ValidateHederaTopicLocator("0.0.12345")
		if err != nil {
			t.Errorf("valid format should pass: %v", err)
		}
	})
}

// =============================================================================
// ValidateKafkaTopicLocator Tests
// =============================================================================

func TestValidateKafkaTopicLocator(t *testing.T) {
	t.Run("empty string fails", func(t *testing.T) {
		err := ValidateKafkaTopicLocator("")
		if err == nil {
			t.Error("empty string should fail validation")
		}
	})

	t.Run("valid name passes", func(t *testing.T) {
		err := ValidateKafkaTopicLocator("my-topic")
		if err != nil {
			t.Errorf("valid name should pass: %v", err)
		}
	})

	t.Run("too long fails", func(t *testing.T) {
		err := ValidateKafkaTopicLocator(strings.Repeat("a", 250))
		if err == nil {
			t.Error("too long name should fail validation")
		}
	})
}
