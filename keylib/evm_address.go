package keylib

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EVMAddress represents a 20-byte Ethereum Virtual Machine address.
// The zero value represents the zero address (0x0000...0000).
type EVMAddress struct {
	addr [20]byte
}

// ZeroEVMAddress is the zero address (0x0000...0000).
var ZeroEVMAddress = EVMAddress{}

// ParseEVMAddress parses an EVM address from a hex string.
// Accepts formats: "0x..." (42 chars) or "..." (40 chars).
// Validates hex characters and length.
func ParseEVMAddress(s string) (EVMAddress, error) {
	const op = "ParseEVMAddress"

	normalized := normalizeHex(s)

	// Validate length
	if len(normalized) != 40 {
		return EVMAddress{}, errInvalidLength(op, 40, len(normalized), "hex characters for EVM address")
	}

	// Validate hex characters
	if pos, char := validateHexString(normalized); pos >= 0 {
		return EVMAddress{}, errInvalidHex(op, char, pos)
	}

	// Decode
	bytes, err := decodeHexStrict(op, normalized)
	if err != nil {
		return EVMAddress{}, err
	}

	var addr EVMAddress
	copy(addr.addr[:], bytes)
	return addr, nil
}

// evmAddressFromBytes creates an EVMAddress from a 20-byte slice.
// This is an internal constructor; use ParseEVMAddress for external input.
func evmAddressFromBytes(b []byte) EVMAddress {
	var addr EVMAddress
	if len(b) >= 20 {
		copy(addr.addr[:], b[:20])
	} else {
		copy(addr.addr[20-len(b):], b)
	}
	return addr
}

// evmAddressFromPublicKeyBytes derives an EVM address from uncompressed public key bytes.
// The input should be 64 bytes (X || Y coordinates, without the 0x04 prefix).
func evmAddressFromPublicKeyBytes(pubKeyBytes []byte) EVMAddress {
	// Keccak256 hash of the public key bytes
	hash := crypto.Keccak256(pubKeyBytes)
	// Take the last 20 bytes
	return evmAddressFromBytes(hash[12:])
}

// IsZero returns true if this is the zero address.
func (a EVMAddress) IsZero() bool {
	return a == ZeroEVMAddress
}

// Bytes returns the raw 20-byte address.
func (a EVMAddress) Bytes() [20]byte {
	return a.addr
}

// Hex returns the address as a lowercase hex string with "0x" prefix.
// Example: "0xe364f2f1e5f4f03d1df682322500b9c68c997ec3"
func (a EVMAddress) Hex() string {
	return encodeHex(a.addr[:], true)
}

// ChecksumHex returns the address with EIP-55 mixed-case checksum encoding.
// Example: "0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"
func (a EVMAddress) ChecksumHex() string {
	// Use go-ethereum's implementation for correctness
	return common.BytesToAddress(a.addr[:]).Hex()
}

// String returns the checksummed hex representation.
// This is an alias for ChecksumHex for convenience.
func (a EVMAddress) String() string {
	return a.ChecksumHex()
}

// Equal returns true if this address equals another.
// Uses constant-time comparison to prevent timing attacks.
func (a EVMAddress) Equal(other EVMAddress) bool {
	return constantTimeEqual(a.addr[:], other.addr[:])
}

// ToCommon converts to go-ethereum's common.Address type.
// Useful for interoperability with go-ethereum libraries.
func (a EVMAddress) ToCommon() common.Address {
	return common.BytesToAddress(a.addr[:])
}

// EVMAddressFromCommon creates an EVMAddress from go-ethereum's common.Address.
func EVMAddressFromCommon(addr common.Address) EVMAddress {
	var a EVMAddress
	copy(a.addr[:], addr.Bytes())
	return a
}
