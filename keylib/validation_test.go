package keylib

import (
	"math/big"
	"testing"
)

// Note: curveOrderN is defined in factory_test.go

// =============================================================================
// normalizeHex Tests
// =============================================================================

func TestNormalizeHex(t *testing.T) {
	t.Run("strips 0x prefix", func(t *testing.T) {
		result := normalizeHex("0xabc123")
		if result != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", result)
		}
	})

	t.Run("strips 0X prefix", func(t *testing.T) {
		result := normalizeHex("0Xabc123")
		if result != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", result)
		}
	})

	t.Run("lowercases input", func(t *testing.T) {
		result := normalizeHex("0xABC123")
		if result != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", result)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := normalizeHex("")
		if result != "" {
			t.Errorf("expected '', got '%s'", result)
		}
	})

	t.Run("handles just 0x", func(t *testing.T) {
		result := normalizeHex("0x")
		if result != "" {
			t.Errorf("expected '', got '%s'", result)
		}
	})

	t.Run("no prefix just lowercases", func(t *testing.T) {
		result := normalizeHex("ABC123")
		if result != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", result)
		}
	})
}

// =============================================================================
// padLeftZeros Tests
// =============================================================================

func TestPadLeftZeros(t *testing.T) {
	t.Run("pads short input", func(t *testing.T) {
		input := []byte{1, 2, 3}
		result := padLeftZeros(input, 5)
		expected := []byte{0, 0, 1, 2, 3}
		if len(result) != 5 {
			t.Errorf("expected length 5, got %d", len(result))
		}
		for i, b := range expected {
			if result[i] != b {
				t.Errorf("byte %d: expected %d, got %d", i, b, result[i])
			}
		}
	})

	t.Run("returns exact length if already correct", func(t *testing.T) {
		input := []byte{1, 2, 3}
		result := padLeftZeros(input, 3)
		if len(result) != 3 {
			t.Errorf("expected length 3, got %d", len(result))
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		input := []byte{}
		result := padLeftZeros(input, 5)
		if len(result) != 5 {
			t.Errorf("expected length 5, got %d", len(result))
		}
		for i, b := range result {
			if b != 0 {
				t.Errorf("byte %d: expected 0, got %d", i, b)
			}
		}
	})

	t.Run("handles target length 0", func(t *testing.T) {
		input := []byte{1, 2, 3}
		result := padLeftZeros(input, 0)
		// Should return input as-is since len > target
		if len(result) < 3 {
			t.Errorf("expected at least 3 bytes")
		}
	})
}

// =============================================================================
// Private Key Scalar Boundary Tests
// =============================================================================

func TestPrivateKeyScalarBoundaries(t *testing.T) {
	t.Run("scalar = 1 (smallest valid)", func(t *testing.T) {
		var b [32]byte
		b[31] = 1 // scalar = 1
		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("scalar=1 should be valid: %v", err)
		}
		if key.IsZero() {
			t.Error("valid key should not be zero")
		}
	})

	t.Run("scalar = N-1 (largest valid)", func(t *testing.T) {
		nMinus1 := new(big.Int).Sub(curveOrderN, big.NewInt(1))
		var b [32]byte
		nMinus1Bytes := nMinus1.Bytes()
		copy(b[32-len(nMinus1Bytes):], nMinus1Bytes)

		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("scalar=N-1 should be valid: %v", err)
		}
		if key.IsZero() {
			t.Error("valid key should not be zero")
		}
	})

	t.Run("scalar = N (invalid - equals curve order)", func(t *testing.T) {
		var b [32]byte
		nBytes := curveOrderN.Bytes()
		copy(b[32-len(nBytes):], nBytes)

		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("scalar=N should be invalid")
		}
	})

	t.Run("scalar = N+1 (invalid - exceeds curve order)", func(t *testing.T) {
		nPlus1 := new(big.Int).Add(curveOrderN, big.NewInt(1))
		var b [32]byte
		nPlus1Bytes := nPlus1.Bytes()
		copy(b[32-len(nPlus1Bytes):], nPlus1Bytes)

		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("scalar=N+1 should be invalid")
		}
	})

	t.Run("scalar = 0 (invalid)", func(t *testing.T) {
		var b [32]byte
		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("scalar=0 should be invalid")
		}
	})

	t.Run("scalar = all 0xFF (very large, invalid)", func(t *testing.T) {
		var b [32]byte
		for i := range b {
			b[i] = 0xFF
		}
		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("all 0xFF should be invalid (exceeds N)")
		}
	})
}

