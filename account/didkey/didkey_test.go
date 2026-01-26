package didkey

import (
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Test Vectors - Known Deterministic Values
// =============================================================================
//
// These vectors are derived from a known secp256k1 public key.
// They ensure that our did:key implementation is deterministic and cross-compatible.

var (
	// Test public key (compressed secp256k1)
	testPubKeyHex = "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"

	// Expected did:key derived from the test public key
	// Format: did:key:z<base58btc(0xe7,0x01 + compressed-pubkey-33-bytes)>
	expectedDIDKey = "did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9"

	// Expected identifier (without "did:key:" prefix)
	expectedIdentifier = "zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9"
)

// getTestPublicKey returns the test public key for testing.
func getTestPublicKey(t *testing.T) keylib.NeuronPublicKey {
	t.Helper()
	pubKey, err := keylib.ParsePublicKeyHex(testPubKeyHex)
	if err != nil {
		t.Fatalf("failed to parse test public key: %v", err)
	}
	return pubKey
}

// =============================================================================
// FromPublicKey Tests
// =============================================================================

func TestFromPublicKey_Comprehensive(t *testing.T) {
	t.Run("creates valid DID from public key", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, err := FromPublicKey(pubKey)
		if err != nil {
			t.Fatalf("FromPublicKey() error = %v", err)
		}
		if did == nil {
			t.Fatal("expected non-nil DID")
		}
		if did.String() == "" {
			t.Error("DID string should not be empty")
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		pubKey := getTestPublicKey(t)

		did1, _ := FromPublicKey(pubKey)
		did2, _ := FromPublicKey(pubKey)

		if did1.String() != did2.String() {
			t.Errorf("FromPublicKey() not deterministic: %v != %v", did1.String(), did2.String())
		}
	})

	t.Run("format starts with did:key:", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		if len(did.String()) < 8 || did.String()[:8] != "did:key:" {
			t.Errorf("DID should start with 'did:key:', got: %s", did.String())
		}
	})

	t.Run("identifier starts with z (base58btc multibase)", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		identifier := did.Identifier()
		if len(identifier) == 0 || identifier[0] != 'z' {
			t.Errorf("identifier should start with 'z', got: %s", identifier)
		}
	})

	t.Run("rejects zero-value public key", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		_, err := FromPublicKey(zeroPubKey)
		if err == nil {
			t.Error("expected error for zero-value public key")
		}
	})
}

func TestFromPublicKey_KnownVector(t *testing.T) {
	t.Run("matches expected did:key", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, err := FromPublicKey(pubKey)
		if err != nil {
			t.Fatalf("FromPublicKey() error = %v", err)
		}

		if did.String() != expectedDIDKey {
			t.Errorf("DID = %v, want %v", did.String(), expectedDIDKey)
		}
	})

	t.Run("identifier matches expected", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		if did.Identifier() != expectedIdentifier {
			t.Errorf("Identifier() = %v, want %v", did.Identifier(), expectedIdentifier)
		}
	})
}

// =============================================================================
// Parse Tests - Valid Formats
// =============================================================================

func TestParse_ValidFormats(t *testing.T) {
	t.Run("parses valid did:key", func(t *testing.T) {
		did, err := Parse(expectedDIDKey)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if did == nil {
			t.Fatal("expected non-nil DID")
		}
	})

	t.Run("parsed DID has correct string representation", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.String() != expectedDIDKey {
			t.Errorf("String() = %v, want %v", did.String(), expectedDIDKey)
		}
	})

	t.Run("parsed DID has correct identifier", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.Identifier() != expectedIdentifier {
			t.Errorf("Identifier() = %v, want %v", did.Identifier(), expectedIdentifier)
		}
	})

	t.Run("parsed DID has correct method", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.Method() != "key" {
			t.Errorf("Method() = %v, want 'key'", did.Method())
		}
	})
}

// =============================================================================
// Parse Tests - Invalid Formats
// =============================================================================

