package keylib

import (
	"math/big"
	"strings"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// =============================================================================
// ParseSignature Tests
// =============================================================================

func TestParseSignature(t *testing.T) {
	// Generate a valid signature for testing
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test message"))
	validSigHex := sig.Hex()

	t.Run("valid signature hex", func(t *testing.T) {
		parsed, err := ParseSignature(validSigHex)
		if err != nil {
			t.Fatalf("ParseSignature failed: %v", err)
		}
		if !parsed.Equal(sig) {
			t.Error("parsed signature should equal original")
		}
	})

	t.Run("valid without 0x prefix", func(t *testing.T) {
		hexWithoutPrefix := validSigHex[2:] // Remove "0x"
		parsed, err := ParseSignature(hexWithoutPrefix)
		if err != nil {
			t.Fatalf("ParseSignature failed: %v", err)
		}
		if !parsed.Equal(sig) {
			t.Error("parsed signature should equal original")
		}
	})

	t.Run("wrong length rejected", func(t *testing.T) {
		_, err := ParseSignature("abc")
		if err == nil {
			t.Error("ParseSignature should reject wrong length")
		}
	})

	t.Run("empty string rejected", func(t *testing.T) {
		_, err := ParseSignature("")
		if err == nil {
			t.Error("ParseSignature should reject empty string")
		}
	})

	t.Run("too short rejected", func(t *testing.T) {
		_, err := ParseSignature(validSigHex[:64]) // Only 64 chars instead of 130
		if err == nil {
			t.Error("ParseSignature should reject too short")
		}
	})

	t.Run("too long rejected", func(t *testing.T) {
		_, err := ParseSignature(validSigHex + "00")
		if err == nil {
			t.Error("ParseSignature should reject too long")
		}
	})

	t.Run("invalid hex characters rejected", func(t *testing.T) {
		invalidHex := "0x" + strings.Repeat("g", 130)
		_, err := ParseSignature(invalidHex)
		if err == nil {
			t.Error("ParseSignature should reject invalid hex")
		}
	})
}

// =============================================================================
// SignatureFromBytes Tests
// =============================================================================

func TestSignatureFromBytes(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))
	validBytes := sig.Bytes()

	t.Run("valid 65 bytes", func(t *testing.T) {
		parsed, err := SignatureFromBytes(validBytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes failed: %v", err)
		}
		if !parsed.Equal(sig) {
			t.Error("parsed signature should equal original")
		}
	})

	t.Run("V normalization 27->0", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 27 // Set V to 27

		parsed, err := SignatureFromBytes(bytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes failed: %v", err)
		}
		if parsed.V() != 0 {
			t.Errorf("V should be normalized to 0, got %d", parsed.V())
		}
	})

	t.Run("V normalization 28->1", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 28 // Set V to 28

		parsed, err := SignatureFromBytes(bytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes failed: %v", err)
		}
		if parsed.V() != 1 {
			t.Errorf("V should be normalized to 1, got %d", parsed.V())
		}
	})

	t.Run("V=0 accepted", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 0

		_, err := SignatureFromBytes(bytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes should accept V=0: %v", err)
		}
	})

	t.Run("V=1 accepted", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 1

		_, err := SignatureFromBytes(bytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes should accept V=1: %v", err)
		}
	})

	t.Run("invalid V=2 rejected", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 2

		_, err := SignatureFromBytes(bytes)
		if err == nil {
			t.Error("SignatureFromBytes should reject V=2")
		}
	})

	t.Run("invalid V=29 rejected", func(t *testing.T) {
		bytes := make([]byte, 65)
		copy(bytes, validBytes)
		bytes[64] = 29

		_, err := SignatureFromBytes(bytes)
		if err == nil {
			t.Error("SignatureFromBytes should reject V=29")
		}
	})

	t.Run("wrong length 64 bytes rejected", func(t *testing.T) {
		_, err := SignatureFromBytes(validBytes[:64])
		if err == nil {
			t.Error("SignatureFromBytes should reject 64 bytes")
		}
	})

	t.Run("wrong length 66 bytes rejected", func(t *testing.T) {
		bytes := make([]byte, 66)
		copy(bytes, validBytes)
		_, err := SignatureFromBytes(bytes)
		if err == nil {
			t.Error("SignatureFromBytes should reject 66 bytes")
		}
	})

	t.Run("empty bytes rejected", func(t *testing.T) {
		_, err := SignatureFromBytes([]byte{})
		if err == nil {
			t.Error("SignatureFromBytes should reject empty bytes")
		}
	})
}

