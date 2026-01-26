package account

import (
	"errors"
	"strings"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Edge Cases Tests - Phases 3-6
// These tests verify proper handling of edge cases, boundary conditions,
// and error recovery scenarios.
// =============================================================================

// =============================================================================
// Phase 3: String/Unicode Edge Cases
// =============================================================================

func TestKafkaTopic_UnicodeCharacters(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "emoji in topic name should fail",
			topic:     "my-topic-🚀",
			wantError: true,
		},
		{
			name:      "chinese characters should fail",
			topic:     "主题-topic",
			wantError: true,
		},
		{
			name:      "japanese characters should fail",
			topic:     "トピック",
			wantError: true,
		},
		{
			name:      "arabic characters should fail",
			topic:     "موضوع",
			wantError: true,
		},
		{
			name:      "combining diacritical marks should fail",
			topic:     "cafe\u0301", // café with combining acute accent
			wantError: true,
		},
		{
			name:      "zero-width characters should fail",
			topic:     "my\u200Btopic", // zero-width space
			wantError: true,
		},
		{
			name:      "valid ascii topic passes",
			topic:     "my-valid-topic_123",
			wantError: false,
		},
		{
			name:      "topic with dots passes",
			topic:     "my.topic.name",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewKafkaTopicAddress(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("NewKafkaTopicAddress(%q) error = %v, wantError %v", tt.topic, err, tt.wantError)
			}
		})
	}
}

func TestHederaTopic_UnicodeInNumbers(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "full-width digits should fail",
			topic:     "０.０.１２３", // full-width digits
			wantError: true,
		},
		{
			name:      "superscript numbers should fail",
			topic:     "0.0.¹²³",
			wantError: true,
		},
		{
			name:      "subscript numbers should fail",
			topic:     "0.0.₁₂₃",
			wantError: true,
		},
		{
			name:      "arabic numerals should fail",
			topic:     "٠.٠.١٢٣",
			wantError: true,
		},
		{
			name:      "valid ascii numbers passes",
			topic:     "0.0.12345",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHederaTopicAddress(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("NewHederaTopicAddress(%q) error = %v, wantError %v", tt.topic, err, tt.wantError)
			}
		})
	}
}

func TestCommAddress_WhitespaceVariations(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "tab character should fail",
			topic:     "0.0.\t123",
			wantError: true,
		},
		{
			name:      "non-breaking space should fail",
			topic:     "0.0.\u00A0123",
			wantError: true,
		},
		{
			name:      "zero-width space should fail",
			topic:     "0.0.\u200B123",
			wantError: true,
		},
		{
			name:      "line break should fail",
			topic:     "0.0.123\n",
			wantError: true,
		},
		{
			name:      "carriage return should fail",
			topic:     "0.0.123\r",
			wantError: true,
		},
		{
			name:      "leading space should fail",
			topic:     " 0.0.123",
			wantError: true,
		},
		{
			name:      "trailing space should fail",
			topic:     "0.0.123 ",
			wantError: true,
		},
		{
			name:      "vertical tab should fail",
			topic:     "0.0.\v123",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHederaTopicAddress(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("NewHederaTopicAddress(%q) error = %v, wantError %v", tt.topic, err, tt.wantError)
			}
		})
	}
}

func TestKafkaTopic_LengthBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		topicLen  int
		wantError bool
	}{
		{
			name:      "1 character valid minimum",
			topicLen:  1,
			wantError: false,
		},
		{
			name:      "0 characters invalid",
			topicLen:  0,
			wantError: true,
		},
		{
			name:      "249 characters valid maximum",
			topicLen:  249,
			wantError: false,
		},
		{
			name:      "250 characters exceeds limit",
			topicLen:  250,
			wantError: true,
		},
		{
			name:      "500 characters way over limit",
			topicLen:  500,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := strings.Repeat("a", tt.topicLen)
			_, err := NewKafkaTopicAddress(topic)
			if (err != nil) != tt.wantError {
				t.Errorf("NewKafkaTopicAddress(len=%d) error = %v, wantError %v", tt.topicLen, err, tt.wantError)
			}
		})
	}
}

