package keylib

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"io"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
)

// NeuronPrivateKey is an immutable, validated ECDSA secp256k1 private key.
// The zero value is invalid; methods return errors on zero value.
//
// This type implements crypto.Signer and provides the unified interface
// for private key operations across Hedera, Ethereum, and libp2p ecosystems.
type NeuronPrivateKey struct {
	key *secp256k1.PrivateKey // unexported, never nil after valid construction
}

// Compile-time check that NeuronPrivateKey implements crypto.Signer.
var _ crypto.Signer = NeuronPrivateKey{}

// IsZero returns true if this is a zero-value (invalid) private key.
func (k NeuronPrivateKey) IsZero() bool {
	return k.key == nil
}

// checkValid returns an error if the key is zero-value.
func (k NeuronPrivateKey) checkValid(op string) error {
	if k.IsZero() {
		return errZeroValue(op, "NeuronPrivateKey")
	}
	return nil
}

// PublicKey derives and returns the corresponding public key.
// Returns zero-value NeuronPublicKey if this key is zero-value.
func (k NeuronPrivateKey) PublicKey() NeuronPublicKey {
	if k.IsZero() {
		return NeuronPublicKey{}
	}
	return NeuronPublicKey{key: k.key.PubKey()}
}

// EVMAddress derives the Ethereum address from this private key.
//
// WARNING: Returns ZeroEVMAddress (0x0000...0000) if key is zero-value.
// The zero address is the Ethereum burn address - funds sent there are lost forever.
// Use EVMAddressSafe() if you need error handling for zero keys.
func (k NeuronPrivateKey) EVMAddress() EVMAddress {
	return k.PublicKey().EVMAddress()
}

// EVMAddressSafe derives the Ethereum address with explicit error handling.
// Returns an error if the private key is zero-value, preventing accidental
// derivation to the burn address (0x0000...0000).
func (k NeuronPrivateKey) EVMAddressSafe() (EVMAddress, error) {
	const op = "NeuronPrivateKey.EVMAddressSafe"
	if err := k.checkValid(op); err != nil {
		return EVMAddress{}, err
	}
	return k.EVMAddress(), nil
}

// PeerID derives the libp2p peer ID from this private key.
// Returns error if key is zero-value.
func (k NeuronPrivateKey) PeerID() (PeerID, error) {
	const op = "NeuronPrivateKey.PeerID"
	if err := k.checkValid(op); err != nil {
		return PeerID{}, err
	}
	return k.PublicKey().PeerID()
}

// Bytes returns the raw 32-byte private key scalar.
// Returns zero array if key is zero-value.
//
// WARNING: Handle the returned bytes with care as they contain secret key material.
func (k NeuronPrivateKey) Bytes() [32]byte {
	var result [32]byte
	if k.IsZero() {
		return result
	}
	keyBytes := k.key.Serialize()
	copy(result[:], keyBytes)
	return result
}

// Hex returns the private key as a lowercase hex string with "0x" prefix.
// Returns empty string if key is zero-value.
//
// WARNING: Handle the returned string with care as it contains secret key material.
func (k NeuronPrivateKey) Hex() string {
	if k.IsZero() {
		return ""
	}
	return encodeHex(k.key.Serialize(), true)
}

// HexWithoutPrefix returns the private key as hex without "0x" prefix.
// Useful for compatibility with systems that expect raw hex.
func (k NeuronPrivateKey) HexWithoutPrefix() string {
	if k.IsZero() {
		return ""
	}
	return encodeHex(k.key.Serialize(), false)
}

// ToECDSA returns the key as a standard library *ecdsa.PrivateKey.
// Returns nil if key is zero-value.
//
// WARNING: The returned key shares the same secret material.
func (k NeuronPrivateKey) ToECDSA() *ecdsa.PrivateKey {
	if k.IsZero() {
		return nil
	}
	return k.key.ToECDSA()
}

// ToHederaPrivateKey returns the key as a Hedera SDK PrivateKey.
// Returns zero-value Hedera key if this key is zero-value or conversion fails.
//
// WARNING: Silently returns empty key on failure. Use ToHederaPrivateKeySafe()
// if you need explicit error handling.
func (k NeuronPrivateKey) ToHederaPrivateKey() hiero.PrivateKey {
	if k.IsZero() {
		return hiero.PrivateKey{}
	}

	keyBytes := k.Bytes()
	hederaKey, err := hiero.PrivateKeyFromBytesECDSA(keyBytes[:])
	if err != nil {
		// This should never happen with a valid key
		return hiero.PrivateKey{}
	}
	return hederaKey
}

// ToHederaPrivateKeySafe converts to Hedera SDK PrivateKey with error handling.
// Returns an error if the key is zero-value or conversion fails.
func (k NeuronPrivateKey) ToHederaPrivateKeySafe() (hiero.PrivateKey, error) {
	const op = "NeuronPrivateKey.ToHederaPrivateKeySafe"

	if err := k.checkValid(op); err != nil {
		return hiero.PrivateKey{}, err
	}

	keyBytes := k.Bytes()
	hederaKey, err := hiero.PrivateKeyFromBytesECDSA(keyBytes[:])
	if err != nil {
		return hiero.PrivateKey{}, wrapError(op, ErrKindDerivation, "failed to create Hedera key", err)
	}
	return hederaKey, nil
}

