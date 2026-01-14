package keylib

import (
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signature represents an ECDSA signature with recovery ID.
// Format: [R (32 bytes) || S (32 bytes) || V (1 byte)] = 65 bytes total.
//
// The V value (recovery ID) allows public key recovery from the signature.
// V can be 0, 1 (raw) or 27, 28 (Ethereum legacy format).
type Signature struct {
	data [65]byte
}

// ZeroSignature is the zero-value signature.
var ZeroSignature = Signature{}

// ParseSignature parses a signature from a hex string.
// Accepts 65-byte signatures with or without "0x" prefix (130 or 132 chars).
func ParseSignature(s string) (Signature, error) {
	const op = "ParseSignature"

	bytes, err := decodeHexStrict(op, s)
	if err != nil {
		return Signature{}, err
	}

	if len(bytes) != 65 {
		return Signature{}, errInvalidLength(op, 65, len(bytes), "bytes for signature")
	}

	var sig Signature
	copy(sig.data[:], bytes)

	// Normalize V value if needed (convert 27/28 to 0/1)
	if sig.data[64] >= 27 {
		sig.data[64] -= 27
	}

	// Validate V is 0 or 1
	if sig.data[64] > 1 {
		return Signature{}, errInvalidFormat(op, "invalid recovery ID (V); must be 0, 1, 27, or 28")
	}

	return sig, nil
}

// SignatureFromBytes creates a signature from raw bytes.
// Returns error if bytes length is not 65.
func SignatureFromBytes(b []byte) (Signature, error) {
	const op = "SignatureFromBytes"

	if len(b) != 65 {
		return Signature{}, errInvalidLength(op, 65, len(b), "bytes for signature")
	}

	var sig Signature
	copy(sig.data[:], b)

	// Normalize V value
	if sig.data[64] >= 27 {
		sig.data[64] -= 27
	}

	if sig.data[64] > 1 {
		return Signature{}, errInvalidFormat(op, "invalid recovery ID (V); must be 0, 1, 27, or 28")
	}

	return sig, nil
}

// signatureFromRSV creates a signature from R, S, and V components.
func signatureFromRSV(r, s *big.Int, v byte) Signature {
	var sig Signature

	// Pad R and S to 32 bytes
	rBytes := padLeftZeros(r.Bytes(), 32)
	sBytes := padLeftZeros(s.Bytes(), 32)

	copy(sig.data[0:32], rBytes)
	copy(sig.data[32:64], sBytes)

	// Normalize V
	if v >= 27 {
		v -= 27
	}
	sig.data[64] = v

	return sig
}

// IsZero returns true if this is a zero-value signature.
func (s Signature) IsZero() bool {
	return s == ZeroSignature
}

// Bytes returns the raw 65-byte signature.
func (s Signature) Bytes() []byte {
	result := make([]byte, 65)
	copy(result, s.data[:])
	return result
}

// Hex returns the signature as a lowercase hex string with "0x" prefix.
func (s Signature) Hex() string {
	return encodeHex(s.data[:], true)
}

// RS returns the R and S components as big.Int.
// Returns nil values if signature is zero-value.
func (s Signature) RS() (r, sx *big.Int) {
	if s.IsZero() {
		return nil, nil
	}
	r = new(big.Int).SetBytes(s.data[0:32])
	sx = new(big.Int).SetBytes(s.data[32:64])
	return r, sx
}

// RSV returns the R, S, and V components.
// R and S are returned as big.Int, V is the raw recovery ID (0 or 1).
func (s Signature) RSV() (r, sx *big.Int, v byte) {
	r, sx = s.RS()
	v = s.data[64]
	return r, sx, v
}

// V returns the recovery ID (0 or 1).
func (s Signature) V() byte {
	return s.data[64]
}

// VEthereum returns the recovery ID in Ethereum legacy format (27 or 28).
func (s Signature) VEthereum() byte {
	return s.data[64] + 27
}

// EthereumBytes returns the signature in Ethereum format [R || S || V+27].
func (s Signature) EthereumBytes() []byte {
	result := make([]byte, 65)
	copy(result, s.data[:])
	result[64] += 27
	return result
}

// Equal returns true if this signature equals another.
// Uses constant-time comparison.
func (s Signature) Equal(other Signature) bool {
	return constantTimeEqual(s.data[:], other.data[:])
}

// RecoverPublicKey recovers the public key from a signature and the original message.
// The message is hashed with Keccak256 before recovery.
func RecoverPublicKey(msg []byte, sig Signature) (NeuronPublicKey, error) {
	const op = "RecoverPublicKey"

	if sig.IsZero() {
		return NeuronPublicKey{}, errZeroValue(op, "Signature")
	}

	hash := crypto.Keccak256(msg)
	return RecoverPublicKeyFromDigest(*(*[32]byte)(hash), sig)
}

// RecoverPublicKeyFromDigest recovers the public key from a signature and a pre-hashed digest.
func RecoverPublicKeyFromDigest(digest [32]byte, sig Signature) (NeuronPublicKey, error) {
	const op = "RecoverPublicKeyFromDigest"

	if sig.IsZero() {
		return NeuronPublicKey{}, errZeroValue(op, "Signature")
	}

	// Use go-ethereum's ecrecover
	// It expects the signature in [R || S || V] format with V as 0 or 1
	pubKeyBytes, err := crypto.Ecrecover(digest[:], sig.data[:])
	if err != nil {
		return NeuronPublicKey{}, wrapError(op, ErrKindInvalidFormat, "failed to recover public key", err)
	}

	// Parse the recovered public key (uncompressed format, 65 bytes)
	secpKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return NeuronPublicKey{}, wrapError(op, ErrKindInvalidKey, "failed to parse recovered public key", err)
	}

	return NeuronPublicKey{key: secpKey}, nil
}
