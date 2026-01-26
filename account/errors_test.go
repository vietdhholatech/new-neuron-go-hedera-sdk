package account

import (
	"errors"
	"fmt"
	"testing"
)

// =============================================================================
// AccountErrorKind.String Tests
// =============================================================================

func TestAccountErrorKind_String(t *testing.T) {
	tests := []struct {
		kind     AccountErrorKind
		expected string
	}{
		{ErrKindInvalidAccount, "InvalidAccount"},
		{ErrKindInvalidTopic, "InvalidTopic"},
		{ErrKindInvalidAddress, "InvalidAddress"},
		{ErrKindInvalidDID, "InvalidDID"},
		{ErrKindPeerIDMismatch, "PeerIDMismatch"},
		{ErrKindMissingRequired, "MissingRequired"},
		{ErrKindValidation, "Validation"},
		{ErrKindZeroValue, "ZeroValue"},
		{ErrKindInvalidHierarchy, "InvalidHierarchy"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}

	t.Run("unknown kind returns Unknown", func(t *testing.T) {
		unknownKind := AccountErrorKind(999)
		got := unknownKind.String()
		if got != "Unknown" {
			t.Errorf("String() for unknown kind = %v, want 'Unknown'", got)
		}
	})

	t.Run("zero kind returns Unknown", func(t *testing.T) {
		var zeroKind AccountErrorKind
		got := zeroKind.String()
		if got != "Unknown" {
			t.Errorf("String() for zero kind = %v, want 'Unknown'", got)
		}
	})
}

// =============================================================================
// AccountError.Error Tests
// =============================================================================

func TestAccountError_Error(t *testing.T) {
	t.Run("error format with underlying error", func(t *testing.T) {
		underlying := fmt.Errorf("underlying error")
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     underlying,
		}
		got := err.Error()
		expected := "account.TestOp: InvalidAccount: test details: underlying error"
		if got != expected {
			t.Errorf("Error() = %v, want %v", got, expected)
		}
	})

	t.Run("error format without underlying error", func(t *testing.T) {
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     nil,
		}
		got := err.Error()
		expected := "account.TestOp: InvalidAccount: test details"
		if got != expected {
			t.Errorf("Error() = %v, want %v", got, expected)
		}
	})

	t.Run("error format without op", func(t *testing.T) {
		err := &AccountError{
			Op:      "",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     nil,
		}
		got := err.Error()
		expected := "account: InvalidAccount: test details"
		if got != expected {
			t.Errorf("Error() = %v, want %v", got, expected)
		}
	})
}

// =============================================================================
// AccountError.Unwrap Tests
// =============================================================================

func TestAccountError_Unwrap(t *testing.T) {
	t.Run("returns underlying error", func(t *testing.T) {
		underlying := fmt.Errorf("underlying error")
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     underlying,
		}
		if err.Unwrap() != underlying {
			t.Error("Unwrap() should return underlying error")
		}
	})

	t.Run("returns nil when no underlying error", func(t *testing.T) {
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     nil,
		}
		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil when no underlying error")
		}
	})

	t.Run("works with errors.Unwrap", func(t *testing.T) {
		underlying := fmt.Errorf("underlying error")
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
			Err:     underlying,
		}
		unwrapped := errors.Unwrap(err)
		if unwrapped != underlying {
			t.Error("errors.Unwrap() should return underlying error")
		}
	})
}

// =============================================================================
// AccountError.Is Tests
// =============================================================================

