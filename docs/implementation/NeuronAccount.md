# NeuronAccount Technical Documentation

## Section 1: Overview

Package `account` provides the `NeuronAccount` abstraction for agent identity and endpoints.

```
Import path: github.com/aspect-build/neuron-go-hedera-sdk/account
Sub-package:  github.com/aspect-build/neuron-go-hedera-sdk/account/didkey
```

### What a NeuronAccount IS

- A portable, blockchain-agnostic representation of an agent's identity and endpoints.
- Identified by a `NeuronPublicKey` (from `keylib`), which is the root of identity.
- All other identifiers are derived from the Neuron key:
  - **PeerID** (libp2p) -- for P2P networking.
  - **EVMAddress** (Ethereum-compatible) -- for blockchain interactions.
  - **DID** (Decentralized Identifier) -- for self-sovereign identity (Parent accounts only).
- The Neuron key is always authoritative; derived identifiers are never authoritative on their own.
- Immutable after construction and safe for concurrent use.

### What a NeuronAccount is NOT

- NOT responsible for communication or message passing.
- It only describes WHERE communication happens and HOW the agent can be reached.
- Actual communication is handled by separate packages.

### Sub-package: `account/didkey`

Provides a `did:key` implementation for secp256k1 public keys. The `did:key` method encodes the public key directly in the DID identifier, making it self-describing, deterministic, and offline-capable.

Format:

```
did:key:z<base58btc(0xe7, 0x01 + compressed-pubkey-33-bytes)>
```

### Thread Safety

- `NeuronAccount` is immutable after construction and safe for concurrent use.
- `AccountBuilder` is NOT thread-safe and should only be used from a single goroutine.
- The backend registry (`RegisterBackend`, `GetBackend`, etc.) is protected by `sync.RWMutex` and is safe for concurrent use.
- The `didParsers` map used by `RegisterDIDParser` is NOT protected by a mutex. Only call `RegisterDIDParser` from `init()` functions.

---

## Section 2: Architecture & Domain Model

### Account Types

Three account types form a strict hierarchy:

| Type   | AccountType         | Public Key | DID        | Parent Ref | Comm Channels    | MultisigKey | Balance Field       |
| ------ | ------------------- | ---------- | ---------- | ---------- | ---------------- | ----------- | ------------------- |
| Parent | `AccountTypeParent` | Required   | Required   | Prohibited | Prohibited       | Prohibited  | `creditBalance`     |
| Child  | `AccountTypeChild`  | Required   | Prohibited | Required   | Required (all 3) | Prohibited  | `balanceAllocation` |
| Shared | `AccountTypeShared` | Prohibited | Prohibited | Prohibited | Prohibited       | Required    | `balance`           |

`AccountType` is defined as:

```go
type AccountType int

const (
    AccountTypeUnspecified AccountType = iota // 0
    AccountTypeParent                         // 1
    AccountTypeChild                          // 2
    AccountTypeShared                         // 3
)
```

Key `AccountType` methods:

| Method                         | Returns `true` for    |
| ------------------------------ | --------------------- |
| `IsValid() bool`               | Parent, Child, Shared |
| `IsParent() bool`              | Parent                |
| `IsChild() bool`               | Child                 |
| `IsShared() bool`              | Shared                |
| `RequiresPublicKey() bool`     | Parent, Child         |
| `RequiresMultisigKey() bool`   | Shared                |
| `RequiresDID() bool`           | Parent                |
| `RequiresParent() bool`        | Child                 |
| `RequiresCommChannels() bool`  | Child                 |
| `ProhibitsCommChannels() bool` | Parent, Shared        |
| `String() string`              | Human-readable name   |
| `Validate() error`             | nil if valid          |

Hierarchy rules:

- Nesting limited to one level (no grandchildren).
- Parent has no parent, may have children.
- Child has exactly one parent, no children.
- Shared has no parent, no children.

### Builder Pattern

`AccountBuilder` uses a fluent API with error accumulation:

```go
func NewParentAccountBuilder(publicKey NeuronPublicKey, did NeuronDID) *AccountBuilder
func NewChildAccountBuilder(publicKey, parentPubKey NeuronPublicKey) *AccountBuilder
func NewSharedAccountBuilder(multisigKey MultisigKey) *AccountBuilder
```

- `Build()` returns the first accumulated error, or `(NeuronAccount, nil)` on success.
- `Errors()` returns all accumulated errors.
- `HasErrors()` returns true if any errors were accumulated.
- `MustBuild()` panics on error (for tests and known-valid initialization).
- During `Build()`, `PeerID` and `EVMAddress` are derived from `publicKey` (for Parent/Child).
- Shared accounts do NOT derive `PeerID`/`EVMAddress` -- they remain zero values.

### NeuronAccount Struct (16 fields)

**Identity:**

| Field         | Type                     | Description                                     |
| ------------- | ------------------------ | ----------------------------------------------- |
| `publicKey`   | `keylib.NeuronPublicKey` | Root of identity; required for Parent and Child |
| `multisigKey` | `*keylib.MultisigKey`    | Threshold signing config; required for Shared   |
| `peerID`      | `keylib.PeerID`          | Derived from publicKey; cached                  |
| `evmAddress`  | `keylib.EVMAddress`      | Derived from publicKey; cached                  |

**Hierarchy:**

| Field          | Type                     | Description                                    |
| -------------- | ------------------------ | ---------------------------------------------- |
| `accountType`  | `AccountType`            | Parent, Child, or Shared                       |
| `did`          | `NeuronDID`              | Required for Parent; nil for Child and Shared  |
| `parentPubKey` | `keylib.NeuronPublicKey` | Set for Child only; zero for Parent and Shared |

**Communication:**

| Field    | Type          | Description                                       |
| -------- | ------------- | ------------------------------------------------- |
| `stdIn`  | `CommAddress` | Where others send messages; required for Child    |
| `stdOut` | `CommAddress` | Where agent publishes outputs; required for Child |
| `stdErr` | `CommAddress` | Where agent publishes errors; required for Child  |

**Reachability:**

| Field            | Type             | Description                        |
| ---------------- | ---------------- | ---------------------------------- |
| `reachableAddrs` | `ReachableAddrs` | Canonical "find/dial me" endpoints |

**Financial:**

| Field               | Type       | Description                               |
| ------------------- | ---------- | ----------------------------------------- |
| `currencySymbol`    | `string`   | Currency identifier (e.g., "HBAR", "ETH") |
| `creditBalance`     | `*big.Int` | Primary balance for Parent accounts       |
| `balanceAllocation` | `*big.Int` | Allocated funds for Child accounts        |
| `balance`           | `*big.Int` | Multisig-controlled balance for Shared    |

**Ledger:**

| Field              | Type                | Description                       |
| ------------------ | ------------------- | --------------------------------- |
| `ledgerAttachment` | `*LedgerAttachment` | Links account to settlement infra |

### Communication Address System

`CommAddress` struct with two fields: `kind` (`CommAddressKind`) and `locator` (`string`).

