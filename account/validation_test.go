package account

import (
	"errors"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// mockDID is a test double for NeuronDID interface
type mockDID struct {
	did      string
	method   string
	id       string
	valid    bool
	matchKey keylib.NeuronPublicKey
}

func (m *mockDID) String() string     { return m.did }
func (m *mockDID) Method() string     { return m.method }
func (m *mockDID) Identifier() string { return m.id }
func (m *mockDID) Validate() error {
	if !m.valid {
		return errors.New("invalid mock DID")
	}
	return nil
}
func (m *mockDID) Equal(other NeuronDID) bool                    { return m.did == other.String() }
func (m *mockDID) PublicKey() (keylib.NeuronPublicKey, error)    { return m.matchKey, nil }
func (m *mockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool { return m.matchKey.Equal(pubKey) }

func newMockDID(pubKey keylib.NeuronPublicKey) *mockDID {
	return &mockDID{
		did:      "did:mock:test",
		method:   "mock",
		id:       "test",
		valid:    true,
		matchKey: pubKey,
	}
}

// =============================================================================
// ValidatePublicKey Tests
// =============================================================================

func TestValidatePublicKey(t *testing.T) {
	t.Run("valid public key passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		if err := ValidatePublicKey(privKey.PublicKey()); err != nil {
			t.Errorf("ValidatePublicKey() error = %v", err)
		}
	})

	t.Run("zero public key fails", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		err := ValidatePublicKey(zeroPubKey)
		if err == nil {
			t.Fatal("expected error for zero public key")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
		if ae.Kind != ErrKindZeroValue {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
		}
	})
}

// =============================================================================
// ValidateAccountType Tests
// =============================================================================

func TestValidateAccountType(t *testing.T) {
	t.Run("Parent is valid", func(t *testing.T) {
		if err := ValidateAccountType(AccountTypeParent); err != nil {
			t.Errorf("ValidateAccountType() error = %v", err)
		}
	})

	t.Run("Child is valid", func(t *testing.T) {
		if err := ValidateAccountType(AccountTypeChild); err != nil {
			t.Errorf("ValidateAccountType() error = %v", err)
		}
	})

	t.Run("Unspecified is invalid", func(t *testing.T) {
		err := ValidateAccountType(AccountTypeUnspecified)
		if err == nil {
			t.Error("expected error for Unspecified")
		}
	})

	t.Run("unknown type is invalid", func(t *testing.T) {
		err := ValidateAccountType(AccountType(999))
		if err == nil {
			t.Error("expected error for unknown type")
		}
	})
}

// =============================================================================
// ValidateParentAccount Tests
// =============================================================================

func TestValidateParentAccount(t *testing.T) {
	t.Run("valid parent account passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		if err := ValidateParentAccount(privKey.PublicKey(), did); err != nil {
			t.Errorf("ValidateParentAccount() error = %v", err)
		}
	})

	t.Run("zero public key fails", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		var zeroPubKey keylib.NeuronPublicKey
		err := ValidateParentAccount(zeroPubKey, did)
		if err == nil {
			t.Fatal("expected error for zero public key")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindZeroValue {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
		}
	})

	t.Run("nil DID fails", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		err := ValidateParentAccount(privKey.PublicKey(), nil)
		if err == nil {
			t.Fatal("expected error for nil DID")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindMissingRequired {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindMissingRequired)
		}
	})
}

// =============================================================================
// ValidateChildAccount Tests
// =============================================================================

func TestValidateChildAccount(t *testing.T) {
	t.Run("valid child account passes", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()

		if err := ValidateChildAccount(childKey.PublicKey(), parentKey.PublicKey()); err != nil {
			t.Errorf("ValidateChildAccount() error = %v", err)
		}
	})

	t.Run("zero child public key fails", func(t *testing.T) {
		parentKey, _ := keylib.GeneratePrivateKey()
		var zeroChildKey keylib.NeuronPublicKey

		err := ValidateChildAccount(zeroChildKey, parentKey.PublicKey())
		if err == nil {
			t.Fatal("expected error for zero child key")
		}
	})

	t.Run("zero parent public key fails", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		var zeroParentKey keylib.NeuronPublicKey

		err := ValidateChildAccount(childKey.PublicKey(), zeroParentKey)
		if err == nil {
			t.Fatal("expected error for zero parent key")
		}
	})

	t.Run("child equals parent fails", func(t *testing.T) {
		key, _ := keylib.GeneratePrivateKey()
		pubKey := key.PublicKey()

		err := ValidateChildAccount(pubKey, pubKey)
		if err == nil {
			t.Fatal("expected error when child equals parent")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindInvalidHierarchy {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidHierarchy)
		}
	})
}

