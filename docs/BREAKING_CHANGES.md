# Breaking Changes - NeuronAccount Specification Alignment

This document describes breaking changes introduced to align the Go implementation with the NeuronAccount technical specification.

## Version Summary

**Affected packages:** `account`, `keylib`

**Impact Level:** HIGH - Compilation errors expected for existing consumers

---

## Breaking Changes

### 1. Parent Account Validation: Communication Channels Prohibited

**Previous behavior:** Parent accounts could optionally have communication channels (stdIn, stdOut, stdErr).

**New behavior:** Parent accounts must NOT have communication channels. Validation rejects any Parent account with non-zero comm channels.

**Error returned:** `ErrKindProhibitedField` with message `"field stdIn is prohibited for Parent accounts"` (or stdOut/stdErr).

#### Migration

```go
// BEFORE (v1.x) - This worked
parent, err := NewParentAccountBuilder(pubKey, did).
    WithStdIn(stdIn).     // ← Now prohibited
    WithStdOut(stdOut).   // ← Now prohibited
    WithStdErr(stdErr).   // ← Now prohibited
    Build()

// AFTER (v2.x) - Remove communication channels
parent, err := NewParentAccountBuilder(pubKey, did).Build()
```

---

### 2. Child Account Validation: All 3 Communication Channels Required

**Previous behavior:** Child accounts could have zero, one, two, or three communication channels.

**New behavior:** Child accounts must have ALL three communication channels (stdIn, stdOut, stdErr).

**Error returned:** `ErrKindMissingRequired` with message `"required field stdIn is missing"` (or stdOut/stdErr).

#### Migration

```go
// BEFORE (v1.x) - Partial channels worked
child, err := NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.111").  // Only one channel
    Build()

// AFTER (v2.x) - All three channels required
child, err := NewChildAccountBuilder(pubKey, parentPubKey).
    WithStdInHedera("0.0.111").
    WithStdOutHedera("0.0.222").
    WithStdErrHedera("0.0.333").
    Build()
```

---

### 3. Child Account Validation: DID Prohibited

**Previous behavior:** Child accounts could optionally have a DID.

**New behavior:** Child accounts must NOT have a DID. DIDs are reserved for Parent accounts only.

**Error returned:** `ErrKindProhibitedField` with message `"field DID is prohibited for Child accounts"`.

#### Migration

If your code attempted to set DID on Child accounts (via internal construction), remove that logic. The builder doesn't expose `WithDID` for Child accounts, so this primarily affects direct struct construction.

---

### 4. Validation Function Signature Changes

#### `ValidateParentAccount`

```go
// BEFORE (v1.x)
func ValidateParentAccount(pubKey keylib.NeuronPublicKey, did NeuronDID) error

// AFTER (v2.x)
func ValidateParentAccount(
    pubKey keylib.NeuronPublicKey,
    did NeuronDID,
    stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey,
) error
```

#### `ValidateChildAccount`

```go
// BEFORE (v1.x)
func ValidateChildAccount(
    pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress,
) error

// AFTER (v2.x)
func ValidateChildAccount(
    pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress,
    did NeuronDID,
) error
```

#### Migration

Update all call sites to pass the additional parameters:

```go
// BEFORE
err := ValidateParentAccount(pubKey, did)

// AFTER
err := ValidateParentAccount(pubKey, did, account.StdIn(), account.StdOut(), account.StdErr(), account.ParentPublicKey())
```

---

### 5. AccountValidator Method Signature Changes

#### `ValidateParentRequirements`

```go
// BEFORE (v1.x)
func (v *AccountValidator) ValidateParentRequirements(
    pubKey keylib.NeuronPublicKey,
    did NeuronDID,
) *AccountValidator

// AFTER (v2.x)
func (v *AccountValidator) ValidateParentRequirements(
    pubKey keylib.NeuronPublicKey,
    did NeuronDID,
    stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey,
) *AccountValidator
```

#### `ValidateChildRequirements`

```go
// BEFORE (v1.x)
func (v *AccountValidator) ValidateChildRequirements(
    pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress,
) *AccountValidator

// AFTER (v2.x)
func (v *AccountValidator) ValidateChildRequirements(
    pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress,
    did NeuronDID,
) *AccountValidator
```

