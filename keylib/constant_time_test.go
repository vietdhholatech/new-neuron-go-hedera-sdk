package keylib

import (
	"testing"
)

// =============================================================================
// constantTimeEqual Tests
// =============================================================================

func TestConstantTimeEqual(t *testing.T) {
	t.Run("equal slices return true", func(t *testing.T) {
		a := []byte{1, 2, 3, 4, 5}
		b := []byte{1, 2, 3, 4, 5}
		if !constantTimeEqual(a, b) {
			t.Error("equal slices should return true")
		}
	})

	t.Run("different slices return false", func(t *testing.T) {
		a := []byte{1, 2, 3, 4, 5}
		b := []byte{1, 2, 3, 4, 6}
		if constantTimeEqual(a, b) {
			t.Error("different slices should return false")
		}
	})

	t.Run("different lengths return false", func(t *testing.T) {
		a := []byte{1, 2, 3}
		b := []byte{1, 2, 3, 4}
		if constantTimeEqual(a, b) {
			t.Error("different length slices should return false")
		}
	})

	t.Run("empty slices are equal", func(t *testing.T) {
		a := []byte{}
		b := []byte{}
		if !constantTimeEqual(a, b) {
			t.Error("empty slices should be equal")
		}
	})

	t.Run("nil slices are equal", func(t *testing.T) {
		var a, b []byte
		if !constantTimeEqual(a, b) {
			t.Error("nil slices should be equal")
		}
	})

	t.Run("nil equals empty", func(t *testing.T) {
		var a []byte
		b := []byte{}
		if !constantTimeEqual(a, b) {
			t.Error("nil and empty slices should be equal")
		}
	})

	t.Run("single byte equal", func(t *testing.T) {
		a := []byte{0xFF}
		b := []byte{0xFF}
		if !constantTimeEqual(a, b) {
			t.Error("equal single byte slices should return true")
		}
	})

	t.Run("single byte different", func(t *testing.T) {
		a := []byte{0xFF}
		b := []byte{0xFE}
		if constantTimeEqual(a, b) {
			t.Error("different single byte slices should return false")
		}
	})

	t.Run("first byte different", func(t *testing.T) {
		a := []byte{0, 2, 3, 4, 5}
		b := []byte{1, 2, 3, 4, 5}
		if constantTimeEqual(a, b) {
			t.Error("slices differing in first byte should return false")
		}
	})

	t.Run("last byte different", func(t *testing.T) {
		a := []byte{1, 2, 3, 4, 0}
		b := []byte{1, 2, 3, 4, 5}
		if constantTimeEqual(a, b) {
			t.Error("slices differing in last byte should return false")
		}
	})

	t.Run("32-byte key comparison", func(t *testing.T) {
		a := make([]byte, 32)
		b := make([]byte, 32)
		for i := range a {
			a[i] = byte(i)
			b[i] = byte(i)
		}
		if !constantTimeEqual(a, b) {
			t.Error("equal 32-byte slices should return true")
		}

		// Change one byte
		b[16] = 0xFF
		if constantTimeEqual(a, b) {
			t.Error("different 32-byte slices should return false")
		}
	})
}

// =============================================================================
// constantTimeEqualStrings Tests
// =============================================================================

func TestConstantTimeEqualStrings(t *testing.T) {
	t.Run("equal strings return true", func(t *testing.T) {
		a := "hello world"
		b := "hello world"
		if !constantTimeEqualStrings(a, b) {
			t.Error("equal strings should return true")
		}
	})

	t.Run("different strings return false", func(t *testing.T) {
		a := "hello world"
		b := "hello worle"
		if constantTimeEqualStrings(a, b) {
			t.Error("different strings should return false")
		}
	})

	t.Run("different length strings return false", func(t *testing.T) {
		a := "hello"
		b := "hello world"
		if constantTimeEqualStrings(a, b) {
			t.Error("different length strings should return false")
		}
	})

	t.Run("empty strings are equal", func(t *testing.T) {
		a := ""
		b := ""
		if !constantTimeEqualStrings(a, b) {
			t.Error("empty strings should be equal")
		}
	})

	t.Run("case sensitive", func(t *testing.T) {
		a := "Hello"
		b := "hello"
		if constantTimeEqualStrings(a, b) {
			t.Error("strings with different case should return false")
		}
	})

	t.Run("hex string comparison", func(t *testing.T) {
		a := "0x1234567890abcdef"
		b := "0x1234567890abcdef"
		if !constantTimeEqualStrings(a, b) {
			t.Error("equal hex strings should return true")
		}
	})
}