func TestParse_InvalidFormats(t *testing.T) {
	invalidCases := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "empty string",
			input:       "",
			errContains: "prefix",
		},
		{
			name:        "missing did prefix",
			input:       "key:zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme",
			errContains: "prefix",
		},
		{
			name:        "wrong method",
			input:       "did:web:example.com",
			errContains: "prefix",
		},
		{
			name:        "empty identifier",
			input:       "did:key:",
			errContains: "empty identifier",
		},
		{
			name:        "wrong multibase prefix (not z)",
			input:       "did:key:bQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme",
			errContains: "multibase",
		},
		{
			name:        "invalid base58 characters",
			input:       "did:key:zQ3shokFTS3br0OIL82RUDfCZESWL1ZdCEJwekUDPQiYBme",
			errContains: "base58",
		},
		{
			name:        "too short (less than 35 bytes decoded)",
			input:       "did:key:zQ3s",
			errContains: "too short",
		},
		{
			name:        "wrong multicodec (not secp256k1)",
			input:       "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
			errContains: "unsupported key type",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// =============================================================================
// MustParse Tests
// =============================================================================

func TestMustParse(t *testing.T) {
	t.Run("valid DID does not panic", func(t *testing.T) {
		did := MustParse(expectedDIDKey)
		if did == nil {
			t.Error("expected non-nil DID")
		}
	})

	t.Run("invalid DID panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic, got none")
			}
		}()
		MustParse("invalid")
	})
}

// =============================================================================
// DIDKey Round-Trip Tests
// =============================================================================

func TestDIDKey_RoundTrip(t *testing.T) {
	t.Run("pubKey -> FromPublicKey -> String -> Parse -> PublicKey -> Equal", func(t *testing.T) {
		originalPubKey := getTestPublicKey(t)

		// Create DID from public key
		did1, err := FromPublicKey(originalPubKey)
		if err != nil {
			t.Fatalf("FromPublicKey() error = %v", err)
		}

		// Convert to string
		didString := did1.String()

		// Parse back
		did2, err := Parse(didString)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		// Extract public key
		extractedPubKey, err := did2.PublicKey()
		if err != nil {
			t.Fatalf("PublicKey() error = %v", err)
		}

		// Compare public keys
		if !originalPubKey.Equal(extractedPubKey) {
			t.Error("round-trip failed: public keys not equal")
		}
	})

	t.Run("DID string round-trip preserves exact value", func(t *testing.T) {
		did1, _ := Parse(expectedDIDKey)
		str1 := did1.String()
		did2, _ := Parse(str1)
		str2 := did2.String()

		if str1 != str2 {
			t.Errorf("string round-trip failed: %v != %v", str1, str2)
		}
	})
}

// =============================================================================
// DIDKey Methods Tests
// =============================================================================

func TestDIDKey_String(t *testing.T) {
	t.Run("returns full DID string", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		s := did.String()
		if s != expectedDIDKey {
			t.Errorf("String() = %v, want %v", s, expectedDIDKey)
		}
	})
}

func TestDIDKey_Method(t *testing.T) {
	t.Run("returns 'key'", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.Method() != "key" {
			t.Errorf("Method() = %v, want 'key'", did.Method())
		}
	})
}

func TestDIDKey_Identifier(t *testing.T) {
	t.Run("returns method-specific identifier without prefix", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		identifier := did.Identifier()
		if identifier != expectedIdentifier {
			t.Errorf("Identifier() = %v, want %v", identifier, expectedIdentifier)
		}
	})

	t.Run("identifier does not contain 'did:key:'", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		identifier := did.Identifier()
		if len(identifier) >= 8 && identifier[:8] == "did:key:" {
			t.Error("Identifier() should not contain 'did:key:' prefix")
		}
	})
}

func TestDIDKey_Validate(t *testing.T) {
	t.Run("valid DID passes validation", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if err := did.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("DID from valid public key passes validation", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)
		if err := did.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})
}