---

### 6. New Error Kind: `ErrKindProhibitedField`

**Added:** `account.ErrKindProhibitedField` for detecting fields that must NOT be present.

```go
// Check for prohibited field errors
var accErr *account.AccountError
if errors.As(err, &accErr) && accErr.Kind == account.ErrKindProhibitedField {
    // Handle prohibited field error
}
```

---

### 7. New Account Type: `AccountTypeShared`

**Added:** `AccountTypeShared` for M-of-N threshold multisig accounts.

**Impact:** `AccountType.IsValid()` now returns true for `AccountTypeShared` (value 3).

```go
// New type value
const AccountTypeShared AccountType = 3

// New helper methods
func (t AccountType) IsShared() bool
func (t AccountType) RequiresMultisigKey() bool
func (t AccountType) ProhibitsCommChannels() bool
```

---

### 8. New Validation Function: `ValidateSharedAccount`

```go
func ValidateSharedAccount(
    multisigKey *keylib.MultisigKey,
    did NeuronDID,
    stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey,
) error
```

---

## Affected Consumers

| Consumer Type | Impact |
|--------------|--------|
| Code calling `ValidateParentAccount` directly | **Compilation error** - signature changed |
| Code calling `ValidateChildAccount` directly | **Compilation error** - signature changed |
| Code using `AccountValidator.ValidateParentRequirements` | **Compilation error** - signature changed |
| Code using `AccountValidator.ValidateChildRequirements` | **Compilation error** - signature changed |
| Code creating Parent accounts with comm channels | **Runtime error** - validation fails |
| Code creating Child accounts with <3 comm channels | **Runtime error** - validation fails |
| Code switching on `AccountType` values | **Logic error** - new `AccountTypeShared` case |

---

## Upgrade Checklist

Use this checklist to verify your codebase is updated:

- [ ] **Search for `ValidateParentAccount(`** - Update all call sites with new parameters
- [ ] **Search for `ValidateChildAccount(`** - Update all call sites with new parameters
- [ ] **Search for `ValidateParentRequirements(`** - Update all call sites with new parameters
- [ ] **Search for `ValidateChildRequirements(`** - Update all call sites with new parameters
- [ ] **Search for `NewParentAccountBuilder(`** - Remove any `.WithStdIn/Out/Err()` calls
- [ ] **Search for `NewChildAccountBuilder(`** - Ensure all three `.WithStdIn/Out/ErrHedera()` calls present
- [ ] **Search for `switch` on `AccountType`** - Add `case AccountTypeShared:` handling
- [ ] **Search for `account.ErrKind`** - Consider handling new `ErrKindProhibitedField`
- [ ] **Run `go build ./...`** - Fix any compilation errors
- [ ] **Run `go test ./...`** - Fix any test failures

---

## Quick Migration Script

Run these grep commands to identify affected code:

```bash
# Find direct validation calls
grep -rn "ValidateParentAccount(" --include="*.go"
grep -rn "ValidateChildAccount(" --include="*.go"
grep -rn "ValidateParentRequirements(" --include="*.go"
grep -rn "ValidateChildRequirements(" --include="*.go"

# Find Parent accounts with comm channels
grep -rn "NewParentAccountBuilder" --include="*.go" | grep -E "WithStd(In|Out|Err)"

# Find Child accounts (verify all have 3 channels)
grep -rn "NewChildAccountBuilder" --include="*.go"

# Find AccountType switches
grep -rn "switch.*AccountType\|case AccountType" --include="*.go"
```

---

## Rationale

These changes align the Go implementation with the NeuronAccount technical specification:

- **FR-012:** Parent accounts represent root identity without operational endpoints
- **FR-013:** Child accounts are operational endpoints requiring full communication setup
- **FR-016:** Shared accounts use MultisigKey for threshold signing scenarios
- **FR-019:** Semantic consistency between account types and their allowed/required fields

---

## Support

If you encounter issues during migration, file an issue at the project repository with:
1. The error message received
2. The code pattern you're trying to migrate
3. The version you're upgrading from