// =============================================================================
// ValidateDID Tests
// =============================================================================

func TestValidateDID(t *testing.T) {
	t.Run("valid DID passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		if err := ValidateDID(did); err != nil {
			t.Errorf("ValidateDID() error = %v", err)
		}
	})

	t.Run("nil DID fails", func(t *testing.T) {
		err := ValidateDID(nil)
		if err == nil {
			t.Fatal("expected error for nil DID")
		}
	})
}

// =============================================================================
// ValidateDIDMatchesKey Tests
// =============================================================================

func TestValidateDIDMatchesKey(t *testing.T) {
	t.Run("matching DID and key passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newMockDID(pubKey)

		if err := ValidateDIDMatchesKey(did, pubKey); err != nil {
			t.Errorf("ValidateDIDMatchesKey() error = %v", err)
		}
	})

	t.Run("non-matching DID and key fails", func(t *testing.T) {
		privKey1, _ := keylib.GeneratePrivateKey()
		privKey2, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey1.PublicKey())

		err := ValidateDIDMatchesKey(did, privKey2.PublicKey())
		if err == nil {
			t.Fatal("expected error for non-matching DID and key")
		}
	})

	t.Run("nil DID fails", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		err := ValidateDIDMatchesKey(nil, privKey.PublicKey())
		if err == nil {
			t.Fatal("expected error for nil DID")
		}
	})

	t.Run("zero public key fails", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		var zeroPubKey keylib.NeuronPublicKey
		err := ValidateDIDMatchesKey(did, zeroPubKey)
		if err == nil {
			t.Fatal("expected error for zero public key")
		}
	})
}

// =============================================================================
// ValidateCommAddress Tests
// =============================================================================

func TestValidateCommAddress_Func(t *testing.T) {
	t.Run("valid address passes", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if err := ValidateCommAddress(addr); err != nil {
			t.Errorf("ValidateCommAddress() error = %v", err)
		}
	})

	t.Run("zero address fails", func(t *testing.T) {
		var addr CommAddress
		err := ValidateCommAddress(addr)
		if err == nil {
			t.Fatal("expected error for zero address")
		}
	})
}

func TestValidateCommAddressNonEmpty(t *testing.T) {
	t.Run("valid address passes", func(t *testing.T) {
		addr, _ := NewHederaTopicAddress("0.0.12345")
		if err := ValidateCommAddressNonEmpty(addr, "stdIn"); err != nil {
			t.Errorf("ValidateCommAddressNonEmpty() error = %v", err)
		}
	})

	t.Run("zero address fails with field name", func(t *testing.T) {
		var addr CommAddress
		err := ValidateCommAddressNonEmpty(addr, "stdIn")
		if err == nil {
			t.Fatal("expected error for zero address")
		}
		// Error should mention the field name
		errStr := err.Error()
		if !containsStr(errStr, "stdIn") {
			t.Errorf("error should mention field name 'stdIn': %s", errStr)
		}
	})
}

// =============================================================================
// ValidationResult Tests
// =============================================================================

func TestNewValidationResult(t *testing.T) {
	t.Run("creates valid result by default", func(t *testing.T) {
		result := NewValidationResult()
		if !result.Valid() {
			t.Error("new result should be valid")
		}
		if len(result.Errors()) != 0 {
			t.Error("new result should have no errors")
		}
	})
}

