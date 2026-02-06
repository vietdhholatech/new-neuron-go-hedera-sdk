package account

import (
	"errors"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// ValidatePublicKey checks if a NeuronPublicKey is valid (non-zero).
func ValidatePublicKey(pubKey keylib.NeuronPublicKey) error {
	const op = "ValidatePublicKey"

	if pubKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	return nil
}

// ValidateAccountType checks if an AccountType is valid.
func ValidateAccountType(t AccountType) error {
	return t.Validate()
}

// ValidateParentAccount validates the requirements for a Parent account.
// Parent accounts must have:
//   - A valid (non-zero) public key
//   - A valid DID
//
// Parent accounts must NOT have:
//   - Communication channels (stdIn, stdOut, stdErr)
//   - A parent reference (parentPubKey)
func ValidateParentAccount(pubKey keylib.NeuronPublicKey, did NeuronDID,
	stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error {
	const op = "ValidateParentAccount"

	if pubKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	if did == nil {
		return errMissingRequired(op, "DID")
	}

	if err := did.Validate(); err != nil {
		return errValidation(op, "DID validation failed", err)
	}

	// Parent accounts must NOT have communication channels
	if !stdIn.IsZero() {
		return errProhibitedField(op, "stdIn", "Parent")
	}
	if !stdOut.IsZero() {
		return errProhibitedField(op, "stdOut", "Parent")
	}
	if !stdErr.IsZero() {
		return errProhibitedField(op, "stdErr", "Parent")
	}

	// Parent accounts must NOT have a parent reference
	if !parentPubKey.IsZero() {
		return errProhibitedField(op, "parentPubKey", "Parent")
	}

	return nil
}

// ValidateChildAccount validates the requirements for a Child account.
// Child accounts must have:
//   - A valid (non-zero) public key
//   - A valid (non-zero) parent public key
//   - Different public keys for self and parent
//   - All three communication channels (stdIn, stdOut, stdErr)
//
// Child accounts must NOT have:
//   - A DID (DIDs are reserved for Parent accounts)
func ValidateChildAccount(pubKey, parentPubKey keylib.NeuronPublicKey,
	stdIn, stdOut, stdErr CommAddress, did NeuronDID) error {
	const op = "ValidateChildAccount"

	if pubKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	if parentPubKey.IsZero() {
		return errMissingRequired(op, "parent public key")
	}

	// A child cannot be its own parent
	if pubKey.Equal(parentPubKey) {
		return errInvalidHierarchy(op, "child public key cannot equal parent public key")
	}

	// Child accounts must have all three communication channels
	if stdIn.IsZero() {
		return errMissingRequired(op, "stdIn")
	}
	if stdOut.IsZero() {
		return errMissingRequired(op, "stdOut")
	}
	if stdErr.IsZero() {
		return errMissingRequired(op, "stdErr")
	}

	// Child accounts must NOT have a DID
	if did != nil {
		return errProhibitedField(op, "DID", "Child")
	}

	return nil
}

// ValidateSharedAccount validates the requirements for a Shared account.
// Shared accounts must have:
//   - A valid (non-zero) MultisigKey
//
// Shared accounts must NOT have:
//   - A DID (DIDs are reserved for Parent accounts)
//   - Communication channels (stdIn, stdOut, stdErr)
//   - A parent reference (parentPubKey)
func ValidateSharedAccount(multisigKey *keylib.MultisigKey,
	did NeuronDID, stdIn, stdOut, stdErr CommAddress,
	parentPubKey keylib.NeuronPublicKey) error {
	const op = "ValidateSharedAccount"

	// Shared accounts must have a valid MultisigKey
	if multisigKey == nil || multisigKey.IsZero() {
		return errMissingRequired(op, "MultisigKey")
	}

	if err := multisigKey.Validate(); err != nil {
		return errValidation(op, "MultisigKey validation failed", err)
	}

	// Shared accounts must NOT have a DID
	if did != nil {
		return errProhibitedField(op, "DID", "Shared")
	}

	// Shared accounts must NOT have communication channels
	if !stdIn.IsZero() {
		return errProhibitedField(op, "stdIn", "Shared")
	}
	if !stdOut.IsZero() {
		return errProhibitedField(op, "stdOut", "Shared")
	}
	if !stdErr.IsZero() {
		return errProhibitedField(op, "stdErr", "Shared")
	}

	// Shared accounts must NOT have a parent reference
	if !parentPubKey.IsZero() {
		return errProhibitedField(op, "parentPubKey", "Shared")
	}

	return nil
}

// ValidateDID checks if a NeuronDID is valid.
func ValidateDID(did NeuronDID) error {
	const op = "ValidateDID"

	if did == nil {
		return errMissingRequired(op, "DID")
	}

	return did.Validate()
}

// ValidateDIDMatchesKey checks if a DID matches a public key.
// This is used to ensure DID-key correspondence for did:key and similar methods.
func ValidateDIDMatchesKey(did NeuronDID, pubKey keylib.NeuronPublicKey) error {
	const op = "ValidateDIDMatchesKey"

	if did == nil {
		return errMissingRequired(op, "DID")
	}

	if pubKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	// Check if the DID supports key matching
	didWithKey, ok := did.(NeuronDIDWithKey)
	if !ok {
		// DID method doesn't support key extraction - cannot validate match
		return nil
	}

	if !didWithKey.MatchesKey(pubKey) {
		return errInvalidDID(op, "DID does not match the provided public key")
	}

	return nil
}

// ValidateCommAddress checks if a CommAddress is valid.
func ValidateCommAddress(addr CommAddress) error {
	return addr.Validate()
}

// ValidateCommAddressNonEmpty checks if a CommAddress is valid and non-empty.
func ValidateCommAddressNonEmpty(addr CommAddress, fieldName string) error {
	const op = "ValidateCommAddressNonEmpty"

	if addr.IsZero() {
		return errMissingRequired(op, fieldName)
	}

	return addr.Validate()
}

// ValidateReachableAddr checks if a ReachableAddr is valid.
func ValidateReachableAddr(addr ReachableAddr) error {
	return addr.Validate()
}

// ValidateReachableAddrForAccount validates a ReachableAddr against an account's PeerID.
func ValidateReachableAddrForAccount(addr ReachableAddr, expectedPeerID keylib.PeerID) error {
	return addr.ValidateForAccount(expectedPeerID)
}

// ValidateReachableAddrs validates a collection of ReachableAddrs.
func ValidateReachableAddrs(addrs ReachableAddrs, expectedPeerID keylib.PeerID) error {
	return addrs.ValidateAll(expectedPeerID)
}

// ValidateCurrencySymbol checks that currency symbol is present when a ledger attachment exists.
// The spec (FR-020) requires each account to have a currency symbol. This validation enforces
// the requirement conditionally: currency is mandatory when attached to a ledger, optional otherwise.
func ValidateCurrencySymbol(currencySymbol string, ledgerAttachment *LedgerAttachment) error {
	const op = "ValidateCurrencySymbol"

	if ledgerAttachment != nil && currencySymbol == "" {
		return errMissingRequired(op, "currencySymbol (required when ledger is attached)")
	}
	return nil
}

// ValidationResult holds the results of validating a NeuronAccount.
// Uses errors.Join (Go 1.20+) for composing multiple validation errors,
// enabling use of errors.As() and errors.Is() on the composed error.
//
// # Thread Safety
//
// ValidationResult is NOT safe for concurrent use. Do not call AddError
// from multiple goroutines concurrently. Each goroutine should use its
// own ValidationResult instance or synchronize access externally.
//
// # Usage with errors.Join
//
// The Error() method returns a composed error that supports error unwrapping:
//
//	result := NewValidationResult()
//	result.AddError(errZeroValue("op", "field1"))
//	result.AddError(errMissingRequired("op", "field2"))
//
//	err := result.Error()
//	// err can be inspected with:
//	// - errors.Is(err, someSpecificError)
//	// - errors.As(err, &accountError)
//	// - result.Errors() for the full list
type ValidationResult struct {
	// errs holds individual validation errors.
	// errors.Join is used when composing the final error.
	errs []error
}

// AddError adds an error to the validation result.
// Nil errors are ignored.
func (r *ValidationResult) AddError(err error) {
	if err != nil {
		r.errs = append(r.errs, err)
	}
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.errs) > 0
}

// Valid returns true if all validations passed (no errors).
func (r *ValidationResult) Valid() bool {
	return len(r.errs) == 0
}

// Error returns the composed error containing all validation failures.
// Returns nil if validation passed.
// The returned error can be inspected with errors.As() and errors.Is().
func (r *ValidationResult) Error() error {
	if len(r.errs) == 0 {
		return nil
	}
	return errors.Join(r.errs...)
}

// FirstError returns the first validation error, or nil if none.
func (r *ValidationResult) FirstError() error {
	if len(r.errs) == 0 {
		return nil
	}
	return r.errs[0]
}

// Errors returns all individual validation errors.
func (r *ValidationResult) Errors() []error {
	return r.errs
}

// NewValidationResult creates a new ValidationResult starting as valid.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{}
}