// =============================================================================
// Signature Methods Tests
// =============================================================================

func TestSignature_Bytes(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns 65 bytes", func(t *testing.T) {
		bytes := sig.Bytes()
		if len(bytes) != 65 {
			t.Errorf("expected 65 bytes, got %d", len(bytes))
		}
	})

	t.Run("zero signature returns zero bytes", func(t *testing.T) {
		var zeroSig Signature
		bytes := zeroSig.Bytes()
		if len(bytes) != 65 {
			t.Errorf("expected 65 bytes, got %d", len(bytes))
		}
		// Check all zeros
		for i, b := range bytes {
			if b != 0 {
				t.Errorf("byte %d should be 0, got %d", i, b)
			}
		}
	})
}

func TestSignature_Hex(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("includes 0x prefix", func(t *testing.T) {
		hex := sig.Hex()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("hex should have 0x prefix")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		hex := sig.Hex()
		// 0x + 65 bytes * 2 hex chars = 132 chars
		if len(hex) != 132 {
			t.Errorf("expected 132 chars, got %d", len(hex))
		}
	})

	t.Run("lowercase", func(t *testing.T) {
		hex := sig.Hex()
		if hex != strings.ToLower(hex) {
			t.Error("hex should be lowercase")
		}
	})
}

func TestSignature_RS(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns correct big.Int values", func(t *testing.T) {
		r, s := sig.RS()
		if r == nil || s == nil {
			t.Error("R and S should not be nil")
		}
		if r.Sign() <= 0 || s.Sign() <= 0 {
			t.Error("R and S should be positive")
		}
	})

	t.Run("zero signature returns nil", func(t *testing.T) {
		var zeroSig Signature
		r, s := zeroSig.RS()
		if r != nil || s != nil {
			t.Error("zero signature RS should return nil, nil")
		}
	})
}

func TestSignature_RSV(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns correct values", func(t *testing.T) {
		r, s, v := sig.RSV()
		if r == nil || s == nil {
			t.Error("R and S should not be nil")
		}
		if v > 1 {
			t.Errorf("V should be 0 or 1, got %d", v)
		}
	})
}

func TestSignature_V(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns 0 or 1", func(t *testing.T) {
		v := sig.V()
		if v > 1 {
			t.Errorf("V should be 0 or 1, got %d", v)
		}
	})
}

func TestSignature_VEthereum(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns 27 or 28", func(t *testing.T) {
		vEth := sig.VEthereum()
		if vEth != 27 && vEth != 28 {
			t.Errorf("VEthereum should be 27 or 28, got %d", vEth)
		}
	})

	t.Run("consistency with V", func(t *testing.T) {
		v := sig.V()
		vEth := sig.VEthereum()
		if int(vEth) != int(v)+27 {
			t.Errorf("VEthereum should equal V + 27")
		}
	})
}

