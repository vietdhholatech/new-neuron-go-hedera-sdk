package keylib

import (
	"crypto/ecdsa"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
)

// NeuronPublicKey is an immutable, validated ECDSA secp256k1 public key.
// The zero value is invalid; methods return errors on zero value.
//
// This type provides the unified interface for public key operations
// across Hedera, Ethereum, and libp2p ecosystems.
type NeuronPublicKey struct {
	key *secp256k1.PublicKey // unexported, never nil after valid construction
}

// IsZero returns true if this is a zero-value (invalid) public key.
func (k NeuronPublicKey) IsZero() bool {
	return k.key == nil
}

// checkValid returns an error if the key is zero-value.
func (k NeuronPublicKey) checkValid(op string) error {
	if k.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}
	return nil
}

// CompressedBytes returns the SEC1 compressed encoding (33 bytes).
// Format: prefix byte (0x02 or 0x03) + 32-byte X coordinate.
// Returns zero array if key is zero-value.
func (k NeuronPublicKey) CompressedBytes() [33]byte {
	var result [33]byte
	if k.IsZero() {
		return result
	}
	compressed := k.key.SerializeCompressed()
	copy(result[:], compressed)
	return result
}

// UncompressedBytes returns the SEC1 uncompressed encoding (65 bytes).
// Format: 0x04 prefix + 32-byte X coordinate + 32-byte Y coordinate.
// Returns zero array if key is zero-value.
func (k NeuronPublicKey) UncompressedBytes() [65]byte {
	var result [65]byte
	if k.IsZero() {
		return result
	}
	uncompressed := k.key.SerializeUncompressed()
	copy(result[:], uncompressed)
	return result
}

// Hex returns the compressed public key as a lowercase hex string with "0x" prefix.
// Returns empty string if key is zero-value.
func (k NeuronPublicKey) Hex() string {
	if k.IsZero() {
		return ""
	}
	return encodeHex(k.key.SerializeCompressed(), true)
}

// HexUncompressed returns the uncompressed public key as hex with "0x" prefix.
// Returns empty string if key is zero-value.
func (k NeuronPublicKey) HexUncompressed() string {
	if k.IsZero() {
		return ""
	}
	return encodeHex(k.key.SerializeUncompressed(), true)
}

// EVMAddress derives the Ethereum address from this public key.
// The address is derived by taking Keccak256(uncompressed_pubkey[1:])
// and using the last 20 bytes.
//
// WARNING: Returns ZeroEVMAddress (0x0000...0000) if key is zero-value.
// The zero address is the Ethereum burn address - funds sent there are lost forever.
// Use EVMAddressSafe() if you need error handling for zero keys.
func (k NeuronPublicKey) EVMAddress() EVMAddress {
	if k.IsZero() {
		return ZeroEVMAddress
	}

	// Get uncompressed bytes without the 0x04 prefix
	uncompressed := k.UncompressedBytes()
	// Skip the 0x04 prefix byte, use only the X and Y coordinates (64 bytes)
	return evmAddressFromPublicKeyBytes(uncompressed[1:])
}

// EVMAddressSafe derives the Ethereum address with explicit error handling.
// Returns an error if the public key is zero-value, preventing accidental
// derivation to the burn address (0x0000...0000).
//
// Use this method when you need to ensure the key is valid before deriving.
func (k NeuronPublicKey) EVMAddressSafe() (EVMAddress, error) {
	const op = "NeuronPublicKey.EVMAddressSafe"
	if err := k.checkValid(op); err != nil {
		return EVMAddress{}, err
	}
	return k.EVMAddress(), nil
}

// PeerID derives the libp2p peer ID from this public key.
// Returns error if key is zero-value.
func (k NeuronPublicKey) PeerID() (PeerID, error) {
	const op = "NeuronPublicKey.PeerID"

	if err := k.checkValid(op); err != nil {
		return PeerID{}, err
	}

	return peerIDFromPublicKeyBytes(k.key.SerializeCompressed())
}

// ToECDSA returns the key as a standard library *ecdsa.PublicKey.
// Returns nil if key is zero-value.
// The returned key uses the secp256k1 curve from go-ethereum.
func (k NeuronPublicKey) ToECDSA() *ecdsa.PublicKey {
	if k.IsZero() {
		return nil
	}

	// Convert to go-ethereum's ECDSA public key
	uncompressed := k.UncompressedBytes()
	ecdsaPubKey, err := crypto.UnmarshalPubkey(uncompressed[:])
	if err != nil {
		// This should never happen with a valid key
		return nil
	}
	return ecdsaPubKey
}