// =============================================================================
// Phase 4: Numeric Boundary Conditions
// =============================================================================

func TestHederaTopic_NumericBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "max uint64 for shard",
			topic:     "18446744073709551615.0.0",
			wantError: false,
		},
		{
			name:      "max uint64 for realm",
			topic:     "0.18446744073709551615.0",
			wantError: false,
		},
		{
			name:      "max uint64 for topic",
			topic:     "0.0.18446744073709551615",
			wantError: false,
		},
		{
			name:      "overflow shard wraps (implementation detail)",
			topic:     "18446744073709551616.0.0",
			wantError: false, // Note: Current implementation wraps on overflow
		},
		{
			name:      "overflow realm wraps (implementation detail)",
			topic:     "0.18446744073709551616.0",
			wantError: false, // Note: Current implementation wraps on overflow
		},
		{
			name:      "overflow topic wraps (implementation detail)",
			topic:     "0.0.18446744073709551616",
			wantError: false, // Note: Current implementation wraps on overflow
		},
		{
			name:      "leading zeros allowed (implementation detail)",
			topic:     "00.00.00012345",
			wantError: false, // Note: Current implementation allows leading zeros
		},
		{
			name:      "leading zero single digit allowed",
			topic:     "0.0.0",
			wantError: false,
		},
		{
			name:      "scientific notation should fail",
			topic:     "1e10.0.0",
			wantError: true,
		},
		{
			name:      "hex notation should fail",
			topic:     "0x0.0x0.0x0",
			wantError: true,
		},
		{
			name:      "negative number should fail",
			topic:     "-1.0.0",
			wantError: true,
		},
		{
			name:      "decimal number should fail",
			topic:     "0.0.123.456",
			wantError: true,
		},
		{
			name:      "empty shard should fail",
			topic:     ".0.0",
			wantError: true,
		},
		{
			name:      "empty realm should fail",
			topic:     "0..0",
			wantError: true,
		},
		{
			name:      "empty topic should fail",
			topic:     "0.0.",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHederaTopicAddress(tt.topic)
			if (err != nil) != tt.wantError {
				t.Errorf("NewHederaTopicAddress(%q) error = %v, wantError %v", tt.topic, err, tt.wantError)
			}
		})
	}
}

func TestMultiaddr_PortBoundaries(t *testing.T) {
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	tests := []struct {
		name      string
		port      string
		wantError bool
	}{
		{
			name:      "port 0 valid but unusual",
			port:      "0",
			wantError: false,
		},
		{
			name:      "port 1 minimum valid",
			port:      "1",
			wantError: false,
		},
		{
			name:      "port 65535 max valid",
			port:      "65535",
			wantError: false,
		},
		{
			name:      "port 65536 exceeds max",
			port:      "65536",
			wantError: true,
		},
		{
			name:      "port 80000 exceeds max",
			port:      "80000",
			wantError: true,
		},
		{
			name:      "port with leading zeros allowed",
			port:      "04001",
			wantError: false, // Note: multiaddr library allows leading zeros
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := "/ip4/192.168.1.1/tcp/" + tt.port + "/p2p/" + testPeerID
			_, err := ParseReachableAddr(addr)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseReachableAddr(port=%s) error = %v, wantError %v", tt.port, err, tt.wantError)
			}
		})
	}
}

// =============================================================================
// Phase 5: Collection Edge Cases
// =============================================================================

