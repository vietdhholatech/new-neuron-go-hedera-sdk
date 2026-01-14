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
func secureZero(b []byte) {
	for i := range b {
		b[i] = 0
	}
	// Use KeepAlive pattern indirectly by touching the slice
	// This helps prevent the compiler from optimizing away the zeroing
	_ = b[0:0]
}

// secureZeroArray32 overwrites a 32-byte array with zeros.
// Specialized version for private key sized arrays.
func secureZeroArray32(b *[32]byte) {
	for i := range b {
		b[i] = 0
	}
}

// constantTimeSelect returns a if selector is 1, b if selector is 0.
// The selection is done in constant time.
func constantTimeSelect(selector int, a, b []byte) []byte {
	if len(a) != len(b) {
		// Different lengths - can't do constant time selection
		// Return based on selector (not constant time, but we have no choice)
		if selector == 1 {
			return a
		}
		return b
	}

	result := make([]byte, len(a))
	for i := range result {
		result[i] = byte(subtle.ConstantTimeSelect(selector, int(a[i]), int(b[i])))
	}
	return result
}