// ToHederaPublicKey returns the key as a Hedera SDK PublicKey.
// Returns zero-value Hedera key if this key is zero-value or conversion fails.
//
// WARNING: Silently returns empty key on failure. Use ToHederaPublicKeySafe()
// if you need explicit error handling.
func (k NeuronPublicKey) ToHederaPublicKey() hiero.PublicKey {
	if k.IsZero() {
		return hiero.PublicKey{}
	}

	// Get compressed bytes and create Hedera public key
	compressed := k.CompressedBytes()
	hederaPubKey, err := hiero.PublicKeyFromBytesECDSA(compressed[:])
	if err != nil {
		// This should never happen with a valid key
		return hiero.PublicKey{}
	}
	return hederaPubKey
}

// ToHederaPublicKeySafe converts to Hedera SDK PublicKey with error handling.
// Returns an error if the key is zero-value or conversion fails.
func (k NeuronPublicKey) ToHederaPublicKeySafe() (hiero.PublicKey, error) {
	const op = "NeuronPublicKey.ToHederaPublicKeySafe"

	if err := k.checkValid(op); err != nil {
		return hiero.PublicKey{}, err
	}

	compressed := k.CompressedBytes()
	hederaPubKey, err := hiero.PublicKeyFromBytesECDSA(compressed[:])
	if err != nil {
		return hiero.PublicKey{}, wrapError(op, ErrKindDerivation, "failed to create Hedera public key", err)
	}
	return hederaPubKey, nil
}

// MatchesEVMAddress checks if this public key derives to the given EVM address.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPublicKey) MatchesEVMAddress(addr EVMAddress) bool {
	if k.IsZero() {
		return false
	}
	derived := k.EVMAddress()
	return derived.Equal(addr)
}

// MatchesPeerID checks if this public key derives to the given peer ID.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPublicKey) MatchesPeerID(pid PeerID) bool {
	if k.IsZero() {
		return false
	}
	derived, err := k.PeerID()
	if err != nil {
		return false
	}
	return derived.Equal(pid)
}

// Verify verifies a signature over a message using this public key.
// The message is hashed with Keccak256 before verification.
// Returns false if key is zero-value or signature is invalid.
func (k NeuronPublicKey) Verify(msg []byte, sig Signature) bool {
	if k.IsZero() || sig.IsZero() {
		return false
	}

	// Hash the message
	hash := crypto.Keccak256(msg)
	return k.VerifyDigest(*(*[32]byte)(hash), sig)
}

// VerifyDigest verifies a signature over a pre-hashed 32-byte digest.
// Returns false if key is zero-value or signature is invalid.
func (k NeuronPublicKey) VerifyDigest(digest [32]byte, sig Signature) bool {
	if k.IsZero() || sig.IsZero() {
		return false
	}

	// Get R and S from signature
	r, s := sig.RS()
	if r == nil || s == nil {
		return false
	}

	// Convert to ecdsa.PublicKey for verification
	ecdsaPubKey := k.ToECDSA()
	if ecdsaPubKey == nil {
		return false
	}

	return ecdsa.Verify(ecdsaPubKey, digest[:], r, s)
}

// Equal returns true if this public key equals another.
// Uses constant-time comparison to prevent timing attacks.
func (k NeuronPublicKey) Equal(other NeuronPublicKey) bool {
	if k.IsZero() || other.IsZero() {
		return k.IsZero() && other.IsZero()
	}
	return constantTimeEqual(k.key.SerializeCompressed(), other.key.SerializeCompressed())
}

// newPublicKeyFromSecp256k1 creates a NeuronPublicKey from a secp256k1.PublicKey.
// This is an internal constructor.
func newPublicKeyFromSecp256k1(key *secp256k1.PublicKey) NeuronPublicKey {
	if key == nil {
		return NeuronPublicKey{}
	}
	return NeuronPublicKey{key: key}
}

// newPublicKeyFromECDSA creates a NeuronPublicKey from an ecdsa.PublicKey.
// This is an internal constructor.
func newPublicKeyFromECDSA(key *ecdsa.PublicKey) NeuronPublicKey {
	if key == nil {
		return NeuronPublicKey{}
	}

	// Marshal to bytes and parse as secp256k1
	pubBytes := crypto.FromECDSAPub(key)
	if pubBytes == nil {
		return NeuronPublicKey{}
	}

	secpKey, err := secp256k1.ParsePubKey(pubBytes)
	if err != nil {
		return NeuronPublicKey{}
	}

	return NeuronPublicKey{key: secpKey}
}
