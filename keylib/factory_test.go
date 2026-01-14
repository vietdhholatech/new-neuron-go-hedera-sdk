package keylib

import (
	"encoding/hex"
	"math/big"
	"testing"
)

// Test vectors and helper values
var (
	// secp256k1 curve order N
	curveOrderN, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	// Valid test key (from integration tests)
	testValidPrivKeyHex = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
)

// =============================================================================
// PrivateKeyFromBytes Tests
// =============================================================================

func TestPrivateKeyFromBytes(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		bytes, _ := hex.DecodeString(testValidPrivKeyHex)
		var b [32]byte
		copy(b[:], bytes)

		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("PrivateKeyFromBytes failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("smallest valid scalar (1)", func(t *testing.T) {
		var b [32]byte
		b[31] = 1 // scalar = 1

		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("PrivateKeyFromBytes(1) failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("largest valid scalar (N-1)", func(t *testing.T) {
		nMinus1 := new(big.Int).Sub(curveOrderN, big.NewInt(1))
		var b [32]byte
		nBytes := nMinus1.Bytes()
		copy(b[32-len(nBytes):], nBytes)

		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("PrivateKeyFromBytes(N-1) failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("all zeros rejected", func(t *testing.T) {
		var b [32]byte // all zeros

		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("PrivateKeyFromBytes should reject all-zeros")
		}
		// Verify error kind
		keyErr, ok := err.(*KeyError)
		if !ok {
			t.Errorf("expected *KeyError, got %T", err)
		} else if keyErr.Kind != ErrKindInvalidKey {
			t.Errorf("expected ErrKindInvalidKey, got %v", keyErr.Kind)
		}
	})

	t.Run("scalar equal to N rejected", func(t *testing.T) {
		var b [32]byte
		nBytes := curveOrderN.Bytes()
		copy(b[32-len(nBytes):], nBytes)

		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("PrivateKeyFromBytes should reject scalar = N")
		}
	})

	t.Run("scalar greater than N rejected", func(t *testing.T) {
		nPlus1 := new(big.Int).Add(curveOrderN, big.NewInt(1))
		var b [32]byte
		nBytes := nPlus1.Bytes()
		copy(b[32-len(nBytes):], nBytes)

		_, err := PrivateKeyFromBytes(b)
		if err == nil {
			t.Error("PrivateKeyFromBytes should reject scalar > N")
		}
	})

	t.Run("leading zeros valid", func(t *testing.T) {
		// Key with many leading zeros but still valid
		var b [32]byte
		b[30] = 0x01
		b[31] = 0x00

		key, err := PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatalf("PrivateKeyFromBytes failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})
}

// =============================================================================
// PublicKeyFromBytes Tests
// =============================================================================

func TestPublicKeyFromBytes(t *testing.T) {
	// Generate a valid key for testing
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	compressedBytes := pubKey.CompressedBytes()
	uncompressedBytes := pubKey.UncompressedBytes()

	t.Run("valid compressed 0x02", func(t *testing.T) {
		// Use actual compressed bytes
		key, err := PublicKeyFromBytes(compressedBytes[:])
		if err != nil {
			t.Fatalf("PublicKeyFromBytes(compressed) failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("valid uncompressed 0x04", func(t *testing.T) {
		key, err := PublicKeyFromBytes(uncompressedBytes[:])
		if err != nil {
			t.Fatalf("PublicKeyFromBytes(uncompressed) failed: %v", err)
		}
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("wrong prefix 0x01 rejected", func(t *testing.T) {
		badBytes := make([]byte, 33)
		copy(badBytes, compressedBytes[:])
		badBytes[0] = 0x01

		_, err := PublicKeyFromBytes(badBytes)
		if err == nil {
			t.Error("PublicKeyFromBytes should reject prefix 0x01")
		}
	})

	t.Run("wrong prefix 0x05 rejected", func(t *testing.T) {
		badBytes := make([]byte, 33)
		copy(badBytes, compressedBytes[:])
		badBytes[0] = 0x05

		_, err := PublicKeyFromBytes(badBytes)
		if err == nil {
			t.Error("PublicKeyFromBytes should reject prefix 0x05")
		}
	})

	t.Run("wrong length 32 bytes rejected", func(t *testing.T) {
		badBytes := make([]byte, 32)

		_, err := PublicKeyFromBytes(badBytes)
		if err == nil {
			t.Error("PublicKeyFromBytes should reject 32 bytes")
		}
		keyErr, ok := err.(*KeyError)
		if !ok {
			t.Errorf("expected *KeyError, got %T", err)
		} else if keyErr.Kind != ErrKindInvalidLength {
			t.Errorf("expected ErrKindInvalidLength, got %v", keyErr.Kind)
		}
	})

	t.Run("wrong length 34 bytes rejected", func(t *testing.T) {
		badBytes := make([]byte, 34)

		_, err := PublicKeyFromBytes(badBytes)
		if err == nil {
			t.Error("PublicKeyFromBytes should reject 34 bytes")
		}
	})

	t.Run("wrong length 64 bytes rejected", func(t *testing.T) {
		badBytes := make([]byte, 64)

		_, err := PublicKeyFromBytes(badBytes)
		if err == nil {
			t.Error("PublicKeyFromBytes should reject 64 bytes")
		}
	})

	t.Run("empty bytes rejected", func(t *testing.T) {
		_, err := PublicKeyFromBytes([]byte{})
		if err == nil {
			t.Error("PublicKeyFromBytes should reject empty bytes")
		}
	})

	t.Run("round-trip compressed", func(t *testing.T) {
		restored, err := PublicKeyFromBytes(compressedBytes[:])
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("round-trip should produce equal keys")
		}
	})

	t.Run("round-trip uncompressed", func(t *testing.T) {
		restored, err := PublicKeyFromBytes(uncompressedBytes[:])
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("round-trip should produce equal keys")
		}
	})
}

// =============================================================================
// PrivateKeyFromHedera Tests
// =============================================================================

func TestPrivateKeyFromHedera(t *testing.T) {
	t.Run("valid ECDSA key conversion", func(t *testing.T) {
		// Generate Neuron key, convert to Hedera, convert back
		original, _ := GeneratePrivateKey()
		hederaKey := original.ToHederaPrivateKey()

		restored, err := PrivateKeyFromHedera(hederaKey)
		if err != nil {
			t.Fatalf("PrivateKeyFromHedera failed: %v", err)
		}

		if !original.Equal(restored) {
			t.Error("round-trip should produce equal keys")
		}
	})

	t.Run("round-trip preserves public key", func(t *testing.T) {
		original, _ := GeneratePrivateKey()
		originalPub := original.PublicKey()

		hederaKey := original.ToHederaPrivateKey()
		restored, _ := PrivateKeyFromHedera(hederaKey)
		restoredPub := restored.PublicKey()

		if !originalPub.Equal(restoredPub) {
			t.Error("public keys should match after round-trip")
		}
	})

	t.Run("round-trip preserves EVM address", func(t *testing.T) {
		original, _ := GeneratePrivateKey()
		originalEVM := original.EVMAddress()

		hederaKey := original.ToHederaPrivateKey()
		restored, _ := PrivateKeyFromHedera(hederaKey)
		restoredEVM := restored.EVMAddress()

		if !originalEVM.Equal(restoredEVM) {
			t.Error("EVM addresses should match after round-trip")
		}
	})
}

// =============================================================================
// PublicKeyFromHedera Tests
// =============================================================================

func TestPublicKeyFromHedera(t *testing.T) {
	t.Run("valid ECDSA public key conversion", func(t *testing.T) {
		privKey, _ := GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		hederaPub := pubKey.ToHederaPublicKey()

		restored, err := PublicKeyFromHedera(hederaPub)
		if err != nil {
			t.Fatalf("PublicKeyFromHedera failed: %v", err)
		}

		if !pubKey.Equal(restored) {
			t.Error("round-trip should produce equal keys")
		}
	})

	t.Run("round-trip preserves EVM address", func(t *testing.T) {
		privKey, _ := GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		originalEVM := pubKey.EVMAddress()

		hederaPub := pubKey.ToHederaPublicKey()
		restored, _ := PublicKeyFromHedera(hederaPub)
		restoredEVM := restored.EVMAddress()

		if !originalEVM.Equal(restoredEVM) {
			t.Error("EVM addresses should match after round-trip")
		}
	})

	t.Run("round-trip preserves PeerID", func(t *testing.T) {
		privKey, _ := GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		originalPeerID, _ := pubKey.PeerID()

		hederaPub := pubKey.ToHederaPublicKey()
		restored, _ := PublicKeyFromHedera(hederaPub)
		restoredPeerID, _ := restored.PeerID()

		if !originalPeerID.Equal(restoredPeerID) {
			t.Error("PeerIDs should match after round-trip")
		}
	})
}

// =============================================================================
// ParsePublicKeyHex Uncompressed Format Tests
// =============================================================================

func TestParsePublicKeyHex_Uncompressed(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	uncompressedHex := pubKey.HexUncompressed()

	t.Run("valid uncompressed with 0x prefix", func(t *testing.T) {
		key, err := ParsePublicKeyHex(uncompressedHex)
		if err != nil {
			t.Fatalf("ParsePublicKeyHex(uncompressed) failed: %v", err)
		}
		if !key.Equal(pubKey) {
			t.Error("parsed key should equal original")
		}
	})

	t.Run("valid uncompressed without prefix", func(t *testing.T) {
		// Remove 0x prefix
		hexWithoutPrefix := uncompressedHex[2:]
		key, err := ParsePublicKeyHex(hexWithoutPrefix)
		if err != nil {
			t.Fatalf("ParsePublicKeyHex failed: %v", err)
		}
		if !key.Equal(pubKey) {
			t.Error("parsed key should equal original")
		}
	})

	t.Run("uncompressed wrong prefix rejected", func(t *testing.T) {
		// Create hex with wrong prefix (0x05 instead of 0x04)
		badHex := "05" + uncompressedHex[4:] // Replace "04" with "05"
		_, err := ParsePublicKeyHex(badHex)
		if err == nil {
			t.Error("ParsePublicKeyHex should reject wrong prefix")
		}
	})
}

// =============================================================================
// GeneratePrivateKey Tests
// =============================================================================

func TestGeneratePrivateKey(t *testing.T) {
	t.Run("generates valid key", func(t *testing.T) {
		key, err := GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GeneratePrivateKey failed: %v", err)
		}
		if key.IsZero() {
			t.Error("generated key should not be zero")
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, _ := GeneratePrivateKey()
		key2, _ := GeneratePrivateKey()

		if key1.Equal(key2) {
			t.Error("generated keys should be unique")
		}
	})

	t.Run("generated key can derive public key", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pub := key.PublicKey()
		if pub.IsZero() {
			t.Error("derived public key should not be zero")
		}
	})

	t.Run("generated key can derive EVM address", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		addr := key.EVMAddress()
		if addr.IsZero() {
			t.Error("derived EVM address should not be zero")
		}
	})

	t.Run("generated key can derive PeerID", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pid, err := key.PeerID()
		if err != nil {
			t.Fatalf("PeerID derivation failed: %v", err)
		}
		if pid.IsZero() {
			t.Error("derived PeerID should not be zero")
		}
	})
}

// =============================================================================
// PublicKeyFromPrivateKey Tests
// =============================================================================

func TestPublicKeyFromPrivateKey(t *testing.T) {
	t.Run("derives correct public key", func(t *testing.T) {
		privKey, _ := GeneratePrivateKey()
		pub1 := privKey.PublicKey()
		pub2 := PublicKeyFromPrivateKey(privKey)

		if !pub1.Equal(pub2) {
			t.Error("PublicKeyFromPrivateKey should match PublicKey()")
		}
	})

	t.Run("zero key returns zero public key", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		pub := PublicKeyFromPrivateKey(zeroKey)
		if !pub.IsZero() {
			t.Error("zero private key should produce zero public key")
		}
	})
}