// =============================================================================
// Public Key Format Validation Tests
// =============================================================================

func TestPublicKeyFormatValidation(t *testing.T) {
	// Generate a valid key to get valid public key bytes
	privKey, _ := GeneratePrivateKey()
	validCompressed := privKey.PublicKey().CompressedBytes()
	validUncompressed := privKey.PublicKey().UncompressedBytes()

	t.Run("accepts compressed 0x02 prefix", func(t *testing.T) {
		if validCompressed[0] == 0x02 {
			_, err := PublicKeyFromBytes(validCompressed[:])
			if err != nil {
				t.Errorf("0x02 prefix should be valid: %v", err)
			}
		}
	})

	t.Run("accepts compressed 0x03 prefix", func(t *testing.T) {
		if validCompressed[0] == 0x03 {
			_, err := PublicKeyFromBytes(validCompressed[:])
			if err != nil {
				t.Errorf("0x03 prefix should be valid: %v", err)
			}
		}
	})

	t.Run("accepts uncompressed 0x04 prefix", func(t *testing.T) {
		_, err := PublicKeyFromBytes(validUncompressed[:])
		if err != nil {
			t.Errorf("0x04 prefix should be valid: %v", err)
		}
	})

	t.Run("rejects 0x00 prefix", func(t *testing.T) {
		invalid := make([]byte, 33)
		invalid[0] = 0x00
		copy(invalid[1:], validCompressed[1:])
		_, err := PublicKeyFromBytes(invalid)
		if err == nil {
			t.Error("0x00 prefix should be invalid")
		}
	})

	t.Run("rejects 0x01 prefix", func(t *testing.T) {
		invalid := make([]byte, 33)
		invalid[0] = 0x01
		copy(invalid[1:], validCompressed[1:])
		_, err := PublicKeyFromBytes(invalid)
		if err == nil {
			t.Error("0x01 prefix should be invalid")
		}
	})

	t.Run("rejects 0x05 prefix", func(t *testing.T) {
		invalid := make([]byte, 33)
		invalid[0] = 0x05
		copy(invalid[1:], validCompressed[1:])
		_, err := PublicKeyFromBytes(invalid)
		if err == nil {
			t.Error("0x05 prefix should be invalid")
		}
	})

	t.Run("rejects wrong length 32", func(t *testing.T) {
		_, err := PublicKeyFromBytes(validCompressed[:32])
		if err == nil {
			t.Error("32 bytes should be invalid")
		}
	})

	t.Run("rejects wrong length 34", func(t *testing.T) {
		invalid := make([]byte, 34)
		copy(invalid, validCompressed[:])
		_, err := PublicKeyFromBytes(invalid)
		if err == nil {
			t.Error("34 bytes should be invalid")
		}
	})

	t.Run("rejects wrong length 64", func(t *testing.T) {
		_, err := PublicKeyFromBytes(validUncompressed[:64])
		if err == nil {
			t.Error("64 bytes should be invalid")
		}
	})

	t.Run("rejects wrong length 66", func(t *testing.T) {
		invalid := make([]byte, 66)
		copy(invalid, validUncompressed[:])
		_, err := PublicKeyFromBytes(invalid)
		if err == nil {
			t.Error("66 bytes should be invalid")
		}
	})
}