func TestAccountError_Is(t *testing.T) {
	t.Run("matches same error kind", func(t *testing.T) {
		err1 := &AccountError{Kind: ErrKindInvalidAccount, Details: "error 1"}
		err2 := &AccountError{Kind: ErrKindInvalidAccount, Details: "error 2"}
		if !errors.Is(err1, err2) {
			t.Error("errors.Is() should match errors with same kind")
		}
	})

	t.Run("does not match different error kind", func(t *testing.T) {
		err1 := &AccountError{Kind: ErrKindInvalidAccount}
		err2 := &AccountError{Kind: ErrKindInvalidTopic}
		if errors.Is(err1, err2) {
			t.Error("errors.Is() should not match errors with different kind")
		}
	})

	t.Run("does not match non-AccountError", func(t *testing.T) {
		err1 := &AccountError{Kind: ErrKindInvalidAccount}
		err2 := fmt.Errorf("regular error")
		if errors.Is(err1, err2) {
			t.Error("errors.Is() should not match non-AccountError")
		}
	})

	t.Run("does not match zero kind target", func(t *testing.T) {
		err1 := &AccountError{Kind: ErrKindInvalidAccount}
		err2 := &AccountError{Kind: 0}
		if errors.Is(err1, err2) {
			t.Error("errors.Is() should not match when target has zero kind")
		}
	})
}

// =============================================================================
// AccountError errors.As Tests
// =============================================================================

func TestAccountError_As(t *testing.T) {
	t.Run("errors.As extracts AccountError", func(t *testing.T) {
		err := &AccountError{
			Op:      "TestOp",
			Kind:    ErrKindInvalidAccount,
			Details: "test details",
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("errors.As() should extract AccountError")
		}
		if ae.Kind != ErrKindInvalidAccount {
			t.Errorf("extracted Kind = %v, want %v", ae.Kind, ErrKindInvalidAccount)
		}
	})

	t.Run("errors.As extracts wrapped AccountError", func(t *testing.T) {
		inner := &AccountError{Kind: ErrKindInvalidAccount, Details: "inner"}
		outer := fmt.Errorf("wrapper: %w", inner)
		var ae *AccountError
		if !errors.As(outer, &ae) {
			t.Error("errors.As() should extract wrapped AccountError")
		}
		if ae.Kind != ErrKindInvalidAccount {
			t.Errorf("extracted Kind = %v, want %v", ae.Kind, ErrKindInvalidAccount)
		}
	})
}

// =============================================================================
// Sentinel Errors Tests
// =============================================================================

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrZeroPublicKey has correct kind", func(t *testing.T) {
		if ErrZeroPublicKey.Kind != ErrKindZeroValue {
			t.Errorf("ErrZeroPublicKey.Kind = %v, want %v", ErrZeroPublicKey.Kind, ErrKindZeroValue)
		}
	})

	t.Run("ErrMissingDID has correct kind", func(t *testing.T) {
		if ErrMissingDID.Kind != ErrKindMissingRequired {
			t.Errorf("ErrMissingDID.Kind = %v, want %v", ErrMissingDID.Kind, ErrKindMissingRequired)
		}
	})

	t.Run("ErrMissingParent has correct kind", func(t *testing.T) {
		if ErrMissingParent.Kind != ErrKindMissingRequired {
			t.Errorf("ErrMissingParent.Kind = %v, want %v", ErrMissingParent.Kind, ErrKindMissingRequired)
		}
	})

	t.Run("sentinel errors can be matched with Is", func(t *testing.T) {
		err := &AccountError{Kind: ErrKindZeroValue, Details: "custom"}
		if !errors.Is(err, ErrZeroPublicKey) {
			t.Error("should match ErrZeroPublicKey by kind")
		}
	})
}

// =============================================================================
// Error Helper Functions Tests
// =============================================================================