// =============================================================================
// secureZero Tests
// =============================================================================

func TestSecureZero(t *testing.T) {
	t.Run("zeros all bytes", func(t *testing.T) {
		b := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		secureZero(b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("zeros single byte", func(t *testing.T) {
		b := []byte{0xFF}
		secureZero(b)
		if b[0] != 0 {
			t.Errorf("byte should be 0, got %d", b[0])
		}
	})

	t.Run("zeros 32-byte key", func(t *testing.T) {
		b := make([]byte, 32)
		for i := range b {
			b[i] = byte(i + 1)
		}
		secureZero(b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("zeros large slice", func(t *testing.T) {
		b := make([]byte, 1024)
		for i := range b {
			b[i] = 0xFF
		}
		secureZero(b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("empty slice no panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("secureZero should not panic on empty slice")
			}
		}()
		b := []byte{}
		secureZero(b)
	})

	t.Run("already zero slice", func(t *testing.T) {
		b := make([]byte, 16)
		secureZero(b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})
}

// =============================================================================
// secureZeroArray32 Tests
// =============================================================================

func TestSecureZeroArray32(t *testing.T) {
	t.Run("zeros all bytes", func(t *testing.T) {
		var b [32]byte
		for i := range b {
			b[i] = byte(i + 1)
		}
		secureZeroArray32(&b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("zeros max values", func(t *testing.T) {
		var b [32]byte
		for i := range b {
			b[i] = 0xFF
		}
		secureZeroArray32(&b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("already zero array", func(t *testing.T) {
		var b [32]byte
		secureZeroArray32(&b)
		for i, v := range b {
			if v != 0 {
				t.Errorf("byte at index %d should be 0, got %d", i, v)
			}
		}
	})

	t.Run("result equals zero array", func(t *testing.T) {
		var b [32]byte
		for i := range b {
			b[i] = byte(i + 1)
		}
		secureZeroArray32(&b)

		expected := [32]byte{}
		if b != expected {
			t.Error("zeroed array should equal zero array")
		}
	})
}

// =============================================================================
// Integration Tests - Security Critical Verification
// =============================================================================

func TestConstantTime_KeyComparison(t *testing.T) {
	// Generate two keys and verify constant-time comparison works
	key1, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key1: %v", err)
	}

	key2, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key2: %v", err)
	}

	t.Run("same key bytes are equal", func(t *testing.T) {
		bytes1 := key1.Bytes()
		bytes2 := key1.Bytes()
		if !constantTimeEqual(bytes1[:], bytes2[:]) {
			t.Error("same key bytes should be equal")
		}
	})

	t.Run("different key bytes are not equal", func(t *testing.T) {
		bytes1 := key1.Bytes()
		bytes2 := key2.Bytes()
		if constantTimeEqual(bytes1[:], bytes2[:]) {
			t.Error("different key bytes should not be equal")
		}
	})

	t.Run("public key comparison", func(t *testing.T) {
		pub1 := key1.PublicKey().CompressedBytes()
		pub2 := key1.PublicKey().CompressedBytes()
		if !constantTimeEqual(pub1[:], pub2[:]) {
			t.Error("same public key bytes should be equal")
		}
	})
}

func TestSecureZero_PrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Get the key bytes
	keyBytes := key.Bytes()

	// Verify key is non-zero
	allZero := true
	for _, b := range keyBytes {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("generated key should not be all zeros")
	}

	// Zero the bytes
	secureZeroArray32(&keyBytes)

	// Verify all bytes are zero
	for i, b := range keyBytes {
		if b != 0 {
			t.Errorf("byte at index %d should be zero after secureZero, got %d", i, b)
		}
	}
}
