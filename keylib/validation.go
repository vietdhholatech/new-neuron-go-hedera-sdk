package keylib

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// secp256k1 curve order N.
// Any valid private key scalar must be in range [1, N-1].
var secp256k1N = secp256k1.S256().N

// normalizeHex strips the "0x" or "0X" prefix and converts to lowercase.
// Returns the normalized hex string.
// Uses strings.CutPrefix (Go 1.20+) for explicit prefix handling.
func normalizeHex(s string) string {
	if after, found := strings.CutPrefix(s, "0x"); found {
		return strings.ToLower(after)
	}
	if after, found := strings.CutPrefix(s, "0X"); found {
		return strings.ToLower(after)
	}
	return strings.ToLower(s)
}

// isValidHexChar checks if a character is a valid hexadecimal digit.
func isValidHexChar(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

// validateHexString validates that a string contains only valid hex characters.
// Returns the position and character of the first invalid character, or -1 if valid.
func validateHexString(s string) (invalidPos int, invalidChar rune) {
	for i, c := range s {
		if !isValidHexChar(c) {
			return i, c
		}
	}
	return -1, 0
}

// decodeHexStrict decodes a hex string with strict validation.
// It normalizes the input (strips 0x prefix) and validates all characters.
func decodeHexStrict(op string, s string) ([]byte, error) {
	normalized := normalizeHex(s)

	// Check for empty string
	if len(normalized) == 0 {
		return nil, errInvalidLength(op, 1, 0, "hex characters (minimum)")
	}

	// Check for odd length
	if len(normalized)%2 != 0 {
		return nil, errInvalidFormat(op, "hex string must have even length")
	}

	// Validate all characters are valid hex
	if pos, char := validateHexString(normalized); pos >= 0 {
		return nil, errInvalidHex(op, char, pos)
	}

	// Decode
	bytes, err := hex.DecodeString(normalized)
	if err != nil {
		// This shouldn't happen after our validation, but handle it anyway
		return nil, wrapError(op, ErrKindInvalidHex, "hex decode failed", err)
	}

	return bytes, nil
}

// isValidSecp256k1Scalar checks if bytes represent a valid secp256k1 private key scalar.
// A valid scalar must be:
// - Exactly 32 bytes
// - Greater than 0
// - Less than the curve order N
func isValidSecp256k1Scalar(b []byte) bool {
	if len(b) != 32 {
		return false
	}

	// Check if all zeros
	allZero := true
	for _, v := range b {
		if v != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return false
	}

	// Check if less than curve order N
	scalar := new(big.Int).SetBytes(b)
	if scalar.Cmp(secp256k1N) >= 0 {
		return false
	}

	return true
}

// validatePrivateKeyBytes validates that bytes are a valid secp256k1 private key.
func validatePrivateKeyBytes(op string, b []byte) error {
	if len(b) != 32 {
		return errInvalidLength(op, 32, len(b), "bytes for private key")
	}

	// Check if all zeros
	allZero := true
	for _, v := range b {
		if v != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return errInvalidKey(op, "zero scalar is not valid on secp256k1")
	}

	// Check if less than curve order N
	scalar := new(big.Int).SetBytes(b)
	if scalar.Cmp(secp256k1N) >= 0 {
		return errInvalidKey(op, "scalar exceeds curve order")
	}

	return nil
}

// validatePublicKeyBytes validates SEC1-encoded public key bytes.
// Accepts both compressed (33 bytes) and uncompressed (65 bytes) formats.
func validatePublicKeyBytes(op string, b []byte) error {
	switch len(b) {
	case 33:
		// Compressed format: prefix byte (02 or 03) + 32-byte X coordinate
		if b[0] != 0x02 && b[0] != 0x03 {
			return errInvalidFormat(op, "invalid SEC1 compressed prefix; expected 0x02 or 0x03")
		}
	case 65:
		// Uncompressed format: 0x04 prefix + 32-byte X + 32-byte Y
		if b[0] != 0x04 {
			return errInvalidFormat(op, "invalid SEC1 uncompressed prefix; expected 0x04")
		}
	default:
		return errInvalidLength(op, 33, len(b), "bytes for compressed public key (or 65 for uncompressed)")
	}

	// Parse the public key to verify it's on the curve
	_, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return errInvalidKey(op, "point not on secp256k1 curve")
	}

	return nil
}

// isValidEVMAddressFormat checks if a string is a valid EVM address format.
// Valid formats: 40 hex chars, or "0x" + 40 hex chars.
func isValidEVMAddressFormat(s string) bool {
	normalized := normalizeHex(s)
	if len(normalized) != 40 {
		return false
	}
	pos, _ := validateHexString(normalized)
	return pos < 0
}

// padLeftZeros pads a byte slice with leading zeros to reach the target length.
// If the slice is already >= target length, it returns the original slice.
func padLeftZeros(b []byte, targetLen int) []byte {
	if len(b) >= targetLen {
		return b
	}
	padded := make([]byte, targetLen)
	copy(padded[targetLen-len(b):], b)
	return padded
}

// encodeHex encodes bytes to a lowercase hex string with optional 0x prefix.
func encodeHex(b []byte, with0xPrefix bool) string {
	s := hex.EncodeToString(b)
	if with0xPrefix {
		return "0x" + s
	}
	return s
}
