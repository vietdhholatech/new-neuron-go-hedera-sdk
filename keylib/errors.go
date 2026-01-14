// Package keylib provides type-safe cryptographic key management for the Neuron SDK.
package keylib

import (
	"errors"
	"fmt"
)

// ErrorKind categorizes key errors for programmatic handling.
type ErrorKind int

const (
	// ErrKindInvalidFormat indicates wrong format or encoding.
	ErrKindInvalidFormat ErrorKind = iota + 1
	// ErrKindInvalidLength indicates wrong byte or character length.
	ErrKindInvalidLength
	// ErrKindInvalidHex indicates invalid hexadecimal characters.
	ErrKindInvalidHex
	// ErrKindInvalidKey indicates key validation failed (e.g., not on curve).
	ErrKindInvalidKey
	// ErrKindZeroValue indicates operation on zero-value type.
	ErrKindZeroValue
	// ErrKindKeyMismatch indicates keys don't match.
	ErrKindKeyMismatch
	// ErrKindEncryption indicates encryption or decryption failure.
	ErrKindEncryption
	// ErrKindMnemonic indicates invalid mnemonic.
	ErrKindMnemonic
	// ErrKindDerivation indicates key derivation failure.
	ErrKindDerivation
	// ErrKindUnsupportedKeyType indicates Ed25519 or other unsupported type.
	ErrKindUnsupportedKeyType
)

// String returns a human-readable name for the error kind.
func (k ErrorKind) String() string {
	switch k {
	case ErrKindInvalidFormat:
		return "InvalidFormat"
	case ErrKindInvalidLength:
		return "InvalidLength"
	case ErrKindInvalidHex:
		return "InvalidHex"
	case ErrKindInvalidKey:
		return "InvalidKey"
	case ErrKindZeroValue:
		return "ZeroValue"
	case ErrKindKeyMismatch:
		return "KeyMismatch"
	case ErrKindEncryption:
		return "Encryption"
	case ErrKindMnemonic:
		return "Mnemonic"
	case ErrKindDerivation:
		return "Derivation"
	case ErrKindUnsupportedKeyType:
		return "UnsupportedKeyType"
	default:
		return "Unknown"
	}
}

// KeyError is the base error type for all key-related errors.
// It provides rich context about what went wrong and why.
type KeyError struct {
	// Op is the operation that failed (e.g., "ParsePrivateKeyHex", "DeriveEVMAddress").
	Op string
	// Kind categorizes the error for programmatic handling.
	Kind ErrorKind
	// Details provides a human-readable explanation of what went wrong.
	Details string
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface with a detailed message.
func (e *KeyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("keylib.%s: %s: %s: %v", e.Op, e.Kind, e.Details, e.Err)
	}
	if e.Op != "" {
		return fmt.Sprintf("keylib.%s: %s: %s", e.Op, e.Kind, e.Details)
	}
	return fmt.Sprintf("keylib: %s: %s", e.Kind, e.Details)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *KeyError) Unwrap() error {
	return e.Err
}

// Is reports whether this error matches the target.
func (e *KeyError) Is(target error) bool {
	var ke *KeyError
	if errors.As(target, &ke) {
		// Match by Kind if both have the same kind
		if ke.Kind != 0 && e.Kind == ke.Kind {
			return true
		}
	}
	return false
}

// newKeyError creates a new KeyError with the given parameters.
func newKeyError(op string, kind ErrorKind, details string, err error) *KeyError {
	return &KeyError{
		Op:      op,
		Kind:    kind,
		Details: details,
		Err:     err,
	}
}

// Sentinel errors for common checks.
// These can be used with errors.Is() for error matching.
var (
	// ErrNilPrivateKey indicates a nil or zero-value private key.
	ErrNilPrivateKey = &KeyError{
		Kind:    ErrKindZeroValue,
		Details: "nil or zero-value private key",
	}

	// ErrNilPublicKey indicates a nil or zero-value public key.
	ErrNilPublicKey = &KeyError{
		Kind:    ErrKindZeroValue,
		Details: "nil or zero-value public key",
	}

	// ErrNilSignature indicates a nil or zero-value signature.
	ErrNilSignature = &KeyError{
		Kind:    ErrKindZeroValue,
		Details: "nil or zero-value signature",
	}

	// ErrEd25519NotAllowed indicates an attempt to use an Ed25519 key.
	ErrEd25519NotAllowed = &KeyError{
		Kind:    ErrKindUnsupportedKeyType,
		Details: "Ed25519 keys are not supported; only secp256k1 ECDSA keys are allowed",
	}

	// ErrInvalidSignature indicates an invalid signature format.
	ErrInvalidSignature = &KeyError{
		Kind:    ErrKindInvalidFormat,
		Details: "invalid signature format",
	}

	// ErrWrongPassword indicates decryption failed due to wrong password.
	ErrWrongPassword = &KeyError{
		Kind:    ErrKindEncryption,
		Details: "wrong password or corrupted data",
	}
)

// Helper functions for creating specific error types.

// errInvalidHex creates an error for invalid hex input.
func errInvalidHex(op string, char rune, position int) *KeyError {
	return newKeyError(op, ErrKindInvalidHex,
		fmt.Sprintf("invalid hex character '%c' at position %d", char, position), nil)
}

// errInvalidLength creates an error for wrong input length.
func errInvalidLength(op string, expected, got int, context string) *KeyError {
	return newKeyError(op, ErrKindInvalidLength,
		fmt.Sprintf("expected %d %s, got %d", expected, context, got), nil)
}

// errInvalidKey creates an error for invalid key data.
func errInvalidKey(op string, reason string) *KeyError {
	return newKeyError(op, ErrKindInvalidKey, reason, nil)
}

// errInvalidFormat creates an error for wrong format.
func errInvalidFormat(op string, reason string) *KeyError {
	return newKeyError(op, ErrKindInvalidFormat, reason, nil)
}

// errZeroValue creates an error for zero-value operations.
func errZeroValue(op string, typeName string) *KeyError {
	return newKeyError(op, ErrKindZeroValue,
		fmt.Sprintf("cannot perform operation on zero-value %s", typeName), nil)
}

// errUnsupportedKeyType creates an error for unsupported key types.
func errUnsupportedKeyType(op string, keyType string) *KeyError {
	return newKeyError(op, ErrKindUnsupportedKeyType,
		fmt.Sprintf("unsupported key type: %s; only secp256k1 ECDSA is supported", keyType), nil)
}

// errEncryption creates an error for encryption/decryption failures.
func errEncryption(op string, reason string, err error) *KeyError {
	return newKeyError(op, ErrKindEncryption, reason, err)
}

// errMnemonic creates an error for mnemonic-related failures.
func errMnemonic(op string, reason string, err error) *KeyError {
	return newKeyError(op, ErrKindMnemonic, reason, err)
}

// errDerivation creates an error for key derivation failures.
func errDerivation(op string, reason string, err error) *KeyError {
	return newKeyError(op, ErrKindDerivation, reason, err)
}

// wrapError wraps an existing error with additional context.
func wrapError(op string, kind ErrorKind, details string, err error) *KeyError {
	return newKeyError(op, kind, details, err)
}
