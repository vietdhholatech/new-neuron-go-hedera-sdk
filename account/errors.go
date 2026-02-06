// Package account provides the NeuronAccount abstraction for agent identity and endpoints.
package account

import (
	"errors"
	"fmt"
)

// AccountErrorKind categorizes account errors for programmatic handling.
type AccountErrorKind int

const (
	// ErrKindInvalidAccount indicates invalid account data or state.
	ErrKindInvalidAccount AccountErrorKind = iota + 1
	// ErrKindInvalidTopic indicates invalid topic configuration.
	ErrKindInvalidTopic
	// ErrKindInvalidAddress indicates invalid communication or reachable address.
	ErrKindInvalidAddress
	// ErrKindInvalidDID indicates invalid DID format or content.
	ErrKindInvalidDID
	// ErrKindPeerIDMismatch indicates PeerID doesn't match expected value.
	ErrKindPeerIDMismatch
	// ErrKindMissingRequired indicates a required field is missing.
	ErrKindMissingRequired
	// ErrKindValidation indicates general validation failure.
	ErrKindValidation
	// ErrKindZeroValue indicates operation on zero-value type.
	ErrKindZeroValue
	// ErrKindInvalidHierarchy indicates invalid parent/child relationship.
	ErrKindInvalidHierarchy
	// ErrKindProhibitedField indicates a field is present that is not allowed.
	ErrKindProhibitedField
)

// String returns a human-readable name for the error kind.
func (k AccountErrorKind) String() string {
	switch k {
	case ErrKindInvalidAccount:
		return "InvalidAccount"
	case ErrKindInvalidTopic:
		return "InvalidTopic"
	case ErrKindInvalidAddress:
		return "InvalidAddress"
	case ErrKindInvalidDID:
		return "InvalidDID"
	case ErrKindPeerIDMismatch:
		return "PeerIDMismatch"
	case ErrKindMissingRequired:
		return "MissingRequired"
	case ErrKindValidation:
		return "Validation"
	case ErrKindZeroValue:
		return "ZeroValue"
	case ErrKindInvalidHierarchy:
		return "InvalidHierarchy"
	case ErrKindProhibitedField:
		return "ProhibitedField"
	default:
		return "Unknown"
	}
}

// AccountError is the base error type for all account-related errors.
// It provides rich context about what went wrong and why.
type AccountError struct {
	// Op is the operation that failed (e.g., "ParseCommAddress", "ValidateAccount").
	Op string
	// Kind categorizes the error for programmatic handling.
	Kind AccountErrorKind
	// Details provides a human-readable explanation of what went wrong.
	Details string
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface with a detailed message.
func (e *AccountError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("account.%s: %s: %s: %v", e.Op, e.Kind, e.Details, e.Err)
	}
	if e.Op != "" {
		return fmt.Sprintf("account.%s: %s: %s", e.Op, e.Kind, e.Details)
	}
	return fmt.Sprintf("account: %s: %s", e.Kind, e.Details)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *AccountError) Unwrap() error {
	return e.Err
}

// Is reports whether this error matches the target.
func (e *AccountError) Is(target error) bool {
	var ae *AccountError
	if errors.As(target, &ae) {
		// Match by Kind if both have the same kind
		if ae.Kind != 0 && e.Kind == ae.Kind {
			return true
		}
	}
	return false
}

// newAccountError creates a new AccountError with the given parameters.
func newAccountError(op string, kind AccountErrorKind, details string, err error) *AccountError {
	return &AccountError{
		Op:      op,
		Kind:    kind,
		Details: details,
		Err:     err,
	}
}

// Sentinel errors for common checks.
var (
	// ErrZeroPublicKey indicates a zero-value public key was provided.
	ErrZeroPublicKey = &AccountError{
		Kind:    ErrKindZeroValue,
		Details: "public key is zero-value",
	}

	// ErrMissingDID indicates a required DID is missing (for Parent accounts).
	ErrMissingDID = &AccountError{
		Kind:    ErrKindMissingRequired,
		Details: "DID is required for Parent accounts",
	}

	// ErrMissingParent indicates a required parent reference is missing (for Child accounts).
	ErrMissingParent = &AccountError{
		Kind:    ErrKindMissingRequired,
		Details: "parent public key is required for Child accounts",
	}
)

// Helper functions for creating specific error types.

// errInvalidAccount creates an error for invalid account data.
func errInvalidAccount(op string, reason string) *AccountError {
	return newAccountError(op, ErrKindInvalidAccount, reason, nil)
}

// errInvalidTopic creates an error for invalid topic configuration.
func errInvalidTopic(op string, reason string) *AccountError {
	return newAccountError(op, ErrKindInvalidTopic, reason, nil)
}

// errInvalidAddress creates an error for invalid address.
func errInvalidAddress(op string, reason string, err error) *AccountError {
	return newAccountError(op, ErrKindInvalidAddress, reason, err)
}

// errInvalidDID creates an error for invalid DID.
func errInvalidDID(op string, reason string) *AccountError {
	return newAccountError(op, ErrKindInvalidDID, reason, nil)
}

// errPeerIDMismatch creates an error for PeerID mismatch.
func errPeerIDMismatch(op string, expected, got string) *AccountError {
	return newAccountError(op, ErrKindPeerIDMismatch,
		fmt.Sprintf("expected PeerID %s, got %s", expected, got), nil)
}

// errMissingRequired creates an error for missing required field.
func errMissingRequired(op string, field string) *AccountError {
	return newAccountError(op, ErrKindMissingRequired,
		fmt.Sprintf("required field %s is missing", field), nil)
}

// errValidation creates an error for validation failures.
func errValidation(op string, reason string, err error) *AccountError {
	return newAccountError(op, ErrKindValidation, reason, err)
}

// errZeroValue creates an error for zero-value operations.
func errZeroValue(op string, typeName string) *AccountError {
	return newAccountError(op, ErrKindZeroValue,
		fmt.Sprintf("cannot perform operation on zero-value %s", typeName), nil)
}

// errInvalidHierarchy creates an error for invalid parent/child relationships.
func errInvalidHierarchy(op string, reason string) *AccountError {
	return newAccountError(op, ErrKindInvalidHierarchy, reason, nil)
}

// errProhibitedField creates an error for prohibited field presence.
func errProhibitedField(op string, field, accountType string) *AccountError {
	return newAccountError(op, ErrKindProhibitedField,
		fmt.Sprintf("field %s is prohibited for %s accounts", field, accountType), nil)
}

// wrapAccountError wraps an existing error with additional context.
func wrapAccountError(op string, kind AccountErrorKind, details string, err error) *AccountError {
	return newAccountError(op, kind, details, err)
}