// AccountValidator provides comprehensive validation for NeuronAccount.
// It collects all validation errors rather than failing on the first one.
type AccountValidator struct {
	result *ValidationResult
}

// NewAccountValidator creates a new AccountValidator.
func NewAccountValidator() *AccountValidator {
	return &AccountValidator{
		result: NewValidationResult(),
	}
}

// ValidatePublicKey adds public key validation to the validator.
func (v *AccountValidator) ValidatePublicKey(pubKey keylib.NeuronPublicKey) *AccountValidator {
	v.result.AddError(ValidatePublicKey(pubKey))
	return v
}

// ValidateAccountType adds account type validation to the validator.
func (v *AccountValidator) ValidateAccountType(t AccountType) *AccountValidator {
	v.result.AddError(ValidateAccountType(t))
	return v
}

// ValidateParentRequirements validates requirements specific to Parent accounts.
func (v *AccountValidator) ValidateParentRequirements(pubKey keylib.NeuronPublicKey, did NeuronDID,
	stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) *AccountValidator {
	v.result.AddError(ValidateParentAccount(pubKey, did, stdIn, stdOut, stdErr, parentPubKey))
	return v
}

// ValidateChildRequirements validates requirements specific to Child accounts.
func (v *AccountValidator) ValidateChildRequirements(pubKey, parentPubKey keylib.NeuronPublicKey,
	stdIn, stdOut, stdErr CommAddress, did NeuronDID) *AccountValidator {
	v.result.AddError(ValidateChildAccount(pubKey, parentPubKey, stdIn, stdOut, stdErr, did))
	return v
}

