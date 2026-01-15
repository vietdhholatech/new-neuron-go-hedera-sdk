package keylib

import (
	"crypto/ecdsa"
	"crypto/rand"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
)

// GeneratePrivateKey generates a new random ECDSA secp256k1 private key.
// Uses crypto/rand as the entropy source.
func GeneratePrivateKey() (NeuronPrivateKey, error) {
	const op = "GeneratePrivateKey"

	key, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
	if err != nil {
		return NeuronPrivateKey{}, wrapError(op, ErrKindDerivation, "failed to generate random key", err)
	}

	return newPrivateKeyFromECDSA(key), nil
}

// ParsePrivateKeyHex parses a private key from a hex string.
// Accepts with or without "0x" prefix (64 or 66 characters).
//
// Validates:
//   - Hex format and characters
//   - Exactly 32 bytes (64 hex chars)
//   - Valid secp256k1 scalar (> 0, < curve order N)
func ParsePrivateKeyHex(s string) (NeuronPrivateKey, error) {
	const op = "ParsePrivateKeyHex"

	// Normalize and decode hex
	normalized := normalizeHex(s)

	// Validate length before decoding
	if len(normalized) != 64 {
		return NeuronPrivateKey{}, errInvalidLength(op, 64, len(normalized), "hex characters for private key")
	}

	bytes, err := decodeHexStrict(op, normalized)
	if err != nil {
		return NeuronPrivateKey{}, err
	}

	return PrivateKeyFromBytes(*(*[32]byte)(bytes))
}

// PrivateKeyFromBytes creates a NeuronPrivateKey from a 32-byte array.
// Validates that the bytes represent a valid secp256k1 scalar.
func PrivateKeyFromBytes(b [32]byte) (NeuronPrivateKey, error) {
	const op = "PrivateKeyFromBytes"

	// Validate the scalar
	if err := validatePrivateKeyBytes(op, b[:]); err != nil {
		return NeuronPrivateKey{}, err
	}

	// Create secp256k1 private key
	key := secp256k1.PrivKeyFromBytes(b[:])

	return NeuronPrivateKey{key: key}, nil
}

// PrivateKeyFromHedera creates a NeuronPrivateKey from a Hedera SDK PrivateKey.
// Only ECDSA secp256k1 keys are supported; Ed25519 keys are rejected.
//
// Detection uses multiple methods:
//  1. DER encoding pattern detection for Ed25519 OID
//  2. Raw bytes length check (Ed25519 = 64 bytes, ECDSA = 32 bytes)
//  3. secp256k1 scalar validation as final safety net
func PrivateKeyFromHedera(hederaKey hiero.PrivateKey) (NeuronPrivateKey, error) {
	const op = "PrivateKeyFromHedera"

	// Check if key is zero/empty first
	if hederaKey.String() == "" {
		return NeuronPrivateKey{}, errZeroValue(op, "hiero.PrivateKey")
	}

	// PRIMARY CHECK: Use Hedera SDK's type detection via DER encoding
	// Ed25519 DER-encoded keys contain specific OID patterns
	// Ed25519 OID: 1.3.101.112 encoded as "06 03 2b 65 70" in DER
	derStr := hederaKey.String()
	if strings.Contains(derStr, "302e") ||
		strings.Contains(derStr, "2b6570") ||
		strings.HasPrefix(derStr, "302a") {
		return NeuronPrivateKey{}, errUnsupportedKeyType(op, "Ed25519 (detected via DER encoding)")
	}

	// SECONDARY CHECK: Raw bytes length
	// Ed25519 private keys have 64-byte raw representation (seed + public)
	// ECDSA secp256k1 private keys have 32-byte raw representation
	rawBytes := hederaKey.BytesRaw()
	if len(rawBytes) == 64 {
		return NeuronPrivateKey{}, errUnsupportedKeyType(op, "Ed25519 (64-byte raw key)")
	}

	if len(rawBytes) != 32 {
		return NeuronPrivateKey{}, errInvalidLength(op, 32, len(rawBytes), "bytes for ECDSA private key")
	}

	// TERTIARY CHECK: Validate it's a valid secp256k1 scalar
	if err := validatePrivateKeyBytes(op, rawBytes); err != nil {
		return NeuronPrivateKey{}, err
	}

	return PrivateKeyFromBytes(*(*[32]byte)(rawBytes))
}

// ParsePublicKeyHex parses a public key from a hex string.
// Accepts both compressed (66 chars) and uncompressed (130 chars) formats,
// with or without "0x" prefix.
//
// Validates:
//   - Hex format and characters
//   - Valid SEC1 encoding (proper prefix bytes)
//   - Point is on the secp256k1 curve
func ParsePublicKeyHex(s string) (NeuronPublicKey, error) {
	const op = "ParsePublicKeyHex"

	normalized := normalizeHex(s)

	// Validate length
	if len(normalized) != 66 && len(normalized) != 130 {
		return NeuronPublicKey{}, errInvalidLength(op, 66, len(normalized),
			"hex characters for compressed public key (or 130 for uncompressed)")
	}

	bytes, err := decodeHexStrict(op, normalized)
	if err != nil {
		return NeuronPublicKey{}, err
	}

	return PublicKeyFromBytes(bytes)
}