```go
type CommAddressKind string
```

- `CommAddressKind` is a string type (e.g., `"hedera-topic"`, `"kafka-topic"`, `"custom"`).
- String format: `"kind:locator"` (e.g., `"hedera-topic:0.0.12345"`).

`CommAddressSet` is a collection of `CommAddress` values:

| Method                      | Description                     |
| --------------------------- | ------------------------------- |
| `Addresses() []CommAddress` | Returns a copy of the addresses |
| `IsEmpty() bool`            | True if no addresses            |
| `Len() int`                 | Number of addresses             |
| `First() CommAddress`       | First address or zero value     |

**TopicTechnology:**

```go
type TopicTechnology string

const (
    TopicTechnologyHedera TopicTechnology = "hedera"
    TopicTechnologyKafka  TopicTechnology = "kafka"
    TopicTechnologyCustom TopicTechnology = "custom"
)
```

### Backend Registry

`Backend` interface:

```go
type Backend interface {
    Kind() string
    ValidateLocator(locator string) error
    ParseLocator(locator string) (string, error)
    Metadata() BackendMetadata
    Technology() TopicTechnology
}
```

`BackendMetadata` struct:

```go
type BackendMetadata struct {
    DisplayName    string
    Description    string
    LocatorFormat  string
    LocatorExample string
    RequiresConfig bool
    Properties     map[string]string
}
```

Registry functions:

| Function                                                  | Description                                           |
| --------------------------------------------------------- | ----------------------------------------------------- |
| `RegisterBackend(b Backend)`                              | Register a backend; panics on duplicate or empty kind |
| `GetBackend(kind string) (Backend, bool)`                 | Retrieve by kind                                      |
| `ListBackends() []string`                                 | All registered kinds, sorted                          |
| `IsRegisteredBackend(kind string) bool`                   | Check if kind is registered                           |
| `GetBackendMetadata(kind string) (BackendMetadata, bool)` | Metadata for a kind                                   |
| `ListBackendsWithMetadata() map[string]BackendMetadata`   | All backends with metadata                            |

- Thread-safe registry (protected by `sync.RWMutex`).
- Backends register via `init()` functions.
- Panics on duplicate registration or empty kind.

Built-in backends:

| Kind             | Constant          | Technology |
| ---------------- | ----------------- | ---------- |
| `"hedera-topic"` | `HederaTopicKind` | `"hedera"` |
| `"kafka-topic"`  | `KafkaTopicKind`  | `"kafka"`  |
| `"custom"`       | `CustomTopicKind` | `"custom"` |

### Reachable Addresses

`ReachableAddr` wraps a libp2p multiaddr with a PeerID:

- Every address MUST contain `/p2p/<PeerID>` matching the account's derived PeerID.
- For circuit relay addresses, the final `/p2p/<TargetPeerID>` after `/p2p-circuit/` is validated.

`ReachableAddr` methods:

| Method                                     | Description                      |
| ------------------------------------------ | -------------------------------- |
| `Multiaddr() multiaddr.Multiaddr`          | Underlying libp2p multiaddr      |
| `PeerID() keylib.PeerID`                   | Extracted PeerID                 |
| `String() string`                          | String representation            |
| `IsZero() bool`                            | True if zero-value               |
| `IsCircuitRelay() bool`                    | True if contains `/p2p-circuit/` |
| `Equal(other ReachableAddr) bool`          | Equality comparison              |
| `Validate() error`                         | Well-formed check                |
| `ValidateForAccount(expectedPeerID) error` | PeerID consistency check         |

`ReachableAddrs` collection methods:

| Method                               | Description                  |
| ------------------------------------ | ---------------------------- |
| `Addrs() []ReachableAddr`            | Copy of addresses            |
| `Strings() []string`                 | String representations       |
| `Multiaddrs() []multiaddr.Multiaddr` | Underlying multiaddrs        |
| `IsEmpty() bool`                     | True if empty                |
| `Len() int`                          | Number of addresses          |
| `First() ReachableAddr`              | First address or zero value  |
| `ValidateAll(expectedPeerID) error`  | Validate all against PeerID  |
| `Contains(addr ReachableAddr) bool`  | Check membership             |
| `DirectAddrs() []ReachableAddr`      | Non-relay addresses only     |
| `RelayAddrs() []ReachableAddr`       | Circuit relay addresses only |

### DID System

`NeuronDID` interface:

```go
type NeuronDID interface {
    String() string
    Method() string
    Identifier() string
    Validate() error
    Equal(other NeuronDID) bool
}
```

`NeuronDIDWithKey` extends `NeuronDID`:

```go
type NeuronDIDWithKey interface {
    NeuronDID
    PublicKey() (keylib.NeuronPublicKey, error)
    MatchesKey(pubKey keylib.NeuronPublicKey) bool
}
```

Supporting types:

```go
type DIDDocument struct {
    Context            []string             `json:"@context"`
    ID                 string               `json:"id"`
    Controller         []string             `json:"controller,omitempty"`
    VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`
    Authentication     []string             `json:"authentication,omitempty"`
    AssertionMethod    []string             `json:"assertionMethod,omitempty"`
    Service            []ServiceEndpoint    `json:"service,omitempty"`
}

type VerificationMethod struct {
    ID                 string         `json:"id"`
    Type               string         `json:"type"`
    Controller         string         `json:"controller"`
    PublicKeyMultibase string         `json:"publicKeyMultibase,omitempty"`
    PublicKeyJwk       map[string]any `json:"publicKeyJwk,omitempty"`
}

type ServiceEndpoint struct {
    ID              string `json:"id"`
    Type            string `json:"type"`
    ServiceEndpoint any    `json:"serviceEndpoint"`
}
```

DID constants and registration:

```go
const DIDMethodKey = "key"

type DIDParser func(didString string) (NeuronDID, error)

func RegisterDIDParser(method string, parser DIDParser)
func ParseDID(didString string) (NeuronDID, error)
```

`RegisterDIDParser` writes to a package-level `map[string]DIDParser` that is NOT protected by a mutex. Only call from `init()` functions.

`ParseDID` parses `"did:<method>:<identifier>"` and dispatches to the registered parser.

**didkey sub-package:**

`DIDKey` struct implements both `NeuronDID` and `NeuronDIDWithKey`:

| Function/Method                                                 | Description                                |
| --------------------------------------------------------------- | ------------------------------------------ |
| `FromPublicKey(pubKey NeuronPublicKey) (*DIDKey, error)`        | Create did:key from public key             |
| `Parse(didString string) (*DIDKey, error)`                      | Parse a did:key string                     |
| `MustParse(didString string) *DIDKey`                           | Parse or panic                             |
| `DIDKeyFromEVMAddress(addr EVMAddress) (*DIDKey, error)`        | Always returns error (one-way derivation)  |
| `(*DIDKey).String() string`                                     | Full DID string                            |
| `(*DIDKey).Method() string`                                     | Returns `"key"`                            |
| `(*DIDKey).Identifier() string`                                 | Method-specific identifier                 |
| `(*DIDKey).Validate() error`                                    | Validates public key is non-zero           |
| `(*DIDKey).Equal(other NeuronDID) bool`                         | String-based equality                      |
| `(*DIDKey).PublicKey() (NeuronPublicKey, error)`                | Extract encoded public key                 |
| `(*DIDKey).MatchesKey(pubKey NeuronPublicKey) bool`             | Check key correspondence                   |
| `(*DIDKey).IsZero() bool`                                       | True if nil, empty raw, or zero public key |
| `(*DIDKey).Resolve() (*DIDDocument, error)`                     | Generate DID Document (self-describing)    |
| `(*DIDKey).VerifySignature(message []byte, sig Signature) bool` | Verify signature against embedded key      |

### LedgerAttachment

`AttachmentState` enum:

```go
type AttachmentState int

