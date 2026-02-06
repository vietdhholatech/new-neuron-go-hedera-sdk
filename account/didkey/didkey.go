// Package didkey provides a did:key implementation for NeuronAccount.
//
// The did:key method is a self-describing DID method that encodes the public key
// directly in the DID identifier. This makes it:
//   - Self-resolving (no external resolution needed)
//   - Deterministic (same key always produces same DID)
//   - Offline-capable (works without network)
//
// Format for secp256k1 keys:
//
//	did:key:z<base58btc(0xe7 + compressed-pubkey-33-bytes)>
//
// Where:
//   - 0xe7 is the multicodec identifier for secp256k1-pub
//   - The compressed public key is 33 bytes (0x02 or 0x03 prefix + 32 bytes)
//   - 'z' prefix indicates base58btc encoding (multibase)
package didkey

import (
	"fmt"
	"strings"

	"github.com/aspect-build/neuron-go-hedera-sdk/account"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
	"github.com/mr-tron/base58"
)

func init() {
	// Register did:key parser with the account package
	account.RegisterDIDParser(account.DIDMethodKey, func(didString string) (account.NeuronDID, error) {
		return Parse(didString)
	})
}

// Ensure DIDKey implements the required interfaces.
var (
	_ account.NeuronDID        = (*DIDKey)(nil)
	_ account.NeuronDIDWithKey = (*DIDKey)(nil)
)

// DIDKey implements the did:key method for secp256k1 public keys.
// It encodes the public key directly in the DID, making it self-describing.
type DIDKey struct {
	// raw is the full DID string (e.g., "did:key:z6Mk...")
	raw string

	// identifier is the method-specific part (e.g., "z6Mk...")
	identifier string

	// publicKey is the decoded NeuronPublicKey
	publicKey keylib.NeuronPublicKey
}

// FromPublicKey creates a did:key DID from a NeuronPublicKey.
// The resulting DID will be deterministic - the same public key always produces the same DID.
func FromPublicKey(pubKey keylib.NeuronPublicKey) (*DIDKey, error) {
	if pubKey.IsZero() {
		return nil, fmt.Errorf("didkey: cannot create DID from zero-value public key")
	}

	// Get compressed public key bytes (33 bytes)
	compressedBytes := pubKey.CompressedBytes()

	// Prepend multicodec identifier for secp256k1-pub
	// Using varint encoding: 0xe7 encodes as two bytes: 0xe7, 0x01
	prefixed := make([]byte, 0, 2+len(compressedBytes))
	prefixed = append(prefixed, 0xe7, 0x01) // secp256k1-pub multicodec varint
	prefixed = append(prefixed, compressedBytes[:]...)

	// Encode with base58btc (multibase 'z' prefix)
	encoded := base58.Encode(prefixed)
	identifier := "z" + encoded

	raw := "did:key:" + identifier

	return &DIDKey{
		raw:        raw,
		identifier: identifier,
		publicKey:  pubKey,
	}, nil
}

// Parse parses a did:key string and validates it.
// Returns error if the string is not a valid did:key or uses an unsupported key type.
func Parse(didString string) (*DIDKey, error) {
	// Check prefix
	if !strings.HasPrefix(didString, "did:key:") {
		return nil, fmt.Errorf("didkey: invalid DID format, expected 'did:key:' prefix, got: %s", didString)
	}

	identifier := strings.TrimPrefix(didString, "did:key:")
	if len(identifier) == 0 {
		return nil, fmt.Errorf("didkey: empty identifier")
	}

	// Check multibase prefix (must be 'z' for base58btc)
	if identifier[0] != 'z' {
		return nil, fmt.Errorf("didkey: unsupported multibase encoding '%c', expected 'z' (base58btc)", identifier[0])
	}

	// Decode base58btc (skip the 'z' prefix)
	decoded, err := base58.Decode(identifier[1:])
	if err != nil {
		return nil, fmt.Errorf("didkey: failed to decode base58btc: %w", err)
	}

	// Check minimum length (2 bytes multicodec + 33 bytes compressed pubkey)
	if len(decoded) < 35 {
		return nil, fmt.Errorf("didkey: decoded data too short, expected at least 35 bytes, got %d", len(decoded))
	}

	// Check multicodec prefix for secp256k1-pub (varint: 0xe7, 0x01)
	if decoded[0] != 0xe7 || decoded[1] != 0x01 {
		return nil, fmt.Errorf("didkey: unsupported key type, expected secp256k1-pub (0xe7, 0x01), got (0x%02x, 0x%02x)", decoded[0], decoded[1])
	}

	// Extract public key bytes (skip 2-byte multicodec prefix)
	pubKeyBytes := decoded[2:]

	// Parse as NeuronPublicKey
	pubKey, err := keylib.PublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("didkey: invalid public key: %w", err)
	}

	return &DIDKey{
		raw:        didString,
		identifier: identifier,
		publicKey:  pubKey,
	}, nil
}