func TestReachableAddrs_EmptyCollectionEdge(t *testing.T) {
	t.Run("NewReachableAddrs with no args", func(t *testing.T) {
		addrs := NewReachableAddrs()
		if !addrs.IsEmpty() {
			t.Error("expected IsEmpty() = true for empty collection")
		}
	})

	t.Run("Addrs returns nil for empty", func(t *testing.T) {
		addrs := NewReachableAddrs()
		result := addrs.Addrs()
		if result != nil && len(result) > 0 {
			t.Error("expected nil or empty slice for Addrs() on empty collection")
		}
	})

	t.Run("First returns zero value for empty", func(t *testing.T) {
		addrs := NewReachableAddrs()
		first := addrs.First()
		if !first.IsZero() {
			t.Error("expected zero value from First() on empty collection")
		}
	})

	t.Run("Strings returns nil for empty", func(t *testing.T) {
		addrs := NewReachableAddrs()
		strs := addrs.Strings()
		if strs != nil && len(strs) > 0 {
			t.Error("expected nil or empty slice for Strings() on empty collection")
		}
	})

	t.Run("Len returns 0 for empty", func(t *testing.T) {
		addrs := NewReachableAddrs()
		if len(addrs.Addrs()) != 0 {
			t.Errorf("expected len(Addrs()) = 0, got %d", len(addrs.Addrs()))
		}
	})
}

func TestReachableAddrs_WithZeroValues(t *testing.T) {
	t.Run("filtering zero values", func(t *testing.T) {
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		validAddr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID)
		var zeroAddr ReachableAddr

		// Create with mix of valid and zero values
		addrs := NewReachableAddrs(validAddr, zeroAddr, validAddr)

		// Zero values should be filtered out
		count := len(addrs.Addrs())
		if count != 2 {
			t.Errorf("expected 2 valid addresses after filtering, got %d", count)
		}
	})
}

func TestCommAddressSet_EmptyCollectionEdge(t *testing.T) {
	t.Run("empty set creation", func(t *testing.T) {
		set := NewCommAddressSet()
		if !set.IsEmpty() {
			t.Error("expected IsEmpty() = true for empty set")
		}
	})

	t.Run("First returns nil for empty", func(t *testing.T) {
		set := NewCommAddressSet()
		first := set.First()
		if !first.IsZero() {
			t.Error("expected zero value from First() on empty set")
		}
	})

	t.Run("Addresses returns nil for empty", func(t *testing.T) {
		set := NewCommAddressSet()
		addrs := set.Addresses()
		if addrs != nil && len(addrs) > 0 {
			t.Error("expected nil or empty slice for Addresses() on empty set")
		}
	})

	t.Run("Len returns 0 for empty", func(t *testing.T) {
		set := NewCommAddressSet()
		if len(set.Addresses()) != 0 {
			t.Errorf("expected len(Addresses()) = 0, got %d", len(set.Addresses()))
		}
	})
}

func TestCommAddressSet_WithZeroValues(t *testing.T) {
	t.Run("filtering zero values", func(t *testing.T) {
		validAddr, _ := NewHederaTopicAddress("0.0.111")
		var zeroAddr CommAddress

		// Create with mix of valid and zero values
		set := NewCommAddressSet(validAddr, zeroAddr, validAddr)

		// Zero values should be filtered out
		count := len(set.Addresses())
		if count != 2 {
			t.Errorf("expected 2 valid addresses after filtering, got %d", count)
		}
	})
}

func TestReachableAddrs_LargeCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large collection test in short mode")
	}

	t.Run("1000 addresses", func(t *testing.T) {
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		var addresses []ReachableAddr

		for i := 0; i < 1000; i++ {
			ip := "192.168." + string(rune('0'+i/256%10)) + "." + string(rune('0'+i%256%10))
			addr, err := ParseReachableAddr("/ip4/" + ip + "/tcp/4001/p2p/" + testPeerID)
			if err == nil {
				addresses = append(addresses, addr)
			}
		}

		addrs := NewReachableAddrs(addresses...)

		// Should handle large collection without panic
		_ = addrs.Addrs()
		_ = addrs.Strings()
		_ = addrs.First()
		_ = len(addrs.Addrs())
	})
}