const (
    AttachmentStateDetached AttachmentState = iota // 0
    AttachmentStateAttached                        // 1
    AttachmentStateVerified                        // 2
)
```

`VerificationStatus` enum:

```go
type VerificationStatus int

const (
    VerificationStatusNone     VerificationStatus = iota // 0
    VerificationStatusPending                            // 1
    VerificationStatusVerified                           // 2
    VerificationStatusFailed                             // 3
)
```

`LedgerAttachment` struct (unexported fields):

| Field                | Type                 |
| -------------------- | -------------------- |
| `ledgerIdentifier`   | `string`             |
| `attachedAddress`    | `string`             |
| `attachmentState`    | `AttachmentState`    |
| `verificationStatus` | `VerificationStatus` |

Constructor:

```go
func NewLedgerAttachment(ledgerID, address string) (*LedgerAttachment, error)
```

Creates in `Detached` state with `VerificationStatusNone`. Returns error if either argument is empty.

All methods are nil-safe (return zero/default values when receiver is nil):

| Method                                    | Description                               |
| ----------------------------------------- | ----------------------------------------- |
| `LedgerIdentifier() string`               | Ledger identifier                         |
| `AttachedAddress() string`                | Ledger-specific account address           |
| `State() AttachmentState`                 | Current attachment state                  |
| `VerificationStatus() VerificationStatus` | Current verification status               |
| `IsAttached() bool`                       | True if Attached or Verified              |
| `IsVerified() bool`                       | True if Verified                          |
| `Validate() error`                        | Validates fields; nil attachment is valid |
| `SetAttached()`                           | Transitions to Attached state             |
| `SetVerified()`                           | Transitions to Verified state and status  |
| `SetVerificationFailed()`                 | Sets verification status to Failed        |

### LedgerVerifier Interface

```go
type LedgerVerifier interface {
    LedgerIdentifier() string
    VerifyAccountExists(address string) (bool, error)
    VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error)
    VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error)
    VerifySemanticConsistency(address string, pubKey keylib.NeuronPublicKey) (bool, error)
    VerifyParentChildRelationship(parentAddr, childAddr string) (bool, error)
}
```

`VerificationResult` struct:

```go
type VerificationResult struct {
    Verified bool
    Message  string
    Error    error
}

func (r *VerificationResult) IsSuccess() bool  // Verified && Error == nil
```

Package-level verification functions:

```go
func VerifyAccountAttachment(account NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error)
func VerifyParentChildLink(parent, child NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error)
```

### TopicMessage

```go
type TopicMessage struct {
    Payload         []byte
    Timestamp       time.Time
    SequenceNumber  uint64
    SenderPublicKey keylib.NeuronPublicKey
    TopicID         string
    Technology      TopicTechnology
    Metadata        map[string]any
}
```

Methods:

| Method                                 | Description                                     |
| -------------------------------------- | ----------------------------------------------- |
| `IsZero() bool`                        | True if empty payload, zero time, zero sequence |
| `HasSender() bool`                     | True if sender public key is non-zero           |
| `GetMetadataString(key string) string` | String value from metadata or `""`              |
| `GetMetadataInt64(key string) int64`   | Int64 value from metadata or `0`                |
| `GetMetadataBytes(key string) []byte`  | Byte slice from metadata or `nil`               |

`TopicMessageBuilder` fluent API:

```go
func NewTopicMessageBuilder() *TopicMessageBuilder
func (b *TopicMessageBuilder) WithPayload(payload []byte) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTimestamp(ts time.Time) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithSequenceNumber(seq uint64) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithSender(pubKey keylib.NeuronPublicKey) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTopicID(id string) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTechnology(tech TopicTechnology) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithMetadata(key string, value any) *TopicMessageBuilder
func (b *TopicMessageBuilder) Build() TopicMessage
```

---

## Section 3: Validation & Error Semantics

### AccountError

```go
type AccountError struct {
    Op      string           // operation that failed
    Kind    AccountErrorKind // category
    Details string           // human-readable explanation
    Err     error            // underlying error
}
```

Error format when all fields are present:

```
account.{Op}: {Kind}: {Details}: {Err}
```

When `Err` is nil:

```
account.{Op}: {Kind}: {Details}
```

When `Op` is also empty:

```
account: {Kind}: {Details}
```

Custom `Is()` method: matches by `Kind` when both errors are `*AccountError` and have the same non-zero `Kind`.

`Unwrap()` returns the underlying `Err` field.

### AccountErrorKind

`AccountErrorKind` is an `int` type with 10 constants:

| Constant                  | Value | String               | Description                          |
| ------------------------- | ----- | -------------------- | ------------------------------------ |
| `ErrKindInvalidAccount`   | 1     | `"InvalidAccount"`   | Invalid account data or state        |
| `ErrKindInvalidTopic`     | 2     | `"InvalidTopic"`     | Invalid topic configuration          |
| `ErrKindInvalidAddress`   | 3     | `"InvalidAddress"`   | Invalid comm or reachable address    |
| `ErrKindInvalidDID`       | 4     | `"InvalidDID"`       | Invalid DID format or content        |
| `ErrKindPeerIDMismatch`   | 5     | `"PeerIDMismatch"`   | PeerID does not match expected value |
| `ErrKindMissingRequired`  | 6     | `"MissingRequired"`  | Required field is missing            |
| `ErrKindValidation`       | 7     | `"Validation"`       | General validation failure           |
| `ErrKindZeroValue`        | 8     | `"ZeroValue"`        | Operation on zero-value type         |
| `ErrKindInvalidHierarchy` | 9     | `"InvalidHierarchy"` | Invalid parent/child relationship    |
| `ErrKindProhibitedField`  | 10    | `"ProhibitedField"`  | Field present that is not allowed    |

### Sentinel Errors

```go
var (
    ErrZeroPublicKey = &AccountError{Kind: ErrKindZeroValue, Details: "public key is zero-value"}
    ErrMissingDID    = &AccountError{Kind: ErrKindMissingRequired, Details: "DID is required for Parent accounts"}
    ErrMissingParent = &AccountError{Kind: ErrKindMissingRequired, Details: "parent public key is required for Child accounts"}
)
```

### Validate() vs ValidateAll()

- `Validate() error` -- returns the first validation error encountered (fail-fast). Wraps errors with `AccountError` context including operation name.
- `ValidateAll() *ValidationResult` -- collects all errors. Returns a `ValidationResult` with all individual errors.
- Both run the same set of checks: account type, public key, type-specific requirements, DID-key match, communication address validation, and reachable address PeerID consistency.

### ValidationResult

```go
type ValidationResult struct {
    errs []error  // unexported
}

