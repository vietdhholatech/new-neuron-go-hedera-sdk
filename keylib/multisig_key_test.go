package keylib

import (
	"strings"
	"testing"
)

// =============================================================================
// Helper Functions
// =============================================================================

// generateTestPublicKeys generates n distinct public keys for testing.
func generateTestPublicKeys(t *testing.T, n int) []NeuronPublicKey {
	t.Helper()
	keys := make([]NeuronPublicKey, n)
	for i := 0; i < n; i++ {
		privKey, err := GeneratePrivateKey()
		if err != nil {
			t.Fatalf("failed to generate private key: %v", err)
		}
		keys[i] = privKey.PublicKey()
	}
	return keys
}

// =============================================================================
// NewMultisigKey Tests
// =============================================================================

func TestNewMultisigKey_ValidConfigurations(t *testing.T) {
	t.Run("1-of-1 configuration", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 1)
		mk, err := NewMultisigKey(keys, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Threshold() != 1 {
			t.Errorf("expected threshold 1, got %d", mk.Threshold())
		}
		if mk.Total() != 1 {
			t.Errorf("expected total 1, got %d", mk.Total())
		}
	})

	t.Run("2-of-3 configuration", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		mk, err := NewMultisigKey(keys, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Threshold() != 2 {
			t.Errorf("expected threshold 2, got %d", mk.Threshold())
		}
		if mk.Total() != 3 {
			t.Errorf("expected total 3, got %d", mk.Total())
		}
	})

	t.Run("3-of-5 configuration", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 5)
		mk, err := NewMultisigKey(keys, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Threshold() != 3 {
			t.Errorf("expected threshold 3, got %d", mk.Threshold())
		}
		if mk.Total() != 5 {
			t.Errorf("expected total 5, got %d", mk.Total())
		}
	})

	t.Run("n-of-n configuration", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 4)
		mk, err := NewMultisigKey(keys, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Threshold() != 4 {
			t.Errorf("expected threshold 4, got %d", mk.Threshold())
		}
		if mk.Total() != 4 {
			t.Errorf("expected total 4, got %d", mk.Total())
		}
	})
}

