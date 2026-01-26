package account

import (
	"sync"
	"testing"
)

// TestRegisteredBackends verifies that the built-in backends are registered.
func TestRegisteredBackends(t *testing.T) {
	// Check Hedera backend is registered
	backend, ok := GetBackend("hedera-topic")
	if !ok {
		t.Error("Hedera backend not registered")
	}
	if backend.Kind() != "hedera-topic" {
		t.Errorf("Hedera backend kind mismatch: got %s, want hedera-topic", backend.Kind())
	}

	// Check Kafka backend is registered
	backend, ok = GetBackend("kafka-topic")
	if !ok {
		t.Error("Kafka backend not registered")
	}
	if backend.Kind() != "kafka-topic" {
		t.Errorf("Kafka backend kind mismatch: got %s, want kafka-topic", backend.Kind())
	}
}

// TestListBackends verifies ListBackends returns all registered backends.
func TestListBackends(t *testing.T) {
	kinds := ListBackends()

	// Should contain at least hedera-topic and kafka-topic
	found := make(map[string]bool)
	for _, kind := range kinds {
		found[kind] = true
	}

	if !found["hedera-topic"] {
		t.Error("ListBackends missing hedera-topic")
	}
	if !found["kafka-topic"] {
		t.Error("ListBackends missing kafka-topic")
	}

	// Verify sorted order
	for i := 1; i < len(kinds); i++ {
		if kinds[i-1] > kinds[i] {
			t.Errorf("ListBackends not sorted: %s > %s", kinds[i-1], kinds[i])
		}
	}
}

// TestIsRegisteredBackend verifies backend existence checking.
func TestIsRegisteredBackend(t *testing.T) {
	tests := []struct {
		kind     string
		expected bool
	}{
		{"hedera-topic", true},
		{"kafka-topic", true},
		{"unknown-backend", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			if got := IsRegisteredBackend(tt.kind); got != tt.expected {
				t.Errorf("IsRegisteredBackend(%q) = %v, want %v", tt.kind, got, tt.expected)
			}
		})
	}
}

// TestBackendMetadata verifies metadata retrieval.
func TestBackendMetadata(t *testing.T) {
	// Test Hedera metadata
	meta, ok := GetBackendMetadata("hedera-topic")
	if !ok {
		t.Fatal("Failed to get Hedera metadata")
	}
	if meta.DisplayName == "" {
		t.Error("Hedera DisplayName is empty")
	}
	if meta.LocatorFormat == "" {
		t.Error("Hedera LocatorFormat is empty")
	}
	if meta.LocatorExample == "" {
		t.Error("Hedera LocatorExample is empty")
	}

	// Test Kafka metadata
	meta, ok = GetBackendMetadata("kafka-topic")
	if !ok {
		t.Fatal("Failed to get Kafka metadata")
	}
	if meta.DisplayName == "" {
		t.Error("Kafka DisplayName is empty")
	}

	// Test unknown backend
	_, ok = GetBackendMetadata("unknown")
	if ok {
		t.Error("GetBackendMetadata should return false for unknown backend")
	}
}

// TestListBackendsWithMetadata verifies metadata listing.
func TestListBackendsWithMetadata(t *testing.T) {
	allMeta := ListBackendsWithMetadata()

	if len(allMeta) < 2 {
		t.Errorf("Expected at least 2 backends, got %d", len(allMeta))
	}

	// Check each backend has valid metadata
	for kind, meta := range allMeta {
		if meta.DisplayName == "" {
			t.Errorf("Backend %q has empty DisplayName", kind)
		}
	}
}

// TestHederaBackendValidation verifies Hedera locator validation.
func TestHederaBackendValidation(t *testing.T) {
	backend, _ := GetBackend("hedera-topic")

	tests := []struct {
		locator   string
		wantError bool
	}{
		// Valid formats
		{"0.0.12345", false},
		{"1.2.3", false},
		{"0.0.0", false},
		{"999.999.999999999", false},

		// Invalid formats
		{"", true},
		{"0.0", true},
		{"0.0.12345.extra", true},
		{"invalid", true},
		{"-1.0.0", true},
		{"a.b.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.locator, func(t *testing.T) {
			err := backend.ValidateLocator(tt.locator)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateLocator(%q) error = %v, wantError %v", tt.locator, err, tt.wantError)
			}
		})
	}
}

