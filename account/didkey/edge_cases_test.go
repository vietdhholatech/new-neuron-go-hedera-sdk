package didkey

import (
	"strings"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// DID:Key Edge Cases - Phase 7
// These tests verify proper handling of edge cases in DID:Key encoding/decoding.
// =============================================================================

// =============================================================================
// Parse Edge Cases
// =============================================================================

func TestParse_MalformedInputs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "only whitespace",
			input:     "   ",
			wantError: true,
		},
		{
			name:      "missing did prefix",
			input:     "key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantError: true,
		},
		{
			name:      "wrong method",
			input:     "did:web:example.com",
			wantError: true,
		},
		{
			name:      "missing identifier",
			input:     "did:key:",
			wantError: true,
		},
		{
			name:      "only did:key",
			input:     "did:key",
			wantError: true,
		},
		{
			name:      "invalid multibase prefix",
			input:     "did:key:a6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantError: true,
		},
		{
			name:      "truncated identifier",
			input:     "did:key:z",
			wantError: true,
		},
		{
			name:      "invalid base58 characters",
			input:     "did:key:z0OIl", // O, I, l are not valid base58btc
			wantError: true,
		},
		{
			name:      "wrong multicodec (not secp256k1)",
			input:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK", // This is ed25519
			wantError: true,
		},
		{
			name:      "extra colons",
			input:     "did:key:z:extra:colons",
			wantError: true,
		},
		{
			name:      "null bytes in input",
			input:     "did:key:z\x00invalid",
			wantError: true,
		},
		{
			name:      "unicode in identifier",
			input:     "did:key:z日本語",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Parse(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

func TestParse_WhitespaceHandling(t *testing.T) {
	// Generate a valid DID for testing
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	validDID, _ := FromPublicKey(pubKey)
	validStr := validDID.String()

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "leading whitespace",
			input:     " " + validStr,
			wantError: true,
		},
		{
			name:      "trailing whitespace",
			input:     validStr + " ",
			wantError: true,
		},
		{
			name:      "leading newline",
			input:     "\n" + validStr,
			wantError: true,
		},
		{
			name:      "trailing newline",
			input:     validStr + "\n",
			wantError: true,
		},
		{
			name:      "tab in middle",
			input:     "did:key:\t" + validStr[8:],
			wantError: true,
		},
		{
			name:      "valid without whitespace",
			input:     validStr,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Parse(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

// =============================================================================
// MustParse Edge Cases
// =============================================================================

func TestMustParse_PanicBehavior(t *testing.T) {
	t.Run("panics on invalid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParse should panic on invalid input")
			}
		}()
		MustParse("invalid-did-key")
	})

	t.Run("does not panic on valid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParse should not panic on valid input: %v", r)
			}
		}()

		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		validDID, _ := FromPublicKey(pubKey)
		MustParse(validDID.String())
	})
}

// =============================================================================
// FromPublicKey Edge Cases
// =============================================================================

func TestFromPublicKey_EdgeCases(t *testing.T) {
	t.Run("zero value public key", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		_, err := FromPublicKey(zeroPubKey)
		if err == nil {
			t.Error("FromPublicKey should fail for zero value public key")
		}
	})

	t.Run("valid public key produces valid DID", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		did, err := FromPublicKey(pubKey)
		if err != nil {
			t.Fatalf("FromPublicKey failed: %v", err)
		}

		// Verify it starts with did:key:z
		if !strings.HasPrefix(did.String(), "did:key:z") {
			t.Errorf("DID should start with did:key:z, got: %s", did.String())
		}

		// Verify identifier is base58btc encoded
		if !strings.HasPrefix(did.Identifier(), "z") {
			t.Errorf("Identifier should start with z (base58btc), got: %s", did.Identifier())
		}
	})

	t.Run("same key produces same DID", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		did1, _ := FromPublicKey(pubKey)
		did2, _ := FromPublicKey(pubKey)

		if did1.String() != did2.String() {
			t.Error("same public key should produce same DID")
		}
	})

	t.Run("different keys produce different DIDs", func(t *testing.T) {
		privKey1, _ := keylib.GeneratePrivateKey()
		privKey2, _ := keylib.GeneratePrivateKey()

		did1, _ := FromPublicKey(privKey1.PublicKey())
		did2, _ := FromPublicKey(privKey2.PublicKey())

		if did1.String() == did2.String() {
			t.Error("different public keys should produce different DIDs")
		}
	})
}

// =============================================================================
// DIDKey Methods Edge Cases
// =============================================================================