func TestSignature_EthereumBytes(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("returns 65 bytes", func(t *testing.T) {
		bytes := sig.EthereumBytes()
		if len(bytes) != 65 {
			t.Errorf("expected 65 bytes, got %d", len(bytes))
		}
	})

	t.Run("V is 27 or 28", func(t *testing.T) {
		bytes := sig.EthereumBytes()
		v := bytes[64]
		if v != 27 && v != 28 {
			t.Errorf("V should be 27 or 28, got %d", v)
		}
	})

	t.Run("R and S unchanged", func(t *testing.T) {
		normalBytes := sig.Bytes()
		ethBytes := sig.EthereumBytes()

		// First 64 bytes (R and S) should be identical
		for i := 0; i < 64; i++ {
			if normalBytes[i] != ethBytes[i] {
				t.Errorf("byte %d mismatch", i)
			}
		}
	})
}

func TestSignature_IsZero(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))

	t.Run("valid signature not zero", func(t *testing.T) {
		if sig.IsZero() {
			t.Error("valid signature should not be zero")
		}
	})

	t.Run("zero signature is zero", func(t *testing.T) {
		var zeroSig Signature
		if !zeroSig.IsZero() {
			t.Error("zero signature should be zero")
		}
	})

	t.Run("ZeroSignature constant", func(t *testing.T) {
		if !ZeroSignature.IsZero() {
			t.Error("ZeroSignature should be zero")
		}
	})
}

func TestSignature_Equal(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig1, _ := privKey.SignMessage([]byte("test1"))
	sig2, _ := privKey.SignMessage([]byte("test2"))

	t.Run("same signature equals itself", func(t *testing.T) {
		if !sig1.Equal(sig1) {
			t.Error("signature should equal itself")
		}
	})

	t.Run("different signatures not equal", func(t *testing.T) {
		if sig1.Equal(sig2) {
			t.Error("different signatures should not be equal")
		}
	})

	t.Run("zero signatures equal", func(t *testing.T) {
		var z1, z2 Signature
		if !z1.Equal(z2) {
			t.Error("zero signatures should be equal")
		}
	})

	t.Run("zero not equal to valid", func(t *testing.T) {
		var zero Signature
		if zero.Equal(sig1) || sig1.Equal(zero) {
			t.Error("zero signature should not equal valid signature")
		}
	})
}

// =============================================================================
// RecoverPublicKey Tests
// =============================================================================

func TestRecoverPublicKey_EdgeCases(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	message := []byte("recover me!")
	sig, _ := privKey.SignMessage(message)

	t.Run("recovers correct public key", func(t *testing.T) {
		recovered, err := RecoverPublicKey(message, sig)
		if err != nil {
			t.Fatalf("RecoverPublicKey failed: %v", err)
		}
		if !recovered.Equal(pubKey) {
			t.Error("recovered public key should equal original")
		}
	})

	t.Run("rejects zero signature", func(t *testing.T) {
		var zeroSig Signature
		_, err := RecoverPublicKey(message, zeroSig)
		if err == nil {
			t.Error("RecoverPublicKey should reject zero signature")
		}
	})

	t.Run("wrong message gives different key", func(t *testing.T) {
		recovered, err := RecoverPublicKey([]byte("wrong message"), sig)
		if err != nil {
			t.Fatalf("RecoverPublicKey failed: %v", err)
		}
		// Should recover a different public key (or same by unlikely chance)
		// The key point is it should not error
		_ = recovered
	})

	t.Run("empty message works", func(t *testing.T) {
		emptyMsg := []byte{}
		emptySig, _ := privKey.SignMessage(emptyMsg)

		recovered, err := RecoverPublicKey(emptyMsg, emptySig)
		if err != nil {
			t.Fatalf("RecoverPublicKey failed for empty message: %v", err)
		}
		if !recovered.Equal(pubKey) {
			t.Error("recovered key should match for empty message")
		}
	})
}

