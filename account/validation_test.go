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
	// Helper to create zero comm addresses
	var zeroComm CommAddress
	var zeroPubKey keylib.NeuronPublicKey

	t.Run("valid parent account passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		// Parent accounts must have pubKey and DID, but NO comm channels or parentPubKey
		if err := ValidateParentAccount(privKey.PublicKey(), did, zeroComm, zeroComm, zeroComm, zeroPubKey); err != nil {
			t.Errorf("ValidateParentAccount() error = %v", err)
		}
	})

	t.Run("zero public key fails", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		err := ValidateParentAccount(zeroPubKey, did, zeroComm, zeroComm, zeroComm, zeroPubKey)
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
		err := ValidateParentAccount(privKey.PublicKey(), nil, zeroComm, zeroComm, zeroComm, zeroPubKey)
		if err == nil {
			t.Fatal("expected error for nil DID")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindMissingRequired {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindMissingRequired)
		}
	})

	t.Run("comm channels prohibited", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())

		// Create a non-zero comm address
		stdIn, _ := NewCommAddress(HederaTopicKind, "0.0.12345")

		err := ValidateParentAccount(privKey.PublicKey(), did, stdIn, zeroComm, zeroComm, zeroPubKey)
		if err == nil {
			t.Fatal("expected error for comm channel on parent")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindProhibitedField {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindProhibitedField)
		}
	})

	t.Run("parent reference prohibited", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newMockDID(privKey.PublicKey())
		otherKey, _ := keylib.GeneratePrivateKey()

		err := ValidateParentAccount(privKey.PublicKey(), did, zeroComm, zeroComm, zeroComm, otherKey.PublicKey())
		if err == nil {
			t.Fatal("expected error for parent reference on parent account")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindProhibitedField {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindProhibitedField)
		}
	})
}

// =============================================================================
// ValidateChildAccount Tests
// =============================================================================

func TestValidateChildAccount(t *testing.T) {
	// Helper function to create required comm addresses for child accounts
	createCommAddrs := func() (CommAddress, CommAddress, CommAddress) {
		stdIn, _ := NewCommAddress(HederaTopicKind, "0.0.111")
		stdOut, _ := NewCommAddress(HederaTopicKind, "0.0.222")
		stdErr, _ := NewCommAddress(HederaTopicKind, "0.0.333")
		return stdIn, stdOut, stdErr
	}

	var zeroComm CommAddress

	t.Run("valid child account passes", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()
		stdIn, stdOut, stdErr := createCommAddrs()

		// Child accounts must have all 3 comm channels and NO DID
		if err := ValidateChildAccount(childKey.PublicKey(), parentKey.PublicKey(), stdIn, stdOut, stdErr, nil); err != nil {
			t.Errorf("ValidateChildAccount() error = %v", err)
		}
	})

	t.Run("zero child public key fails", func(t *testing.T) {
		parentKey, _ := keylib.GeneratePrivateKey()
		var zeroChildKey keylib.NeuronPublicKey
		stdIn, stdOut, stdErr := createCommAddrs()

		err := ValidateChildAccount(zeroChildKey, parentKey.PublicKey(), stdIn, stdOut, stdErr, nil)
		if err == nil {
			t.Fatal("expected error for zero child key")
		}
	})

	t.Run("zero parent public key fails", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		var zeroParentKey keylib.NeuronPublicKey
		stdIn, stdOut, stdErr := createCommAddrs()

		err := ValidateChildAccount(childKey.PublicKey(), zeroParentKey, stdIn, stdOut, stdErr, nil)
		if err == nil {
			t.Fatal("expected error for zero parent key")
		}
	})

	t.Run("child equals parent fails", func(t *testing.T) {
		key, _ := keylib.GeneratePrivateKey()
		pubKey := key.PublicKey()
		stdIn, stdOut, stdErr := createCommAddrs()

		err := ValidateChildAccount(pubKey, pubKey, stdIn, stdOut, stdErr, nil)
		if err == nil {
			t.Fatal("expected error when child equals parent")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindInvalidHierarchy {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidHierarchy)
		}
	})

	t.Run("missing stdIn fails", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()
		_, stdOut, stdErr := createCommAddrs()

		err := ValidateChildAccount(childKey.PublicKey(), parentKey.PublicKey(), zeroComm, stdOut, stdErr, nil)
		if err == nil {
			t.Fatal("expected error for missing stdIn")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindMissingRequired {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindMissingRequired)
		}
	})

	t.Run("DID prohibited", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()
		stdIn, stdOut, stdErr := createCommAddrs()
		did := newMockDID(childKey.PublicKey())

		err := ValidateChildAccount(childKey.PublicKey(), parentKey.PublicKey(), stdIn, stdOut, stdErr, did)
		if err == nil {
			t.Fatal("expected error for DID on child account")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindProhibitedField {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindProhibitedField)
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
	var zeroComm CommAddress
	var zeroPubKey keylib.NeuronPublicKey

	t.Run("valid parent passes", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newMockDID(pubKey)

		// Parent accounts must have pubKey and DID, but NO comm channels or parentPubKey
		v := NewAccountValidator().ValidateParentRequirements(pubKey, did, zeroComm, zeroComm, zeroComm, zeroPubKey)
		if !v.IsValid() {
			t.Errorf("error: %v", v.Error())
		}
	})
}

