package keylib

import (
	"strings"
	"testing"
)

// =============================================================================
// NeuronPublicKey.IsZero Tests
// =============================================================================

func TestNeuronPublicKey_IsZero(t *testing.T) {
	t.Run("derived key is not zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pubKey := key.PublicKey()
		if pubKey.IsZero() {
			t.Error("derived public key should not be zero")
		}
	})

	t.Run("zero value key is zero", func(t *testing.T) {
		var pubKey NeuronPublicKey
		if !pubKey.IsZero() {
			t.Error("zero value key should be zero")
		}
	})

	t.Run("parsed key is not zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.PublicKey().Hex()
		parsed, _ := ParsePublicKeyHex(hex)
		if parsed.IsZero() {
			t.Error("parsed key should not be zero")
		}
	})
}

// =============================================================================
// NeuronPublicKey.CompressedBytes Tests
// =============================================================================

func TestNeuronPublicKey_CompressedBytes(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns 33 bytes", func(t *testing.T) {
		bytes := pubKey.CompressedBytes()
		if len(bytes) != 33 {
			t.Errorf("expected 33 bytes, got %d", len(bytes))
		}
	})

	t.Run("starts with 02 or 03", func(t *testing.T) {
		bytes := pubKey.CompressedBytes()
		if bytes[0] != 0x02 && bytes[0] != 0x03 {
			t.Errorf("compressed key should start with 0x02 or 0x03, got 0x%02x", bytes[0])
		}
	})

	t.Run("round trip bytes", func(t *testing.T) {
		bytes := pubKey.CompressedBytes()
		restored, err := PublicKeyFromBytes(bytes[:])
		if err != nil {
			t.Fatalf("failed to restore from bytes: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})

	t.Run("zero key returns zero bytes", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		bytes := zeroPub.CompressedBytes()
		expected := [33]byte{}
		if bytes != expected {
			t.Error("zero key should return zero bytes")
		}
	})
}

// =============================================================================
// NeuronPublicKey.UncompressedBytes Tests
// =============================================================================

func TestNeuronPublicKey_UncompressedBytes(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns 65 bytes", func(t *testing.T) {
		bytes := pubKey.UncompressedBytes()
		if len(bytes) != 65 {
			t.Errorf("expected 65 bytes, got %d", len(bytes))
		}
	})

	t.Run("starts with 04", func(t *testing.T) {
		bytes := pubKey.UncompressedBytes()
		if bytes[0] != 0x04 {
			t.Errorf("uncompressed key should start with 0x04, got 0x%02x", bytes[0])
		}
	})

	t.Run("round trip bytes", func(t *testing.T) {
		bytes := pubKey.UncompressedBytes()
		restored, err := PublicKeyFromBytes(bytes[:])
		if err != nil {
			t.Fatalf("failed to restore from bytes: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})

	t.Run("zero key returns zero bytes", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		bytes := zeroPub.UncompressedBytes()
		expected := [65]byte{}
		if bytes != expected {
			t.Error("zero key should return zero bytes")
		}
	})
}

// =============================================================================
// NeuronPublicKey.Hex Tests
// =============================================================================

func TestNeuronPublicKey_Hex(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("includes 0x prefix", func(t *testing.T) {
		hex := pubKey.Hex()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("hex should start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		hex := pubKey.Hex()
		// 0x + 66 chars (33 bytes * 2)
		if len(hex) != 68 {
			t.Errorf("expected 68 chars, got %d", len(hex))
		}
	})

	t.Run("lowercase", func(t *testing.T) {
		hex := pubKey.Hex()
		if strings.ToLower(hex) != hex {
			t.Error("hex should be lowercase")
		}
	})

	t.Run("round trip hex", func(t *testing.T) {
		hex := pubKey.Hex()
		restored, err := ParsePublicKeyHex(hex)
		if err != nil {
			t.Fatalf("failed to restore from hex: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})

	t.Run("zero key returns empty", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		hex := zeroPub.Hex()
		if hex != "" {
			t.Errorf("zero key should return empty string, got %s", hex)
		}
	})
}

// =============================================================================
// NeuronPublicKey.HexUncompressed Tests
// =============================================================================

func TestNeuronPublicKey_HexUncompressed(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("includes 0x prefix", func(t *testing.T) {
		hex := pubKey.HexUncompressed()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("hex should start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		hex := pubKey.HexUncompressed()
		// 0x + 130 chars (65 bytes * 2)
		if len(hex) != 132 {
			t.Errorf("expected 132 chars, got %d", len(hex))
		}
	})

	t.Run("starts with 0x04", func(t *testing.T) {
		hex := pubKey.HexUncompressed()
		if !strings.HasPrefix(hex, "0x04") {
			t.Error("uncompressed hex should start with 0x04")
		}
	})

	t.Run("round trip hex", func(t *testing.T) {
		hex := pubKey.HexUncompressed()
		restored, err := ParsePublicKeyHex(hex)
		if err != nil {
			t.Fatalf("failed to restore from hex: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})

	t.Run("zero key returns empty", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		hex := zeroPub.HexUncompressed()
		if hex != "" {
			t.Errorf("zero key should return empty string, got %s", hex)
		}
	})
}

// =============================================================================
// NeuronPublicKey.EVMAddress Tests
// =============================================================================

func TestNeuronPublicKey_EVMAddress(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns valid EVM address", func(t *testing.T) {
		addr := pubKey.EVMAddress()
		if addr.IsZero() {
			t.Error("EVM address should not be zero")
		}
	})

	t.Run("EVM address is deterministic", func(t *testing.T) {
		addr1 := pubKey.EVMAddress()
		addr2 := pubKey.EVMAddress()
		if !addr1.Equal(addr2) {
			t.Error("EVM address should be deterministic")
		}
	})

	t.Run("zero key returns zero address", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		addr := zeroPub.EVMAddress()
		if !addr.IsZero() {
			t.Error("zero public key should return zero EVM address")
		}
	})

	t.Run("different keys have different addresses", func(t *testing.T) {
		key2, _ := GeneratePrivateKey()
		addr1 := pubKey.EVMAddress()
		addr2 := key2.PublicKey().EVMAddress()
		if addr1.Equal(addr2) {
			t.Error("different public keys should have different EVM addresses")
		}
	})
}

// =============================================================================
// NeuronPublicKey.PeerID Tests
// =============================================================================

func TestNeuronPublicKey_PeerID(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns valid peer ID", func(t *testing.T) {
		pid, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("PeerID failed: %v", err)
		}
		if pid.IsZero() {
			t.Error("peer ID should not be zero")
		}
	})

	t.Run("peer ID is deterministic", func(t *testing.T) {
		pid1, _ := pubKey.PeerID()
		pid2, _ := pubKey.PeerID()
		if !pid1.Equal(pid2) {
			t.Error("peer ID should be deterministic")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		_, err := zeroPub.PeerID()
		if err == nil {
			t.Error("zero public key should return error for PeerID")
		}
	})
}

// =============================================================================
// NeuronPublicKey.ToECDSA Tests
// =============================================================================

func TestNeuronPublicKey_ToECDSA(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns valid ECDSA key", func(t *testing.T) {
		ecdsaKey := pubKey.ToECDSA()
		if ecdsaKey == nil {
			t.Error("should return non-nil ECDSA key")
		}
	})

	t.Run("ECDSA key has correct curve", func(t *testing.T) {
		ecdsaKey := pubKey.ToECDSA()
		if ecdsaKey.Curve == nil {
			t.Error("ECDSA key should have curve")
		}
	})

	t.Run("zero key returns nil", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		ecdsaKey := zeroPub.ToECDSA()
		if ecdsaKey != nil {
			t.Error("zero key should return nil ECDSA key")
		}
	})
}

// =============================================================================
// NeuronPublicKey.ToHederaPublicKey Tests
// =============================================================================

func TestNeuronPublicKey_ToHederaPublicKey(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("returns valid Hedera key", func(t *testing.T) {
		hederaKey := pubKey.ToHederaPublicKey()
		// Hedera key string should be non-empty for valid key
		if hederaKey.String() == "" {
			t.Error("Hedera key string should not be empty")
		}
	})

	t.Run("round trip conversion", func(t *testing.T) {
		hederaKey := pubKey.ToHederaPublicKey()
		restored, err := PublicKeyFromHedera(hederaKey)
		if err != nil {
			t.Fatalf("failed to restore from Hedera key: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})
}

// =============================================================================
// NeuronPublicKey.MatchesEVMAddress Tests
// =============================================================================

func TestNeuronPublicKey_MatchesEVMAddress(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()
	pubKey := key1.PublicKey()

	t.Run("matches own EVM address", func(t *testing.T) {
		if !pubKey.MatchesEVMAddress(pubKey.EVMAddress()) {
			t.Error("public key should match its own EVM address")
		}
	})

	t.Run("does not match other EVM address", func(t *testing.T) {
		if pubKey.MatchesEVMAddress(key2.PublicKey().EVMAddress()) {
			t.Error("public key should not match different EVM address")
		}
	})

	t.Run("zero key does not match", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		if zeroPub.MatchesEVMAddress(pubKey.EVMAddress()) {
			t.Error("zero public key should not match any EVM address")
		}
	})

	t.Run("does not match zero EVM address", func(t *testing.T) {
		if pubKey.MatchesEVMAddress(ZeroEVMAddress) {
			t.Error("public key should not match zero EVM address")
		}
	})
}

// =============================================================================
// NeuronPublicKey.MatchesPeerID Tests
// =============================================================================

func TestNeuronPublicKey_MatchesPeerID(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()
	pubKey := key1.PublicKey()

	t.Run("matches own peer ID", func(t *testing.T) {
		pid, _ := pubKey.PeerID()
		if !pubKey.MatchesPeerID(pid) {
			t.Error("public key should match its own peer ID")
		}
	})

	t.Run("does not match other peer ID", func(t *testing.T) {
		pid2, _ := key2.PublicKey().PeerID()
		if pubKey.MatchesPeerID(pid2) {
			t.Error("public key should not match different peer ID")
		}
	})

	t.Run("zero key does not match", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		pid, _ := pubKey.PeerID()
		if zeroPub.MatchesPeerID(pid) {
			t.Error("zero public key should not match any peer ID")
		}
	})

	t.Run("does not match zero peer ID", func(t *testing.T) {
		var zeroPID PeerID
		if pubKey.MatchesPeerID(zeroPID) {
			t.Error("public key should not match zero peer ID")
		}
	})
}

// =============================================================================
// NeuronPublicKey.Verify Tests
// =============================================================================

func TestNeuronPublicKey_Verify(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()
	message := []byte("hello world")

	t.Run("verifies valid signature", func(t *testing.T) {
		sig, _ := key.SignMessage(message)
		if !pubKey.Verify(message, sig) {
			t.Error("should verify valid signature")
		}
	})

	t.Run("rejects wrong message", func(t *testing.T) {
		sig, _ := key.SignMessage(message)
		if pubKey.Verify([]byte("wrong message"), sig) {
			t.Error("should reject signature for wrong message")
		}
	})

	t.Run("rejects wrong key signature", func(t *testing.T) {
		key2, _ := GeneratePrivateKey()
		sig, _ := key2.SignMessage(message)
		if pubKey.Verify(message, sig) {
			t.Error("should reject signature from different key")
		}
	})

	t.Run("rejects zero signature", func(t *testing.T) {
		var zeroSig Signature
		if pubKey.Verify(message, zeroSig) {
			t.Error("should reject zero signature")
		}
	})

	t.Run("zero key returns false", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		sig, _ := key.SignMessage(message)
		if zeroPub.Verify(message, sig) {
			t.Error("zero public key should return false")
		}
	})

	t.Run("empty message works", func(t *testing.T) {
		emptyMsg := []byte{}
		sig, _ := key.SignMessage(emptyMsg)
		if !pubKey.Verify(emptyMsg, sig) {
			t.Error("empty message signature should verify")
		}
	})
}

// =============================================================================
// NeuronPublicKey.VerifyDigest Tests
// =============================================================================

func TestNeuronPublicKey_VerifyDigest(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()
	digest := [32]byte{1, 2, 3, 4, 5}

	t.Run("verifies valid signature", func(t *testing.T) {
		sig, _ := key.SignDigest(digest)
		if !pubKey.VerifyDigest(digest, sig) {
			t.Error("should verify valid signature")
		}
	})

	t.Run("rejects wrong digest", func(t *testing.T) {
		sig, _ := key.SignDigest(digest)
		wrongDigest := [32]byte{5, 4, 3, 2, 1}
		if pubKey.VerifyDigest(wrongDigest, sig) {
			t.Error("should reject signature for wrong digest")
		}
	})

	t.Run("rejects zero signature", func(t *testing.T) {
		var zeroSig Signature
		if pubKey.VerifyDigest(digest, zeroSig) {
			t.Error("should reject zero signature")
		}
	})

	t.Run("zero key returns false", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		sig, _ := key.SignDigest(digest)
		if zeroPub.VerifyDigest(digest, sig) {
			t.Error("zero public key should return false")
		}
	})
}

// =============================================================================
// NeuronPublicKey.Equal Tests
// =============================================================================

func TestNeuronPublicKey_Equal(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()
	pub1 := key1.PublicKey()
	pub2 := key2.PublicKey()

	t.Run("key equals itself", func(t *testing.T) {
		if !pub1.Equal(pub1) {
			t.Error("key should equal itself")
		}
	})

	t.Run("different keys not equal", func(t *testing.T) {
		if pub1.Equal(pub2) {
			t.Error("different keys should not be equal")
		}
	})

	t.Run("parsed key equals original", func(t *testing.T) {
		hex := pub1.Hex()
		parsed, _ := ParsePublicKeyHex(hex)
		if !pub1.Equal(parsed) {
			t.Error("parsed key should equal original")
		}
	})

	t.Run("restored from compressed bytes equals original", func(t *testing.T) {
		bytes := pub1.CompressedBytes()
		restored, _ := PublicKeyFromBytes(bytes[:])
		if !pub1.Equal(restored) {
			t.Error("restored key should equal original")
		}
	})

	t.Run("restored from uncompressed bytes equals original", func(t *testing.T) {
		bytes := pub1.UncompressedBytes()
		restored, _ := PublicKeyFromBytes(bytes[:])
		if !pub1.Equal(restored) {
			t.Error("restored key should equal original")
		}
	})

	t.Run("zero keys are equal", func(t *testing.T) {
		var zero1, zero2 NeuronPublicKey
		if !zero1.Equal(zero2) {
			t.Error("zero keys should be equal")
		}
	})

	t.Run("zero key not equal to valid key", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		if zeroPub.Equal(pub1) || pub1.Equal(zeroPub) {
			t.Error("zero key should not equal valid key")
		}
	})
}

// =============================================================================
// Known Test Vector Tests
// =============================================================================

func TestNeuronPublicKey_KnownVector(t *testing.T) {
	// Known test vector from integration tests
	pubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	expectedPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	expectedEVM := "0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"

	pubKey, err := ParsePublicKeyHex(pubKeyHex)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	t.Run("derives correct peer ID", func(t *testing.T) {
		pid, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("PeerID failed: %v", err)
		}
		if pid.String() != expectedPeerID {
			t.Errorf("expected peer ID %s, got %s", expectedPeerID, pid.String())
		}
	})

	t.Run("derives correct EVM address", func(t *testing.T) {
		addr := pubKey.EVMAddress()
		// Compare case-insensitively
		if !strings.EqualFold(addr.Hex(), expectedEVM) {
			t.Errorf("expected EVM address %s, got %s", expectedEVM, addr.Hex())
		}
	})

	t.Run("matches expected peer ID", func(t *testing.T) {
		pid, _ := ParsePeerID(expectedPeerID)
		if !pubKey.MatchesPeerID(pid) {
			t.Error("public key should match expected peer ID")
		}
	})

	t.Run("matches expected EVM address", func(t *testing.T) {
		addr, _ := ParseEVMAddress(expectedEVM)
		if !pubKey.MatchesEVMAddress(addr) {
			t.Error("public key should match expected EVM address")
		}
	})
}

// =============================================================================
// NeuronPublicKey.EVMAddressSafe Tests
// =============================================================================

func TestNeuronPublicKey_EVMAddressSafe(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("valid key returns address", func(t *testing.T) {
		addr, err := pubKey.EVMAddressSafe()
		if err != nil {
			t.Fatalf("EVMAddressSafe should succeed for valid key: %v", err)
		}
		if addr.IsZero() {
			t.Error("address should not be zero")
		}
		// Should match the non-safe version
		if !addr.Equal(pubKey.EVMAddress()) {
			t.Error("EVMAddressSafe should match EVMAddress for valid key")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		_, err := zeroPub.EVMAddressSafe()
		if err == nil {
			t.Error("EVMAddressSafe should return error for zero key")
		}
		// Verify error kind
		keyErr, ok := err.(*KeyError)
		if !ok {
			t.Errorf("expected *KeyError, got %T", err)
		} else if keyErr.Kind != ErrKindZeroValue {
			t.Errorf("expected ErrKindZeroValue, got %v", keyErr.Kind)
		}
	})
}

// =============================================================================
// NeuronPublicKey.ToHederaPublicKeySafe Tests
// =============================================================================

func TestNeuronPublicKey_ToHederaPublicKeySafe(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("valid key converts", func(t *testing.T) {
		hederaKey, err := pubKey.ToHederaPublicKeySafe()
		if err != nil {
			t.Fatalf("ToHederaPublicKeySafe should succeed: %v", err)
		}
		if hederaKey.String() == "" {
			t.Error("Hedera key should not be empty")
		}
	})

	t.Run("round-trip with safe method", func(t *testing.T) {
		hederaKey, _ := pubKey.ToHederaPublicKeySafe()
		restored, err := PublicKeyFromHedera(hederaKey)
		if err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if !pubKey.Equal(restored) {
			t.Error("keys should match after round-trip")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		_, err := zeroPub.ToHederaPublicKeySafe()
		if err == nil {
			t.Error("ToHederaPublicKeySafe should return error for zero key")
		}
		// Verify error kind
		keyErr, ok := err.(*KeyError)
		if !ok {
			t.Errorf("expected *KeyError, got %T", err)
		} else if keyErr.Kind != ErrKindZeroValue {
			t.Errorf("expected ErrKindZeroValue, got %v", keyErr.Kind)
		}
	})
}
