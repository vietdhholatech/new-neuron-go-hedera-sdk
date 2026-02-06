# NeuronAccount Implementation Gap Analysis

**Document Version:** 1.0
**Analysis Date:** 2026-01-27
**Specification Reference:** [docs/neuron-sdk-spec/NeuronAccount.md](neuron-sdk-spec/NeuronAccount.md)
**Implementation Version:** v1.3.0

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Account Types](#2-account-types)
3. [NeuronAccount Data Structure](#3-neuronaccount-data-structure)
4. [Functional Requirements Analysis](#4-functional-requirements-analysis)
5. [Key Entities](#5-key-entities)
6. [Validation Rules](#6-validation-rules)
7. [Specification Conflicts](#7-specification-conflicts)
8. [Implementation Roadmap](#8-implementation-roadmap)

---

## 1. Executive Summary

This document provides a detailed technical comparison between the **NeuronAccount Module Technical Specification** and the **current Go implementation**. The analysis identifies implementation gaps, specification conflicts, and missing features that must be addressed to achieve full specification compliance.

### 1.1 Overall Compliance Status

| Category                   | Compliance Level | Assessment                                                                 |
| -------------------------- | ---------------- | -------------------------------------------------------------------------- |
| Core Identity Management   | 95%              | Fully functional; public key derivation, PeerID, and EVMAddress working    |
| Account Type System        | 67%              | Parent and Child implemented; Shared account type not implemented          |
| Communication Channels     | 80%              | CommAddress system complete; validation rules differ from specification    |
| Financial/Balance Features | 0%               | No balance fields, currency symbols, or financial tracking implemented     |
| Ledger Integration         | 0%               | LedgerAttachment entity and verification methods not implemented           |
| Backend Plugin System      | 100%             | Registry pattern fully implemented with Hedera, Kafka, and Custom backends |

### 1.2 Critical Gaps Summary

The following features specified in the requirements document are completely absent from the implementation:

1. **Shared Account Type** - Multisig threshold accounts requiring M-of-N signatures
2. **MultisigKey** - Cryptographic key type for threshold signing operations
3. **LedgerAttachment** - Entity for linking accounts to settlement infrastructure
4. **Financial Fields** - Currency symbols, credit balances, and balance allocations
5. **Ledger Verification** - Methods to prove account existence and key ownership on ledgers

---

## 2. Account Types

### 2.1 Specification Requirements

The specification mandates four account type values:

| Type Value | Constant Name          | Description                                                                                         |
| ---------- | ---------------------- | --------------------------------------------------------------------------------------------------- |
| 0          | AccountTypeUnspecified | Invalid state; must be rejected during validation                                                   |
| 1          | AccountTypeParent      | Root identity account with DID document, credit balance capability, and no communication channels   |
| 2          | AccountTypeChild       | Subordinate account with parent reference, balance allocation, and mandatory communication channels |
| 3          | AccountTypeShared      | Multisig threshold account using MultisigKey instead of single public key; no DID or comm channels  |

### 2.2 Current Implementation

**File:** `account/account_type.go` (Lines 16-34)

The implementation defines only three constants:

```go
type AccountType int

const (
    AccountTypeUnspecified AccountType = iota  // Value: 0
    AccountTypeParent                          // Value: 1
    AccountTypeChild                           // Value: 2
)
```

**Missing:** `AccountTypeShared` with value 3 is not defined.

### 2.3 Detailed Gap Analysis

| Requirement                      | Specification                               | Implementation                        | Status          |
| -------------------------------- | ------------------------------------------- | ------------------------------------- | --------------- |
| Parent account type (value 1)    | Must be defined and validatable             | Defined as `AccountTypeParent`        | Implemented     |
| Child account type (value 2)     | Must be defined and validatable             | Defined as `AccountTypeChild`         | Implemented     |
| Shared account type (value 3)    | Must be defined for multisig accounts       | Not defined                           | Not Implemented |
| Unspecified rejection (value 0)  | Must reject during validation               | `IsValid()` returns false for value 0 | Implemented     |
| String representation            | Human-readable type names                   | `String()` method returns type names  | Implemented     |
| Type-specific requirement checks | `RequiresDID()`, `RequiresParent()` methods | Implemented for Parent and Child only | Partial         |

### 2.4 Required Implementation Changes

The following changes are required in `account/account_type.go`:

```go
const (
    AccountTypeUnspecified AccountType = iota
    AccountTypeParent
    AccountTypeChild
    AccountTypeShared  // Add this constant
)

// Update String() method
func (t AccountType) String() string {
    switch t {
    case AccountTypeUnspecified:
        return "Unspecified"
    case AccountTypeParent:
        return "Parent"
    case AccountTypeChild:
        return "Child"
    case AccountTypeShared:
        return "Shared"
    default:
        return fmt.Sprintf("Unknown(%d)", int(t))
    }
}

// Update IsValid() method
func (t AccountType) IsValid() bool {
    return t == AccountTypeParent || t == AccountTypeChild || t == AccountTypeShared
}

// Add new method
func (t AccountType) IsShared() bool {
    return t == AccountTypeShared
}

// Add new method
func (t AccountType) RequiresMultisigKey() bool {
    return t == AccountTypeShared
}
```

---

## 3. NeuronAccount Data Structure

### 3.1 Specification Requirements

The specification defines the following field requirements organized by account type:

**Universal Fields (All Account Types):**

| Field            | Type                   | Description                                |
| ---------------- | ---------------------- | ------------------------------------------ |
| publicKey        | keylib.NeuronPublicKey | Root identity key (Parent/Child only)      |
| multisigKey      | keylib.MultisigKey     | Threshold signing key (Shared only)        |
| peerID           | keylib.PeerID          | Derived libp2p peer identifier             |
| evmAddress       | keylib.EVMAddress      | Derived Ethereum-compatible address        |
| accountType      | AccountType            | Discriminator for account behavior         |
| currencySymbol   | string                 | Currency identifier for balance tracking   |
| ledgerAttachment | \*LedgerAttachment     | Optional link to settlement infrastructure |

**Parent Account Fields:**

| Field         | Type      | Requirement | Description                                 |
| ------------- | --------- | ----------- | ------------------------------------------- |
| did           | NeuronDID | Required    | Decentralized identifier document           |
| creditBalance | \*big.Int | Required    | Primary financial account balance           |
| stdIn         | -         | Prohibited  | Parent accounts must not have comm channels |
| stdOut        | -         | Prohibited  | Parent accounts must not have comm channels |
| stdErr        | -         | Prohibited  | Parent accounts must not have comm channels |

**Child Account Fields:**

| Field             | Type                   | Requirement | Description                              |
| ----------------- | ---------------------- | ----------- | ---------------------------------------- |
| parentPubKey      | keylib.NeuronPublicKey | Required    | Reference to parent account's public key |
| balanceAllocation | \*big.Int              | Required    | Allocated funds for operational expenses |
| stdIn             | CommAddress            | Required    | Inbound message channel                  |
| stdOut            | CommAddress            | Required    | Outbound message channel                 |
| stdErr            | CommAddress            | Required    | Error/diagnostic channel                 |
| reachableAddrs    | ReachableAddrs         | Optional    | P2P connection endpoints                 |

**Shared Account Fields:**

| Field       | Type               | Requirement | Description                                 |
| ----------- | ------------------ | ----------- | ------------------------------------------- |
| multisigKey | keylib.MultisigKey | Required    | Threshold key configuration (M-of-N)        |
| balance     | \*big.Int          | Required    | Multisig-controlled balance                 |
| did         | -                  | Prohibited  | Shared accounts must not have DID           |
| stdIn       | -                  | Prohibited  | Shared accounts must not have comm channels |

### 3.2 Current Implementation

**File:** `account/account.go` (Lines 20-64)

```go
type NeuronAccount struct {
    // Identity fields
    publicKey      keylib.NeuronPublicKey   // Implemented
    peerID         keylib.PeerID            // Implemented
    evmAddress     keylib.EVMAddress        // Implemented

    // Hierarchy fields
    accountType    AccountType              // Implemented
    did            NeuronDID                // Implemented
    parentPubKey   keylib.NeuronPublicKey   // Implemented

    // Communication fields
    stdIn          CommAddress              // Implemented
    stdOut         CommAddress              // Implemented
    stdErr         CommAddress              // Implemented
    reachableAddrs ReachableAddrs           // Implemented
}
```

### 3.3 Field-by-Field Gap Analysis

| Field             | Spec Requirement                                 | Current State       | Gap Type |
| ----------------- | ------------------------------------------------ | ------------------- | -------- |
| publicKey         | Required for Parent/Child                        | Present             | None     |
| peerID            | Derived from publicKey                           | Present, derived    | None     |
| evmAddress        | Derived from publicKey                           | Present, derived    | None     |
| accountType       | Required, values 0-3                             | Present, values 0-2 | Partial  |
| did               | Required for Parent only                         | Present             | None     |
| parentPubKey      | Required for Child only                          | Present             | None     |
| stdIn             | Required for Child, prohibited for Parent/Shared | Present, optional   | Conflict |
| stdOut            | Required for Child, prohibited for Parent/Shared | Present, optional   | Conflict |
| stdErr            | Required for Child, prohibited for Parent/Shared | Present, optional   | Conflict |
| reachableAddrs    | Optional for Child                               | Present             | None     |
| multisigKey       | Required for Shared                              | Not present         | Missing  |
| currencySymbol    | Required for all                                 | Not present         | Missing  |
| creditBalance     | Required for Parent                              | Not present         | Missing  |
| balanceAllocation | Required for Child                               | Not present         | Missing  |
| balance           | Required for Shared                              | Not present         | Missing  |
| ledgerAttachment  | Optional for all                                 | Not present         | Missing  |

### 3.4 Required Struct Modifications

```go
type NeuronAccount struct {
    // === Identity Layer ===
    publicKey      keylib.NeuronPublicKey   // For Parent/Child accounts
    multisigKey    *keylib.MultisigKey      // NEW: For Shared accounts (mutually exclusive with publicKey)
    peerID         keylib.PeerID
    evmAddress     keylib.EVMAddress

    // === Hierarchy Layer ===
    accountType    AccountType
    did            NeuronDID
    parentPubKey   keylib.NeuronPublicKey

    // === Communication Layer ===
    stdIn          CommAddress
    stdOut         CommAddress
    stdErr         CommAddress
    reachableAddrs ReachableAddrs

    // === Financial Layer (NEW) ===
    currencySymbol    string                // NEW: Currency identifier
    creditBalance     *big.Int              // NEW: For Parent accounts
    balanceAllocation *big.Int              // NEW: For Child accounts
    balance           *big.Int              // NEW: For Shared accounts

    // === Ledger Integration Layer (NEW) ===
    ledgerAttachment  *LedgerAttachment     // NEW: Optional ledger link
}
```

---

## 4. Functional Requirements Analysis

### 4.1 Fully Implemented Requirements

| Requirement | Description                                        | Implementation Location                     | Notes                                      |
| ----------- | -------------------------------------------------- | ------------------------------------------- | ------------------------------------------ |
| FR-003      | Three communication channels (stdIn/stdOut/stdErr) | `account/comm_address.go`                   | CommAddress type with kind/locator pattern |
| FR-004      | Backend-specific validation rules                  | `backend_hedera.go`, `backend_kafka.go`     | Each backend implements ValidateLocator()  |
| FR-005      | Extensible backend plugin system                   | `account/registry.go`                       | Thread-safe registry with init() pattern   |
| FR-008      | Derive PeerID and EVMAddress from public key       | `account/builder.go:283-287`                | Derivation occurs in Build() method        |
| FR-009      | P2P endpoints with libp2p multiaddresses           | `account/reachable.go`                      | Full multiaddr parsing and validation      |
| FR-010      | PeerID consistency enforcement                     | `account/reachable.go:ValidateForAccount()` | Validates extracted PeerID matches account |
| FR-011      | Fluent builder API for account construction        | `account/builder.go`                        | Method chaining with error accumulation    |
| FR-012      | DID:key format identifier generation               | `account/didkey/didkey.go`                  | Multicodec 0xe7,0x01 + base58btc encoding  |
| FR-014      | Clear, actionable error messages                   | `account/errors.go`                         | AccountError with Op, Kind, Details, Err   |
| FR-015      | JSON serialization and deserialization             | `account/account.go:MarshalJSON()`          | Custom JSON handling with field omission   |

### 4.2 Partially Implemented Requirements

| Requirement | Description                                      | Gap Description                                                                   |
| ----------- | ------------------------------------------------ | --------------------------------------------------------------------------------- |
| FR-001      | Parent account creation with all features        | Missing: creditBalance, ledgerAttachment; Conflict: allows communication channels |
| FR-002      | Child account creation with all features         | Missing: balanceAllocation, ledgerAttachment; Conflict: comm channels optional    |
| FR-006      | Validate Parent accounts have no comm channels   | Not enforced: validation does not reject comm channels on Parent accounts         |
| FR-007      | Validate Child accounts have all 3 comm channels | Not enforced: all communication channels are optional in validation               |
| FR-013      | AccountType enumeration with values 0-3          | Only values 0-2 defined; Shared (3) is missing                                    |

### 4.3 Not Implemented Requirements

| Requirement | Description                                                | Required Implementation                                                |
| ----------- | ---------------------------------------------------------- | ---------------------------------------------------------------------- |
| FR-007a     | Validate Shared accounts have MultisigKey                  | Add ValidateSharedAccount() function checking for valid MultisigKey    |
| FR-013a     | Three valid account types with distinct validation         | Add AccountTypeShared constant and type-specific validation rules      |
| FR-016      | Credit balance management for Parent, allocation for Child | Add balance fields with getter/setter methods and builder support      |
| FR-017      | Parent-child relationship verification against ledger      | Implement VerifyParentChildRelationship() with ledger/registry queries |
| FR-018      | Ledger attachment for accounts                             | Create LedgerAttachment struct with attachment state management        |
| FR-019      | Prove ledger attachment validity                           | Implement ProveAttachment() verifying account existence and key match  |
| FR-020      | Currency symbol identification for accounts                | Add currencySymbol field with validation for supported currencies      |
| FR-021      | Shared accounts with MultisigKey and ledger attachment     | Complete Shared account implementation with all specified features     |

---

## 5. Key Entities

### 5.1 Implemented Entities

| Entity          | File Location              | Implementation Status | Compliance Notes                          |
| --------------- | -------------------------- | --------------------- | ----------------------------------------- |
| NeuronAccount   | `account/account.go`       | Partial               | Missing financial and ledger fields       |
| CommAddress     | `account/comm_address.go`  | Complete              | Kind/locator pattern fully implemented    |
| CommAddressKind | `account/comm_address.go`  | Complete              | Dynamic validation via backend registry   |
| Backend         | `account/backend.go`       | Complete              | Interface with 5 required methods         |
| BackendMetadata | `account/backend.go`       | Complete              | Metadata struct for backend introspection |
| ReachableAddr   | `account/reachable.go`     | Complete              | Multiaddr parsing with PeerID extraction  |
| ReachableAddrs  | `account/reachable.go`     | Complete              | Collection type with validation methods   |
| NeuronDID       | `account/did.go`           | Complete              | Interface for DID implementations         |
| DIDKey          | `account/didkey/didkey.go` | Complete              | did:key method with secp256k1 support     |
| AccountError    | `account/errors.go`        | Complete              | Structured error with 9 error kinds       |
| TopicMessage    | `account/message.go`       | Complete              | Message structure for topic communication |
| TopicTechnology | `account/topic.go`         | Complete              | Enum for messaging technology types       |

### 5.2 Missing Entities

#### 5.2.1 LedgerAttachment

**Specification Requirement:**

The LedgerAttachment entity represents the link between a NeuronAccount and settlement infrastructure (blockchain or ledger).

**Required Fields:**

| Field              | Type   | Description                                              |
| ------------------ | ------ | -------------------------------------------------------- |
| ledgerIdentifier   | string | Ledger name (e.g., "ethereum-mainnet", "hedera-mainnet") |
| attachedAddress    | string | Ethereum address or ledger-specific derived address      |
| attachmentState    | enum   | Current state: Detached, Attached, Verified              |
| verificationStatus | enum   | Verification result: None, Pending, Verified, Failed     |

**Required Methods:**

| Method            | Description                                                     |
| ----------------- | --------------------------------------------------------------- |
| IsAttached()      | Returns true if account is linked to a ledger                   |
| Verify()          | Verifies account exists on ledger and key ownership matches     |
| ProveOwnership()  | Generates proof that Neuron key controls the ledger account     |
| VerifySemantics() | Ensures public key derivation matches between Neuron and ledger |

#### 5.2.2 MultisigKey (in keylib package)

**Specification Requirement:**

The MultisigKey type represents a threshold signing configuration for Shared accounts.

**Required Fields:**

| Field      | Type              | Description                               |
| ---------- | ----------------- | ----------------------------------------- |
| publicKeys | []NeuronPublicKey | List of participating public keys         |
| threshold  | int               | Minimum signatures required (M in M-of-N) |
| total      | int               | Total number of keys (N in M-of-N)        |

**Required Methods:**

| Method          | Description                                       |
| --------------- | ------------------------------------------------- |
| Validate()      | Ensures threshold <= total and all keys are valid |
| IsZero()        | Returns true if MultisigKey is uninitialized      |
| ContainsKey()   | Checks if a specific public key is in the set     |
| DeriveAddress() | Derives a deterministic address from the key set  |

### 5.3 New Files Required

| File Path                           | Purpose                                                   |
| ----------------------------------- | --------------------------------------------------------- |
| `keylib/multisig_key.go`            | MultisigKey type definition and methods                   |
| `keylib/multisig_key_test.go`       | Unit tests for MultisigKey                                |
| `account/ledger_attachment.go`      | LedgerAttachment entity definition                        |
| `account/ledger_attachment_test.go` | Unit tests for LedgerAttachment                           |
| `account/ledger_verification.go`    | Verification interface and implementations                |
| `account/builder_shared.go`         | NewSharedAccountBuilder function (optional separate file) |

---

## 6. Validation Rules

### 6.1 Current Validation Implementation

**Files:** `account/account.go` (Lines 152-208), `account/validation.go`

The current implementation provides the following validation:

| Validation Rule               | Implementation                                          | Triggered By            |
| ----------------------------- | ------------------------------------------------------- | ----------------------- |
| Public key non-zero           | `pubKey.IsZero()` check                                 | `Validate()`, `Build()` |
| Account type valid            | `accountType.IsValid()` check                           | `Validate()`, `Build()` |
| DID present for Parent        | `did == nil` check in ValidateParentAccount()           | `Build()`               |
| DID validates                 | `did.Validate()` call                                   | `Build()`               |
| DID matches public key        | `ValidateDIDMatchesKey()` call                          | `Build()`               |
| Parent key present for Child  | `parentPubKey.IsZero()` check in ValidateChildAccount() | `Build()`               |
| Child key differs from parent | `pubKey.Equal(parentPubKey)` check                      | `Build()`               |
| CommAddress format valid      | `addr.Validate()` call                                  | `Validate()` (if set)   |
| ReachableAddr PeerID match    | `ValidateForAccount()` call                             | `Build()`, `Validate()` |

### 6.2 Specification vs Implementation Comparison

| Validation Rule                     | Specification Requirement        | Current Implementation   | Compliance    |
| ----------------------------------- | -------------------------------- | ------------------------ | ------------- |
| Parent: public key required         | Must be non-zero                 | Validated                | Compliant     |
| Parent: DID required                | Must be present and valid        | Validated                | Compliant     |
| Parent: DID matches key             | DID must derive from public key  | Validated                | Compliant     |
| Parent: no communication channels   | stdIn/stdOut/stdErr must be zero | Not validated            | Non-compliant |
| Parent: no parent reference         | parentPubKey must be zero        | Not validated            | Non-compliant |
| Child: public key required          | Must be non-zero                 | Validated                | Compliant     |
| Child: parent reference required    | Must be non-zero                 | Validated                | Compliant     |
| Child: key differs from parent      | Must not equal parent key        | Validated                | Compliant     |
| Child: all 3 comm channels required | stdIn/stdOut/stdErr must be set  | Not validated (optional) | Non-compliant |
| Child: no DID                       | DID must be nil                  | Not validated            | Non-compliant |
| Shared: MultisigKey required        | Must be present and valid        | Not implemented          | Missing       |
| Shared: no DID                      | DID must be nil                  | Not implemented          | Missing       |
| Shared: no communication channels   | stdIn/stdOut/stdErr must be zero | Not implemented          | Missing       |
| Shared: no parent reference         | parentPubKey must be zero        | Not implemented          | Missing       |

### 6.3 Required Validation Additions

#### 6.3.1 Updated ValidateParentAccount

```go
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

    // NEW: Parent must not have communication channels
    if !stdIn.IsZero() || !stdOut.IsZero() || !stdErr.IsZero() {
        return errInvalidAccount(op, "parent accounts must not have communication channels")
    }

    // NEW: Parent must not have parent reference
    if !parentPubKey.IsZero() {
        return errInvalidAccount(op, "parent accounts must not have parent reference")
    }

    return nil
}
```

#### 6.3.2 Updated ValidateChildAccount

```go
func ValidateChildAccount(pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress, did NeuronDID) error {
    const op = "ValidateChildAccount"

    if pubKey.IsZero() {
        return errZeroValue(op, "NeuronPublicKey")
    }

    if parentPubKey.IsZero() {
        return errMissingRequired(op, "parent public key")
    }

    if pubKey.Equal(parentPubKey) {
        return errInvalidHierarchy(op, "child public key cannot equal parent public key")
    }

    // NEW: Child must have all three communication channels
    if stdIn.IsZero() {
        return errMissingRequired(op, "stdIn is required for child accounts")
    }
    if stdOut.IsZero() {
        return errMissingRequired(op, "stdOut is required for child accounts")
    }
    if stdErr.IsZero() {
        return errMissingRequired(op, "stdErr is required for child accounts")
    }

    // NEW: Child must not have DID
    if did != nil {
        return errInvalidAccount(op, "child accounts must not have DID")
    }

    return nil
}
```

#### 6.3.3 New ValidateSharedAccount

```go
func ValidateSharedAccount(multisigKey *keylib.MultisigKey,
    did NeuronDID, stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey) error {
    const op = "ValidateSharedAccount"

    // Shared account requires MultisigKey
    if multisigKey == nil {
        return errMissingRequired(op, "MultisigKey is required for shared accounts")
    }

    if err := multisigKey.Validate(); err != nil {
        return errValidation(op, "invalid MultisigKey", err)
    }

    // Shared account must not have DID
    if did != nil {
        return errInvalidAccount(op, "shared accounts must not have DID")
    }

    // Shared account must not have communication channels
    if !stdIn.IsZero() || !stdOut.IsZero() || !stdErr.IsZero() {
        return errInvalidAccount(op, "shared accounts must not have communication channels")
    }

    // Shared account must not have parent reference
    if !parentPubKey.IsZero() {
        return errInvalidAccount(op, "shared accounts must not have parent reference")
    }

    return nil
}
```

---

## 7. Specification Conflicts

### 7.1 Communication Channel Constraints

The specification defines strict rules about which account types may have communication channels. The current implementation does not enforce these constraints.

| Account Type | Specification Constraint               | Current Implementation    | Required Change                          |
| ------------ | -------------------------------------- | ------------------------- | ---------------------------------------- |
| Parent       | Must NOT have stdIn, stdOut, stdErr    | Allows all three channels | Add validation to reject non-zero values |
| Child        | Must have ALL of stdIn, stdOut, stdErr | All three are optional    | Add validation to require all three      |
| Shared       | Must NOT have stdIn, stdOut, stdErr    | Type not implemented      | Implement with rejection validation      |

**Impact Assessment:**

This conflict affects API compatibility. Existing code that creates Parent accounts with communication channels will fail validation after the fix. A migration strategy or deprecation period may be needed.

### 7.2 DID Constraints

The specification defines strict rules about DID presence based on account type.

| Account Type | Specification Constraint | Current Implementation             | Required Change              |
| ------------ | ------------------------ | ---------------------------------- | ---------------------------- |
| Parent       | DID required             | Validated in ValidateParentAccount | None                         |
| Child        | DID prohibited           | Not validated                      | Add validation to reject DID |
| Shared       | DID prohibited           | Type not implemented               | Implement with rejection     |

### 7.3 Financial Field Absence

The specification requires financial tracking capabilities that are completely absent:

| Feature            | Account Type | Specification Requirement           | Current State   |
| ------------------ | ------------ | ----------------------------------- | --------------- |
| Currency Symbol    | All          | Required field identifying currency | Not implemented |
| Credit Balance     | Parent       | Primary account balance for revenue | Not implemented |
| Balance Allocation | Child        | Operational expense allocation      | Not implemented |
| Balance            | Shared       | Multisig-controlled balance         | Not implemented |

### 7.4 Ledger Integration Absence

The specification requires ledger attachment and verification capabilities:

| Feature                 | Specification Requirement                               | Current State   |
| ----------------------- | ------------------------------------------------------- | --------------- |
| Ledger Attachment       | Link account to settlement infrastructure               | Not implemented |
| Account Existence Proof | Verify account exists on specified ledger               | Not implemented |
| Key Ownership Proof     | Prove Neuron key controls ledger account                | Not implemented |
| Semantic Verification   | Ensure key derivation matches between Neuron and ledger | Not implemented |

---

## 8. Implementation Roadmap

### Phase 1: Core Type Extensions

**Priority:** High
**Dependencies:** None
**Files Affected:** `account/account_type.go`, `keylib/multisig_key.go` (new), `account/account.go`

**Tasks:**

1. Add `AccountTypeShared = 3` constant to account_type.go
2. Update `String()`, `IsValid()` methods for new type
3. Add `IsShared()`, `RequiresMultisigKey()` methods
4. Create `keylib/multisig_key.go` with MultisigKey type
5. Implement MultisigKey validation and utility methods
6. Add `multisigKey` field to NeuronAccount struct
7. Update NeuronAccount accessors for new field

### Phase 2: Financial Fields

**Priority:** High
**Dependencies:** Phase 1
**Files Affected:** `account/account.go`, `account/builder.go`

**Tasks:**

1. Add `currencySymbol` field to NeuronAccount
2. Add `creditBalance` field for Parent accounts
3. Add `balanceAllocation` field for Child accounts
4. Add `balance` field for Shared accounts
5. Create accessor methods for all new fields
6. Update builder with financial field setters
7. Update JSON serialization to include new fields

### Phase 3: Shared Account Builder

**Priority:** High
**Dependencies:** Phase 1, Phase 2
**Files Affected:** `account/builder.go` or `account/builder_shared.go` (new)

**Tasks:**

1. Create `NewSharedAccountBuilder()` constructor
2. Accept MultisigKey instead of single publicKey
3. Implement WithBalance() method
4. Implement WithCurrencySymbol() method
5. Add Shared-specific validation in Build()
6. Add MustBuild() variant for Shared accounts

### Phase 4: Ledger Integration

**Priority:** Medium
**Dependencies:** Phase 1
**Files Affected:** `account/ledger_attachment.go` (new), `account/ledger_verification.go` (new)

**Tasks:**

1. Define LedgerAttachment struct with required fields
2. Define AttachmentState and VerificationStatus enums
3. Implement IsAttached(), GetAttachment() methods
4. Define LedgerVerifier interface
5. Implement basic verification methods
6. Add ledgerAttachment field to NeuronAccount
7. Update builder with ledger attachment methods

### Phase 5: Validation Alignment

**Priority:** Medium
**Dependencies:** Phase 1, Phase 2, Phase 3
**Files Affected:** `account/validation.go`, `account/account.go`, `account/builder.go`

**Tasks:**

1. Update ValidateParentAccount to reject comm channels
2. Update ValidateChildAccount to require all 3 comm channels
3. Update ValidateChildAccount to reject DID
4. Create ValidateSharedAccount function
5. Update NeuronAccount.Validate() for all account types
6. Update builder Build() method with new validation
7. Add comprehensive tests for all validation rules

### Phase 6: Testing and Documentation

**Priority:** Low
**Dependencies:** All previous phases
**Files Affected:** Test files, documentation

**Tasks:**

1. Add unit tests for MultisigKey
2. Add unit tests for LedgerAttachment
3. Add unit tests for Shared account validation
4. Add integration tests for account type interactions
5. Update docs/neuron-account-specification.md
6. Update README.md with new features
7. Add migration guide for breaking changes

---

## Appendix A: Files to Create

| File Path                           | Purpose                                    | Phase |
| ----------------------------------- | ------------------------------------------ | ----- |
| `keylib/multisig_key.go`            | MultisigKey type and methods               | 1     |
| `keylib/multisig_key_test.go`       | MultisigKey unit tests                     | 1     |
| `account/ledger_attachment.go`      | LedgerAttachment entity                    | 4     |
| `account/ledger_attachment_test.go` | LedgerAttachment unit tests                | 4     |
| `account/ledger_verification.go`    | Verification interface and implementations | 4     |

## Appendix B: Files to Modify

| File Path                    | Changes Required                                        | Phase   |
| ---------------------------- | ------------------------------------------------------- | ------- |
| `account/account_type.go`    | Add AccountTypeShared, update methods                   | 1       |
| `account/account.go`         | Add new fields, update validation, update serialization | 1, 2    |
| `account/builder.go`         | Add Shared builder, financial methods, validation       | 2, 3, 5 |
| `account/validation.go`      | Update Parent/Child validation, add Shared validation   | 5       |
| `account/errors.go`          | Add error kinds if needed                               | 5       |
| `account/account_test.go`    | Add tests for new features                              | 6       |
| `account/builder_test.go`    | Add tests for new builder methods                       | 6       |
| `account/validation_test.go` | Add tests for new validation rules                      | 6       |

---

_Document maintained by the development team. Update as implementation progresses._