// ValidateSharedRequirements validates requirements specific to Shared accounts.
func (v *AccountValidator) ValidateSharedRequirements(multisigKey *keylib.MultisigKey,
	did NeuronDID, stdIn, stdOut, stdErr CommAddress,
	parentPubKey keylib.NeuronPublicKey) *AccountValidator {
	v.result.AddError(ValidateSharedAccount(multisigKey, did, stdIn, stdOut, stdErr, parentPubKey))
	return v
}

// ValidateDID adds DID validation to the validator.
func (v *AccountValidator) ValidateDID(did NeuronDID) *AccountValidator {
	v.result.AddError(ValidateDID(did))
	return v
}

// ValidateDIDMatchesKey adds DID-key match validation.
func (v *AccountValidator) ValidateDIDMatchesKey(did NeuronDID, pubKey keylib.NeuronPublicKey) *AccountValidator {
	v.result.AddError(ValidateDIDMatchesKey(did, pubKey))
	return v
}

// ValidateCommAddress adds CommAddress validation.
func (v *AccountValidator) ValidateCommAddress(addr CommAddress) *AccountValidator {
	v.result.AddError(ValidateCommAddress(addr))
	return v
}

// ValidateReachableAddrs adds ReachableAddrs validation.
func (v *AccountValidator) ValidateReachableAddrs(addrs ReachableAddrs, expectedPeerID keylib.PeerID) *AccountValidator {
	v.result.AddError(ValidateReachableAddrs(addrs, expectedPeerID))
	return v
}

// Result returns the validation result.
func (v *AccountValidator) Result() *ValidationResult {
	return v.result
}

// IsValid returns true if all validations passed.
func (v *AccountValidator) IsValid() bool {
	return v.result.Valid()
}

// Error returns the first validation error, or nil if valid.
func (v *AccountValidator) Error() error {
	return v.result.FirstError()
}