func TestDIDKey_Equal(t *testing.T) {
	t.Run("same DID equals itself", func(t *testing.T) {
		did1, _ := Parse(expectedDIDKey)
		if !did1.Equal(did1) {
			t.Error("DID should equal itself")
		}
	})

	t.Run("equivalent DIDs are equal", func(t *testing.T) {
		did1, _ := Parse(expectedDIDKey)
		did2, _ := Parse(expectedDIDKey)
		if !did1.Equal(did2) {
			t.Error("equivalent DIDs should be equal")
		}
	})

	t.Run("different DIDs are not equal", func(t *testing.T) {
		did1, _ := Parse(expectedDIDKey)
		// Create a different DID by generating a new key
		privKey, _ := keylib.GeneratePrivateKey()
		did2, _ := FromPublicKey(privKey.PublicKey())
		if did1.Equal(did2) {
			t.Error("different DIDs should not be equal")
		}
	})

	t.Run("DID does not equal nil", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.Equal(nil) {
			t.Error("DID should not equal nil")
		}
	})
}

// =============================================================================
// DIDKey.PublicKey Tests
// =============================================================================

func TestDIDKey_PublicKey(t *testing.T) {
	t.Run("returns correct public key from parsed DID", func(t *testing.T) {
		originalPubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(originalPubKey)

		extractedPubKey, err := did.PublicKey()
		if err != nil {
			t.Fatalf("PublicKey() error = %v", err)
		}

		if !originalPubKey.Equal(extractedPubKey) {
			t.Error("extracted public key does not match original")
		}
	})

	t.Run("returns correct public key from string-parsed DID", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		pubKey, err := did.PublicKey()
		if err != nil {
			t.Fatalf("PublicKey() error = %v", err)
		}

		// Verify by comparing with expected public key
		expectedPubKey := getTestPublicKey(t)
		if !pubKey.Equal(expectedPubKey) {
			t.Error("public key from parsed DID does not match expected")
		}
	})
}

// =============================================================================
// DIDKey.MatchesKey Tests
// =============================================================================

func TestDIDKey_MatchesKey(t *testing.T) {
	t.Run("matches correct public key", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		if !did.MatchesKey(pubKey) {
			t.Error("MatchesKey() should return true for correct key")
		}
	})

	t.Run("does not match different public key", func(t *testing.T) {
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		// Generate a different key
		otherPrivKey, _ := keylib.GeneratePrivateKey()
		otherPubKey := otherPrivKey.PublicKey()

		if did.MatchesKey(otherPubKey) {
			t.Error("MatchesKey() should return false for different key")
		}
	})

	t.Run("does not match zero public key", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		var zeroPubKey keylib.NeuronPublicKey
		if did.MatchesKey(zeroPubKey) {
			t.Error("MatchesKey() should return false for zero key")
		}
	})
}

// =============================================================================
// DIDKey.IsZero Tests
// =============================================================================

func TestDIDKey_IsZero(t *testing.T) {
	t.Run("nil is zero", func(t *testing.T) {
		var did *DIDKey
		if !did.IsZero() {
			t.Error("nil DIDKey should be zero")
		}
	})

	t.Run("valid DID is not zero", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		if did.IsZero() {
			t.Error("valid DID should not be zero")
		}
	})
}

// =============================================================================
// DIDKey.Resolve Tests
// =============================================================================

func TestDIDKey_Resolve(t *testing.T) {
	t.Run("resolves to valid DID document", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, err := did.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if doc == nil {
			t.Fatal("expected non-nil document")
		}
	})

	t.Run("document has correct ID", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if doc.ID != expectedDIDKey {
			t.Errorf("document ID = %v, want %v", doc.ID, expectedDIDKey)
		}
	})

	t.Run("document has context", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.Context) == 0 {
			t.Error("document should have context")
		}
	})

	t.Run("document has verification method", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.VerificationMethod) == 0 {
			t.Error("document should have verification method")
		}
	})

	t.Run("verification method has correct type", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.VerificationMethod) > 0 {
			vm := doc.VerificationMethod[0]
			if vm.Type != "EcdsaSecp256k1VerificationKey2019" {
				t.Errorf("verification method type = %v, want EcdsaSecp256k1VerificationKey2019", vm.Type)
			}
		}
	})

	t.Run("verification method ID contains fragment", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.VerificationMethod) > 0 {
			vm := doc.VerificationMethod[0]
			if !containsString(vm.ID, "#") {
				t.Error("verification method ID should contain fragment (#)")
			}
		}
	})

	t.Run("document has authentication", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.Authentication) == 0 {
			t.Error("document should have authentication")
		}
	})

	t.Run("document has assertion method", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.AssertionMethod) == 0 {
			t.Error("document should have assertion method")
		}
	})

	t.Run("document has controller set to self", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		doc, _ := did.Resolve()
		if len(doc.Controller) == 0 || doc.Controller[0] != expectedDIDKey {
			t.Error("document controller should be self")
		}
	})
}

