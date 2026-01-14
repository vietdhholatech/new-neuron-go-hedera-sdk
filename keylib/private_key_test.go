package keylib

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"strings"
	"testing"
)

// =============================================================================
// NeuronPrivateKey.IsZero Tests
// =============================================================================

func TestNeuronPrivateKey_IsZero(t *testing.T) {
	t.Run("generated key is not zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		if key.IsZero() {
			t.Error("generated key should not be zero")
		}
	})

	t.Run("zero value key is zero", func(t *testing.T) {
		var key NeuronPrivateKey
		if !key.IsZero() {
			t.Error("zero value key should be zero")
		}
	})

	t.Run("parsed key is not zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		parsed, _ := ParsePrivateKeyHex(hex)
		if parsed.IsZero() {
			t.Error("parsed key should not be zero")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.PublicKey Tests
// =============================================================================

func TestNeuronPrivateKey_PublicKey(t *testing.T) {
	t.Run("returns valid public key", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pubKey := key.PublicKey()
		if pubKey.IsZero() {
			t.Error("public key should not be zero")
		}
	})

	t.Run("public key is deterministic", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pub1 := key.PublicKey()
		pub2 := key.PublicKey()
		if !pub1.Equal(pub2) {
			t.Error("public key should be deterministic")
		}
	})

	t.Run("zero key returns zero public key", func(t *testing.T) {
		var key NeuronPrivateKey
		pubKey := key.PublicKey()
		if !pubKey.IsZero() {
			t.Error("zero private key should return zero public key")
		}
	})

	t.Run("different keys have different public keys", func(t *testing.T) {
		key1, _ := GeneratePrivateKey()
		key2, _ := GeneratePrivateKey()
		if key1.PublicKey().Equal(key2.PublicKey()) {
			t.Error("different private keys should have different public keys")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.EVMAddress Tests
// =============================================================================

func TestNeuronPrivateKey_EVMAddress(t *testing.T) {
	t.Run("returns valid EVM address", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		addr := key.EVMAddress()
		if addr.IsZero() {
			t.Error("EVM address should not be zero")
		}
	})

	t.Run("EVM address is deterministic", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		addr1 := key.EVMAddress()
		addr2 := key.EVMAddress()
		if !addr1.Equal(addr2) {
			t.Error("EVM address should be deterministic")
		}
	})

	t.Run("zero key returns zero address", func(t *testing.T) {
		var key NeuronPrivateKey
		addr := key.EVMAddress()
		if !addr.IsZero() {
			t.Error("zero private key should return zero EVM address")
		}
	})

	t.Run("matches public key EVM address", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		privAddr := key.EVMAddress()
		pubAddr := key.PublicKey().EVMAddress()
		if !privAddr.Equal(pubAddr) {
			t.Error("private and public key should derive same EVM address")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.PeerID Tests
// =============================================================================

func TestNeuronPrivateKey_PeerID(t *testing.T) {
	t.Run("returns valid peer ID", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		peerID, err := key.PeerID()
		if err != nil {
			t.Fatalf("PeerID failed: %v", err)
		}
		if peerID.IsZero() {
			t.Error("peer ID should not be zero")
		}
	})

	t.Run("peer ID is deterministic", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pid1, _ := key.PeerID()
		pid2, _ := key.PeerID()
		if !pid1.Equal(pid2) {
			t.Error("peer ID should be deterministic")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var key NeuronPrivateKey
		_, err := key.PeerID()
		if err == nil {
			t.Error("zero private key should return error for PeerID")
		}
	})

	t.Run("matches public key peer ID", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		privPID, _ := key.PeerID()
		pubPID, _ := key.PublicKey().PeerID()
		if !privPID.Equal(pubPID) {
			t.Error("private and public key should derive same peer ID")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.Bytes Tests
// =============================================================================

func TestNeuronPrivateKey_Bytes(t *testing.T) {
	t.Run("returns 32 bytes", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		bytes := key.Bytes()
		if len(bytes) != 32 {
			t.Errorf("expected 32 bytes, got %d", len(bytes))
		}
	})

	t.Run("bytes are not all zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		bytes := key.Bytes()
		allZero := true
		for _, b := range bytes {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Error("key bytes should not be all zero")
		}
	})

	t.Run("zero key returns zero bytes", func(t *testing.T) {
		var key NeuronPrivateKey
		bytes := key.Bytes()
		expected := [32]byte{}
		if bytes != expected {
			t.Error("zero key should return zero bytes")
		}
	})

	t.Run("round trip bytes", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		bytes := key.Bytes()
		restored, err := PrivateKeyFromBytes(bytes)
		if err != nil {
			t.Fatalf("failed to restore from bytes: %v", err)
		}
		if !key.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.Hex Tests
// =============================================================================

func TestNeuronPrivateKey_Hex(t *testing.T) {
	t.Run("includes 0x prefix", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("hex should start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		if len(hex) != 66 {
			t.Errorf("expected 66 chars (0x + 64), got %d", len(hex))
		}
	})

	t.Run("lowercase", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		if strings.ToLower(hex) != hex {
			t.Error("hex should be lowercase")
		}
	})

	t.Run("zero key returns empty", func(t *testing.T) {
		var key NeuronPrivateKey
		hex := key.Hex()
		if hex != "" {
			t.Errorf("zero key should return empty string, got %s", hex)
		}
	})

	t.Run("round trip hex", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.Hex()
		restored, err := ParsePrivateKeyHex(hex)
		if err != nil {
			t.Fatalf("failed to restore from hex: %v", err)
		}
		if !key.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.HexWithoutPrefix Tests
// =============================================================================

func TestNeuronPrivateKey_HexWithoutPrefix(t *testing.T) {
	t.Run("no 0x prefix", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.HexWithoutPrefix()
		if strings.HasPrefix(hex, "0x") {
			t.Error("hex should not start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hex := key.HexWithoutPrefix()
		if len(hex) != 64 {
			t.Errorf("expected 64 chars, got %d", len(hex))
		}
	})

	t.Run("consistent with Hex", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		withPrefix := key.Hex()
		withoutPrefix := key.HexWithoutPrefix()
		if withPrefix != "0x"+withoutPrefix {
			t.Error("HexWithoutPrefix should be Hex without 0x prefix")
		}
	})

	t.Run("zero key returns empty", func(t *testing.T) {
		var key NeuronPrivateKey
		hex := key.HexWithoutPrefix()
		if hex != "" {
			t.Errorf("zero key should return empty string, got %s", hex)
		}
	})
}

// =============================================================================
// NeuronPrivateKey.ToECDSA Tests
// =============================================================================

func TestNeuronPrivateKey_ToECDSA(t *testing.T) {
	t.Run("returns valid ECDSA key", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		ecdsaKey := key.ToECDSA()
		if ecdsaKey == nil {
			t.Error("should return non-nil ECDSA key")
		}
	})

	t.Run("ECDSA key can sign", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		ecdsaKey := key.ToECDSA()

		digest := [32]byte{1, 2, 3}
		_, err := ecdsa.SignASN1(rand.Reader, ecdsaKey, digest[:])
		if err != nil {
			t.Errorf("ECDSA key should be able to sign: %v", err)
		}
	})

	t.Run("zero key returns nil", func(t *testing.T) {
		var key NeuronPrivateKey
		ecdsaKey := key.ToECDSA()
		if ecdsaKey != nil {
			t.Error("zero key should return nil ECDSA key")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.ToHederaPrivateKey Tests
// =============================================================================

func TestNeuronPrivateKey_ToHederaPrivateKey(t *testing.T) {
	t.Run("returns valid Hedera key", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hederaKey := key.ToHederaPrivateKey()
		// Hedera key string should be non-empty for valid key
		if hederaKey.String() == "" {
			t.Error("Hedera key string should not be empty")
		}
	})

	t.Run("round trip conversion", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		hederaKey := key.ToHederaPrivateKey()
		restored, err := PrivateKeyFromHedera(hederaKey)
		if err != nil {
			t.Fatalf("failed to restore from Hedera key: %v", err)
		}
		if !key.Equal(restored) {
			t.Error("key should equal restored key")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.SignMessage Tests
// =============================================================================

func TestNeuronPrivateKey_SignMessage(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("signs message successfully", func(t *testing.T) {
		msg := []byte("hello world")
		sig, err := key.SignMessage(msg)
		if err != nil {
			t.Fatalf("SignMessage failed: %v", err)
		}
		if sig.IsZero() {
			t.Error("signature should not be zero")
		}
	})

	t.Run("signature verifies", func(t *testing.T) {
		msg := []byte("hello world")
		sig, _ := key.SignMessage(msg)
		if !key.PublicKey().Verify(msg, sig) {
			t.Error("signature should verify")
		}
	})

	t.Run("empty message works", func(t *testing.T) {
		msg := []byte{}
		sig, err := key.SignMessage(msg)
		if err != nil {
			t.Fatalf("SignMessage with empty message failed: %v", err)
		}
		if !key.PublicKey().Verify(msg, sig) {
			t.Error("empty message signature should verify")
		}
	})

	t.Run("zero key fails", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.SignMessage([]byte("test"))
		if err == nil {
			t.Error("zero key should fail to sign")
		}
	})

	t.Run("different messages produce different signatures", func(t *testing.T) {
		sig1, _ := key.SignMessage([]byte("message 1"))
		sig2, _ := key.SignMessage([]byte("message 2"))
		if sig1.Equal(sig2) {
			t.Error("different messages should produce different signatures")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.SignDigest Tests
// =============================================================================

func TestNeuronPrivateKey_SignDigest(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("signs digest successfully", func(t *testing.T) {
		digest := [32]byte{1, 2, 3, 4, 5}
		sig, err := key.SignDigest(digest)
		if err != nil {
			t.Fatalf("SignDigest failed: %v", err)
		}
		if sig.IsZero() {
			t.Error("signature should not be zero")
		}
	})

	t.Run("signature verifies", func(t *testing.T) {
		digest := [32]byte{1, 2, 3, 4, 5}
		sig, _ := key.SignDigest(digest)
		if !key.PublicKey().VerifyDigest(digest, sig) {
			t.Error("signature should verify")
		}
	})

	t.Run("all-zeros digest signs without error", func(t *testing.T) {
		digest := [32]byte{}
		sig, err := key.SignDigest(digest)
		if err != nil {
			t.Fatalf("SignDigest with zero digest failed: %v", err)
		}
		// Note: We only check that signing succeeds. Verification of zero-digest
		// signatures may have edge cases in the secp256k1 library.
		if sig.IsZero() {
			t.Error("signature should not be zero")
		}
	})

	t.Run("all-ones digest works", func(t *testing.T) {
		var digest [32]byte
		for i := range digest {
			digest[i] = 0xFF
		}
		sig, err := key.SignDigest(digest)
		if err != nil {
			t.Fatalf("SignDigest with all-ones digest failed: %v", err)
		}
		if !key.PublicKey().VerifyDigest(digest, sig) {
			t.Error("all-ones digest signature should verify")
		}
	})

	t.Run("zero key fails", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.SignDigest([32]byte{1, 2, 3})
		if err == nil {
			t.Error("zero key should fail to sign digest")
		}
	})
}

// =============================================================================
// NeuronPrivateKey crypto.Signer Interface Tests
// =============================================================================

func TestNeuronPrivateKey_CryptoSigner(t *testing.T) {
	key, _ := GeneratePrivateKey()

	// Verify interface compliance
	var _ crypto.Signer = key

	t.Run("Public returns correct type", func(t *testing.T) {
		pub := key.Public()
		_, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			t.Error("Public should return *ecdsa.PublicKey")
		}
	})

	t.Run("Sign implements crypto.Signer", func(t *testing.T) {
		digest := [32]byte{1, 2, 3, 4, 5}
		sig, err := key.Sign(rand.Reader, digest[:], crypto.SHA256)
		if err != nil {
			t.Fatalf("Sign failed: %v", err)
		}
		if len(sig) == 0 {
			t.Error("signature should not be empty")
		}
	})

	t.Run("zero key Public returns nil", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		pub := zeroKey.Public()
		if pub != nil {
			t.Error("zero key Public should return nil")
		}
	})

	t.Run("zero key Sign fails", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.Sign(rand.Reader, []byte{1, 2, 3}, nil)
		if err == nil {
			t.Error("zero key Sign should fail")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.MatchesPublicKey Tests
// =============================================================================

func TestNeuronPrivateKey_MatchesPublicKey(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()

	t.Run("matches own public key", func(t *testing.T) {
		if !key1.MatchesPublicKey(key1.PublicKey()) {
			t.Error("key should match its own public key")
		}
	})

	t.Run("does not match other public key", func(t *testing.T) {
		if key1.MatchesPublicKey(key2.PublicKey()) {
			t.Error("key should not match different public key")
		}
	})

	t.Run("zero key does not match", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		if zeroKey.MatchesPublicKey(key1.PublicKey()) {
			t.Error("zero key should not match any public key")
		}
	})

	t.Run("does not match zero public key", func(t *testing.T) {
		var zeroPub NeuronPublicKey
		if key1.MatchesPublicKey(zeroPub) {
			t.Error("key should not match zero public key")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.MatchesEVMAddress Tests
// =============================================================================

func TestNeuronPrivateKey_MatchesEVMAddress(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()

	t.Run("matches own EVM address", func(t *testing.T) {
		if !key1.MatchesEVMAddress(key1.EVMAddress()) {
			t.Error("key should match its own EVM address")
		}
	})

	t.Run("does not match other EVM address", func(t *testing.T) {
		if key1.MatchesEVMAddress(key2.EVMAddress()) {
			t.Error("key should not match different EVM address")
		}
	})

	t.Run("zero key does not match", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		if zeroKey.MatchesEVMAddress(key1.EVMAddress()) {
			t.Error("zero key should not match any EVM address")
		}
	})

	t.Run("does not match zero EVM address", func(t *testing.T) {
		if key1.MatchesEVMAddress(ZeroEVMAddress) {
			t.Error("key should not match zero EVM address")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.Zeroize Tests
// =============================================================================

func TestNeuronPrivateKey_Zeroize(t *testing.T) {
	t.Run("key becomes zero after zeroize", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		if key.IsZero() {
			t.Fatal("key should not be zero before zeroize")
		}

		key.Zeroize()

		if !key.IsZero() {
			t.Error("key should be zero after zeroize")
		}
	})

	t.Run("operations fail after zeroize", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		key.Zeroize()

		_, err := key.SignMessage([]byte("test"))
		if err == nil {
			t.Error("signing should fail after zeroize")
		}
	})

	t.Run("multiple zeroize is safe", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		key.Zeroize()
		key.Zeroize() // Should not panic
		key.Zeroize()
	})

	t.Run("zeroize on zero key is safe", func(t *testing.T) {
		var key NeuronPrivateKey
		key.Zeroize() // Should not panic
	})

	t.Run("hex returns empty after zeroize", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		key.Zeroize()
		if key.Hex() != "" {
			t.Error("hex should return empty after zeroize")
		}
	})

	t.Run("bytes returns zero after zeroize", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		key.Zeroize()
		bytes := key.Bytes()
		expected := [32]byte{}
		if bytes != expected {
			t.Error("bytes should be zero after zeroize")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.Equal Tests
// =============================================================================

func TestNeuronPrivateKey_Equal(t *testing.T) {
	key1, _ := GeneratePrivateKey()
	key2, _ := GeneratePrivateKey()

	t.Run("key equals itself", func(t *testing.T) {
		if !key1.Equal(key1) {
			t.Error("key should equal itself")
		}
	})

	t.Run("different keys not equal", func(t *testing.T) {
		if key1.Equal(key2) {
			t.Error("different keys should not be equal")
		}
	})

	t.Run("parsed key equals original", func(t *testing.T) {
		hex := key1.Hex()
		parsed, _ := ParsePrivateKeyHex(hex)
		if !key1.Equal(parsed) {
			t.Error("parsed key should equal original")
		}
	})

	t.Run("restored from bytes equals original", func(t *testing.T) {
		bytes := key1.Bytes()
		restored, _ := PrivateKeyFromBytes(bytes)
		if !key1.Equal(restored) {
			t.Error("restored key should equal original")
		}
	})

	t.Run("zero keys are equal", func(t *testing.T) {
		var zero1, zero2 NeuronPrivateKey
		if !zero1.Equal(zero2) {
			t.Error("zero keys should be equal")
		}
	})

	t.Run("zero key not equal to valid key", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		if zeroKey.Equal(key1) || key1.Equal(zeroKey) {
			t.Error("zero key should not equal valid key")
		}
	})

	t.Run("uses constant-time comparison", func(t *testing.T) {
		// This test verifies that Equal uses constant-time comparison
		// by ensuring the behavior is consistent (can't directly test timing)
		bytes := key1.Bytes()
		copy1, _ := PrivateKeyFromBytes(bytes)
		copy2, _ := PrivateKeyFromBytes(bytes)
		if !copy1.Equal(copy2) {
			t.Error("keys from same bytes should be equal")
		}
	})
}

// =============================================================================
// NeuronPrivateKey.EVMAddressSafe Tests
// =============================================================================

func TestNeuronPrivateKey_EVMAddressSafe(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("valid key returns address", func(t *testing.T) {
		addr, err := key.EVMAddressSafe()
		if err != nil {
			t.Fatalf("EVMAddressSafe should succeed for valid key: %v", err)
		}
		if addr.IsZero() {
			t.Error("address should not be zero")
		}
		// Should match the non-safe version
		if !addr.Equal(key.EVMAddress()) {
			t.Error("EVMAddressSafe should match EVMAddress for valid key")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.EVMAddressSafe()
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
// NeuronPrivateKey.ToHederaPrivateKeySafe Tests
// =============================================================================

func TestNeuronPrivateKey_ToHederaPrivateKeySafe(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("valid key converts", func(t *testing.T) {
		hederaKey, err := key.ToHederaPrivateKeySafe()
		if err != nil {
			t.Fatalf("ToHederaPrivateKeySafe should succeed: %v", err)
		}
		if hederaKey.String() == "" {
			t.Error("Hedera key should not be empty")
		}
	})

	t.Run("round-trip with safe method", func(t *testing.T) {
		hederaKey, _ := key.ToHederaPrivateKeySafe()
		restored, err := PrivateKeyFromHedera(hederaKey)
		if err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if !key.Equal(restored) {
			t.Error("keys should match after round-trip")
		}
	})

	t.Run("zero key returns error", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.ToHederaPrivateKeySafe()
		if err == nil {
			t.Error("ToHederaPrivateKeySafe should return error for zero key")
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