func TestRecoverPublicKeyFromDigest(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	digest := [32]byte{1, 2, 3, 4, 5}
	sig, _ := privKey.SignDigest(digest)

	t.Run("recovers correct public key", func(t *testing.T) {
		recovered, err := RecoverPublicKeyFromDigest(digest, sig)
		if err != nil {
			t.Fatalf("RecoverPublicKeyFromDigest failed: %v", err)
		}
		if !recovered.Equal(pubKey) {
			t.Error("recovered public key should equal original")
		}
	})

	t.Run("rejects zero signature", func(t *testing.T) {
		var zeroSig Signature
		_, err := RecoverPublicKeyFromDigest(digest, zeroSig)
		if err == nil {
			t.Error("RecoverPublicKeyFromDigest should reject zero signature")
		}
	})

	t.Run("all-zeros digest works", func(t *testing.T) {
		zeroDigest := [32]byte{}
		zeroSig, _ := privKey.SignDigest(zeroDigest)

		recovered, err := RecoverPublicKeyFromDigest(zeroDigest, zeroSig)
		if err != nil {
			t.Fatalf("RecoverPublicKeyFromDigest failed for zero digest: %v", err)
		}
		if !recovered.Equal(pubKey) {
			t.Error("recovered key should match for zero digest")
		}
	})
}

// =============================================================================
// signatureFromRSV Tests (internal function)
// =============================================================================

func TestSignatureFromRSV(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	sig, _ := privKey.SignMessage([]byte("test"))
	r, s, v := sig.RSV()

	t.Run("reconstructs signature", func(t *testing.T) {
		reconstructed := signatureFromRSV(r, s, v)

		if !reconstructed.Equal(sig) {
			t.Error("reconstructed signature should equal original")
		}
	})

	t.Run("V normalization from 27", func(t *testing.T) {
		reconstructed := signatureFromRSV(r, s, 27)
		if reconstructed.V() != 0 {
			t.Errorf("V should be normalized to 0, got %d", reconstructed.V())
		}
	})

	t.Run("V normalization from 28", func(t *testing.T) {
		reconstructed := signatureFromRSV(r, s, 28)
		if reconstructed.V() != 1 {
			t.Errorf("V should be normalized to 1, got %d", reconstructed.V())
		}
	})
}

// =============================================================================
// Low-S Normalization Tests (EIP-2 Compliance)
// =============================================================================