func TestNilVsZeroValue_CommAddress(t *testing.T) {
	t.Run("zero value behavior", func(t *testing.T) {
		var zeroAddr CommAddress

		if !zeroAddr.IsZero() {
			t.Error("zero value CommAddress should return IsZero() = true")
		}

		// String should return empty for zero value
		if zeroAddr.String() != "" {
			t.Errorf("zero value String() should be empty, got %q", zeroAddr.String())
		}
	})

	t.Run("valid vs zero", func(t *testing.T) {
		validAddr, _ := NewHederaTopicAddress("0.0.111")
		var zeroAddr CommAddress

		if validAddr.IsZero() {
			t.Error("valid address should not be zero")
		}
		if !zeroAddr.IsZero() {
			t.Error("zero address should be zero")
		}
		if validAddr.Equal(zeroAddr) {
			t.Error("valid address should not equal zero address")
		}
	})
}

func TestNilVsZeroValue_ReachableAddr(t *testing.T) {
	t.Run("zero value behavior", func(t *testing.T) {
		var zeroAddr ReachableAddr

		if !zeroAddr.IsZero() {
			t.Error("zero value ReachableAddr should return IsZero() = true")
		}

		// String should return empty or handle gracefully
		str := zeroAddr.String()
		if str != "" {
			t.Logf("zero value String() = %q (may vary by implementation)", str)
		}
	})

	t.Run("valid vs zero", func(t *testing.T) {
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		validAddr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID)
		var zeroAddr ReachableAddr

		if validAddr.IsZero() {
			t.Error("valid address should not be zero")
		}
		if !zeroAddr.IsZero() {
			t.Error("zero address should be zero")
		}
	})
}

// =============================================================================
// Phase 6: Error Recovery & State Consistency
// =============================================================================

func TestBuilder_StateAfterError(t *testing.T) {
	t.Run("builder state after invalid StdIn", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		builder := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("invalid-topic") // Invalid topic format

		// Builder should have accumulated error
		if !builder.HasErrors() {
			t.Error("expected builder to have errors after invalid StdIn")
		}

		// Build should fail
		_, err := builder.Build()
		if err == nil {
			t.Error("expected Build() to fail after invalid StdIn")
		}
	})

	t.Run("builder state after invalid ReachableAddr", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		builder := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr("not-a-valid-multiaddr")

		if !builder.HasErrors() {
			t.Error("expected builder to have errors after invalid ReachableAddr")
		}
	})

	t.Run("can continue building after error", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		peerID, _ := pubKey.PeerID()

		builder := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("invalid-topic").                                     // Error 1
			WithStdOutHedera("0.0.222").                                          // Valid
			WithReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()) // Valid

		// Should have accumulated the first error
		if !builder.HasErrors() {
			t.Error("expected builder to have errors")
		}

		// Errors should be accessible
		errs := builder.Errors()
		if len(errs) == 0 {
			t.Error("expected at least one error")
		}
	})

	t.Run("multiple errors accumulate", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		builder := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("bad1").
			WithStdOutHedera("bad2").
			WithStdErrHedera("bad3")

		errs := builder.Errors()
		if len(errs) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(errs))
		}
	})
}

func TestBuilder_StateAfterBuildFailure(t *testing.T) {
	t.Run("builder state after Build() returns error", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		builder := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("invalid")

		// First build attempt
		_, err1 := builder.Build()
		if err1 == nil {
			t.Fatal("expected first Build() to fail")
		}

		// Second build attempt should also fail (errors accumulated)
		_, err2 := builder.Build()
		if err2 == nil {
			t.Error("expected second Build() to also fail")
		}
	})
}