// =============================================================================
// Hex Validation Tests
// =============================================================================

func TestHexValidation(t *testing.T) {
	t.Run("valid hex chars accepted", func(t *testing.T) {
		// Generate a key and get its hex
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		_, err := ParsePrivateKeyHex(hex)
		if err != nil {
			t.Errorf("valid hex should be accepted: %v", err)
		}
	})

	t.Run("rejects 'g' character", func(t *testing.T) {
		invalid := "0xg364f2f1e5f4f03d1df682322500b9c68c997ec3a1234567890abcdef12345678"
		_, err := ParsePrivateKeyHex(invalid)
		if err == nil {
			t.Error("'g' should be rejected")
		}
	})

	t.Run("rejects 'z' character", func(t *testing.T) {
		invalid := "0xz364f2f1e5f4f03d1df682322500b9c68c997ec3a1234567890abcdef12345678"
		_, err := ParsePrivateKeyHex(invalid)
		if err == nil {
			t.Error("'z' should be rejected")
		}
	})

	t.Run("rejects space character", func(t *testing.T) {
		invalid := "0xe364 f2f1e5f4f03d1df682322500b9c68c997ec3a1234567890abcdef12345678"
		_, err := ParsePrivateKeyHex(invalid)
		if err == nil {
			t.Error("space should be rejected")
		}
	})

	t.Run("accepts 0-9", func(t *testing.T) {
		valid := "0x0123456789012345678901234567890123456789012345678901234567890001"
		_, err := ParsePrivateKeyHex(valid)
		if err != nil {
			t.Errorf("0-9 should be accepted: %v", err)
		}
	})

	t.Run("accepts a-f lowercase", func(t *testing.T) {
		valid := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef0001"
		_, err := ParsePrivateKeyHex(valid)
		if err != nil {
			t.Errorf("a-f should be accepted: %v", err)
		}
	})

	t.Run("accepts A-F uppercase", func(t *testing.T) {
		valid := "0xABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0001"
		_, err := ParsePrivateKeyHex(valid)
		if err != nil {
			t.Errorf("A-F should be accepted: %v", err)
		}
	})
}

// =============================================================================
// Error Kind Tests
// =============================================================================

func TestErrorKindString(t *testing.T) {
	tests := []struct {
		kind     ErrorKind
		expected string
	}{
		{ErrKindInvalidFormat, "InvalidFormat"},
		{ErrKindInvalidLength, "InvalidLength"},
		{ErrKindInvalidHex, "InvalidHex"},
		{ErrKindInvalidKey, "InvalidKey"},
		{ErrKindZeroValue, "ZeroValue"},
		{ErrKindKeyMismatch, "KeyMismatch"},
		{ErrKindEncryption, "Encryption"},
		{ErrKindMnemonic, "Mnemonic"},
		{ErrKindDerivation, "Derivation"},
		{ErrKindUnsupportedKeyType, "UnsupportedKeyType"},
		{ErrorKind(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.kind.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.kind.String())
			}
		})
	}
}

// =============================================================================
// KeyError Tests
// =============================================================================

func TestKeyError(t *testing.T) {
	t.Run("Error format with underlying error", func(t *testing.T) {
		err := &KeyError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidKey,
			Details: "test details",
			Err:     errInvalidKey("inner", "inner reason"),
		}
		errStr := err.Error()
		if errStr == "" {
			t.Error("error string should not be empty")
		}
	})

	t.Run("Error format without underlying error", func(t *testing.T) {
		err := &KeyError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidKey,
			Details: "test details",
		}
		errStr := err.Error()
		if errStr == "" {
			t.Error("error string should not be empty")
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		inner := errInvalidKey("inner", "inner reason")
		err := &KeyError{
			Op:  "TestOp",
			Err: inner,
		}
		unwrapped := err.Unwrap()
		if unwrapped != inner {
			t.Error("Unwrap should return underlying error")
		}
	})
}
