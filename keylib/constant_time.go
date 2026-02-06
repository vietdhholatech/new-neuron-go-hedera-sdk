package keylib

import (
	"crypto/subtle"
)

// constantTimeEqual compares two byte slices in constant time.
// This prevents timing attacks when comparing secret values like keys.
// Returns true if the slices are equal, false otherwise.
//
// Important: If the slices have different lengths, this function
// returns false in constant time relative to the shorter slice.
func constantTimeEqual(a, b []byte) bool {
	// subtle.ConstantTimeCompare returns 1 if equal, 0 otherwise
	// It also handles different length slices by returning 0
	return subtle.ConstantTimeCompare(a, b) == 1
}

// constantTimeEqualStrings compares two strings in constant time.
// This is a convenience wrapper around constantTimeEqual for strings.
func constantTimeEqualStrings(a, b string) bool {
	return constantTimeEqual([]byte(a), []byte(b))
}

// secureZero overwrites a byte slice with zeros.
// This should be used to clear sensitive data like private keys from memory.
//
// Note: Due to Go's garbage collector and potential compiler optimizations,
// this may not guarantee complete removal from memory, but it represents
// a best-effort approach to minimize exposure window.
//
// Uses the clear() builtin (Go 1.21+) which is compiler-optimized and
// may be more resistant to dead-store elimination than manual loops.
func secureZero(b []byte) {
	clear(b)
}

// secureZeroArray32 overwrites a 32-byte array with zeros.
// Specialized version for private key sized arrays.
// Uses the clear() builtin (Go 1.21+) for compiler-optimized zeroing.
func secureZeroArray32(b *[32]byte) {
	clear(b[:])
}