func TestErrorHelpers(t *testing.T) {
	t.Run("errInvalidAccount creates correct error", func(t *testing.T) {
		err := errInvalidAccount("TestOp", "test reason")
		if err.Kind != ErrKindInvalidAccount {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidAccount)
		}
		if err.Op != "TestOp" {
			t.Errorf("Op = %v, want 'TestOp'", err.Op)
		}
	})

	t.Run("errInvalidTopic creates correct error", func(t *testing.T) {
		err := errInvalidTopic("TestOp", "test reason")
		if err.Kind != ErrKindInvalidTopic {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidTopic)
		}
	})

	t.Run("errInvalidAddress creates correct error", func(t *testing.T) {
		underlying := fmt.Errorf("underlying")
		err := errInvalidAddress("TestOp", "test reason", underlying)
		if err.Kind != ErrKindInvalidAddress {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidAddress)
		}
		if err.Err != underlying {
			t.Error("underlying error not preserved")
		}
	})

	t.Run("errInvalidDID creates correct error", func(t *testing.T) {
		err := errInvalidDID("TestOp", "test reason")
		if err.Kind != ErrKindInvalidDID {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidDID)
		}
	})

	t.Run("errPeerIDMismatch includes expected and got values", func(t *testing.T) {
		err := errPeerIDMismatch("TestOp", "expected123", "got456")
		if err.Kind != ErrKindPeerIDMismatch {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindPeerIDMismatch)
		}
		errStr := err.Error()
		if !contains(errStr, "expected123") || !contains(errStr, "got456") {
			t.Errorf("error message should contain both expected and got values: %s", errStr)
		}
	})

	t.Run("errMissingRequired includes field name", func(t *testing.T) {
		err := errMissingRequired("TestOp", "testField")
		if err.Kind != ErrKindMissingRequired {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindMissingRequired)
		}
		if !contains(err.Error(), "testField") {
			t.Errorf("error message should contain field name: %s", err.Error())
		}
	})

	t.Run("errValidation wraps underlying error", func(t *testing.T) {
		underlying := fmt.Errorf("validation failed")
		err := errValidation("TestOp", "test reason", underlying)
		if err.Kind != ErrKindValidation {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindValidation)
		}
		if err.Err != underlying {
			t.Error("underlying error not preserved")
		}
	})

	t.Run("errZeroValue includes type name", func(t *testing.T) {
		err := errZeroValue("TestOp", "TestType")
		if err.Kind != ErrKindZeroValue {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindZeroValue)
		}
		if !contains(err.Error(), "TestType") {
			t.Errorf("error message should contain type name: %s", err.Error())
		}
	})

	t.Run("errInvalidHierarchy creates correct error", func(t *testing.T) {
		err := errInvalidHierarchy("TestOp", "test reason")
		if err.Kind != ErrKindInvalidHierarchy {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidHierarchy)
		}
	})

	t.Run("wrapAccountError preserves context", func(t *testing.T) {
		underlying := fmt.Errorf("underlying")
		err := wrapAccountError("TestOp", ErrKindValidation, "test details", underlying)
		if err.Kind != ErrKindValidation {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindValidation)
		}
		if err.Op != "TestOp" {
			t.Errorf("Op = %v, want 'TestOp'", err.Op)
		}
		if err.Details != "test details" {
			t.Errorf("Details = %v, want 'test details'", err.Details)
		}
		if err.Err != underlying {
			t.Error("underlying error not preserved")
		}
	})
}

// =============================================================================
// newAccountError Tests
// =============================================================================

func TestNewAccountError(t *testing.T) {
	t.Run("creates error with all fields", func(t *testing.T) {
		underlying := fmt.Errorf("underlying")
		err := newAccountError("TestOp", ErrKindInvalidAccount, "test details", underlying)
		if err.Op != "TestOp" {
			t.Errorf("Op = %v, want 'TestOp'", err.Op)
		}
		if err.Kind != ErrKindInvalidAccount {
			t.Errorf("Kind = %v, want %v", err.Kind, ErrKindInvalidAccount)
		}
		if err.Details != "test details" {
			t.Errorf("Details = %v, want 'test details'", err.Details)
		}
		if err.Err != underlying {
			t.Error("Err not set correctly")
		}
	})

	t.Run("creates error without underlying", func(t *testing.T) {
		err := newAccountError("TestOp", ErrKindInvalidAccount, "test details", nil)
		if err.Err != nil {
			t.Error("Err should be nil")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