func TestDIDKey_ZeroValueEdgeCases(t *testing.T) {
	t.Run("zero value IsZero returns true", func(t *testing.T) {
		var zeroDID DIDKey
		if !zeroDID.IsZero() {
			t.Error("zero value DIDKey should return IsZero() = true")
		}
	})

	t.Run("zero value String returns empty", func(t *testing.T) {
		var zeroDID DIDKey
		if zeroDID.String() != "" {
			t.Errorf("zero value String() should be empty, got %q", zeroDID.String())
		}
	})

	t.Run("zero value Method returns key (constant)", func(t *testing.T) {
		var zeroDID DIDKey
		// Method() always returns "key" since it's a constant for DID:Key
		if zeroDID.Method() != "key" {
			t.Errorf("zero value Method() should return 'key', got %q", zeroDID.Method())
		}
	})

	t.Run("zero value Identifier returns empty", func(t *testing.T) {
		var zeroDID DIDKey
		if zeroDID.Identifier() != "" {
			t.Errorf("zero value Identifier() should be empty, got %q", zeroDID.Identifier())
		}
	})

	t.Run("zero value Validate returns error", func(t *testing.T) {
		var zeroDID DIDKey
		if err := zeroDID.Validate(); err == nil {
			t.Error("zero value Validate() should return error")
		}
	})

	t.Run("zero value PublicKey returns zero key", func(t *testing.T) {
		var zeroDID DIDKey
		pubKey, err := zeroDID.PublicKey()
		if err == nil && !pubKey.IsZero() {
			t.Error("zero value PublicKey() should return error or zero key")
		}
	})
}

func TestDIDKey_EqualEdgeCases(t *testing.T) {
	privKey1, _ := keylib.GeneratePrivateKey()
	privKey2, _ := keylib.GeneratePrivateKey()
	did1, _ := FromPublicKey(privKey1.PublicKey())
	did2, _ := FromPublicKey(privKey2.PublicKey())
	did1Copy, _ := FromPublicKey(privKey1.PublicKey())

	t.Run("same DID equals itself", func(t *testing.T) {
		if !did1.Equal(did1) {
			t.Error("DID should equal itself")
		}
	})

	t.Run("same key DIDs are equal", func(t *testing.T) {
		if !did1.Equal(did1Copy) {
			t.Error("DIDs from same key should be equal")
		}
	})

	t.Run("different key DIDs not equal", func(t *testing.T) {
		if did1.Equal(did2) {
			t.Error("DIDs from different keys should not be equal")
		}
	})

	t.Run("zero value equals zero value", func(t *testing.T) {
		var zero1 DIDKey
		var zero2 DIDKey
		// Two zero values should be equal (both empty)
		if !zero1.Equal(&zero2) {
			t.Error("two zero value DIDs should be equal")
		}
	})

	t.Run("valid DID not equal to zero", func(t *testing.T) {
		var zeroDID DIDKey
		if did1.Equal(&zeroDID) {
			t.Error("valid DID should not equal zero value")
		}
	})

	t.Run("nil comparison", func(t *testing.T) {
		if did1.Equal(nil) {
			t.Error("DID should not equal nil")
		}
	})
}

func TestDIDKey_MatchesKeyEdgeCases(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did, _ := FromPublicKey(pubKey)

	otherPrivKey, _ := keylib.GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()

	t.Run("matches original key", func(t *testing.T) {
		if !did.MatchesKey(pubKey) {
			t.Error("DID should match its original public key")
		}
	})

	t.Run("does not match different key", func(t *testing.T) {
		if did.MatchesKey(otherPubKey) {
			t.Error("DID should not match different public key")
		}
	})

	t.Run("does not match zero key", func(t *testing.T) {
		var zeroKey keylib.NeuronPublicKey
		if did.MatchesKey(zeroKey) {
			t.Error("DID should not match zero public key")
		}
	})

	t.Run("zero DID matches nothing", func(t *testing.T) {
		var zeroDID DIDKey
		if zeroDID.MatchesKey(pubKey) {
			t.Error("zero DID should not match any key")
		}
	})
}

// =============================================================================
// Round-Trip Tests
// =============================================================================

func TestDIDKey_RoundTripEdgeCases(t *testing.T) {
	t.Run("parse then string", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		originalDID, _ := FromPublicKey(pubKey)
		originalStr := originalDID.String()

		// Parse the string
		parsedDID, err := Parse(originalStr)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// Should produce same string
		if parsedDID.String() != originalStr {
			t.Errorf("round-trip failed: %s != %s", parsedDID.String(), originalStr)
		}
	})

	t.Run("key extraction after round-trip", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		originalDID, _ := FromPublicKey(pubKey)

		// Parse the string representation
		parsedDID, _ := Parse(originalDID.String())

		// Extract public key
		extractedKey, err := parsedDID.PublicKey()
		if err != nil {
			t.Fatalf("PublicKey() failed: %v", err)
		}

		// Should match original
		if !extractedKey.Equal(pubKey) {
			t.Error("extracted public key should match original")
		}
	})

	t.Run("multiple round-trips stable", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did, _ := FromPublicKey(pubKey)

		str1 := did.String()
		parsed1, _ := Parse(str1)
		str2 := parsed1.String()
		parsed2, _ := Parse(str2)
		str3 := parsed2.String()

		if str1 != str2 || str2 != str3 {
			t.Error("multiple round-trips should produce stable results")
		}
	})
}