func TestValidation_PartialFailure(t *testing.T) {
	t.Run("ValidateAll captures all errors", func(t *testing.T) {
		// Create an account with multiple potential validation issues
		// This tests that ValidateAll doesn't short-circuit
		var account NeuronAccount // zero value has multiple issues

		result := account.ValidateAll()

		// Should capture multiple errors (HasErrors returns true if invalid)
		if !result.HasErrors() {
			t.Error("zero value account should have validation errors")
		}

		if len(result.Errors()) == 0 {
			t.Error("expected multiple validation errors for zero value")
		}
	})

	t.Run("ValidationResult AddError ignores nil", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError(nil)
		result.AddError(errors.New("real error"))
		result.AddError(nil)

		if len(result.Errors()) != 1 {
			t.Errorf("expected 1 error (nil ignored), got %d", len(result.Errors()))
		}
	})
}

func TestAccount_StateWithInvalidFields(t *testing.T) {
	t.Run("accessors on zero value account", func(t *testing.T) {
		var account NeuronAccount

		// All accessors should work without panic on zero value
		_ = account.PublicKey()
		_ = account.PeerID()
		_ = account.EVMAddress()
		_ = account.AccountType()
		_ = account.DID()
		_ = account.ParentPublicKey()
		_ = account.StdIn()
		_ = account.StdOut()
		_ = account.StdErr()
		_ = account.ReachableAddrs()
		_ = account.FirstReachableAddr()
		_ = account.IsZero()
		_ = account.String()
	})

	t.Run("methods on zero value account", func(t *testing.T) {
		var account NeuronAccount

		// Validate should return error for zero value
		err := account.Validate()
		if err == nil {
			t.Error("Validate should return error for zero value account")
		}

		// ValidateAll should return errors
		result := account.ValidateAll()
		if !result.HasErrors() {
			t.Error("ValidateAll should have errors for zero value")
		}

		// MarshalJSON should not panic
		_, err = account.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON should not error for zero value, got: %v", err)
		}
	})

	t.Run("Equal with zero values", func(t *testing.T) {
		var zero1 NeuronAccount
		var zero2 NeuronAccount

		// Two zero values should not be equal (by design)
		if zero1.Equal(zero2) {
			t.Error("two zero value accounts should not be equal")
		}

		// Zero should not equal valid account
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		valid, _ := NewParentAccountBuilder(pubKey, did).Build()

		if zero1.Equal(valid) {
			t.Error("zero value should not equal valid account")
		}
		if valid.Equal(zero1) {
			t.Error("valid account should not equal zero value")
		}
	})
}

// =============================================================================
// Circuit Relay Edge Cases (Phase 8)
// =============================================================================

func TestCircuitRelay_ComplexFormats(t *testing.T) {
	targetPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	relayPeerID := "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"

	tests := []struct {
		name      string
		addr      string
		wantError bool
		isRelay   bool
	}{
		{
			name:      "valid relay address",
			addr:      "/ip4/192.168.1.1/tcp/4001/p2p/" + relayPeerID + "/p2p-circuit/p2p/" + targetPeerID,
			wantError: false,
			isRelay:   true,
		},
		{
			name:      "relay with IPv6",
			addr:      "/ip6/::1/tcp/4001/p2p/" + relayPeerID + "/p2p-circuit/p2p/" + targetPeerID,
			wantError: false,
			isRelay:   true,
		},
		{
			name:      "direct address (no circuit)",
			addr:      "/ip4/192.168.1.1/tcp/4001/p2p/" + targetPeerID,
			wantError: false,
			isRelay:   false,
		},
		{
			name:      "dns address direct",
			addr:      "/dns4/example.com/tcp/4001/p2p/" + targetPeerID,
			wantError: false,
			isRelay:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseReachableAddr(tt.addr)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseReachableAddr() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if addr.IsCircuitRelay() != tt.isRelay {
					t.Errorf("IsRelay() = %v, want %v", addr.IsCircuitRelay(), tt.isRelay)
				}
			}
		})
	}
}