func NewValidationResult() *ValidationResult

func (r *ValidationResult) AddError(err error)     // nil errors are ignored
func (r *ValidationResult) HasErrors() bool
func (r *ValidationResult) Valid() bool             // true if no errors
func (r *ValidationResult) Error() error            // errors.Join of all errors; nil if valid
func (r *ValidationResult) FirstError() error       // first error or nil
func (r *ValidationResult) Errors() []error         // all individual errors
```

`ValidationResult` is NOT safe for concurrent use. Each goroutine should use its own instance.

The `Error()` method uses `errors.Join` (Go 1.20+), enabling use of `errors.As()` and `errors.Is()` on the composed error.

### AccountValidator

`AccountValidator` provides fluent validation with field-level granularity:

```go
func NewAccountValidator() *AccountValidator

func (v *AccountValidator) ValidatePublicKey(pubKey keylib.NeuronPublicKey) *AccountValidator
func (v *AccountValidator) ValidateAccountType(t AccountType) *AccountValidator
func (v *AccountValidator) ValidateParentRequirements(pubKey keylib.NeuronPublicKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) *AccountValidator
func (v *AccountValidator) ValidateChildRequirements(pubKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress, did NeuronDID) *AccountValidator
func (v *AccountValidator) ValidateSharedRequirements(multisigKey *keylib.MultisigKey,
    did NeuronDID, stdIn, stdOut, stdErr CommAddress,
    parentPubKey keylib.NeuronPublicKey) *AccountValidator