// =============================================================================
// Signature Verification Edge Cases
// =============================================================================

func TestDIDKey_VerifySignatureEdgeCases(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did, _ := FromPublicKey(pubKey)

	message := []byte("test message")
	signature, _ := privKey.SignMessage(message)

	t.Run("valid signature verifies", func(t *testing.T) {
		if !did.VerifySignature(message, signature) {
			t.Error("valid signature should verify")
		}
	})

	t.Run("wrong message fails", func(t *testing.T) {
		if did.VerifySignature([]byte("wrong message"), signature) {
			t.Error("wrong message should fail verification")
		}
	})

	t.Run("empty message with valid signature", func(t *testing.T) {
		emptyMsg := []byte{}
		emptySig, err := privKey.SignMessage(emptyMsg)
		if err != nil {
			t.Skipf("signing empty message not supported: %v", err)
		}

		if !did.VerifySignature(emptyMsg, emptySig) {
			t.Error("empty message with valid signature should verify")
		}
	})

	t.Run("zero DID verification fails", func(t *testing.T) {
		var zeroDID DIDKey
		if zeroDID.VerifySignature(message, signature) {
			t.Error("zero DID should not verify any signature")
		}
	})
}

// =============================================================================
// DIDKeyFromEVMAddress Edge Cases
// =============================================================================

func TestDIDKeyFromEVMAddress_EdgeCases(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	evmAddr := pubKey.EVMAddress()

	t.Run("EVM address cannot create DID (one-way derivation)", func(t *testing.T) {
		// DID:Key requires the full public key, but EVM addresses are one-way
		// derived from the public key (keccak256 hash). Therefore, we cannot
		// create a valid DID:Key from just an EVM address.
		_, err := DIDKeyFromEVMAddress(evmAddr)
		if err == nil {
			t.Error("DIDKeyFromEVMAddress should fail - EVM addresses are one-way derived")
		}
	})

	t.Run("zero EVM address also fails", func(t *testing.T) {
		var zeroAddr keylib.EVMAddress
		_, err := DIDKeyFromEVMAddress(zeroAddr)
		if err == nil {
			t.Error("DIDKeyFromEVMAddress should fail for zero address")
		}
	})
}

// =============================================================================
// Resolve Edge Cases
// =============================================================================

func TestDIDKey_ResolveEdgeCases(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did, _ := FromPublicKey(pubKey)

	t.Run("resolves to valid document", func(t *testing.T) {
		doc, err := did.Resolve()
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if doc == nil {
			t.Fatal("Resolve should return non-nil document")
		}
	})

	t.Run("zero DID resolve behavior", func(t *testing.T) {
		var zeroDID DIDKey
		_, err := zeroDID.Resolve()
		// Depending on implementation, may return error
		if err == nil {
			t.Log("zero DID Resolve returned no error")
		}
	})
}

// =============================================================================
// Encoding Consistency Tests
// =============================================================================

func TestDIDKey_EncodingConsistency(t *testing.T) {
	// Generate many keys and verify encoding is consistent
	for i := 0; i < 100; i++ {
		privKey, err := keylib.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("key generation failed: %v", err)
		}
		pubKey := privKey.PublicKey()

		did1, _ := FromPublicKey(pubKey)
		did2, _ := FromPublicKey(pubKey)

		// Must be identical
		if did1.String() != did2.String() {
			t.Errorf("iteration %d: encoding inconsistent", i)
		}

		// Must be parseable
		parsed, err := Parse(did1.String())
		if err != nil {
			t.Errorf("iteration %d: parse failed: %v", i, err)
		}

		// Must match
		if !parsed.Equal(did1) {
			t.Errorf("iteration %d: round-trip failed", i)
		}
	}
}

func TestDIDKey_IdentifierFormat(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did, _ := FromPublicKey(pubKey)

	identifier := did.Identifier()

	t.Run("starts with z (base58btc)", func(t *testing.T) {
		if !strings.HasPrefix(identifier, "z") {
			t.Errorf("identifier should start with 'z', got: %s", identifier)
		}
	})

	t.Run("contains only valid base58btc characters", func(t *testing.T) {
		// Base58btc alphabet: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
		validChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

		for i, c := range identifier {
			if i == 0 {
				continue // skip 'z' prefix
			}
			if !strings.ContainsRune(validChars, c) {
				t.Errorf("invalid character %q at position %d in identifier", c, i)
			}
		}
	})

	t.Run("reasonable length", func(t *testing.T) {
		// secp256k1 compressed public key + multicodec = about 35 bytes
		// base58btc encoding is roughly 1.37x the byte length
		// So expect identifier (including 'z') to be roughly 48-52 chars
		if len(identifier) < 40 || len(identifier) > 60 {
			t.Errorf("unexpected identifier length: %d", len(identifier))
		}
	})
}