// =============================================================================
// Must* Functions Tests
// =============================================================================

func TestMustParsePrivateKeyHex(t *testing.T) {
	t.Run("valid key succeeds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParsePrivateKeyHex panicked: %v", r)
			}
		}()

		key := MustParsePrivateKeyHex("0x" + testValidPrivKeyHex)
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("invalid key panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParsePrivateKeyHex should panic on invalid input")
			}
		}()

		MustParsePrivateKeyHex("invalid")
	})
}

func TestMustParsePublicKeyHex(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubHex := privKey.PublicKey().Hex()

	t.Run("valid key succeeds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParsePublicKeyHex panicked: %v", r)
			}
		}()

		key := MustParsePublicKeyHex(pubHex)
		if key.IsZero() {
			t.Error("key should not be zero")
		}
	})

	t.Run("invalid key panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParsePublicKeyHex should panic on invalid input")
			}
		}()

		MustParsePublicKeyHex("invalid")
	})
}

func TestMustParseEVMAddress(t *testing.T) {
	t.Run("valid address succeeds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParseEVMAddress panicked: %v", r)
			}
		}()

		addr := MustParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
		if addr.IsZero() {
			t.Error("address should not be zero")
		}
	})

	t.Run("invalid address panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseEVMAddress should panic on invalid input")
			}
		}()

		MustParseEVMAddress("invalid")
	})
}

func TestMustParsePeerID(t *testing.T) {
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	t.Run("valid peer ID succeeds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParsePeerID panicked: %v", r)
			}
		}()

		pid := MustParsePeerID(testPeerID)
		if pid.IsZero() {
			t.Error("peer ID should not be zero")
		}
	})

	t.Run("invalid peer ID panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParsePeerID should panic on invalid input")
			}
		}()

		MustParsePeerID("invalid")
	})
}

// =============================================================================
// IsEd25519* Helper Functions Tests
// =============================================================================

func TestIsEd25519Key(t *testing.T) {
	// Note: These tests verify the heuristic detection functions
	// Actual Ed25519 key testing would require creating Ed25519 keys via Hedera SDK

	t.Run("ECDSA key returns false", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hederaKey := key.ToHederaPrivateKey()

		if IsEd25519Key(hederaKey) {
			t.Error("ECDSA key should not be detected as Ed25519")
		}
	})
}

func TestIsEd25519PublicKey(t *testing.T) {
	t.Run("ECDSA public key returns false", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hederaPub := key.PublicKey().ToHederaPublicKey()

		if IsEd25519PublicKey(hederaPub) {
			t.Error("ECDSA public key should not be detected as Ed25519")
		}
	})
}