func (v *AccountValidator) ValidateDID(did NeuronDID) *AccountValidator
func (v *AccountValidator) ValidateDIDMatchesKey(did NeuronDID, pubKey keylib.NeuronPublicKey) *AccountValidator
func (v *AccountValidator) ValidateCommAddress(addr CommAddress) *AccountValidator
func (v *AccountValidator) ValidateReachableAddrs(addrs ReachableAddrs, expectedPeerID keylib.PeerID) *AccountValidator
func (v *AccountValidator) Result() *ValidationResult
func (v *AccountValidator) IsValid() bool
func (v *AccountValidator) Error() error  // returns first error
```

### Type-Specific Validation Functions

```go
func ValidatePublicKey(key keylib.NeuronPublicKey) error
func ValidateAccountType(t AccountType) error
func ValidateParentAccount(publicKey keylib.NeuronPublicKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error
func ValidateChildAccount(publicKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress, did NeuronDID) error
func ValidateSharedAccount(multisigKey *keylib.MultisigKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error
func ValidateDID(did NeuronDID) error
func ValidateDIDMatchesKey(did NeuronDID, pubKey keylib.NeuronPublicKey) error
func ValidateCommAddress(addr CommAddress) error
func ValidateCommAddressNonEmpty(addr CommAddress, fieldName string) error
func ValidateReachableAddr(addr ReachableAddr) error
func ValidateReachableAddrForAccount(addr ReachableAddr, expectedPeerID keylib.PeerID) error
func ValidateReachableAddrs(addrs ReachableAddrs, expectedPeerID keylib.PeerID) error
```

`ValidateChildAccount` also checks that the child public key is not equal to the parent public key.

---

## Section 4: Public API Reference

### Account Types

```go
type AccountType int

const (
    AccountTypeUnspecified AccountType = iota
    AccountTypeParent
    AccountTypeChild
    AccountTypeShared
)

func (t AccountType) String() string
func (t AccountType) IsValid() bool
func (t AccountType) IsParent() bool
func (t AccountType) IsChild() bool
func (t AccountType) IsShared() bool
func (t AccountType) RequiresDID() bool
func (t AccountType) RequiresParent() bool
func (t AccountType) RequiresMultisigKey() bool
func (t AccountType) RequiresPublicKey() bool
func (t AccountType) RequiresCommChannels() bool
func (t AccountType) ProhibitsCommChannels() bool
func (t AccountType) Validate() error
```

### NeuronAccount

```go
type NeuronAccount struct { /* unexported fields */ }

// Identity
func (a NeuronAccount) PublicKey() keylib.NeuronPublicKey
func (a NeuronAccount) PeerID() keylib.PeerID
func (a NeuronAccount) EVMAddress() keylib.EVMAddress

// Hierarchy
func (a NeuronAccount) AccountType() AccountType
func (a NeuronAccount) IsParent() bool
func (a NeuronAccount) IsChild() bool
func (a NeuronAccount) IsShared() bool
func (a NeuronAccount) MultisigKey() *keylib.MultisigKey
func (a NeuronAccount) DID() NeuronDID
func (a NeuronAccount) ParentPublicKey() keylib.NeuronPublicKey

// Communication
func (a NeuronAccount) StdIn() CommAddress
func (a NeuronAccount) StdOut() CommAddress
func (a NeuronAccount) StdErr() CommAddress

// Reachability
func (a NeuronAccount) ReachableAddrs() ReachableAddrs
func (a NeuronAccount) FirstReachableAddr() ReachableAddr

// Financial
func (a NeuronAccount) CurrencySymbol() string
func (a NeuronAccount) CreditBalance() *big.Int
func (a NeuronAccount) BalanceAllocation() *big.Int
func (a NeuronAccount) Balance() *big.Int

// Ledger
func (a NeuronAccount) LedgerAttachment() *LedgerAttachment

// Validation
func (a NeuronAccount) IsZero() bool
func (a NeuronAccount) Validate() error
func (a NeuronAccount) ValidateAll() *ValidationResult

// Comparison
func (a NeuronAccount) Equal(other NeuronAccount) bool

// Serialization
func (a NeuronAccount) MarshalJSON() ([]byte, error)
func (a *NeuronAccount) UnmarshalJSON(data []byte) error
func (a NeuronAccount) String() string
```

### AccountBuilder

```go
type AccountBuilder struct { /* unexported fields */ }

// Constructors
func NewParentAccountBuilder(publicKey keylib.NeuronPublicKey, did NeuronDID) *AccountBuilder
func NewChildAccountBuilder(publicKey, parentPubKey keylib.NeuronPublicKey) *AccountBuilder
func NewSharedAccountBuilder(multisigKey keylib.MultisigKey) *AccountBuilder

// Communication -- individual
func (b *AccountBuilder) WithStdIn(addr CommAddress) *AccountBuilder
func (b *AccountBuilder) WithStdInBackend(kind, locator string) *AccountBuilder
func (b *AccountBuilder) WithStdInHedera(topicID string) *AccountBuilder
func (b *AccountBuilder) WithStdInKafka(topicName string) *AccountBuilder

func (b *AccountBuilder) WithStdOut(addr CommAddress) *AccountBuilder
func (b *AccountBuilder) WithStdOutBackend(kind, locator string) *AccountBuilder
func (b *AccountBuilder) WithStdOutHedera(topicID string) *AccountBuilder
func (b *AccountBuilder) WithStdOutKafka(topicName string) *AccountBuilder

func (b *AccountBuilder) WithStdErr(addr CommAddress) *AccountBuilder
func (b *AccountBuilder) WithStdErrBackend(kind, locator string) *AccountBuilder
func (b *AccountBuilder) WithStdErrHedera(topicID string) *AccountBuilder
func (b *AccountBuilder) WithStdErrKafka(topicName string) *AccountBuilder

// Communication -- batch
func (b *AccountBuilder) WithBackendTopics(kind, stdInLocator, stdOutLocator, stdErrLocator string) *AccountBuilder
func (b *AccountBuilder) WithHederaTopics(stdInTopic, stdOutTopic, stdErrTopic string) *AccountBuilder
func (b *AccountBuilder) WithKafkaTopics(stdInTopic, stdOutTopic, stdErrTopic string) *AccountBuilder

// Reachability
func (b *AccountBuilder) WithReachableAddr(multiaddr string) *AccountBuilder
func (b *AccountBuilder) WithReachableAddrs(multiaddrs ...string) *AccountBuilder
func (b *AccountBuilder) WithReachableAddrValidated(multiaddr string) *AccountBuilder

// Financial
func (b *AccountBuilder) WithCurrencySymbol(symbol string) *AccountBuilder
func (b *AccountBuilder) WithCreditBalance(balance *big.Int) *AccountBuilder
func (b *AccountBuilder) WithBalanceAllocation(allocation *big.Int) *AccountBuilder
func (b *AccountBuilder) WithBalance(balance *big.Int) *AccountBuilder

// Ledger
func (b *AccountBuilder) WithLedgerAttachment(ledgerID, address string) *AccountBuilder

// Build
func (b *AccountBuilder) Build() (NeuronAccount, error)
func (b *AccountBuilder) MustBuild() NeuronAccount
func (b *AccountBuilder) Errors() []error
func (b *AccountBuilder) HasErrors() bool
```

### CommAddress

```go
type CommAddressKind string
type CommAddress struct { /* unexported fields */ }

// CommAddressKind methods
func (k CommAddressKind) String() string
func (k CommAddressKind) IsValid() bool
func (k CommAddressKind) TopicTechnologyForKind() TopicTechnology

// CommAddress constructors
func NewCommAddress(kind string, locator string) (CommAddress, error)
func NewCustomAddress(kind CommAddressKind, locator string) (CommAddress, error)  // Deprecated
func ParseCommAddress(s string) (CommAddress, error)
func MustParseCommAddress(s string) CommAddress

// CommAddress methods
func (a CommAddress) Kind() CommAddressKind
func (a CommAddress) Locator() string
func (a CommAddress) Technology() TopicTechnology
func (a CommAddress) String() string
func (a CommAddress) IsZero() bool
func (a CommAddress) Validate() error
func (a CommAddress) Equal(other CommAddress) bool

// CommAddressSet
type CommAddressSet struct { /* unexported fields */ }
func NewCommAddressSet(addrs ...CommAddress) CommAddressSet
func (s CommAddressSet) Addresses() []CommAddress
func (s CommAddressSet) IsEmpty() bool
func (s CommAddressSet) Len() int
func (s CommAddressSet) First() CommAddress
```

### Hedera Backend

```go
const HederaTopicKind = "hedera-topic"

func NewHederaTopicAddress(locator string) (CommAddress, error)
func ValidateHederaTopicLocator(locator string) error
```

### Kafka Backend

```go
const KafkaTopicKind = "kafka-topic"
const MaxKafkaTopicLength = 249

func NewKafkaTopicAddress(locator string) (CommAddress, error)
func ValidateKafkaTopicLocator(locator string) error
```

### Custom Backend

```go
const CustomTopicKind = "custom"

func NewCustomTopicAddress(locator string) (CommAddress, error)
```

### ReachableAddr

```go
type ReachableAddr struct { /* unexported fields */ }

func ParseReachableAddr(s string) (ReachableAddr, error)
func MustParseReachableAddr(s string) ReachableAddr
func NewReachableAddrWithValidation(ma multiaddr.Multiaddr, expectedPeerID keylib.PeerID) (ReachableAddr, error)
func NewReachableAddrFromStringWithValidation(s string, expectedPeerID keylib.PeerID) (ReachableAddr, error)

func (a ReachableAddr) Multiaddr() multiaddr.Multiaddr
func (a ReachableAddr) PeerID() keylib.PeerID
func (a ReachableAddr) String() string
func (a ReachableAddr) IsZero() bool
func (a ReachableAddr) IsCircuitRelay() bool
func (a ReachableAddr) Equal(other ReachableAddr) bool
func (a ReachableAddr) Validate() error
func (a ReachableAddr) ValidateForAccount(expectedPeerID keylib.PeerID) error
```

### ReachableAddrs

```go
type ReachableAddrs struct { /* unexported fields */ }

func NewReachableAddrs(addrs ...ReachableAddr) ReachableAddrs
func ParseReachableAddrs(strs ...string) (ReachableAddrs, error)
func NewReachableAddrsWithValidation(strs []string, expectedPeerID keylib.PeerID) (ReachableAddrs, error)

func (r ReachableAddrs) Addrs() []ReachableAddr
func (r ReachableAddrs) Strings() []string
func (r ReachableAddrs) Multiaddrs() []multiaddr.Multiaddr
func (r ReachableAddrs) IsEmpty() bool
func (r ReachableAddrs) Len() int
func (r ReachableAddrs) First() ReachableAddr
func (r ReachableAddrs) ValidateAll(expectedPeerID keylib.PeerID) error
func (r ReachableAddrs) Contains(addr ReachableAddr) bool
func (r ReachableAddrs) DirectAddrs() []ReachableAddr
func (r ReachableAddrs) RelayAddrs() []ReachableAddr
```

### DID

```go
type NeuronDID interface {
    String() string
    Method() string
    Identifier() string
    Validate() error
    Equal(other NeuronDID) bool
}

type NeuronDIDWithKey interface {
    NeuronDID
    PublicKey() (keylib.NeuronPublicKey, error)
    MatchesKey(pubKey keylib.NeuronPublicKey) bool
}

const DIDMethodKey = "key"

type DIDParser func(didString string) (NeuronDID, error)

func RegisterDIDParser(method string, parser DIDParser)
func ParseDID(didString string) (NeuronDID, error)
```

### DID Document Types

```go
type DIDDocument struct { ... }      // See Section 2 for field details
type VerificationMethod struct { ... }
type ServiceEndpoint struct { ... }
```

### didkey Sub-package

```go
package didkey

type DIDKey struct { /* unexported fields */ }

func FromPublicKey(pubKey keylib.NeuronPublicKey) (*DIDKey, error)
func Parse(didString string) (*DIDKey, error)
func MustParse(didString string) *DIDKey
func DIDKeyFromEVMAddress(addr keylib.EVMAddress) (*DIDKey, error)

func (d *DIDKey) String() string
func (d *DIDKey) Method() string
func (d *DIDKey) Identifier() string
func (d *DIDKey) Validate() error
func (d *DIDKey) Equal(other account.NeuronDID) bool
func (d *DIDKey) PublicKey() (keylib.NeuronPublicKey, error)
func (d *DIDKey) MatchesKey(pubKey keylib.NeuronPublicKey) bool
func (d *DIDKey) IsZero() bool
func (d *DIDKey) Resolve() (*account.DIDDocument, error)
func (d *DIDKey) VerifySignature(message []byte, signature keylib.Signature) bool
```

### TopicTechnology

```go
type TopicTechnology string

const (
    TopicTechnologyHedera TopicTechnology = "hedera"
    TopicTechnologyKafka  TopicTechnology = "kafka"
    TopicTechnologyCustom TopicTechnology = "custom"
)

func (t TopicTechnology) String() string
func (t TopicTechnology) IsValid() bool
```

### TopicMessage

```go
type TopicMessage struct {
    Payload         []byte
    Timestamp       time.Time
    SequenceNumber  uint64
    SenderPublicKey keylib.NeuronPublicKey
    TopicID         string
    Technology      TopicTechnology
    Metadata        map[string]any
}

func (m TopicMessage) IsZero() bool
func (m TopicMessage) HasSender() bool
func (m TopicMessage) GetMetadataString(key string) string
func (m TopicMessage) GetMetadataInt64(key string) int64
func (m TopicMessage) GetMetadataBytes(key string) []byte

type TopicMessageBuilder struct { /* unexported fields */ }

func NewTopicMessageBuilder() *TopicMessageBuilder
func (b *TopicMessageBuilder) WithPayload(payload []byte) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTimestamp(ts time.Time) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithSequenceNumber(seq uint64) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithSender(pubKey keylib.NeuronPublicKey) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTopicID(id string) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithTechnology(tech TopicTechnology) *TopicMessageBuilder
func (b *TopicMessageBuilder) WithMetadata(key string, value any) *TopicMessageBuilder
func (b *TopicMessageBuilder) Build() TopicMessage
```

### LedgerAttachment

```go
type AttachmentState int

const (
    AttachmentStateDetached AttachmentState = iota
    AttachmentStateAttached
    AttachmentStateVerified
)

func (s AttachmentState) String() string
func (s AttachmentState) IsValid() bool

type VerificationStatus int

const (
    VerificationStatusNone     VerificationStatus = iota
    VerificationStatusPending
    VerificationStatusVerified
    VerificationStatusFailed
)

func (s VerificationStatus) String() string

type LedgerAttachment struct { /* unexported fields */ }

func NewLedgerAttachment(ledgerID, address string) (*LedgerAttachment, error)
func (a *LedgerAttachment) LedgerIdentifier() string
func (a *LedgerAttachment) AttachedAddress() string
func (a *LedgerAttachment) State() AttachmentState
func (a *LedgerAttachment) VerificationStatus() VerificationStatus
func (a *LedgerAttachment) IsAttached() bool
func (a *LedgerAttachment) IsVerified() bool
func (a *LedgerAttachment) Validate() error
func (a *LedgerAttachment) SetAttached()
func (a *LedgerAttachment) SetVerified()
func (a *LedgerAttachment) SetVerificationFailed()
```

### LedgerVerifier

```go
type LedgerVerifier interface {
    LedgerIdentifier() string
    VerifyAccountExists(address string) (bool, error)
    VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error)
    VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error)
    VerifySemanticConsistency(address string, pubKey keylib.NeuronPublicKey) (bool, error)
    VerifyParentChildRelationship(parentAddr, childAddr string) (bool, error)
}