// MustParse parses a did:key string and panics on error.
// This is intended for use in tests and initialization of known-valid DIDs.
func MustParse(didString string) *DIDKey {
	d, err := Parse(didString)
	if err != nil {
		panic(err)
	}
	return d
}

// String returns the full DID string.
// Example: "did:key:zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme"
func (d *DIDKey) String() string {
	return d.raw
}

// Method returns "key".
func (d *DIDKey) Method() string {
	return account.DIDMethodKey
}

// Identifier returns the method-specific identifier.
// Example: "zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme"
func (d *DIDKey) Identifier() string {
	return d.identifier
}

// Validate checks if the DID is well-formed.
// For did:key, this verifies the public key is valid on the secp256k1 curve.
func (d *DIDKey) Validate() error {
	if d.publicKey.IsZero() {
		return fmt.Errorf("didkey: invalid DID, public key is zero-value")
	}
	return nil
}

// Equal compares this DID with another for equality.
func (d *DIDKey) Equal(other account.NeuronDID) bool {
	if other == nil {
		return false
	}
	// Compare by string representation for simplicity and correctness
	return d.raw == other.String()
}

// PublicKey returns the NeuronPublicKey encoded in this DID.
func (d *DIDKey) PublicKey() (keylib.NeuronPublicKey, error) {
	if d.publicKey.IsZero() {
		return keylib.NeuronPublicKey{}, fmt.Errorf("didkey: public key is zero-value")
	}
	return d.publicKey, nil
}

// MatchesKey checks if this DID corresponds to the given public key.
func (d *DIDKey) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	if d.publicKey.IsZero() || pubKey.IsZero() {
		return false
	}
	// Use constant-time comparison from keylib
	return d.publicKey.Equal(pubKey)
}

// IsZero returns true if this is a zero-value DIDKey.
func (d *DIDKey) IsZero() bool {
	return d == nil || d.raw == "" || d.publicKey.IsZero()
}

// Resolve generates a DID Document for this did:key.
// did:key is self-describing, so no external resolution is needed.
func (d *DIDKey) Resolve() (*account.DIDDocument, error) {
	if d.publicKey.IsZero() {
		return nil, fmt.Errorf("didkey: cannot resolve DID with zero-value public key")
	}

	// The verification method ID is the DID itself with a fragment
	verificationMethodID := d.raw + "#" + d.identifier

	// Get the compressed public key as multibase
	compressedBytes := d.publicKey.CompressedBytes()
	prefixed := make([]byte, 0, 2+len(compressedBytes))
	prefixed = append(prefixed, 0xe7, 0x01)
	prefixed = append(prefixed, compressedBytes[:]...)
	publicKeyMultibase := "z" + base58.Encode(prefixed)

	return &account.DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/secp256k1-2019/v1",
		},
		ID:         d.raw,
		Controller: []string{d.raw},
		VerificationMethod: []account.VerificationMethod{
			{
				ID:                 verificationMethodID,
				Type:               "EcdsaSecp256k1VerificationKey2019",
				Controller:         d.raw,
				PublicKeyMultibase: publicKeyMultibase,
			},
		},
		Authentication:  []string{verificationMethodID},
		AssertionMethod: []string{verificationMethodID},
	}, nil
}

// VerifySignature verifies a signature against the public key in this DID.
func (d *DIDKey) VerifySignature(message []byte, signature keylib.Signature) bool {
	if d.publicKey.IsZero() {
		return false
	}
	return d.publicKey.Verify(message, signature)
}

// DIDKeyFromEVMAddress creates a did:key by looking up or deriving from an EVM address.
// Note: This is not directly possible as EVM addresses are derived from public keys,
// not the other way around. This function is provided for documentation purposes
// and will return an error indicating the limitation.
func DIDKeyFromEVMAddress(addr keylib.EVMAddress) (*DIDKey, error) {
	return nil, fmt.Errorf("didkey: cannot create did:key from EVM address; public key is required (EVM addresses are one-way derived)")
}