func TestCircuitRelay_EdgeFormats(t *testing.T) {
	// These tests document current behavior for edge cases
	tests := []struct {
		name      string
		addr      string
		wantError bool
	}{
		{
			name:      "circuit without explicit target (relay's PeerID used)",
			addr:      "/ip4/192.168.1.1/tcp/4001/p2p/12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN/p2p-circuit",
			wantError: false, // Note: multiaddr library accepts this
		},
		{
			name:      "circuit with explicit target",
			addr:      "/ip4/192.168.1.1/tcp/4001/p2p-circuit/p2p/16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq",
			wantError: false, // Note: multiaddr library accepts this
		},
		{
			name:      "completely invalid multiaddr",
			addr:      "not-a-valid-multiaddr-at-all",
			wantError: true,
		},
		{
			name:      "missing transport protocol still parses",
			addr:      "/ip4/192.168.1.1/p2p/16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq",
			wantError: false, // Note: multiaddr library accepts this (QUIC can work this way)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReachableAddr(tt.addr)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseReachableAddr(%q) error = %v, wantError %v", tt.addr, err, tt.wantError)
			}
		})
	}
}

// =============================================================================
// AccountType Edge Cases
// =============================================================================

func TestAccountType_EdgeCases(t *testing.T) {
	t.Run("unspecified type validation", func(t *testing.T) {
		err := AccountTypeUnspecified.Validate()
		if err == nil {
			t.Error("Unspecified account type should fail validation")
		}
	})

	t.Run("invalid type value validation", func(t *testing.T) {
		invalidType := AccountType(99)
		err := invalidType.Validate()
		if err == nil {
			t.Error("invalid account type should fail validation")
		}
		if invalidType.IsValid() {
			t.Error("invalid account type IsValid() should return false")
		}
	})

	t.Run("type string for unknown value", func(t *testing.T) {
		unknownType := AccountType(99)
		str := unknownType.String()
		if !strings.Contains(str, "Unknown") {
			t.Errorf("unknown type string should contain 'Unknown', got %s", str)
		}
	})
}

// =============================================================================
// Validation Edge Cases
// =============================================================================

func TestValidateDIDMatchesKey_EdgeCases(t *testing.T) {
	t.Run("nil DID", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		err := ValidateDIDMatchesKey(nil, pubKey)
		if err == nil {
			t.Error("ValidateDIDMatchesKey should fail for nil DID")
		}
	})

	t.Run("zero public key", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		var zeroPubKey keylib.NeuronPublicKey

		err := ValidateDIDMatchesKey(did, zeroPubKey)
		if err == nil {
			t.Error("ValidateDIDMatchesKey should fail for zero public key")
		}
	})
}

func TestValidateParentAccount_EdgeCases(t *testing.T) {
	t.Run("nil DID for parent", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		err := ValidateParentAccount(pubKey, nil)
		if err == nil {
			t.Error("ValidateParentAccount should fail for nil DID")
		}
	})

	t.Run("zero public key for parent", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		err := ValidateParentAccount(zeroPubKey, did)
		if err == nil {
			t.Error("ValidateParentAccount should fail for zero public key")
		}
	})
}

func TestValidateChildAccount_EdgeCases(t *testing.T) {
	t.Run("same child and parent key", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		err := ValidateChildAccount(pubKey, pubKey)
		if err == nil {
			t.Error("ValidateChildAccount should fail when child equals parent")
		}
	})

	t.Run("zero child key", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()
		var zeroChildKey keylib.NeuronPublicKey

		err := ValidateChildAccount(zeroChildKey, parentPubKey)
		if err == nil {
			t.Error("ValidateChildAccount should fail for zero child key")
		}
	})

	t.Run("zero parent key", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()
		var zeroParentKey keylib.NeuronPublicKey

		err := ValidateChildAccount(childPubKey, zeroParentKey)
		if err == nil {
			t.Error("ValidateChildAccount should fail for zero parent key")
		}
	})
}