func TestValidationResult_AddError(t *testing.T) {
	t.Run("adding error marks result invalid", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError(errors.New("test error"))

		if result.Valid() {
			t.Error("result should be invalid after adding error")
		}
		if len(result.Errors()) != 1 {
			t.Errorf("Errors len = %d, want 1", len(result.Errors()))
		}
	})

	t.Run("adding nil does not mark invalid", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError(nil)

		if !result.Valid() {
			t.Error("result should remain valid after adding nil")
		}
		if len(result.Errors()) != 0 {
			t.Error("nil should not be added to errors")
		}
	})

	t.Run("multiple errors accumulated", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError(errors.New("error 1"))
		result.AddError(errors.New("error 2"))

		if len(result.Errors()) != 2 {
			t.Errorf("Errors len = %d, want 2", len(result.Errors()))
		}
	})
}

func TestValidationResult_HasErrors(t *testing.T) {
	t.Run("no errors returns false", func(t *testing.T) {
		result := NewValidationResult()
		if result.HasErrors() {
			t.Error("HasErrors() should return false")
		}
	})

	t.Run("with errors returns true", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError(errors.New("test"))
		if !result.HasErrors() {
			t.Error("HasErrors() should return true")
		}
	})
}

func TestValidationResult_FirstError(t *testing.T) {
	t.Run("no errors returns nil", func(t *testing.T) {
		result := NewValidationResult()
		if result.FirstError() != nil {
			t.Error("FirstError() should return nil")
		}
	})

	t.Run("returns first error", func(t *testing.T) {
		result := NewValidationResult()
		first := errors.New("first")
		result.AddError(first)
		result.AddError(errors.New("second"))

		if result.FirstError() != first {
			t.Error("FirstError() should return first error")
		}
	})
}

// =============================================================================
// AccountValidator Tests
// =============================================================================

func TestNewAccountValidator(t *testing.T) {
	t.Run("creates validator with valid result", func(t *testing.T) {
		v := NewAccountValidator()
		if v == nil {
			t.Fatal("expected non-nil validator")
		}
		if !v.IsValid() {
			t.Error("new validator should be valid")
		}
	})
}

func TestAccountValidator_Fluent(t *testing.T) {
	t.Run("supports fluent chaining", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newMockDID(pubKey)
		addr, _ := NewHederaTopicAddress("0.0.12345")

		v := NewAccountValidator().
			ValidatePublicKey(pubKey).
			ValidateAccountType(AccountTypeParent).
			ValidateDID(did).
			ValidateDIDMatchesKey(did, pubKey).
			ValidateCommAddress(addr)

		if !v.IsValid() {
			t.Errorf("validation should pass, got error: %v", v.Error())
		}
	})

	t.Run("collects multiple errors", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		v := NewAccountValidator().
			ValidatePublicKey(zeroPubKey).
			ValidateAccountType(AccountTypeUnspecified).
			ValidateDID(nil)

		if v.IsValid() {
			t.Error("validation should fail with multiple errors")
		}
		result := v.Result()
		if len(result.Errors()) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors()))
		}
	})
}

func TestAccountValidator_ValidateParentRequirements(t *testing.T) {
	t.Run("valid parent passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newMockDID(pubKey)

		v := NewAccountValidator().ValidateParentRequirements(pubKey, did)
		if !v.IsValid() {
			t.Errorf("error: %v", v.Error())
		}
	})
}

func TestAccountValidator_ValidateChildRequirements(t *testing.T) {
	t.Run("valid child passes", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()

		v := NewAccountValidator().ValidateChildRequirements(childKey.PublicKey(), parentKey.PublicKey())
		if !v.IsValid() {
			t.Errorf("error: %v", v.Error())
		}
	})
}

func TestAccountValidator_Result(t *testing.T) {
	t.Run("returns validation result", func(t *testing.T) {
		v := NewAccountValidator()
		result := v.Result()
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if !result.Valid() {
			t.Error("result should be valid")
		}
	})
}

func TestAccountValidator_Error(t *testing.T) {
	t.Run("returns nil when valid", func(t *testing.T) {
		v := NewAccountValidator()
		if v.Error() != nil {
			t.Error("Error() should return nil when valid")
		}
	})

	t.Run("returns first error when invalid", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		v := NewAccountValidator().ValidatePublicKey(zeroPubKey)
		if v.Error() == nil {
			t.Error("Error() should return error when invalid")
		}
	})
}

// =============================================================================
// Helper
// =============================================================================

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