// =============================================================================
// DIDKey.VerifySignature Tests
// =============================================================================

func TestDIDKey_VerifySignature(t *testing.T) {
	t.Run("verifies valid signature", func(t *testing.T) {
		// Generate a key pair
		privKey, err := keylib.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GenerateKey() error = %v", err)
		}

		// Create DID from public key
		did, err := FromPublicKey(privKey.PublicKey())
		if err != nil {
			t.Fatalf("FromPublicKey() error = %v", err)
		}

		// Sign a message
		message := []byte("test message")
		signature, err := privKey.SignMessage(message)
		if err != nil {
			t.Fatalf("Sign() error = %v", err)
		}

		// Verify signature using DID
		if !did.VerifySignature(message, signature) {
			t.Error("VerifySignature() should return true for valid signature")
		}
	})

	t.Run("rejects signature from different key", func(t *testing.T) {
		// Generate two key pairs
		privKey1, _ := keylib.GeneratePrivateKey()
		privKey2, _ := keylib.GeneratePrivateKey()

		// Create DID from first key
		did, _ := FromPublicKey(privKey1.PublicKey())

		// Sign with second key
		message := []byte("test message")
		signature, _ := privKey2.SignMessage(message)

		// Verify should fail
		if did.VerifySignature(message, signature) {
			t.Error("VerifySignature() should return false for signature from different key")
		}
	})

	t.Run("rejects invalid message", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did, _ := FromPublicKey(privKey.PublicKey())

		originalMessage := []byte("original message")
		signature, _ := privKey.SignMessage(originalMessage)

		tamperedMessage := []byte("tampered message")
		if did.VerifySignature(tamperedMessage, signature) {
			t.Error("VerifySignature() should return false for tampered message")
		}
	})
}

// =============================================================================
// DIDKeyFromEVMAddress Tests
// =============================================================================

func TestDIDKeyFromEVMAddress(t *testing.T) {
	t.Run("returns error (not supported)", func(t *testing.T) {
		// Create an EVM address
		privKey, _ := keylib.GeneratePrivateKey()
		evmAddr := privKey.PublicKey().EVMAddress()

		_, err := DIDKeyFromEVMAddress(evmAddr)
		if err == nil {
			t.Error("DIDKeyFromEVMAddress() should return error")
		}
	})
}

// =============================================================================
// Interface Implementation Tests
// =============================================================================

func TestDIDKey_ImplementsNeuronDID(t *testing.T) {
	// This test is compile-time verified by the var _ declarations in didkey.go
	// but we can also verify at runtime
	t.Run("implements NeuronDID interface", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)

		// Verify all interface methods are callable
		_ = did.String()
		_ = did.Method()
		_ = did.Identifier()
		_ = did.Validate()
		_ = did.Equal(did)
	})
}

func TestDIDKey_ImplementsNeuronDIDWithKey(t *testing.T) {
	t.Run("implements NeuronDIDWithKey interface", func(t *testing.T) {
		did, _ := Parse(expectedDIDKey)
		pubKey := getTestPublicKey(t)

		// Verify extended interface methods
		_, _ = did.PublicKey()
		_ = did.MatchesKey(pubKey)
	})
}

// =============================================================================
// Multicodec Tests
// =============================================================================

func TestMulticodec_Secp256k1(t *testing.T) {
	t.Run("uses correct multicodec prefix", func(t *testing.T) {
		// The secp256k1-pub multicodec is 0xe7, encoded as varint: 0xe7, 0x01
		pubKey := getTestPublicKey(t)
		did, _ := FromPublicKey(pubKey)

		// Parse and verify the decoded bytes have correct prefix
		did2, _ := Parse(did.String())
		extractedPubKey, _ := did2.PublicKey()

		// If we can extract the public key correctly, the multicodec was handled properly
		if !pubKey.Equal(extractedPubKey) {
			t.Error("multicodec handling failed: public keys don't match")
		}
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