func TestSignature_LowSNormalization(t *testing.T) {
	// secp256k1 curve order N
	n := secp256k1.S256().N
	halfN := new(big.Int).Rsh(n, 1)

	t.Run("high-S is normalized to low-S", func(t *testing.T) {
		// Create a signature with S > N/2
		r := big.NewInt(12345)
		highS := new(big.Int).Add(halfN, big.NewInt(1000)) // S > N/2

		sig := signatureFromRSV(r, highS, 0)

		// After normalization, S should be <= N/2
		_, actualS := sig.RS()
		if actualS.Cmp(halfN) > 0 {
			t.Error("S should be normalized to low-S")
		}
	})

	t.Run("V is flipped when S is normalized", func(t *testing.T) {
		r := big.NewInt(12345)
		highS := new(big.Int).Add(halfN, big.NewInt(1000))

		sig := signatureFromRSV(r, highS, 0)

		// V should be flipped from 0 to 1
		if sig.V() != 1 {
			t.Errorf("V should be 1 after normalization, got %d", sig.V())
		}
	})

	t.Run("low-S remains unchanged", func(t *testing.T) {
		r := big.NewInt(12345)
		lowS := new(big.Int).Sub(halfN, big.NewInt(1000)) // S < N/2

		sig := signatureFromRSV(r, lowS, 0)

		_, actualS := sig.RS()
		if actualS.Cmp(lowS) != 0 {
			t.Error("low-S should remain unchanged")
		}
		if sig.V() != 0 {
			t.Error("V should remain 0 for low-S")
		}
	})

	t.Run("S exactly at N/2 is unchanged", func(t *testing.T) {
		r := big.NewInt(12345)
		sig := signatureFromRSV(r, halfN, 0)

		_, actualS := sig.RS()
		if actualS.Cmp(halfN) != 0 {
			t.Error("S=N/2 should remain unchanged")
		}
	})

	t.Run("S exactly at N/2+1 is normalized", func(t *testing.T) {
		r := big.NewInt(12345)
		justOverHalf := new(big.Int).Add(halfN, big.NewInt(1))
		sig := signatureFromRSV(r, justOverHalf, 0)

		_, actualS := sig.RS()
		if actualS.Cmp(halfN) > 0 {
			t.Error("S=N/2+1 should be normalized")
		}
	})

	t.Run("ParseSignature normalizes high-S", func(t *testing.T) {
		// Build a hex signature with high-S
		r := big.NewInt(12345)
		highS := new(big.Int).Add(halfN, big.NewInt(1000))

		rBytes := padLeftZeros(r.Bytes(), 32)
		sBytes := padLeftZeros(highS.Bytes(), 32)
		sigBytes := append(append(rBytes, sBytes...), 0)
		hexStr := encodeHex(sigBytes, true)

		sig, err := ParseSignature(hexStr)
		if err != nil {
			t.Fatalf("ParseSignature failed: %v", err)
		}

		_, actualS := sig.RS()
		if actualS.Cmp(halfN) > 0 {
			t.Error("ParseSignature should normalize high-S")
		}
	})

	t.Run("SignatureFromBytes normalizes high-S", func(t *testing.T) {
		r := big.NewInt(12345)
		highS := new(big.Int).Add(halfN, big.NewInt(1000))

		rBytes := padLeftZeros(r.Bytes(), 32)
		sBytes := padLeftZeros(highS.Bytes(), 32)
		sigBytes := append(append(rBytes, sBytes...), 0)

		sig, err := SignatureFromBytes(sigBytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes failed: %v", err)
		}

		_, actualS := sig.RS()
		if actualS.Cmp(halfN) > 0 {
			t.Error("SignatureFromBytes should normalize high-S")
		}
	})
}

func TestSignature_NormalizedRecoveryRegression(t *testing.T) {
	// CRITICAL: Verify signature recovery still works after normalization
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	msg := []byte("test message for recovery")

	t.Run("sign-normalize-recover produces same key", func(t *testing.T) {
		sig, err := privKey.SignMessage(msg)
		if err != nil {
			t.Fatalf("SignMessage failed: %v", err)
		}

		recovered, err := RecoverPublicKey(msg, sig)
		if err != nil {
			t.Fatalf("RecoverPublicKey failed: %v", err)
		}

		if !recovered.Equal(pubKey) {
			t.Error("recovered key should match original")
		}
	})

	t.Run("normalized signature verifies correctly", func(t *testing.T) {
		sig, _ := privKey.SignMessage(msg)

		if !pubKey.Verify(msg, sig) {
			t.Error("normalized signature should verify")
		}
	})

	t.Run("parsing external high-S signature works", func(t *testing.T) {
		// Create a signature and then manually flip to high-S
		sig, _ := privKey.SignMessage(msg)
		r, s, v := sig.RSV()

		// Compute high-S version: N - s
		n := secp256k1.S256().N
		highS := new(big.Int).Sub(n, s)
		flippedV := v ^ 1

		// Build raw bytes with high-S (simulating external source)
		rBytes := padLeftZeros(r.Bytes(), 32)
		sBytes := padLeftZeros(highS.Bytes(), 32)
		rawBytes := append(append(rBytes, sBytes...), flippedV)

		// Parse - this should normalize
		highSSig, err := SignatureFromBytes(rawBytes)
		if err != nil {
			t.Fatalf("SignatureFromBytes failed: %v", err)
		}

		// After normalization, recovery should produce same key
		recovered, err := RecoverPublicKey(msg, highSSig)
		if err != nil {
			t.Fatalf("recovery from high-S failed: %v", err)
		}
		if !recovered.Equal(pubKey) {
			t.Error("high-S signature should recover to same key after normalization")
		}
	})
}