func TestNewMultisigKey_InvalidThreshold(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)

	t.Run("threshold exceeds total", func(t *testing.T) {
		_, err := NewMultisigKey(keys, 4)
		if err == nil {
			t.Fatal("expected error for threshold > total")
		}
		if !strings.Contains(err.Error(), "threshold cannot exceed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("threshold is zero", func(t *testing.T) {
		_, err := NewMultisigKey(keys, 0)
		if err == nil {
			t.Fatal("expected error for zero threshold")
		}
		if !strings.Contains(err.Error(), "threshold must be at least 1") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("negative threshold", func(t *testing.T) {
		_, err := NewMultisigKey(keys, -1)
		if err == nil {
			t.Fatal("expected error for negative threshold")
		}
		if !strings.Contains(err.Error(), "threshold must be at least 1") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestNewMultisigKey_EmptyPublicKeys(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		_, err := NewMultisigKey(nil, 1)
		if err == nil {
			t.Fatal("expected error for nil keys")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		_, err := NewMultisigKey([]NeuronPublicKey{}, 1)
		if err == nil {
			t.Fatal("expected error for empty keys")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestNewMultisigKey_DuplicateKeys(t *testing.T) {
	keys := generateTestPublicKeys(t, 2)

	t.Run("duplicate key in list", func(t *testing.T) {
		duplicated := []NeuronPublicKey{keys[0], keys[1], keys[0]}
		_, err := NewMultisigKey(duplicated, 2)
		if err == nil {
			t.Fatal("expected error for duplicate keys")
		}
		if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestNewMultisigKey_ZeroValueKey(t *testing.T) {
	keys := generateTestPublicKeys(t, 2)

	t.Run("zero-value key in slice", func(t *testing.T) {
		withZero := []NeuronPublicKey{keys[0], {}, keys[1]}
		_, err := NewMultisigKey(withZero, 2)
		if err == nil {
			t.Fatal("expected error for zero-value key")
		}
		if !strings.Contains(err.Error(), "zero-value") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// =============================================================================
// MultisigKey.IsZero Tests
// =============================================================================

func TestMultisigKey_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var mk MultisigKey
		if !mk.IsZero() {
			t.Error("zero-value MultisigKey should be zero")
		}
	})

	t.Run("valid key is not zero", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 2)
		mk, err := NewMultisigKey(keys, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.IsZero() {
			t.Error("valid MultisigKey should not be zero")
		}
	})
}

// =============================================================================
// MultisigKey.Validate Tests
// =============================================================================

func TestMultisigKey_Validate(t *testing.T) {
	t.Run("valid key passes validation", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		mk, err := NewMultisigKey(keys, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := mk.Validate(); err != nil {
			t.Errorf("valid key should pass validation: %v", err)
		}
	})

	t.Run("zero value fails validation", func(t *testing.T) {
		var mk MultisigKey
		if err := mk.Validate(); err == nil {
			t.Error("zero-value should fail validation")
		}
	})
}

// =============================================================================
// MultisigKey.PublicKeys Tests
// =============================================================================

func TestMultisigKey_PublicKeys_DefensiveCopy(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)
	mk, err := NewMultisigKey(keys, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("returns all keys", func(t *testing.T) {
		returned := mk.PublicKeys()
		if len(returned) != 3 {
			t.Errorf("expected 3 keys, got %d", len(returned))
		}
	})

	t.Run("modifying returned slice does not affect original", func(t *testing.T) {
		returned := mk.PublicKeys()
		newKey, _ := GeneratePrivateKey()
		returned[0] = newKey.PublicKey()

		// Get keys again and verify they're unchanged
		returned2 := mk.PublicKeys()
		if returned2[0].Equal(returned[0]) {
			t.Error("modifying returned slice should not affect original")
		}
	})

	t.Run("zero value returns nil", func(t *testing.T) {
		var mk MultisigKey
		if mk.PublicKeys() != nil {
			t.Error("zero-value should return nil")
		}
	})
}

// =============================================================================
// MultisigKey.ContainsKey Tests
// =============================================================================

func TestMultisigKey_ContainsKey(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)
	mk, err := NewMultisigKey(keys, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("contains included key", func(t *testing.T) {
		for i, key := range keys {
			if !mk.ContainsKey(key) {
				t.Errorf("should contain key at index %d", i)
			}
		}
	})

	t.Run("does not contain other key", func(t *testing.T) {
		otherKey, _ := GeneratePrivateKey()
		if mk.ContainsKey(otherKey.PublicKey()) {
			t.Error("should not contain unrelated key")
		}
	})

	t.Run("does not contain zero key", func(t *testing.T) {
		var zeroKey NeuronPublicKey
		if mk.ContainsKey(zeroKey) {
			t.Error("should not contain zero key")
		}
	})

	t.Run("zero MultisigKey does not contain any key", func(t *testing.T) {
		var mk MultisigKey
		if mk.ContainsKey(keys[0]) {
			t.Error("zero MultisigKey should not contain any key")
		}
	})
}

// =============================================================================
// MultisigKey.Equal Tests
// =============================================================================

func TestMultisigKey_Equal(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)

	t.Run("same keys same threshold are equal", func(t *testing.T) {
		mk1, _ := NewMultisigKey(keys, 2)
		mk2, _ := NewMultisigKey(keys, 2)
		if !mk1.Equal(mk2) {
			t.Error("identical MultisigKeys should be equal")
		}
	})

	t.Run("same keys different threshold are not equal", func(t *testing.T) {
		mk1, _ := NewMultisigKey(keys, 2)
		mk2, _ := NewMultisigKey(keys, 3)
		if mk1.Equal(mk2) {
			t.Error("different thresholds should not be equal")
		}
	})

	t.Run("different keys same threshold are not equal", func(t *testing.T) {
		otherKeys := generateTestPublicKeys(t, 3)
		mk1, _ := NewMultisigKey(keys, 2)
		mk2, _ := NewMultisigKey(otherKeys, 2)
		if mk1.Equal(mk2) {
			t.Error("different keys should not be equal")
		}
	})

	t.Run("reordered keys are equal (deterministic order)", func(t *testing.T) {
		reordered := []NeuronPublicKey{keys[2], keys[0], keys[1]}
		mk1, _ := NewMultisigKey(keys, 2)
		mk2, _ := NewMultisigKey(reordered, 2)
		if !mk1.Equal(mk2) {
			t.Error("same keys in different order should be equal after sorting")
		}
	})

	t.Run("zero values are equal", func(t *testing.T) {
		var mk1, mk2 MultisigKey
		if !mk1.Equal(mk2) {
			t.Error("zero values should be equal")
		}
	})

	t.Run("zero and non-zero are not equal", func(t *testing.T) {
		var mk1 MultisigKey
		mk2, _ := NewMultisigKey(keys, 2)
		if mk1.Equal(mk2) {
			t.Error("zero and non-zero should not be equal")
		}
	})

	t.Run("same keys same threshold different protocol are not equal", func(t *testing.T) {
		mk1, _ := NewMultisigKey(keys, 2)
		mk2, _ := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		if mk1.Equal(mk2) {
			t.Error("different protocols should not be equal")
		}
	})

	t.Run("same protocol are equal", func(t *testing.T) {
		mk1, _ := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		mk2, _ := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		if !mk1.Equal(mk2) {
			t.Error("same protocol should be equal")
		}
	})
}

// =============================================================================
// MultisigKey.String Tests
// =============================================================================

func TestMultisigKey_String(t *testing.T) {
	t.Run("2-of-3 format with protocol", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		mk, _ := NewMultisigKey(keys, 2)
		str := mk.String()
		if !strings.Contains(str, "2-of-3") {
			t.Errorf("expected '2-of-3' in string, got: %s", str)
		}
		if !strings.Contains(str, "protocol=secp256k1-aggregated") {
			t.Errorf("expected protocol in string, got: %s", str)
		}
	})

	t.Run("custom protocol in string", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		mk, _ := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		str := mk.String()
		if !strings.Contains(str, "protocol=hedera-threshold") {
			t.Errorf("expected hedera-threshold protocol in string, got: %s", str)
		}
	})

	t.Run("zero value format", func(t *testing.T) {
		var mk MultisigKey
		str := mk.String()
		if !strings.Contains(str, "zero-value") {
			t.Errorf("expected 'zero-value' in string, got: %s", str)
		}
	})
}

// =============================================================================
// MustNewMultisigKey Tests
// =============================================================================

func TestMustNewMultisigKey(t *testing.T) {
	t.Run("valid configuration does not panic", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("should not panic for valid config: %v", r)
			}
		}()
		mk := MustNewMultisigKey(keys, 2)
		if mk.IsZero() {
			t.Error("should return valid MultisigKey")
		}
	})

	t.Run("invalid configuration panics", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 3)
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic for invalid config")
			}
		}()
		_ = MustNewMultisigKey(keys, 5) // threshold > total
	})
}