type VerificationResult struct {
    Verified bool
    Message  string
    Error    error
}

func (r *VerificationResult) IsSuccess() bool

func VerifyAccountAttachment(account NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error)
func VerifyParentChildLink(parent, child NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error)
```

### Backend

```go
type Backend interface {
    Kind() string
    ValidateLocator(locator string) error
    ParseLocator(locator string) (string, error)
    Metadata() BackendMetadata
    Technology() TopicTechnology
}

type BackendMetadata struct {
    DisplayName    string
    Description    string
    LocatorFormat  string
    LocatorExample string
    RequiresConfig bool
    Properties     map[string]string
}

func RegisterBackend(b Backend)
func GetBackend(kind string) (Backend, bool)
func ListBackends() []string
func IsRegisteredBackend(kind string) bool
func GetBackendMetadata(kind string) (BackendMetadata, bool)
func ListBackendsWithMetadata() map[string]BackendMetadata
```

### Validation

```go
func ValidatePublicKey(key keylib.NeuronPublicKey) error
func ValidateAccountType(t AccountType) error
func ValidateParentAccount(publicKey keylib.NeuronPublicKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error
func ValidateChildAccount(publicKey, parentPubKey keylib.NeuronPublicKey,
    stdIn, stdOut, stdErr CommAddress, did NeuronDID) error
func ValidateSharedAccount(multisigKey *keylib.MultisigKey, did NeuronDID,
    stdIn, stdOut, stdErr CommAddress, parentPubKey keylib.NeuronPublicKey) error
func ValidateDID(did NeuronDID) error
func ValidateDIDMatchesKey(did NeuronDID, pubKey keylib.NeuronPublicKey) error
func ValidateCommAddress(addr CommAddress) error
func ValidateCommAddressNonEmpty(addr CommAddress, fieldName string) error
func ValidateReachableAddr(addr ReachableAddr) error
func ValidateReachableAddrForAccount(addr ReachableAddr, expectedPeerID keylib.PeerID) error
func ValidateReachableAddrs(addrs ReachableAddrs, expectedPeerID keylib.PeerID) error

func NewValidationResult() *ValidationResult
func NewAccountValidator() *AccountValidator
```

### Errors

```go
type AccountErrorKind int

const (
    ErrKindInvalidAccount  AccountErrorKind = iota + 1
    ErrKindInvalidTopic
    ErrKindInvalidAddress
    ErrKindInvalidDID
    ErrKindPeerIDMismatch
    ErrKindMissingRequired
    ErrKindValidation
    ErrKindZeroValue
    ErrKindInvalidHierarchy
    ErrKindProhibitedField
)

func (k AccountErrorKind) String() string

type AccountError struct {
    Op      string
    Kind    AccountErrorKind
    Details string
    Err     error
}

func (e *AccountError) Error() string
func (e *AccountError) Unwrap() error
func (e *AccountError) Is(target error) bool

var ErrZeroPublicKey *AccountError
var ErrMissingDID    *AccountError
var ErrMissingParent *AccountError
```

---

## Section 5: Usage Guide

### Creating a Parent Account

```go
import (
    "math/big"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

privKey, err := keylib.GeneratePrivateKey()
if err != nil {
    return err
}
pubKey := privKey.PublicKey()
did, err := didkey.FromPublicKey(pubKey)
if err != nil {
    return err
}

parent, err := account.NewParentAccountBuilder(pubKey, did).
    WithCurrencySymbol("HBAR").
    WithCreditBalance(big.NewInt(1000)).
    Build()
if err != nil {
    return err
}
```

Parent accounts:

- MUST have a public key and DID.
- Must NOT have communication channels, parent reference, or MultisigKey.
- PeerID and EVMAddress are automatically derived from the public key during `Build()`.

### Creating a Child Account

```go
childPrivKey, err := keylib.GeneratePrivateKey()
if err != nil {
    return err
}
childPubKey := childPrivKey.PublicKey()

child, err := account.NewChildAccountBuilder(childPubKey, pubKey).
    WithStdInHedera("0.0.1001").
    WithStdOutHedera("0.0.1002").
    WithStdErrHedera("0.0.1003").
    WithCurrencySymbol("HBAR").
    WithBalanceAllocation(big.NewInt(100)).
    Build()
if err != nil {
    return err
}
```

Child accounts:

- MUST have a public key and parent's public key.
- MUST have all three communication channels (stdIn, stdOut, stdErr).
- Must NOT have a DID.
- The child's public key must differ from the parent's public key.

### Creating a Shared Account

```go
keys := make([]keylib.NeuronPublicKey, 3)
for i := range keys {
    priv, _ := keylib.GeneratePrivateKey()
    keys[i] = priv.PublicKey()
}
multisigKey, err := keylib.NewMultisigKey(keys, 2) // 2-of-3 threshold
if err != nil {
    return err
}

shared, err := account.NewSharedAccountBuilder(multisigKey).
    WithCurrencySymbol("HBAR").
    WithBalance(big.NewInt(5000)).
    Build()
if err != nil {
    return err
}
```

Shared accounts:

- MUST have a valid MultisigKey (passed by value to `NewSharedAccountBuilder`).
- Must NOT have a public key, DID, communication channels, or parent reference.
- Do NOT derive PeerID or EVMAddress (they remain zero values).
- Cannot have reachable addresses (no PeerID to validate against).

### JSON Serialization

```go
// Marshal
data, err := json.Marshal(parent)
if err != nil {
    return err
}

// Unmarshal
var restored account.NeuronAccount
if err := json.Unmarshal(data, &restored); err != nil {
    return err
}
```

JSON field names (from the internal `neuronAccountJSON` struct):

| Go Field           | JSON Key             | `omitempty` | Notes                                           |
| ------------------ | -------------------- | ----------- | ----------------------------------------------- |
| PublicKey          | `publicKey`          | yes         | Hex-encoded                                     |
| PeerID             | `peerId`             | yes         | String representation                           |
| EVMAddress         | `evmAddress`         | yes         | Hex-encoded                                     |
| AccountType        | `accountType`        | no          | `"Parent"`, `"Child"`, `"Shared"`               |
| MultisigThreshold  | `multisigThreshold`  | yes         | Integer                                         |
| MultisigTotal      | `multisigTotal`      | yes         | Integer                                         |
| MultisigKeys       | `multisigKeys`       | yes         | Array of hex-encoded keys                       |
| DID                | `did`                | yes         | Full DID string                                 |
| ParentPublicKey    | `parentPublicKey`    | yes         | Hex-encoded                                     |
| StdIn              | `stdIn`              | yes         | `"kind:locator"` format                         |
| StdOut             | `stdOut`             | yes         | `"kind:locator"` format                         |
| StdErr             | `stdErr`             | yes         | `"kind:locator"` format                         |
| ReachableAddrs     | `reachableAddrs`     | yes         | Array of multiaddr strings                      |
| CurrencySymbol     | `currencySymbol`     | yes         | String                                          |
| CreditBalance      | `creditBalance`      | yes         | Decimal string (base-10)                        |
| BalanceAllocation  | `balanceAllocation`  | yes         | Decimal string (base-10)                        |
| Balance            | `balance`            | yes         | Decimal string (base-10)                        |
| LedgerIdentifier   | `ledgerIdentifier`   | yes         | String                                          |
| AttachedAddress    | `attachedAddress`    | yes         | String                                          |
| AttachmentState    | `attachmentState`    | yes         | `"Detached"`, `"Attached"`, `"Verified"`        |
| VerificationStatus | `verificationStatus` | yes         | `"None"`, `"Pending"`, `"Verified"`, `"Failed"` |

`*big.Int` values are serialized as decimal strings (base-10) via `(*big.Int).String()` and deserialized via `(*big.Int).SetString(s, 10)`.

During `UnmarshalJSON`, PeerID and EVMAddress are re-derived from the public key rather than taken from the JSON payload.

### Registering a Custom Backend

```go
func init() {
    account.RegisterBackend(&myBackend{})
}

type myBackend struct{}

func (m *myBackend) Kind() string           { return "my-backend" }
func (m *myBackend) ValidateLocator(locator string) error {
    if locator == "" {
        return fmt.Errorf("empty locator")
    }
    return nil
}
func (m *myBackend) ParseLocator(locator string) (string, error) {
    if err := m.ValidateLocator(locator); err != nil {
        return "", err
    }
    return locator, nil
}
func (m *myBackend) Metadata() account.BackendMetadata {
    return account.BackendMetadata{
        DisplayName:    "My Backend",
        Description:    "Custom messaging backend",
        LocatorFormat:  "endpoint-id",
        LocatorExample: "my-endpoint-123",
    }
}
func (m *myBackend) Technology() account.TopicTechnology {
    return account.TopicTechnologyCustom
}
```

After registration, the backend can be used with:

```go
builder.WithStdInBackend("my-backend", "my-endpoint-123")
```

### Using Validation

```go
// Fail-fast validation
if err := acct.Validate(); err != nil {
    var ae *account.AccountError
    if errors.As(err, &ae) {
        fmt.Printf("Operation: %s, Kind: %s\n", ae.Op, ae.Kind)
    }
}

// Collect all errors
result := acct.ValidateAll()
if result.HasErrors() {
    for _, err := range result.Errors() {
        fmt.Println(err)
    }
}

// Fluent validator
result := account.NewAccountValidator().
    ValidatePublicKey(pubKey).
    ValidateAccountType(account.AccountTypeParent).
    ValidateDID(did).
    Result()

if !result.Valid() {
    return result.Error()
}
```

### Ledger Attachment and Verification

```go
// Attach to a ledger via the builder
parent, err := account.NewParentAccountBuilder(pubKey, did).
    WithLedgerAttachment("hedera-mainnet", "0.0.98765").
    Build()

// Check attachment state
if att := parent.LedgerAttachment(); att != nil {
    fmt.Println(att.State())              // "Detached"
    fmt.Println(att.LedgerIdentifier())   // "hedera-mainnet"
    fmt.Println(att.AttachedAddress())     // "0.0.98765"
}

// Verify with a LedgerVerifier implementation
result, err := account.VerifyAccountAttachment(parent, myVerifier)
if err != nil {
    return err
}
if result.IsSuccess() {
    parent.LedgerAttachment().SetVerified()
}
```

---

## Section 6: Edge Cases & Common Misuse

### 1. Zero-value NeuronAccount

`IsZero()` returns `true` for the default struct. Always construct via a builder.

```go
var acct account.NeuronAccount  // zero-value
fmt.Println(acct.IsZero())      // true
fmt.Println(acct.String())      // "NeuronAccount{zero-value}"
```

A zero-value `NeuronAccount` will fail validation. Do not use it as a sentinel -- check `IsZero()` explicitly.

### 2. IsZero() for Shared accounts

For Shared accounts, `IsZero()` checks `multisigKey == nil || multisigKey.IsZero()` instead of `publicKey.IsZero()`. A valid Shared account has a zero `publicKey` by design.

```go
func (a NeuronAccount) IsZero() bool {
    if a.accountType == AccountTypeShared {
        return a.multisigKey == nil || a.multisigKey.IsZero()
    }
    return a.publicKey.IsZero()
}
```

This means a default `NeuronAccount{}` (with `AccountTypeUnspecified`) is considered zero based on `publicKey`, not `multisigKey`.

### 3. Equal() for Shared accounts

`Equal()` compares `MultisigKey` configuration (via `MultisigKey.Equal()`) for Shared accounts, not `publicKey`. Comparing a Shared account with a non-Shared account always returns `false`. Both accounts must be non-zero.

```go
func (a NeuronAccount) Equal(other NeuronAccount) bool {
    if a.IsZero() || other.IsZero() {
        return false
    }
    if a.accountType == AccountTypeShared && other.accountType == AccountTypeShared {
        if a.multisigKey == nil || other.multisigKey == nil {
            return false
        }
        return a.multisigKey.Equal(*other.multisigKey)
    }
    if a.accountType != AccountTypeShared && other.accountType != AccountTypeShared {
        return a.publicKey.Equal(other.publicKey)
    }
    return false  // different account types
}
```

### 4. `*big.Int` pointer exposure

`CreditBalance()`, `BalanceAllocation()`, and `Balance()` return `*big.Int` pointers. Callers can mutate the internal state of the account through these pointers. Always check for `nil` before use, and copy if you need to modify the value.

```go
bal := acct.CreditBalance()
if bal == nil {
    // handle nil -- not a Parent account or balance not set
    return
}

// WRONG: mutates the account's internal state
bal.Add(bal, big.NewInt(100))

// RIGHT: copy first
copied := new(big.Int).Set(bal)
copied.Add(copied, big.NewInt(100))
```

### 5. Comm channel requirements

Child accounts MUST have all 3 channels (stdIn, stdOut, stdErr). Parent and Shared accounts MUST NOT have any channels. The builder enforces this at `Build()` time.

```go
// This will fail at Build() -- Parent accounts prohibit comm channels
_, err := account.NewParentAccountBuilder(pubKey, did).
    WithStdInHedera("0.0.1001").  // error accumulated
    Build()
// err: "parent validation failed: field stdIn is prohibited for Parent accounts"

// This will fail at Build() -- Child missing required channels
_, err := account.NewChildAccountBuilder(childPubKey, pubKey).
    WithStdInHedera("0.0.1001").
    // Missing stdOut and stdErr
    Build()
// err: "child validation failed: required field stdOut is missing"
```

### 6. PeerID consistency

Every `ReachableAddr` must contain `/p2p/<PeerID>` matching the account's derived PeerID. This is enforced during `Build()` and `Validate()`.

```go
// Addresses with wrong PeerID will cause Build() to fail
_, err := account.NewChildAccountBuilder(childPubKey, pubKey).
    WithHederaTopics("0.0.1001", "0.0.1002", "0.0.1003").
    WithReachableAddr("/ip4/1.2.3.4/tcp/4001/p2p/WRONG_PEER_ID").
    Build()
// err: "reachable address at index 0 has wrong PeerID"
```

Shared accounts cannot have reachable addresses at all because they have no PeerID to validate against.

### 7. DID-key mismatch

Parent accounts require a DID that matches their `publicKey`. `ValidateDIDMatchesKey` checks this. The builder validates at `Build()` time. If the DID method implements `NeuronDIDWithKey` (like `did:key`), the key is extracted and compared. If the DID method does not implement `NeuronDIDWithKey`, the check passes (cannot validate).

```go
// This will fail -- DID was created from a different key
otherPriv, _ := keylib.GeneratePrivateKey()
otherDID, _ := didkey.FromPublicKey(otherPriv.PublicKey())

_, err := account.NewParentAccountBuilder(pubKey, otherDID).Build()
// err: "DID-key mismatch: DID does not match the provided public key"
```

### 8. `didParsers` map is NOT protected by a mutex

`RegisterDIDParser` writes to a package-level `map[string]DIDParser`. It is not protected by any synchronization primitive. Only call it from `init()` functions to ensure registration happens during single-threaded package initialization.

```go
// CORRECT: register in init()
func init() {
    account.RegisterDIDParser("mymethod", myParser)
}

// WRONG: registering at runtime from concurrent goroutines
go func() {
    account.RegisterDIDParser("method1", parser1) // data race
}()
go func() {
    account.RegisterDIDParser("method2", parser2) // data race
}()
```

### 9. Builder error accumulation

Errors accumulate during `With*()` calls and constructor calls. `Build()` returns the first error. Use `Errors()` to see all accumulated errors.

```go
builder := account.NewParentAccountBuilder(keylib.NeuronPublicKey{}, nil) // both invalid
builder.WithCreditBalance(big.NewInt(-1)) // negative balance

// Build returns first error only
_, err := builder.Build()
fmt.Println(err) // first error: zero-value NeuronPublicKey

// Errors() returns all accumulated errors
for _, e := range builder.Errors() {
    fmt.Println(e)
}
// Error 1: zero-value NeuronPublicKey
// Error 2: required field DID is missing
// Error 3: credit balance cannot be negative
```

### 10. Validate() vs ValidateAll()

`Validate()` returns the first error and wraps it with `AccountError` context (operation name, error kind, details). `ValidateAll()` returns a `ValidationResult` containing all errors. They run the same set of checks.

```go
// Validate() -- fail-fast, returns first error with full context
err := acct.Validate()
// Example: "account.NeuronAccount.Validate: Validation: parent account validation failed:
//           account.ValidateParentAccount: MissingRequired: required field DID is missing"

// ValidateAll() -- collects everything
result := acct.ValidateAll()
fmt.Println(len(result.Errors())) // total number of validation issues

// The composed error supports errors.As
err = result.Error()
var ae *account.AccountError
if errors.As(err, &ae) {
    fmt.Println(ae.Kind) // kind of the first matching AccountError
}
```