// SignMessage signs a message with this private key.
// The message is hashed with Keccak256 before signing.
// Returns a recoverable signature (65 bytes: R || S || V).
func (k NeuronPrivateKey) SignMessage(msg []byte) (Signature, error) {
	const op = "NeuronPrivateKey.SignMessage"

	if err := k.checkValid(op); err != nil {
		return Signature{}, err
	}

	// Hash the message
	hash := ethcrypto.Keccak256(msg)
	return k.SignDigest(*(*[32]byte)(hash))
}

// SignDigest signs a pre-hashed 32-byte digest with this private key.
// Returns a recoverable signature (65 bytes: R || S || V).
func (k NeuronPrivateKey) SignDigest(digest [32]byte) (Signature, error) {
	const op = "NeuronPrivateKey.SignDigest"

	if err := k.checkValid(op); err != nil {
		return Signature{}, err
	}

	// Sign using go-ethereum's secp256k1 implementation
	ecdsaKey := k.ToECDSA()
	sigBytes, err := ethcrypto.Sign(digest[:], ecdsaKey)
	if err != nil {
		return Signature{}, wrapError(op, ErrKindInvalidKey, "signing failed", err)
	}

	return SignatureFromBytes(sigBytes)
}

// Sign implements the crypto.Signer interface.
// It signs the given digest using ECDSA and returns the ASN.1 DER encoded signature.
//
// The rand parameter is ignored as we use deterministic signing (RFC 6979).
// The opts parameter is also ignored for ECDSA signatures.
func (k NeuronPrivateKey) Sign(randReader io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	const op = "NeuronPrivateKey.Sign"

	if err := k.checkValid(op); err != nil {
		return nil, err
	}

	if len(digest) != 32 {
		return nil, errInvalidLength(op, 32, len(digest), "bytes for digest")
	}

	ecdsaKey := k.ToECDSA()

	// Use Go's standard ECDSA signing
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, digest)
	if err != nil {
		return nil, wrapError(op, ErrKindInvalidKey, "ECDSA signing failed", err)
	}

	// Return signature as R || S (64 bytes)
	sig := signatureFromRSV(r, s, 0)
	// Return first 64 bytes (R and S only, no V for crypto.Signer)
	return sig.Bytes()[:64], nil
}

// Public implements the crypto.Signer interface.
// Returns the corresponding public key as crypto.PublicKey.
func (k NeuronPrivateKey) Public() crypto.PublicKey {
	if k.IsZero() {
		return nil
	}
	return k.ToECDSA().Public()
}

// MatchesPublicKey checks if this private key corresponds to the given public key.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPrivateKey) MatchesPublicKey(pub NeuronPublicKey) bool {
	if k.IsZero() || pub.IsZero() {
		return false
	}
	derived := k.PublicKey()
	return derived.Equal(pub)
}

// MatchesEVMAddress checks if this private key corresponds to the given EVM address.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPrivateKey) MatchesEVMAddress(addr EVMAddress) bool {
	if k.IsZero() {
		return false
	}
	return k.PublicKey().MatchesEVMAddress(addr)
}

// Zeroize overwrites the private key material with zeros.
// After calling Zeroize, the key becomes a zero-value key.
//
// This provides a best-effort approach to clearing sensitive data from memory.
// Due to Go's garbage collector, complete removal is not guaranteed.
func (k *NeuronPrivateKey) Zeroize() {
	if k == nil || k.key == nil {
		return
	}

	// Get the key bytes and zero them
	keyBytes := k.key.Serialize()
	secureZero(keyBytes)

	// Set the key to nil
	k.key = nil
}

// Equal returns true if this private key equals another.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPrivateKey) Equal(other NeuronPrivateKey) bool {
	if k.IsZero() || other.IsZero() {
		return k.IsZero() && other.IsZero()
	}
	return constantTimeEqual(k.key.Serialize(), other.key.Serialize())
}

// newPrivateKeyFromSecp256k1 creates a NeuronPrivateKey from a secp256k1.PrivateKey.
// This is an internal constructor.
func newPrivateKeyFromSecp256k1(key *secp256k1.PrivateKey) NeuronPrivateKey {
	if key == nil {
		return NeuronPrivateKey{}
	}
	return NeuronPrivateKey{key: key}
}

// newPrivateKeyFromECDSA creates a NeuronPrivateKey from an ecdsa.PrivateKey.
// This is an internal constructor.
func newPrivateKeyFromECDSA(key *ecdsa.PrivateKey) NeuronPrivateKey {
	if key == nil {
		return NeuronPrivateKey{}
	}

	// Get the D value and create secp256k1 private key
	keyBytes := padLeftZeros(key.D.Bytes(), 32)
	secpKey := secp256k1.PrivKeyFromBytes(keyBytes)

	return NeuronPrivateKey{key: secpKey}
}
