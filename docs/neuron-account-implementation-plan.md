# NeuronAccount Specification Implementation Plan

**Document Version:** 1.0
**Created:** 2026-01-28
**Based On:** [NeuronAccount Implementation Gap Analysis](neuron-account-implementation-gap-analysis.md)
**Specification Reference:** [docs/neuron-sdk-spec/NeuronAccount.md](neuron-sdk-spec/NeuronAccount.md)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Phase 1: AccountTypeShared and MultisigKey Foundation](#2-phase-1-accounttypeshared-and-multisigkey-foundation)
3. [Phase 2: Financial Fields](#3-phase-2-financial-fields)
4. [Phase 3: Shared Account Builder](#4-phase-3-shared-account-builder)
5. [Phase 4: LedgerAttachment Entity](#5-phase-4-ledgerattachment-entity)
6. [Phase 5: Validation Alignment](#6-phase-5-validation-alignment)
7. [Phase 6: Testing and Documentation](#7-phase-6-testing-and-documentation)
8. [Breaking Changes and Migration](#8-breaking-changes-and-migration)
9. [Verification Procedures](#9-verification-procedures)
10. [Risk Assessment](#10-risk-assessment)

---

## 1. Executive Summary

This document provides a detailed implementation plan to address the gaps identified between the NeuronAccount technical specification and the current Go implementation. The plan is organized into six phases with clear dependencies, file modifications, and verification procedures.

### 1.1 Implementation Overview

| Phase | Description                                  | Priority | Breaking Changes |
| ----- | -------------------------------------------- | -------- | ---------------- |
| 1     | AccountTypeShared and MultisigKey Foundation | High     | None             |
| 2     | Financial Fields                             | High     | None (additive)  |
| 3     | Shared Account Builder                       | High     | None             |
| 4     | LedgerAttachment Entity                      | Medium   | None             |
| 5     | Validation Alignment                         | Medium   | Yes              |
| 6     | Testing and Documentation                    | Low      | None             |

### 1.2 Files Summary

**New Files to Create (6):**

- `keylib/multisig_key.go`
- `keylib/multisig_key_test.go`
- `account/ledger_attachment.go`
- `account/ledger_attachment_test.go`
- `account/ledger_verification.go`
- `account/ledger_verification_test.go`

**Existing Files to Modify (7):**

- `account/account_type.go`
- `account/account.go`
- `account/builder.go`
- `account/validation.go`
- `account/errors.go`
- `keylib/errors.go`
- `keylib/doc.go`

---

## 2. Phase 1: AccountTypeShared and MultisigKey Foundation

### 2.1 Overview

| Attribute        | Value |
| ---------------- | ----- |
| Priority         | High  |
| Dependencies     | None  |
| Breaking Changes | None  |
| New Files        | 2     |
| Modified Files   | 2     |

This phase establishes the foundation for Shared accounts by adding the `AccountTypeShared` constant and implementing the `MultisigKey` type in the keylib package.

### 2.2 New File: keylib/multisig_key.go

**Purpose:** Define the MultisigKey type for M-of-N threshold signing configurations.

**Type Definition:**

```go
// MultisigKey represents an M-of-N threshold signing configuration.
// This type is immutable after construction.
// The zero value is invalid; methods return errors on zero value.
type MultisigKey struct {
    publicKeys []NeuronPublicKey  // participating public keys (unexported, defensive copy on access)
    threshold  int                // M in M-of-N (minimum signatures required)
    total      int                // N in M-of-N (total number of keys)
}
```

**Constructor Functions:**

| Function           | Signature                                                                               | Description                           |
| ------------------ | --------------------------------------------------------------------------------------- | ------------------------------------- |
| NewMultisigKey     | `func NewMultisigKey(publicKeys []NeuronPublicKey, threshold int) (MultisigKey, error)` | Creates a MultisigKey with validation |
| MustNewMultisigKey | `func MustNewMultisigKey(publicKeys []NeuronPublicKey, threshold int) MultisigKey`      | Panics on error (for tests)           |

**Methods:**

| Method      | Signature                                                       | Description                       |
| ----------- | --------------------------------------------------------------- | --------------------------------- |
| IsZero      | `func (k MultisigKey) IsZero() bool`                            | Returns true if uninitialized     |
| Validate    | `func (k MultisigKey) Validate() error`                         | Validates threshold configuration |
| Threshold   | `func (k MultisigKey) Threshold() int`                          | Returns M (minimum signatures)    |
| Total       | `func (k MultisigKey) Total() int`                              | Returns N (total keys)            |
| PublicKeys  | `func (k MultisigKey) PublicKeys() []NeuronPublicKey`           | Returns defensive copy of keys    |
| ContainsKey | `func (k MultisigKey) ContainsKey(pubKey NeuronPublicKey) bool` | Checks if key is in set           |
| Equal       | `func (k MultisigKey) Equal(other MultisigKey) bool`            | Constant-time equality            |
| String      | `func (k MultisigKey) String() string`                          | Human-readable representation     |

**Validation Rules:**

| Rule                   | Error Condition                    |
| ---------------------- | ---------------------------------- |
| Non-empty keys         | `len(publicKeys) == 0`             |
| Valid threshold        | `threshold < 1`                    |
| Threshold not exceeded | `threshold > len(publicKeys)`      |
| No duplicate keys      | Any key appears more than once     |
| All keys valid         | Any key returns `IsZero() == true` |

**Implementation Notes:**

1. Store defensive copy of publicKeys slice in constructor
2. Return defensive copy in PublicKeys() accessor
3. Use constant-time comparison for Equal() method
4. Validate all keys are non-zero during construction

### 2.3 New File: keylib/multisig_key_test.go

**Test Cases:**

| Test Function                            | Description                                       |
| ---------------------------------------- | ------------------------------------------------- |
| TestNewMultisigKey_ValidConfiguration    | Test 2-of-3, 3-of-5, 1-of-1 configurations        |
| TestNewMultisigKey_InvalidThreshold      | Threshold > total returns error                   |
| TestNewMultisigKey_ZeroThreshold         | Threshold = 0 returns error                       |
| TestNewMultisigKey_EmptyPublicKeys       | Empty slice returns error                         |
| TestNewMultisigKey_DuplicateKeys         | Duplicate keys return error                       |
| TestNewMultisigKey_ZeroValueKey          | Zero-value key in slice returns error             |
| TestMultisigKey_IsZero                   | Zero value returns true                           |
| TestMultisigKey_Validate                 | Validation on valid/invalid configs               |
| TestMultisigKey_PublicKeys_DefensiveCopy | Modifying returned slice does not affect original |
| TestMultisigKey_ContainsKey              | Key presence detection                            |
| TestMultisigKey_Equal                    | Equality comparison                               |
| TestMultisigKey_String                   | String representation format                      |

### 2.4 Modified File: account/account_type.go

**Changes Required:**

1. Add `AccountTypeShared` constant with value 3
2. Update `String()` method to handle Shared type
3. Update `IsValid()` method to include Shared type
4. Add `IsShared()` method
5. Add `RequiresMultisigKey()` method

**Code Changes:**

```go
// Constants section - add AccountTypeShared
const (
    AccountTypeUnspecified AccountType = iota  // 0
    AccountTypeParent                          // 1
    AccountTypeChild                           // 2
    AccountTypeShared                          // 3 - NEW
)

// String() method - add case for Shared
func (t AccountType) String() string {
    switch t {
    case AccountTypeUnspecified:
        return "Unspecified"
    case AccountTypeParent:
        return "Parent"
    case AccountTypeChild:
        return "Child"
    case AccountTypeShared:  // NEW
        return "Shared"
    default:
        return fmt.Sprintf("Unknown(%d)", int(t))
    }
}

// IsValid() method - add Shared to valid types
func (t AccountType) IsValid() bool {
    return t == AccountTypeParent || t == AccountTypeChild || t == AccountTypeShared
}

// NEW: IsShared() method
func (t AccountType) IsShared() bool {
    return t == AccountTypeShared
}

// NEW: RequiresMultisigKey() method
func (t AccountType) RequiresMultisigKey() bool {
    return t == AccountTypeShared
}
```

### 2.5 Modified File: keylib/errors.go

**Changes Required:**

Add new error kinds for MultisigKey validation.

```go
const (
    // ... existing error kinds ...
    ErrKindInvalidThreshold   ErrorKind = iota + <next_value>  // NEW: threshold configuration invalid
)

// NEW: Error helper function
func errInvalidThreshold(op string, threshold, total int) *KeyError {
    return &KeyError{
        Op:      op,
        Kind:    ErrKindInvalidThreshold,
        Details: fmt.Sprintf("invalid threshold: %d-of-%d (threshold must be >= 1 and <= total)", threshold, total),
    }
}
```

---

## 3. Phase 2: Financial Fields

### 3.1 Overview

| Attribute        | Value                |
| ---------------- | -------------------- |
| Priority         | High                 |
| Dependencies     | Phase 1 complete     |
| Breaking Changes | None (additive only) |
| New Files        | 0                    |
| Modified Files   | 2                    |

This phase adds financial fields and the MultisigKey field to NeuronAccount, along with corresponding builder methods.

### 3.2 Modified File: account/account.go

**Struct Field Additions:**

```go
type NeuronAccount struct {
    // === Identity (from keylib) ===
    publicKey  keylib.NeuronPublicKey
    peerID     keylib.PeerID
    evmAddress keylib.EVMAddress

    // NEW: For Shared accounts (nil for Parent/Child)
    multisigKey *keylib.MultisigKey

    // === Hierarchy ===
    accountType  AccountType
    did          NeuronDID
    parentPubKey keylib.NeuronPublicKey

    // === Communication Endpoints ===
    stdIn          CommAddress
    stdOut         CommAddress
    stdErr         CommAddress
    reachableAddrs ReachableAddrs

    // === Financial Layer (NEW) ===
    currencySymbol    string    // Currency identifier (e.g., "HBAR", "ETH")
    creditBalance     *big.Int  // For Parent accounts only
    balanceAllocation *big.Int  // For Child accounts only
    balance           *big.Int  // For Shared accounts only

    // === Ledger Integration Layer (NEW) ===
    ledgerAttachment *LedgerAttachment  // Optional ledger link
}
```

**New Accessor Methods:**

| Method              | Signature                                                     | Description                                   |
| ------------------- | ------------------------------------------------------------- | --------------------------------------------- |
| MultisigKey         | `func (a NeuronAccount) MultisigKey() *keylib.MultisigKey`    | Returns multisig key (nil for Parent/Child)   |
| CurrencySymbol      | `func (a NeuronAccount) CurrencySymbol() string`              | Returns currency identifier                   |
| CreditBalance       | `func (a NeuronAccount) CreditBalance() *big.Int`             | Returns credit balance (nil for Child/Shared) |
| BalanceAllocation   | `func (a NeuronAccount) BalanceAllocation() *big.Int`         | Returns allocation (nil for Parent/Shared)    |
| Balance             | `func (a NeuronAccount) Balance() *big.Int`                   | Returns balance (nil for Parent/Child)        |
| LedgerAttachment    | `func (a NeuronAccount) LedgerAttachment() *LedgerAttachment` | Returns ledger attachment                     |
| HasLedgerAttachment | `func (a NeuronAccount) HasLedgerAttachment() bool`           | Returns true if attached                      |

**JSON Serialization Updates:**

```go
type neuronAccountJSON struct {
    // ... existing fields ...

    // NEW: MultisigKey fields
    MultisigThreshold int      `json:"multisigThreshold,omitempty"`
    MultisigTotal     int      `json:"multisigTotal,omitempty"`
    MultisigKeys      []string `json:"multisigKeys,omitempty"`

    // NEW: Financial fields
    CurrencySymbol    string `json:"currencySymbol,omitempty"`
    CreditBalance     string `json:"creditBalance,omitempty"`      // big.Int as string
    BalanceAllocation string `json:"balanceAllocation,omitempty"`  // big.Int as string
    Balance           string `json:"balance,omitempty"`            // big.Int as string

    // NEW: Ledger fields
    LedgerIdentifier string `json:"ledgerIdentifier,omitempty"`
    AttachedAddress  string `json:"attachedAddress,omitempty"`
}
```

### 3.3 Modified File: account/builder.go

**Builder Field Additions:**

```go
type AccountBuilder struct {
    // ... existing fields ...

    // NEW fields
    multisigKey       *keylib.MultisigKey
    currencySymbol    string
    creditBalance     *big.Int
    balanceAllocation *big.Int
    balance           *big.Int
    ledgerAttachment  *LedgerAttachment
}
```

**New Builder Methods:**

| Method                | Signature                                                                                     | Description                 |
| --------------------- | --------------------------------------------------------------------------------------------- | --------------------------- |
| WithCurrencySymbol    | `func (b *AccountBuilder) WithCurrencySymbol(symbol string) *AccountBuilder`                  | Set currency symbol         |
| WithCreditBalance     | `func (b *AccountBuilder) WithCreditBalance(balance *big.Int) *AccountBuilder`                | Set credit balance (Parent) |
| WithBalanceAllocation | `func (b *AccountBuilder) WithBalanceAllocation(allocation *big.Int) *AccountBuilder`         | Set allocation (Child)      |
| WithBalance           | `func (b *AccountBuilder) WithBalance(balance *big.Int) *AccountBuilder`                      | Set balance (Shared)        |
| WithLedgerAttachment  | `func (b *AccountBuilder) WithLedgerAttachment(attachment *LedgerAttachment) *AccountBuilder` | Set ledger attachment       |

**Method Implementation Pattern:**

```go
func (b *AccountBuilder) WithCurrencySymbol(symbol string) *AccountBuilder {
    if symbol == "" {
        // Empty symbol is allowed (optional field)
    }
    b.currencySymbol = symbol
    return b
}

func (b *AccountBuilder) WithCreditBalance(balance *big.Int) *AccountBuilder {
    if b.accountType != AccountTypeParent {
        b.addError(errInvalidAccount("WithCreditBalance",
            "credit balance can only be set on parent accounts"))
        return b
    }
    b.creditBalance = balance
    return b
}
```

---

## 4. Phase 3: Shared Account Builder

### 4.1 Overview

| Attribute        | Value                        |
| ---------------- | ---------------------------- |
| Priority         | High                         |
| Dependencies     | Phase 1 and Phase 2 complete |
| Breaking Changes | None                         |
| New Files        | 0                            |
| Modified Files   | 1                            |

This phase adds the `NewSharedAccountBuilder` constructor and updates the `Build()` method to handle Shared accounts.

### 4.2 Modified File: account/builder.go

**New Constructor:**

```go
// NewSharedAccountBuilder creates a builder for a Shared NeuronAccount.
// Shared accounts require a MultisigKey and do not have a single public key.
//
// Shared accounts:
//   - Use MultisigKey for threshold signing (M-of-N)
//   - Do NOT have a DID
//   - Do NOT have communication channels
//   - Do NOT have a parent reference
func NewSharedAccountBuilder(multisigKey keylib.MultisigKey) *AccountBuilder {
    b := &AccountBuilder{
        accountType: AccountTypeShared,
        multisigKey: &multisigKey,
    }

    // Validate MultisigKey upfront (non-blocking, accumulate errors)
    if multisigKey.IsZero() {
        b.addError(errZeroValue("NewSharedAccountBuilder", "MultisigKey"))
    } else if err := multisigKey.Validate(); err != nil {
        b.addError(wrapAccountError("NewSharedAccountBuilder", ErrKindValidation,
            "invalid MultisigKey", err))
    }

    return b
}
```

**Build() Method Updates:**

```go
func (b *AccountBuilder) Build() (NeuronAccount, error) {
    const op = "AccountBuilder.Build"

    // Check for accumulated errors
    if len(b.errors) > 0 {
        return NeuronAccount{}, b.errors[0]
    }

    // Final validation based on account type
    switch b.accountType {
    case AccountTypeParent:
        // ... existing validation ...

    case AccountTypeChild:
        // ... existing validation ...

    case AccountTypeShared:  // NEW
        if err := ValidateSharedAccount(b.multisigKey, b.did,
            b.stdIn, b.stdOut, b.stdErr, b.parentPubKey); err != nil {
            return NeuronAccount{}, wrapAccountError(op, ErrKindValidation,
                "shared validation failed", err)
        }

    default:
        return NeuronAccount{}, errInvalidAccount(op, "invalid account type")
    }

    // Derive identifiers based on account type
    var peerID keylib.PeerID
    var evmAddress keylib.EVMAddress

    if b.accountType == AccountTypeShared {
        // Shared accounts don't have single public key
        // PeerID and EVMAddress may be zero for Shared accounts
        // unless derived from MultisigKey (future enhancement)
    } else {
        peerID, err = b.publicKey.PeerID()
        if err != nil {
            return NeuronAccount{}, wrapAccountError(op, ErrKindValidation,
                "failed to derive PeerID", err)
        }
        evmAddress = b.publicKey.EVMAddress()
    }

    // ... validation of reachable addresses (skip for Shared) ...

    // Construct account with new fields
    account := NeuronAccount{
        publicKey:         b.publicKey,
        multisigKey:       b.multisigKey,  // NEW
        peerID:            peerID,
        evmAddress:        evmAddress,
        accountType:       b.accountType,
        did:               b.did,
        parentPubKey:      b.parentPubKey,
        stdIn:             b.stdIn,
        stdOut:            b.stdOut,
        stdErr:            b.stdErr,
        reachableAddrs:    NewReachableAddrs(b.reachableAddrs...),
        currencySymbol:    b.currencySymbol,     // NEW
        creditBalance:     b.creditBalance,      // NEW
        balanceAllocation: b.balanceAllocation,  // NEW
        balance:           b.balance,            // NEW
        ledgerAttachment:  b.ledgerAttachment,   // NEW
    }

    return account, nil
}
```

---

## 5. Phase 4: LedgerAttachment Entity

### 5.1 Overview

| Attribute        | Value            |
| ---------------- | ---------------- |
| Priority         | Medium           |
| Dependencies     | Phase 1 complete |
| Breaking Changes | None             |
| New Files        | 4                |
| Modified Files   | 0                |

This phase creates the LedgerAttachment entity and verification interface.

### 5.2 New File: account/ledger_attachment.go

**Type Definitions:**

```go
// AttachmentState represents the current state of a ledger attachment.
type AttachmentState int

const (
    AttachmentStateDetached AttachmentState = iota  // Not linked to ledger
    AttachmentStateAttached                         // Linked but not verified
    AttachmentStateVerified                         // Linked and verified
)

func (s AttachmentState) String() string
func (s AttachmentState) IsValid() bool

// VerificationStatus represents the result of ledger verification.
type VerificationStatus int

const (
    VerificationStatusNone VerificationStatus = iota     // Not attempted
    VerificationStatusPending                            // In progress
    VerificationStatusVerified                           // Successfully verified
    VerificationStatusFailed                             // Verification failed
)

func (s VerificationStatus) String() string
```

**LedgerAttachment Struct:**

```go
// LedgerAttachment represents the link between a NeuronAccount
// and settlement infrastructure (blockchain or ledger).
type LedgerAttachment struct {
    ledgerIdentifier   string              // e.g., "ethereum-mainnet", "hedera-mainnet"
    attachedAddress    string              // Ledger-specific account address
    attachmentState    AttachmentState
    verificationStatus VerificationStatus
    verifiedAt         time.Time           // When verification completed (zero if not verified)
    lastError          string              // Last verification error message
}
```

**Constructor and Methods:**

| Function/Method     | Signature                                                                       | Description                       |
| ------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| NewLedgerAttachment | `func NewLedgerAttachment(ledgerID, address string) (*LedgerAttachment, error)` | Create in Detached state          |
| LedgerIdentifier    | `func (a *LedgerAttachment) LedgerIdentifier() string`                          | Get ledger identifier             |
| AttachedAddress     | `func (a *LedgerAttachment) AttachedAddress() string`                           | Get attached address              |
| State               | `func (a *LedgerAttachment) State() AttachmentState`                            | Get current state                 |
| VerificationStatus  | `func (a *LedgerAttachment) VerificationStatus() VerificationStatus`            | Get verification status           |
| IsAttached          | `func (a *LedgerAttachment) IsAttached() bool`                                  | Returns true if state >= Attached |
| IsVerified          | `func (a *LedgerAttachment) IsVerified() bool`                                  | Returns true if state == Verified |
| Validate            | `func (a *LedgerAttachment) Validate() error`                                   | Validate configuration            |

### 5.3 New File: account/ledger_verification.go

**LedgerVerifier Interface:**

```go
// LedgerVerifier defines the interface for verifying account existence
// and key ownership on a specific ledger.
type LedgerVerifier interface {
    // LedgerIdentifier returns the identifier for this ledger.
    LedgerIdentifier() string

    // VerifyAccountExists checks if the account exists on the ledger.
    VerifyAccountExists(address string) (bool, error)

    // VerifyKeyOwnership verifies that the given key controls the account.
    VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error)

    // VerifyMultisigOwnership verifies multisig key ownership.
    VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error)
}
```

**Verification Function:**

```go
// VerifyAttachment performs full verification of a ledger attachment.
// This function:
//   1. Checks that the account exists on the ledger
//   2. Verifies that the provided key controls the account
//   3. Updates the attachment state and verification status
func VerifyAttachment(attachment *LedgerAttachment, verifier LedgerVerifier,
    pubKey keylib.NeuronPublicKey) error
```

### 5.4 New Files: Test Files

- `account/ledger_attachment_test.go` - Tests for LedgerAttachment
- `account/ledger_verification_test.go` - Tests for verification (with mock verifier)

---

## 6. Phase 5: Validation Alignment

### 6.1 Overview

| Attribute        | Value               |
| ---------------- | ------------------- |
| Priority         | Medium              |
| Dependencies     | Phases 1-4 complete |
| Breaking Changes | Yes (see Section 8) |
| New Files        | 0                   |
| Modified Files   | 3                   |

This phase updates validation rules to match the specification exactly, introducing breaking changes.

### 6.2 Modified File: account/validation.go

**Updated ValidateParentAccount:**

```go
// ValidateParentAccount validates the requirements for a Parent account.
//
// Parent accounts MUST have:
//   - A valid (non-zero) public key
//   - A valid DID that matches the public key
//
// Parent accounts MUST NOT have:
//   - Communication channels (stdIn, stdOut, stdErr)
//   - A parent reference
func ValidateParentAccount(pubKey keylib.NeuronPublicKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error {
    const op = "ValidateParentAccount"

    // Existing validations
    if pubKey.IsZero() {
        return errZeroValue(op, "NeuronPublicKey")
    }
    if did == nil {
        return errMissingRequired(op, "DID")
    }
    if err := did.Validate(); err != nil {
        return errValidation(op, "DID validation failed", err)
    }

    // NEW: Parent must NOT have communication channels
    if !stdIn.IsZero() {
        return errProhibitedField(op, "stdIn", "parent accounts must not have communication channels")
    }
    if !stdOut.IsZero() {
        return errProhibitedField(op, "stdOut", "parent accounts must not have communication channels")
    }
    if !stdErr.IsZero() {
        return errProhibitedField(op, "stdErr", "parent accounts must not have communication channels")
    }

    // NEW: Parent must NOT have parent reference
    if !parentPubKey.IsZero() {
        return errProhibitedField(op, "parentPubKey", "parent accounts must not have parent reference")
    }

    return nil
}
```

**Updated ValidateChildAccount:**

```go
// ValidateChildAccount validates the requirements for a Child account.
//
// Child accounts MUST have:
//   - A valid (non-zero) public key
//   - A valid (non-zero) parent public key (different from own key)
//   - ALL 3 communication channels (stdIn, stdOut, stdErr)
//
// Child accounts MUST NOT have:
//   - A DID
func ValidateChildAccount(pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress, did NeuronDID) error {
    const op = "ValidateChildAccount"

    // Existing validations
    if pubKey.IsZero() {
        return errZeroValue(op, "NeuronPublicKey")
    }
    if parentPubKey.IsZero() {
        return errMissingRequired(op, "parent public key")
    }
    if pubKey.Equal(parentPubKey) {
        return errInvalidHierarchy(op, "child public key cannot equal parent public key")
    }

    // NEW: Child MUST have all three communication channels
    if stdIn.IsZero() {
        return errMissingRequired(op, "stdIn (required for child accounts)")
    }
    if stdOut.IsZero() {
        return errMissingRequired(op, "stdOut (required for child accounts)")
    }
    if stdErr.IsZero() {
        return errMissingRequired(op, "stdErr (required for child accounts)")
    }

    // NEW: Child must NOT have DID
    if did != nil {
        return errProhibitedField(op, "DID", "child accounts must not have DID")
    }

    return nil
}
```

**New ValidateSharedAccount:**

```go
// ValidateSharedAccount validates the requirements for a Shared account.
//
// Shared accounts MUST have:
//   - A valid MultisigKey with threshold configuration
//
// Shared accounts MUST NOT have:
//   - A DID
//   - Communication channels (stdIn, stdOut, stdErr)
//   - A parent reference
//   - A single public key (uses MultisigKey instead)
func ValidateSharedAccount(multisigKey *keylib.MultisigKey,
    did NeuronDID, stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey) error {
    const op = "ValidateSharedAccount"

    // Shared account requires MultisigKey
    if multisigKey == nil {
        return errMissingRequired(op, "MultisigKey")
    }
    if err := multisigKey.Validate(); err != nil {
        return errValidation(op, "invalid MultisigKey", err)
    }

    // Shared account must NOT have DID
    if did != nil {
        return errProhibitedField(op, "DID", "shared accounts must not have DID")
    }

    // Shared account must NOT have communication channels
    if !stdIn.IsZero() {
        return errProhibitedField(op, "stdIn", "shared accounts must not have communication channels")
    }
    if !stdOut.IsZero() {
        return errProhibitedField(op, "stdOut", "shared accounts must not have communication channels")
    }
    if !stdErr.IsZero() {
        return errProhibitedField(op, "stdErr", "shared accounts must not have communication channels")
    }

    // Shared account must NOT have parent reference
    if !parentPubKey.IsZero() {
        return errProhibitedField(op, "parentPubKey", "shared accounts must not have parent reference")
    }

    return nil
}
```

### 6.3 Modified File: account/errors.go

**New Error Kind:**

```go
const (
    // ... existing error kinds ...
    ErrKindProhibitedField  // NEW: field set when it should be empty/nil
)

// NEW: Error helper function
func errProhibitedField(op, fieldName, reason string) *AccountError {
    return &AccountError{
        Op:      op,
        Kind:    ErrKindProhibitedField,
        Details: fmt.Sprintf("prohibited field %s: %s", fieldName, reason),
    }
}
```

### 6.4 Modified File: account/builder.go

**Build() Method Updates:**

Update validation calls to pass all required parameters:

```go
case AccountTypeParent:
    if err := ValidateParentAccount(b.publicKey, b.did,
        b.stdIn, b.stdOut, b.stdErr, b.parentPubKey); err != nil {
        return NeuronAccount{}, wrapAccountError(op, ErrKindValidation,
            "parent validation failed", err)
    }

case AccountTypeChild:
    if err := ValidateChildAccount(b.publicKey, b.parentPubKey,
        b.stdIn, b.stdOut, b.stdErr, b.did); err != nil {
        return NeuronAccount{}, wrapAccountError(op, ErrKindValidation,
            "child validation failed", err)
    }
```

---

## 7. Phase 6: Testing and Documentation

### 7.1 Overview

| Attribute        | Value                        |
| ---------------- | ---------------------------- |
| Priority         | Low                          |
| Dependencies     | All previous phases complete |
| Breaking Changes | None                         |
| New Files        | 0                            |
| Modified Files   | 6+                           |

### 7.2 Test Coverage Targets

| Package/File                 | Target Coverage | Focus Areas                             |
| ---------------------------- | --------------- | --------------------------------------- |
| keylib/multisig_key.go       | >90%            | Validation edge cases, immutability     |
| account/ledger_attachment.go | >85%            | State transitions, validation           |
| account/validation.go        | >95%            | All account types, all error conditions |
| account/builder.go           | >85%            | All builders, all With\* methods        |

### 7.3 Test Files to Update

| File                           | New Tests                                                 |
| ------------------------------ | --------------------------------------------------------- |
| `account/account_test.go`      | Tests for new fields, Shared account accessors            |
| `account/account_type_test.go` | Tests for IsShared(), RequiresMultisigKey()               |
| `account/builder_test.go`      | Tests for NewSharedAccountBuilder, financial methods      |
| `account/validation_test.go`   | Tests for updated validation rules, ValidateSharedAccount |
| `keylib/multisig_key_test.go`  | Comprehensive MultisigKey tests (created in Phase 1)      |

### 7.4 Documentation Updates

| File             | Updates                                             |
| ---------------- | --------------------------------------------------- |
| `keylib/doc.go`  | Add MultisigKey section with usage examples         |
| `account/doc.go` | Add Shared account section, update validation rules |
| `README.md`      | Update feature list, add migration section          |

---

## 8. Breaking Changes and Migration

### 8.1 Breaking Changes Summary

| Change                                     | Affected Code                                                       | Error Message                                          |
| ------------------------------------------ | ------------------------------------------------------------------- | ------------------------------------------------------ |
| Parent accounts reject comm channels       | Code that sets stdIn/stdOut/stdErr on Parent                        | "parent accounts must not have communication channels" |
| Child accounts require all 3 comm channels | Code that builds Child without all channels                         | "stdIn/stdOut/stdErr (required for child accounts)"    |
| Child accounts reject DID                  | Code that sets DID on Child accounts                                | "child accounts must not have DID"                     |
| Validation function signatures change      | Code that calls ValidateParentAccount/ValidateChildAccount directly | Compile error (additional parameters)                  |

### 8.2 Migration Strategy

**Recommended Approach: Major Version Bump**

1. **v1.x Release (Current):**
   - Add deprecation warnings in builder methods
   - Log warnings when prohibited fields are set
   - Continue to allow invalid configurations

2. **v2.0 Release (This Plan):**
   - Enforce strict validation
   - Breaking changes take effect
   - Provide migration guide

### 8.3 Migration Guide Template

````markdown
# Migrating to v2.0

## Parent Account Changes

Before (v1.x):

```go
account, _ := NewParentAccountBuilder(pubKey, did).
    WithStdInHedera("0.0.123").  // Was allowed, now prohibited
    Build()
```
````

After (v2.0):

```go
account, _ := NewParentAccountBuilder(pubKey, did).
    // No communication channels for Parent accounts
    Build()
```

## Child Account Changes

Before (v1.x):

```go
account, _ := NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.123").  // Only stdIn was optional
    Build()
```

After (v2.0):

```go
account, _ := NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.123").   // Required
    WithStdOutHedera("0.0.124").  // Required
    WithStdErrHedera("0.0.125").  // Required
    Build()
```

````

---

## 9. Verification Procedures

### 9.1 After Each Phase

Execute the following verification steps:

```bash
# 1. Run all tests with race detector
go test ./keylib/... ./account/... -v -race

# 2. Check test coverage
go test ./keylib/... -coverprofile=keylib.out
go tool cover -func=keylib.out | grep -E "total:|multisig"

go test ./account/... -coverprofile=account.out
go tool cover -func=account.out | grep -E "total:|ledger|validation"

# 3. Verify no vet errors
go vet ./...

# 4. Format code
go fmt ./...

# 5. Tidy dependencies
go mod tidy
````

### 9.2 Final Integration Verification

After all phases complete, verify end-to-end scenarios:

**Scenario 1: Parent Account**

```go
// Should succeed: Parent with DID, no comm channels
account, err := NewParentAccountBuilder(pubKey, did).
    WithCurrencySymbol("HBAR").
    WithCreditBalance(big.NewInt(1000)).
    Build()
// Verify: err == nil, account.IsParent() == true
```

**Scenario 2: Child Account**

```go
// Should succeed: Child with parent ref, all 3 comm channels
account, err := NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.123").
    WithStdOutHedera("0.0.124").
    WithStdErrHedera("0.0.125").
    WithCurrencySymbol("HBAR").
    WithBalanceAllocation(big.NewInt(100)).
    Build()
// Verify: err == nil, account.IsChild() == true
```

**Scenario 3: Shared Account**

```go
// Should succeed: Shared with MultisigKey, no DID, no comm channels
multisigKey, _ := keylib.NewMultisigKey([]keylib.NeuronPublicKey{pk1, pk2, pk3}, 2)
account, err := NewSharedAccountBuilder(multisigKey).
    WithCurrencySymbol("ETH").
    WithBalance(big.NewInt(5000)).
    Build()
// Verify: err == nil, account.IsShared() == true
```

**Scenario 4: Invalid Configurations**

```go
// Should fail: Parent with comm channels
_, err := NewParentAccountBuilder(pubKey, did).
    WithStdInHedera("0.0.123").
    Build()
// Verify: err != nil, error contains "must not have communication channels"

// Should fail: Child without all comm channels
_, err = NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.123").
    Build()
// Verify: err != nil, error contains "required for child accounts"

// Should fail: Shared with DID
_, err = NewSharedAccountBuilder(multisigKey).
    WithDID(did).  // If such method exists, or attempt to set
    Build()
// Verify: err != nil, error contains "must not have DID"
```

---

## 10. Risk Assessment

### 10.1 Risk Matrix

| Risk                                     | Probability | Impact | Mitigation                                                |
| ---------------------------------------- | ----------- | ------ | --------------------------------------------------------- |
| Breaking changes affect downstream users | High        | High   | Major version bump, migration guide, deprecation warnings |
| MultisigKey complexity introduces bugs   | Medium      | Medium | Extensive unit tests, table-driven tests for edge cases   |
| big.Int nil pointer dereferences         | Medium      | Medium | Nil checks in all accessors, defensive programming        |
| Test coverage gaps                       | Low         | Medium | CI coverage gates, review coverage reports                |
| JSON serialization incompatibility       | Low         | High   | Versioned JSON schema, backward-compatible parsing        |

### 10.2 Mitigation Strategies

**For Breaking Changes:**

- Release deprecation warnings in v1.x
- Provide comprehensive migration guide
- Consider feature flag for gradual rollout

**For MultisigKey Complexity:**

- Start with basic 2-of-3 functionality
- Add more complex configurations incrementally
- Extensive property-based testing

**For nil Pointer Safety:**

- All accessor methods check for nil before dereferencing
- Return nil for inappropriate account types
- Document nil-safety in method comments

---

## Appendix A: Complete File Change Summary

### New Files

| File                                  | Phase | Purpose                  |
| ------------------------------------- | ----- | ------------------------ |
| `keylib/multisig_key.go`              | 1     | MultisigKey type         |
| `keylib/multisig_key_test.go`         | 1     | MultisigKey tests        |
| `account/ledger_attachment.go`        | 4     | LedgerAttachment type    |
| `account/ledger_attachment_test.go`   | 4     | LedgerAttachment tests   |
| `account/ledger_verification.go`      | 4     | LedgerVerifier interface |
| `account/ledger_verification_test.go` | 4     | Verification tests       |

### Modified Files

| File                      | Phases | Changes                                                        |
| ------------------------- | ------ | -------------------------------------------------------------- |
| `account/account_type.go` | 1      | Add AccountTypeShared, new methods                             |
| `account/account.go`      | 2      | Add fields, accessors, JSON serialization                      |
| `account/builder.go`      | 2, 3   | Add NewSharedAccountBuilder, financial methods, update Build() |
| `account/validation.go`   | 5      | Update validation functions, add ValidateSharedAccount         |
| `account/errors.go`       | 5      | Add ErrKindProhibitedField                                     |
| `keylib/errors.go`        | 1      | Add ErrKindInvalidThreshold                                    |
| `keylib/doc.go`           | 6      | Add MultisigKey documentation                                  |

---

_Document maintained by the development team. Update as implementation progresses._