// =============================================================================
// NewMultisigKey Protocol Tests
// =============================================================================

func TestNewMultisigKey_DefaultProtocol(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)
	mk, err := NewMultisigKey(keys, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mk.Protocol() != MultisigProtocolSecp256k1Aggregated {
		t.Errorf("expected default protocol %q, got %q", MultisigProtocolSecp256k1Aggregated, mk.Protocol())
	}
}

// =============================================================================
// NewMultisigKeyWithProtocol Tests
// =============================================================================

func TestNewMultisigKeyWithProtocol(t *testing.T) {
	keys := generateTestPublicKeys(t, 3)

	t.Run("hedera-threshold protocol", func(t *testing.T) {
		mk, err := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Protocol() != MultisigProtocolHederaThreshold {
			t.Errorf("expected protocol %q, got %q", MultisigProtocolHederaThreshold, mk.Protocol())
		}
		if mk.Threshold() != 2 {
			t.Errorf("expected threshold 2, got %d", mk.Threshold())
		}
	})

	t.Run("frost protocol", func(t *testing.T) {
		mk, err := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolFROST)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Protocol() != MultisigProtocolFROST {
			t.Errorf("expected protocol %q, got %q", MultisigProtocolFROST, mk.Protocol())
		}
	})

	t.Run("bls protocol", func(t *testing.T) {
		mk, err := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolBLS)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Protocol() != MultisigProtocolBLS {
			t.Errorf("expected protocol %q, got %q", MultisigProtocolBLS, mk.Protocol())
		}
	})

	t.Run("custom protocol", func(t *testing.T) {
		mk, err := NewMultisigKeyWithProtocol(keys, 2, "custom-threshold")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mk.Protocol() != "custom-threshold" {
			t.Errorf("expected protocol %q, got %q", "custom-threshold", mk.Protocol())
		}
	})

	t.Run("empty protocol is rejected", func(t *testing.T) {
		_, err := NewMultisigKeyWithProtocol(keys, 2, "")
		if err == nil {
			t.Fatal("expected error for empty protocol")
		}
		if !strings.Contains(err.Error(), "protocol must not be empty") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("invalid threshold still fails", func(t *testing.T) {
		_, err := NewMultisigKeyWithProtocol(keys, 5, MultisigProtocolHederaThreshold)
		if err == nil {
			t.Fatal("expected error for invalid threshold")
		}
	})
}

// =============================================================================
// MultisigKey.Protocol Tests
// =============================================================================

func TestMultisigKey_Protocol(t *testing.T) {
	t.Run("default protocol accessor", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 2)
		mk, _ := NewMultisigKey(keys, 2)
		if mk.Protocol() != MultisigProtocolSecp256k1Aggregated {
			t.Errorf("expected %q, got %q", MultisigProtocolSecp256k1Aggregated, mk.Protocol())
		}
	})

	t.Run("custom protocol accessor", func(t *testing.T) {
		keys := generateTestPublicKeys(t, 2)
		mk, _ := NewMultisigKeyWithProtocol(keys, 2, MultisigProtocolHederaThreshold)
		if mk.Protocol() != MultisigProtocolHederaThreshold {
			t.Errorf("expected %q, got %q", MultisigProtocolHederaThreshold, mk.Protocol())
		}
	})

	t.Run("zero value returns empty string", func(t *testing.T) {
		var mk MultisigKey
		if mk.Protocol() != "" {
			t.Errorf("expected empty string for zero value, got %q", mk.Protocol())
		}
	})
}