func TestAccountValidator_ValidateChildRequirements(t *testing.T) {
	t.Run("valid child passes", func(t *testing.T) {
		childKey, _ := keylib.GeneratePrivateKey()
		parentKey, _ := keylib.GeneratePrivateKey()

		// Child accounts must have all 3 comm channels and NO DID
		stdIn, _ := NewCommAddress(HederaTopicKind, "0.0.111")
		stdOut, _ := NewCommAddress(HederaTopicKind, "0.0.222")
		stdErr, _ := NewCommAddress(HederaTopicKind, "0.0.333")

		v := NewAccountValidator().ValidateChildRequirements(childKey.PublicKey(), parentKey.PublicKey(), stdIn, stdOut, stdErr, nil)
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
// ValidateSharedAccount Tests
// =============================================================================

func TestValidateSharedAccount(t *testing.T) {
	// Helper to create valid MultisigKey
	createValidMultisigKey := func() *keylib.MultisigKey {
		priv1, _ := keylib.GeneratePrivateKey()
		priv2, _ := keylib.GeneratePrivateKey()
		priv3, _ := keylib.GeneratePrivateKey()
		mk, _ := keylib.NewMultisigKey(
			[]keylib.NeuronPublicKey{priv1.PublicKey(), priv2.PublicKey(), priv3.PublicKey()},
			2,
		)
		return &mk
	}

	t.Run("valid shared account passes", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		err := ValidateSharedAccount(mk, nil, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)
		if err != nil {
			t.Errorf("ValidateSharedAccount() error = %v", err)
		}
	})

	t.Run("nil MultisigKey fails", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		err := ValidateSharedAccount(nil, nil, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)
		if err == nil {
			t.Error("expected error for nil MultisigKey")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindMissingRequired {
			t.Errorf("expected ErrKindMissingRequired, got %v", ae.Kind)
		}
	})

	t.Run("zero MultisigKey fails", func(t *testing.T) {
		var zeroMk keylib.MultisigKey
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		err := ValidateSharedAccount(&zeroMk, nil, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)
		if err == nil {
			t.Error("expected error for zero MultisigKey")
		}
	})

	t.Run("DID prohibited", func(t *testing.T) {
		mk := createValidMultisigKey()
		priv, _ := keylib.GeneratePrivateKey()
		did := newMockDID(priv.PublicKey())
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		err := ValidateSharedAccount(mk, did, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)
		if err == nil {
			t.Error("expected error when DID is present")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindProhibitedField {
			t.Errorf("expected ErrKindProhibitedField, got %v", ae.Kind)
		}
	})

	t.Run("stdIn prohibited", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress
		stdIn, _ := NewHederaTopicAddress("0.0.111")

		err := ValidateSharedAccount(mk, nil, stdIn, zeroAddr, zeroAddr, zeroPubKey)
		if err == nil {
			t.Error("expected error when stdIn is present")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindProhibitedField {
			t.Errorf("expected ErrKindProhibitedField, got %v", ae.Kind)
		}
	})

	t.Run("stdOut prohibited", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress
		stdOut, _ := NewHederaTopicAddress("0.0.222")

		err := ValidateSharedAccount(mk, nil, zeroAddr, stdOut, zeroAddr, zeroPubKey)
		if err == nil {
			t.Error("expected error when stdOut is present")
		}
	})

	t.Run("stdErr prohibited", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress
		stdErr, _ := NewHederaTopicAddress("0.0.333")

		err := ValidateSharedAccount(mk, nil, zeroAddr, zeroAddr, stdErr, zeroPubKey)
		if err == nil {
			t.Error("expected error when stdErr is present")
		}
	})

	t.Run("parentPubKey prohibited", func(t *testing.T) {
		mk := createValidMultisigKey()
		priv, _ := keylib.GeneratePrivateKey()
		var zeroAddr CommAddress

		err := ValidateSharedAccount(mk, nil, zeroAddr, zeroAddr, zeroAddr, priv.PublicKey())
		if err == nil {
			t.Error("expected error when parentPubKey is present")
		}
	})
}

func TestAccountValidator_ValidateSharedRequirements(t *testing.T) {
	createValidMultisigKey := func() *keylib.MultisigKey {
		priv1, _ := keylib.GeneratePrivateKey()
		priv2, _ := keylib.GeneratePrivateKey()
		mk, _ := keylib.NewMultisigKey(
			[]keylib.NeuronPublicKey{priv1.PublicKey(), priv2.PublicKey()},
			2,
		)
		return &mk
	}

	t.Run("valid shared requirements", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		v := NewAccountValidator().
			ValidateSharedRequirements(mk, nil, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)

		if !v.IsValid() {
			t.Errorf("expected valid, got error: %v", v.Error())
		}
	})

	t.Run("returns self for chaining", func(t *testing.T) {
		mk := createValidMultisigKey()
		var zeroPubKey keylib.NeuronPublicKey
		var zeroAddr CommAddress

		v := NewAccountValidator()
		result := v.ValidateSharedRequirements(mk, nil, zeroAddr, zeroAddr, zeroAddr, zeroPubKey)

		if result != v {
			t.Error("ValidateSharedRequirements should return self for chaining")
		}
	})
}

// =============================================================================
// ValidateCurrencySymbol Tests
// =============================================================================

func TestValidateCurrencySymbol(t *testing.T) {
	t.Run("no ledger attachment and no currency is valid", func(t *testing.T) {
		if err := ValidateCurrencySymbol("", nil); err != nil {
			t.Errorf("should pass without ledger attachment: %v", err)
		}
	})

	t.Run("no ledger attachment with currency is valid", func(t *testing.T) {
		if err := ValidateCurrencySymbol("HBAR", nil); err != nil {
			t.Errorf("should pass with currency but no ledger: %v", err)
		}
	})

	t.Run("ledger attachment with currency is valid", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("hedera-mainnet", "0.0.12345")
		if err := ValidateCurrencySymbol("HBAR", attachment); err != nil {
			t.Errorf("should pass with both currency and ledger: %v", err)
		}
	})

	t.Run("ledger attachment without currency fails", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("hedera-mainnet", "0.0.12345")
		err := ValidateCurrencySymbol("", attachment)
		if err == nil {
			t.Error("should fail when ledger is attached but currency is empty")
		}
		if !containsStr(err.Error(), "currencySymbol") {
			t.Errorf("error should mention currencySymbol, got: %v", err)
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