// TestKafkaBackendValidation verifies Kafka locator validation.
func TestKafkaBackendValidation(t *testing.T) {
	backend, _ := GetBackend("kafka-topic")

	tests := []struct {
		locator   string
		wantError bool
	}{
		// Valid formats
		{"my-topic", false},
		{"topic.name", false},
		{"topic_name", false},
		{"topic-name-123", false},
		{"a", false},

		// Invalid formats
		{"", true},
		{"topic name", true},              // spaces not allowed
		{"topic@name", true},              // special chars not allowed
		{string(make([]byte, 250)), true}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.locator, func(t *testing.T) {
			err := backend.ValidateLocator(tt.locator)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateLocator(%q) error = %v, wantError %v", tt.locator, err, tt.wantError)
			}
		})
	}
}

// TestBackendParseLocator verifies locator parsing.
func TestBackendParseLocator(t *testing.T) {
	// Hedera
	hederaBackend, _ := GetBackend("hedera-topic")
	parsed, err := hederaBackend.ParseLocator("0.0.12345")
	if err != nil {
		t.Errorf("Hedera ParseLocator failed: %v", err)
	}
	if parsed != "0.0.12345" {
		t.Errorf("Hedera ParseLocator returned %q, want %q", parsed, "0.0.12345")
	}

	// Invalid locator
	_, err = hederaBackend.ParseLocator("invalid")
	if err == nil {
		t.Error("Hedera ParseLocator should fail for invalid locator")
	}

	// Kafka
	kafkaBackend, _ := GetBackend("kafka-topic")
	parsed, err = kafkaBackend.ParseLocator("my-topic")
	if err != nil {
		t.Errorf("Kafka ParseLocator failed: %v", err)
	}
	if parsed != "my-topic" {
		t.Errorf("Kafka ParseLocator returned %q, want %q", parsed, "my-topic")
	}
}

// TestNewCommAddressWithRegistry verifies NewCommAddress uses the registry.
func TestNewCommAddressWithRegistry(t *testing.T) {
	// Valid Hedera address
	addr, err := NewCommAddress("hedera-topic", "0.0.12345")
	if err != nil {
		t.Errorf("NewCommAddress failed: %v", err)
	}
	if addr.Kind() != CommAddressKind(HederaTopicKind) {
		t.Errorf("Kind mismatch: got %s, want %s", addr.Kind(), HederaTopicKind)
	}
	if addr.Locator() != "0.0.12345" {
		t.Errorf("Locator mismatch: got %s, want 0.0.12345", addr.Locator())
	}

	// Invalid Hedera address
	_, err = NewCommAddress("hedera-topic", "invalid")
	if err == nil {
		t.Error("NewCommAddress should fail for invalid Hedera locator")
	}

	// Valid Kafka address
	addr, err = NewCommAddress("kafka-topic", "my-topic")
	if err != nil {
		t.Errorf("NewCommAddress failed: %v", err)
	}
	if addr.Kind() != CommAddressKind(KafkaTopicKind) {
		t.Errorf("Kind mismatch: got %s, want %s", addr.Kind(), KafkaTopicKind)
	}

	// Unknown backend (should succeed with valid locator)
	addr, err = NewCommAddress("unknown-backend", "some-locator")
	if err != nil {
		t.Errorf("NewCommAddress should succeed for unknown backend with valid locator: %v", err)
	}
	if string(addr.Kind()) != "unknown-backend" {
		t.Errorf("Kind mismatch: got %s, want unknown-backend", addr.Kind())
	}

	// Unknown backend with empty locator (should fail)
	_, err = NewCommAddress("unknown-backend", "")
	if err == nil {
		t.Error("NewCommAddress should fail for unknown backend with empty locator")
	}
}

// TestRegistryConcurrentAccess verifies thread-safety of registry operations.
func TestRegistryConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			GetBackend("hedera-topic")
		}()
		go func() {
			defer wg.Done()
			ListBackends()
		}()
		go func() {
			defer wg.Done()
			IsRegisteredBackend("kafka-topic")
		}()
	}

	wg.Wait()
}