// PublicKeyFromBytes creates a NeuronPublicKey from SEC1-encoded bytes.
// Accepts both compressed (33 bytes) and uncompressed (65 bytes) formats.
func PublicKeyFromBytes(b []byte) (NeuronPublicKey, error) {
	const op = "PublicKeyFromBytes"

	// Validate the public key bytes
	if err := validatePublicKeyBytes(op, b); err != nil {
		return NeuronPublicKey{}, err
	}

	// Parse the public key
	key, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return NeuronPublicKey{}, errInvalidKey(op, "point not on secp256k1 curve")
	}

	return NeuronPublicKey{key: key}, nil
}

// PublicKeyFromHedera creates a NeuronPublicKey from a Hedera SDK PublicKey.
// Only ECDSA secp256k1 keys are supported; Ed25519 keys are rejected.
func PublicKeyFromHedera(hederaKey hiero.PublicKey) (NeuronPublicKey, error) {
	const op = "PublicKeyFromHedera"

	// Get raw bytes from Hedera key
	rawBytes := hederaKey.BytesRaw()

	// ECDSA compressed public key is 33 bytes
	// Ed25519 public key is 32 bytes
	if len(rawBytes) == 32 {
		return NeuronPublicKey{}, errUnsupportedKeyType(op, "Ed25519")
	}

	if len(rawBytes) != 33 && len(rawBytes) != 65 {
		return NeuronPublicKey{}, errInvalidLength(op, 33, len(rawBytes),
			"bytes for ECDSA public key (compressed)")
	}

	return PublicKeyFromBytes(rawBytes)
}

// PublicKeyFromPrivateKey derives a NeuronPublicKey from a NeuronPrivateKey.
// This is a convenience function equivalent to privKey.PublicKey().
func PublicKeyFromPrivateKey(privKey NeuronPrivateKey) NeuronPublicKey {
	return privKey.PublicKey()
}

// MustParsePrivateKeyHex parses a private key hex or panics.
// Only use this for hardcoded test values.
func MustParsePrivateKeyHex(s string) NeuronPrivateKey {
	key, err := ParsePrivateKeyHex(s)
	if err != nil {
		panic("keylib.MustParsePrivateKeyHex: " + err.Error())
	}
	return key
}

// MustParsePublicKeyHex parses a public key hex or panics.
// Only use this for hardcoded test values.
func MustParsePublicKeyHex(s string) NeuronPublicKey {
	key, err := ParsePublicKeyHex(s)
	if err != nil {
		panic("keylib.MustParsePublicKeyHex: " + err.Error())
	}
	return key
}

// MustParseEVMAddress parses an EVM address or panics.
// Only use this for hardcoded test values.
func MustParseEVMAddress(s string) EVMAddress {
	addr, err := ParseEVMAddress(s)
	if err != nil {
		panic("keylib.MustParseEVMAddress: " + err.Error())
	}
	return addr
}

// MustParsePeerID parses a peer ID or panics.
// Only use this for hardcoded test values.
func MustParsePeerID(s string) PeerID {
	pid, err := ParsePeerID(s)
	if err != nil {
		panic("keylib.MustParsePeerID: " + err.Error())
	}
	return pid
}

// IsEd25519Key detects if a Hedera private key is Ed25519 type.
// Uses multiple detection methods for reliability:
//  1. DER encoding pattern detection for Ed25519 OID
//  2. Raw bytes length check (Ed25519 = 64 bytes)
func IsEd25519Key(hederaKey hiero.PrivateKey) bool {
	// Method 1: Check DER encoding for Ed25519 OID
	// Ed25519 OID: 1.3.101.112 encoded as "2b6570" in hex
	derStr := hederaKey.String()
	if strings.Contains(derStr, "302e") ||
		strings.Contains(derStr, "2b6570") ||
		strings.HasPrefix(derStr, "302a") {
		return true
	}

	// Method 2: Ed25519 raw bytes are 64 bytes (seed + public key)
	// ECDSA raw bytes are 32 bytes (scalar only)
	if len(hederaKey.BytesRaw()) == 64 {
		return true
	}

	return false
}

// IsEd25519PublicKey detects if a Hedera public key is Ed25519 type.
// Ed25519 public keys are 32 bytes raw, while ECDSA are 33 (compressed) or 65 (uncompressed).
func IsEd25519PublicKey(hederaKey hiero.PublicKey) bool {
	// Ed25519 public keys are 32 bytes raw
	// ECDSA public keys are 33 (compressed) or 65 (uncompressed) bytes
	rawLen := len(hederaKey.BytesRaw())
	return rawLen == 32
}
