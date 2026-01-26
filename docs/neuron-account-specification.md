# NeuronAccount Module Technical Specification

**Version:** 1.3.0
**Status:** Implemented and Tested
**Test Coverage:** account 92.0%, didkey 87.0%

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Backend Interface](#3-backend-interface)
4. [Data Structures](#4-data-structures)
5. [Identity Derivation](#5-identity-derivation)
6. [DID:Key Encoding](#6-didkey-encoding)
7. [Communication Addresses](#7-communication-addresses)
8. [Reachable Addresses](#8-reachable-addresses)
9. [Account Builder](#9-account-builder)
10. [Validation Rules](#10-validation-rules)
11. [Error Taxonomy](#11-error-taxonomy)
12. [Serialization Format](#12-serialization-format)
13. [Test Vectors](#13-test-vectors)
14. [API Reference](#14-api-reference)
- [Appendix A: Test File Summary](#appendix-a-test-file-summary)
- [Appendix B: Code Cleanup History](#appendix-b-code-cleanup-history)

---

## 1. Overview

The NeuronAccount module implements a blockchain-agnostic agent identity and endpoint descriptor. The module does NOT perform communication; it describes WHERE and HOW an agent can be reached.

### Core Responsibilities

- Cryptographic identity management (secp256k1 public keys)
- Hierarchical account relationships (Parent/Child)
- Communication channel specification (stdIn/stdOut/stdErr)
- Peer-to-peer address publication with PeerID consistency enforcement

### Non-Responsibilities

- Message transport or routing
- Key generation or storage
- Network connectivity

---

## 2. Architecture

### Module Dependencies

```
account/
    |
    +---> keylib/
    |       |-- NeuronPublicKey (secp256k1 compressed, 33 bytes)
    |       |-- NeuronPrivateKey
    |       |-- PeerID (libp2p identity)
    |       +-- EVMAddress (Ethereum address)
    |
    +---> multiaddr/ (libp2p multiaddr parsing)
```

### Package Structure

```
account/
|-- account.go           NeuronAccount struct
|-- account_type.go      AccountType enumeration
|-- backend.go           Backend interface definition
|-- backend_custom.go    Custom backend implementation (v1.3.0)
|-- backend_hedera.go    Hedera backend implementation
|-- backend_kafka.go     Kafka backend implementation
|-- builder.go           Fluent construction API
|-- comm_address.go      Communication address types (technology-agnostic)
|-- did.go               NeuronDID interface and DID document types
|-- errors.go            Error types and constructors
|-- message.go           Topic message structures
|-- reachable.go         Reachable address handling
|-- registry.go          Backend registry (plugin system)
|-- topic.go             TopicTechnology type
|-- validation.go        Validation functions
+-- didkey/
    +-- didkey.go        did:key method implementation
```

### Component Relationships

```
+-------------------+
| AccountBuilder    |
+--------+----------+
         |
         | Build()
         v
+-------------------+      +-------------------+
| NeuronAccount     |<---->| Validation System |
+-------------------+      +-------------------+
         |
         +---> CommAddress (stdIn, stdOut, stdErr)
         |           |
         |           v
         |     +-------------------+      +-------------------+
         |     | Backend Registry  |<---->| Backend Interface |
         |     +-------------------+      +-------------------+
         |           |                           |
         |           +---> HederaBackend         |
         |           +---> KafkaBackend          |
         |           +---> CustomBackend         |
         |           +---> (External Backends...)|
         |
         +---> ReachableAddrs (connection endpoints)
         +---> NeuronDID (identity document)
```

---

## 3. Backend Interface

The Backend Interface provides an extensible plugin system for communication technologies. New messaging backends can be added without modifying the core account package.

### 3.1 Design Principles

1. **Registry Pattern** - Backends register themselves via `init()` functions
2. **Interface-driven** - All backends implement the `Backend` interface
3. **Forward Compatibility** - Unknown backend kinds are accepted without validation
4. **Thread-safe** - Registry operations are protected by `sync.RWMutex`

### 3.2 Backend Interface

```go
type Backend interface {
    // Kind returns the unique identifier for this backend
    // Examples: "hedera-topic", "kafka-topic", "pulsar-topic"
    Kind() string

    // ValidateLocator checks if a locator string is valid
    ValidateLocator(locator string) error

    // ParseLocator validates and normalizes a locator string
    ParseLocator(locator string) (string, error)

    // Metadata returns descriptive information about this backend
    Metadata() BackendMetadata

    // Technology returns the TopicTechnology for this backend (v1.3.0)
    // This enables dynamic mapping from CommAddressKind to TopicTechnology
    // without requiring hardcoded switch statements in core files.
    Technology() TopicTechnology
}
```

### 3.3 Backend Metadata

```go
type BackendMetadata struct {
    DisplayName     string            // Human-readable name
    Description     string            // Short description
    LocatorFormat   string            // Expected format
    LocatorExample  string            // Example locator
    RequiresConfig  bool              // Needs additional configuration
    Properties      map[string]string // Backend-specific properties
}
```

### 3.4 Backend Registry

```go
// Register a new backend (typically in init())
func RegisterBackend(b Backend)

// Retrieve a backend by kind
func GetBackend(kind string) (Backend, bool)

// List all registered backend kinds
func ListBackends() []string

// Check if a backend kind is registered
func IsRegisteredBackend(kind string) bool

// Get metadata for a backend
func GetBackendMetadata(kind string) (BackendMetadata, bool)

// List all backends with their metadata
func ListBackendsWithMetadata() map[string]BackendMetadata
```

### 3.5 Built-in Backends

| Backend | Kind | Locator Format | Example | Constant |
|---------|------|----------------|---------|----------|
| Hedera Consensus Service | `hedera-topic` | `shard.realm.topic` | `0.0.12345` | `HederaTopicKind` |
| Apache Kafka | `kafka-topic` | `topic-name` | `my-events-topic` | `KafkaTopicKind` |
| Custom/Third-party | `custom` | implementation-defined | `my-custom-endpoint` | `CustomTopicKind` |

**Note (v1.3.0):** Backend kind constants (`HederaTopicKind`, `KafkaTopicKind`, `CustomTopicKind`) are now defined in their respective backend files (`backend_hedera.go`, `backend_kafka.go`, `backend_custom.go`), not in `comm_address.go`. This enables full decoupling where adding new backends requires only creating a new file.

### 3.6 Adding Custom Backends

To add a new backend, implement the `Backend` interface and register it:

```go
// backends/pulsar/pulsar.go
package pulsar

import "github.com/aspect-build/neuron-go-hedera-sdk/account"

// PulsarTopicKind is the unique identifier for the Pulsar backend.
const PulsarTopicKind = "pulsar-topic"

func init() {
    account.RegisterBackend(&pulsarBackend{})
}

type pulsarBackend struct{}

func (p *pulsarBackend) Kind() string { return PulsarTopicKind }

func (p *pulsarBackend) ValidateLocator(locator string) error {
    // Validate Pulsar topic format: persistent://tenant/namespace/topic
    if !strings.HasPrefix(locator, "persistent://") &&
       !strings.HasPrefix(locator, "non-persistent://") {
        return fmt.Errorf("invalid Pulsar topic format")
    }
    return nil
}

func (p *pulsarBackend) ParseLocator(locator string) (string, error) {
    if err := p.ValidateLocator(locator); err != nil {
        return "", err
    }
    return locator, nil
}

func (p *pulsarBackend) Metadata() account.BackendMetadata {
    return account.BackendMetadata{
        DisplayName:    "Apache Pulsar",
        Description:    "Cloud-native multi-tenant streaming",
        LocatorFormat:  "persistent://tenant/namespace/topic",
        LocatorExample: "persistent://public/default/my-topic",
    }
}

// Technology returns the TopicTechnology for Pulsar (v1.3.0)
func (p *pulsarBackend) Technology() account.TopicTechnology {
    return account.TopicTechnologyCustom // Or define a new TopicTechnology if needed
}

// NewPulsarTopicAddress creates a CommAddress for a Pulsar topic.
func NewPulsarTopicAddress(locator string) (account.CommAddress, error) {
    return account.NewCommAddress(PulsarTopicKind, locator)
}
```

### 3.7 Supported Messaging Technologies (2026)

The extensible backend system supports these modern messaging technologies:

| Technology | Use Case | Backend Kind |
|------------|----------|--------------|
| Apache Kafka | High-throughput enterprise streaming | `kafka-topic` |
| Apache Pulsar | Cloud-native multi-tenant streaming | `pulsar-topic` |
| NATS/JetStream | Lightweight microservices & IoT | `nats-topic` |
| Redis Streams | Low-latency in-memory messaging | `redis-stream` |
| AWS SQS/SNS | Managed cloud queuing | `aws-sqs` |
| Google Cloud Pub/Sub | Managed cloud messaging | `gcp-pubsub` |
| RabbitMQ | Traditional message broker | `rabbitmq-topic` |
| Hedera Consensus Service | Blockchain-based consensus | `hedera-topic` |

---

## 4. Data Structures

### 4.1 NeuronAccount

```go
type NeuronAccount struct {
    // Identity layer - derived from keylib
    publicKey    keylib.NeuronPublicKey   // 33-byte compressed secp256k1
    peerID       keylib.PeerID            // libp2p peer identifier
    evmAddress   keylib.EVMAddress        // 20-byte Ethereum address

    // Hierarchy layer
    accountType  AccountType              // Parent (1) or Child (2)
    did          NeuronDID                // Required for Parent, nil for Child
    parentPubKey keylib.NeuronPublicKey   // Required for Child, zero for Parent

    // Communication layer
    stdIn        CommAddress              // Inbound message channel
    stdOut       CommAddress              // Outbound message channel
    stdErr       CommAddress              // Error/diagnostic channel

    // Reachability layer
    reachableAddrs ReachableAddrs         // P2P connection endpoints
}
```

### 4.2 AccountType

```go
type AccountType int

const (
    AccountTypeUnspecified AccountType = 0   // Invalid state
    AccountTypeParent      AccountType = 1   // Root identity with DID
    AccountTypeChild       AccountType = 2   // Derived identity with parent reference
)
```

| Type            | DID Required | Parent Required | Valid |
| --------------- | ------------ | --------------- | ----- |
| Unspecified (0) | -            | -               | No    |
| Parent (1)      | Yes          | No              | Yes   |
| Child (2)       | No           | Yes             | Yes   |

### 4.3 CommAddress

```go
type CommAddress struct {
    kind    CommAddressKind   // Address type discriminator
    locator string            // Technology-specific address
}

type CommAddressKind string

// Kind constants are defined in their respective backend files (v1.3.0):
//   - HederaTopicKind = "hedera-topic"  (backend_hedera.go)
//   - KafkaTopicKind  = "kafka-topic"   (backend_kafka.go)
//   - CustomTopicKind = "custom"        (backend_custom.go)

// IsValid checks if this is a registered address kind (v1.3.0)
// Uses the backend registry instead of hardcoded switch statement.
func (k CommAddressKind) IsValid() bool {
    return IsRegisteredBackend(string(k))
}

// TopicTechnologyForKind returns the corresponding TopicTechnology (v1.3.0)
// Queries the registered backend instead of using a hardcoded switch.
func (k CommAddressKind) TopicTechnologyForKind() TopicTechnology {
    backend, ok := GetBackend(string(k))
    if !ok {
        return TopicTechnologyCustom // Default fallback
    }
    return backend.Technology()
}
```

### 4.4 ReachableAddr

```go
type ReachableAddr struct {
    ma     multiaddr.Multiaddr   // Parsed libp2p multiaddr
    peerID keylib.PeerID         // Extracted from /p2p/ component
}

type ReachableAddrs struct {
    addrs []ReachableAddr
}
```

---

## 5. Identity Derivation

### Derivation Chain

```
NeuronPrivateKey (secp256k1)
        |
        | .PublicKey()
        v
NeuronPublicKey (33 bytes, compressed)
        |
        +-----> .PeerID()     --> PeerID (base58btc, libp2p)
        |
        +-----> .EVMAddress() --> EVMAddress (20 bytes, keccak256)
        |
        +-----> didkey.FromPublicKey() --> did:key (multicodec + base58btc)
```

### PeerID Derivation

1. Hash public key bytes using identity multihash (for keys <= 42 bytes)
2. Encode as base58btc with libp2p peer ID prefix

### EVMAddress Derivation

1. Take uncompressed public key (64 bytes, without 0x04 prefix)
2. Compute Keccak-256 hash
3. Take last 20 bytes

---

## 6. DID:Key Encoding

### Multicodec Specification

The did:key method uses multicodec to identify the key type.

| Key Type      | Multicodec Prefix | Hex  |
| ------------- | ----------------- | ---- |
| secp256k1-pub | 0xe7, 0x01        | e701 |

### Encoding Algorithm

```
Input:  33-byte compressed secp256k1 public key
Output: did:key:z<base58btc-encoded-data>

1. Prepend multicodec prefix [0xe7, 0x01] to public key bytes
2. Encode result with base58btc (Bitcoin alphabet)
3. Prepend 'z' multibase prefix
4. Prepend "did:key:" scheme
```

### Decoding Algorithm

```
Input:  "did:key:z<identifier>"
Output: 33-byte compressed secp256k1 public key

1. Verify "did:key:" prefix
2. Extract identifier after "did:key:"
3. Verify 'z' multibase prefix (base58btc)
4. Decode base58btc
5. Verify multicodec prefix [0xe7, 0x01]
6. Extract remaining 33 bytes as public key
7. Validate public key is on secp256k1 curve
```

### Example

```
Public Key (hex):  02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae
With Multicodec:   e70102759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae
Base58btc:         Q3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9
With 'z' prefix:   zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9
Full DID:          did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9
```

---

## 7. Communication Addresses

Communication addresses use the extensible Backend Interface (Section 3) for validation and normalization. New messaging technologies can be added by registering custom backends.

### 7.1 Creating Addresses

```go
// Using the generic constructor (recommended)
addr, err := NewCommAddress("hedera-topic", "0.0.12345")
addr, err := NewCommAddress("kafka-topic", "my-events")
addr, err := NewCommAddress("pulsar-topic", "persistent://tenant/ns/topic")

// Using convenience constructors (built-in backends)
addr, err := NewHederaTopicAddress("0.0.12345")
addr, err := NewKafkaTopicAddress("my-events")

// Parsing from string format
addr, err := ParseCommAddress("hedera-topic:0.0.12345")
```

### 7.2 Hedera Topic Address

Format: `shard.realm.num`

```
Validation regex: ^[0-9]+\.[0-9]+\.[0-9]+$

Valid examples:
  0.0.12345
  0.0.999999999
  1.0.54321

Invalid examples:
  0.0           (only 2 parts)
  0.0.0.0       (4 parts)
  a.b.c         (non-numeric)
  0.0.-1        (negative)
  0.0.12 345    (contains space)
```

### 7.3 Kafka Topic Address

Format: Alphanumeric with dots, underscores, hyphens. Max 249 characters.

```
Validation regex: ^[a-zA-Z0-9._-]{1,249}$

Valid examples:
  my-topic
  agent.events.v1
  test_topic_123

Invalid examples:
  (empty string)
  topic with spaces
  topic@special#chars
  (string > 249 chars)
```

### 7.4 Serialized Format

```
<kind>:<locator>

Examples:
  hedera-topic:0.0.12345
  kafka-topic:agent.events
  custom:redis://localhost:6379/0
```

---

## 8. Reachable Addresses

### 8.1 PeerID Consistency Rule (SPEC 4.2)

All reachable addresses MUST contain a PeerID that matches the account's derived PeerID.

```
Account PeerID (derived from publicKey): 16Uiu2HAm...

Reachable Address:
  /ip4/192.168.1.1/tcp/4001/p2p/16Uiu2HAm...
                               ^^^^^^^^^^^^
                               Must match account PeerID
```

Enforcement points:

1. `WithReachableAddrValidated()` - validates immediately during builder construction
2. `Build()` - validates all addresses before returning NeuronAccount
3. `Validate()` - validates on existing accounts

### 8.2 Circuit Relay PeerID Extraction (SPEC 4.5)

For circuit relay addresses, the TARGET PeerID is extracted, not the relay's PeerID.

```
Format:
  /ip4/<relay-ip>/tcp/<port>/p2p/<RELAY_PEER>/p2p-circuit/p2p/<TARGET_PEER>

Extraction:
  - Find the LAST /p2p/<peer> component
  - If /p2p-circuit/ exists, take the PeerID AFTER it
  - This is the TARGET peer (the account being reached)
```

Example:

```
Address:
  /ip4/192.168.1.100/tcp/4001/p2p/12D3KooWDpJ7.../p2p-circuit/p2p/16Uiu2HAm...

Components:
  - Relay IP:      192.168.1.100
  - Relay Port:    4001
  - Relay PeerID:  12D3KooWDpJ7... (IGNORED for validation)
  - Target PeerID: 16Uiu2HAm...    (EXTRACTED and validated)
```

### 8.3 Supported Multiaddr Protocols

| Protocol Stack | Example                                          |
| -------------- | ------------------------------------------------ |
| IPv4/TCP       | `/ip4/192.168.1.1/tcp/4001/p2p/<PeerID>`         |
| IPv6/TCP       | `/ip6/::1/tcp/4001/p2p/<PeerID>`                 |
| IPv4/UDP/QUIC  | `/ip4/192.168.1.1/udp/4001/quic-v1/p2p/<PeerID>` |
| IPv4/TCP/WS    | `/ip4/192.168.1.1/tcp/443/ws/p2p/<PeerID>`       |
| DNS/TCP        | `/dns/example.com/tcp/4001/p2p/<PeerID>`         |
| DNS4/TCP       | `/dns4/example.com/tcp/4001/p2p/<PeerID>`        |
| Circuit Relay  | `/ip4/.../p2p/<Relay>/p2p-circuit/p2p/<Target>`  |

Required: All addresses MUST terminate with `/p2p/<PeerID>`.

---

## 9. Account Builder

### 9.1 Construction Flow

```
1. NewParentAccountBuilder(publicKey, did) or NewChildAccountBuilder(publicKey, parentKey)
   - Validates required parameters
   - Records errors if invalid (does not fail immediately)

2. WithStdIn*/WithStdOut*/WithStdErr* methods
   - Parses and validates address format
   - Records errors if invalid

3. WithReachableAddr*/WithReachableAddrs methods
   - Parses multiaddr format
   - Records errors if invalid
   - WithReachableAddrValidated also validates PeerID immediately

4. Build()
   - Checks for accumulated errors (returns first if any)
   - Validates account-type-specific requirements
   - Validates DID matches public key (Parent accounts)
   - Validates child key differs from parent key (Child accounts)
   - Derives PeerID and EVMAddress from public key
   - Validates all reachable address PeerIDs match derived PeerID
   - Constructs and returns NeuronAccount

5. MustBuild()
   - Calls Build()
   - Panics if error (for tests and initialization of known-valid data)
```

### 9.2 Error Accumulation

The builder collects errors rather than failing on first error:

```go
b := NewParentAccountBuilder(zeroPubKey, nil)     // Records 2 errors
b.WithStdInHedera("invalid")                       // Records 1 error
b.WithReachableAddr("bad")                         // Records 1 error

// Query errors before Build()
b.HasErrors()     // true
b.Errors()        // []error with 4 entries

// Build() returns first error only
_, err := b.Build()  // Returns error about zero public key
```

### 9.3 Builder Methods

| Method                       | Parameters                    | Validates                       |
| ---------------------------- | ----------------------------- | ------------------------------- |
| `WithStdIn`                  | CommAddress                   | Address.Validate()              |
| `WithStdInBackend`           | kind, locator strings         | Backend-specific validation     |
| `WithStdInHedera`            | topicID string                | Hedera format                   |
| `WithStdInKafka`             | topicName string              | Kafka format                    |
| `WithStdOut`                 | CommAddress                   | Address.Validate()              |
| `WithStdOutBackend`          | kind, locator strings         | Backend-specific validation     |
| `WithStdOutHedera`           | topicID string                | Hedera format                   |
| `WithStdOutKafka`            | topicName string              | Kafka format                    |
| `WithStdErr`                 | CommAddress                   | Address.Validate()              |
| `WithStdErrBackend`          | kind, locator strings         | Backend-specific validation     |
| `WithStdErrHedera`           | topicID string                | Hedera format                   |
| `WithStdErrKafka`            | topicName string              | Kafka format                    |
| `WithBackendTopics`          | kind, in, out, err strings    | All same backend formats        |
| `WithHederaTopics`           | in, out, err strings          | All Hedera formats              |
| `WithKafkaTopics`            | in, out, err strings          | All Kafka formats               |
| `WithReachableAddr`          | multiaddr string              | Multiaddr parse                 |
| `WithReachableAddrs`         | ...strings                    | Each multiaddr                  |
| `WithReachableAddrValidated` | multiaddr string              | Multiaddr + PeerID match        |

---

## 10. Validation Rules

### 10.1 Rule Matrix

| Validation                 | Parent Account    | Child Account     | Error Kind              |
| -------------------------- | ----------------- | ----------------- | ----------------------- |
| PublicKey is non-zero      | Required          | Required          | ErrKindZeroValue        |
| AccountType is valid       | Required          | Required          | ErrKindInvalidAccount   |
| DID is present             | Required          | Must be nil       | ErrKindMissingRequired  |
| DID validates              | Required          | N/A               | ErrKindInvalidDID       |
| DID matches PublicKey      | Required          | N/A               | ErrKindInvalidDID       |
| ParentPubKey is present    | Must be zero      | Required          | ErrKindMissingRequired  |
| Child != Parent key        | N/A               | Required          | ErrKindInvalidHierarchy |
| CommAddress valid (if set) | Optional          | Optional          | ErrKindInvalidAddress   |
| ReachableAddr PeerID match | Required (if any) | Required (if any) | ErrKindPeerIDMismatch   |

### 10.2 Validation Functions

```go
// Individual validations
func ValidatePublicKey(key keylib.NeuronPublicKey) error
func ValidateAccountType(at AccountType) error
func ValidateDID(did NeuronDID) error
func ValidateDIDMatchesKey(did NeuronDID, key keylib.NeuronPublicKey) error
func ValidateParentAccount(key keylib.NeuronPublicKey, did NeuronDID) error
func ValidateChildAccount(key, parentKey keylib.NeuronPublicKey) error
func ValidateCommAddress(addr CommAddress) error
func ValidateCommAddressNonEmpty(addr CommAddress, fieldName string) error

// Account-level validation
func (a NeuronAccount) Validate() error        // Returns first error
func (a NeuronAccount) ValidateAll() *ValidationResult  // Returns all errors
```

### 10.3 Fluent Validator

```go
v := NewAccountValidator().
    ValidatePublicKey(pubKey).
    ValidateAccountType(AccountTypeParent).
    ValidateDID(did).
    ValidateDIDMatchesKey(did, pubKey).
    ValidateCommAddress(addr)

v.IsValid()       // bool
v.Error()         // First error or nil
v.Result()        // *ValidationResult with all errors
v.Result().Errors // []error
```

---

## 11. Error Taxonomy

### 11.1 AccountError Structure

```go
type AccountError struct {
    Op      string            // Operation that failed (e.g., "Build", "ValidateDID")
    Kind    AccountErrorKind  // Error category
    Details string            // Human-readable description
    Err     error             // Underlying error (for wrapping)
}

func (e *AccountError) Error() string
func (e *AccountError) Unwrap() error
func (e *AccountError) Is(target error) bool  // Matches by Kind
```

### 11.2 Error Kinds

| Kind                    | Value | Description                               |
| ----------------------- | ----- | ----------------------------------------- |
| ErrKindInvalidAccount   | 1     | Account configuration is invalid          |
| ErrKindInvalidTopic     | 2     | Topic format is invalid                   |
| ErrKindInvalidAddress   | 3     | Address format is invalid                 |
| ErrKindInvalidDID       | 4     | DID is invalid or doesn't match key       |
| ErrKindPeerIDMismatch   | 5     | PeerID doesn't match expected value       |
| ErrKindMissingRequired  | 6     | Required field is missing or zero         |
| ErrKindValidation       | 7     | General validation failure                |
| ErrKindZeroValue        | 8     | Zero value where non-zero required        |
| ErrKindInvalidHierarchy | 9     | Hierarchy rule violation (child = parent) |

### 11.3 Sentinel Errors

```go
var (
    ErrZeroPublicKey = &AccountError{Kind: ErrKindZeroValue, Details: "public key is zero"}
    ErrMissingDID    = &AccountError{Kind: ErrKindMissingRequired, Details: "DID is required"}
    ErrMissingParent = &AccountError{Kind: ErrKindMissingRequired, Details: "parent key is required"}
)
```

### 11.4 Error Matching

```go
// Match by kind using errors.Is
if errors.Is(err, &AccountError{Kind: ErrKindPeerIDMismatch}) {
    // Handle PeerID mismatch
}

// Extract AccountError using errors.As
var ae *AccountError
if errors.As(err, &ae) {
    fmt.Printf("Operation: %s, Kind: %s\n", ae.Op, ae.Kind)
}
```

---

## 12. Serialization Format

### 12.1 JSON Schema

```json
{
  "publicKey": "string (hex with 0x prefix)",
  "peerId": "string (base58btc)",
  "evmAddress": "string (hex with 0x prefix)",
  "accountType": "string (Parent|Child)",
  "did": "string (optional, DID URI)",
  "parentPublicKey": "string (optional, hex with 0x prefix)",
  "stdIn": "string (optional, kind:locator)",
  "stdOut": "string (optional, kind:locator)",
  "stdErr": "string (optional, kind:locator)",
  "reachableAddrs": ["string (multiaddr)", "..."]
}
```

### 12.2 Field Presence Rules

| Field           | Parent Account    | Child Account    |
| --------------- | ----------------- | ---------------- |
| publicKey       | Always            | Always           |
| peerId          | Always            | Always           |
| evmAddress      | Always            | Always           |
| accountType     | Always ("Parent") | Always ("Child") |
| did             | Always            | Omitted          |
| parentPublicKey | Omitted           | Always           |
| stdIn           | If set            | If set           |
| stdOut          | If set            | If set           |
| stdErr          | If set            | If set           |
| reachableAddrs  | If non-empty      | If non-empty     |

### 12.3 Example: Parent Account

```json
{
  "publicKey": "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae",
  "peerId": "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq",
  "evmAddress": "0x742d35Cc6634C0532925a3b844Bc9e7595f2bD23",
  "accountType": "Parent",
  "did": "did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9",
  "stdIn": "hedera-topic:0.0.111",
  "stdOut": "hedera-topic:0.0.222",
  "stdErr": "hedera-topic:0.0.333",
  "reachableAddrs": [
    "/ip4/192.168.1.1/tcp/4001/p2p/16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
  ]
}
```

### 12.4 Example: Child Account

```json
{
  "publicKey": "0x03abc123...",
  "peerId": "16Uiu2HAm...",
  "evmAddress": "0xdef456...",
  "accountType": "Child",
  "parentPublicKey": "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae",
  "stdIn": "hedera-topic:0.0.444"
}
```

---

## 13. Test Vectors

### 13.1 Reference Public Key

```
Compressed Public Key (hex):
  0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae

Derived Values:
  PeerID:     16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq
  DID:Key:    did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9
```

### 13.2 DID:Key Encoding Vector

```
Input (public key):
  02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae

Step 1 - Add multicodec prefix (0xe7, 0x01):
  e70102759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae

Step 2 - Base58btc encode:
  Q3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9

Step 3 - Add multibase prefix 'z':
  zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9

Step 4 - Add did:key scheme:
  did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9
```

### 13.3 Circuit Relay Vector

```
Input:
  /ip4/192.168.1.100/tcp/4001/p2p/12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN/p2p-circuit/p2p/16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq

Parsing:
  Relay IP:        192.168.1.100
  Relay Port:      4001
  Relay PeerID:    12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN
  Circuit marker:  /p2p-circuit/
  Target PeerID:   16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq

Extracted PeerID for validation:
  16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq
```

### 13.4 Hedera Topic Address Vectors

| Input           | Valid | Notes              |
| --------------- | ----- | ------------------ |
| `0.0.12345`     | Yes   | Standard format    |
| `0.0.999999999` | Yes   | Large topic number |
| `1.0.54321`     | Yes   | Non-zero shard     |
| `0.1.54321`     | Yes   | Non-zero realm     |
| `0.0.0`         | Yes   | Zero topic         |
| `0.0`           | No    | Only 2 parts       |
| `0.0.0.0`       | No    | 4 parts            |
| `a.b.c`         | No    | Non-numeric        |
| `-1.0.0`        | No    | Negative number    |
| `0.0.12345 `    | No    | Trailing space     |

---

## 14. API Reference

### 14.1 Constructors

```go
// Parent account builder
func NewParentAccountBuilder(publicKey keylib.NeuronPublicKey, did NeuronDID) *AccountBuilder

// Child account builder
func NewChildAccountBuilder(publicKey, parentPubKey keylib.NeuronPublicKey) *AccountBuilder

// Address constructors (generic)
func NewCommAddress(kind, locator string) (CommAddress, error)
func ParseCommAddress(s string) (CommAddress, error)
func MustParseCommAddress(s string) CommAddress

// Address constructors (convenience - defined in backend files v1.3.0)
func NewHederaTopicAddress(topicID string) (CommAddress, error)    // backend_hedera.go
func NewKafkaTopicAddress(topic string) (CommAddress, error)       // backend_kafka.go
func NewCustomTopicAddress(locator string) (CommAddress, error)    // backend_custom.go
func NewCustomAddress(kind CommAddressKind, locator string) (CommAddress, error)  // Deprecated

// Reachable address constructors
func ParseReachableAddr(multiaddr string) (ReachableAddr, error)
func MustParseReachableAddr(multiaddr string) ReachableAddr
func NewReachableAddrFromStringWithValidation(multiaddr string, expectedPeerID keylib.PeerID) (ReachableAddr, error)

// Backend registry functions
func RegisterBackend(b Backend)
func GetBackend(kind string) (Backend, bool)
func ListBackends() []string
func IsRegisteredBackend(kind string) bool
func GetBackendMetadata(kind string) (BackendMetadata, bool)
func ListBackendsWithMetadata() map[string]BackendMetadata

// Backend-specific validation (defined in backend files v1.3.0)
func ValidateHederaTopicLocator(locator string) error   // backend_hedera.go
func ValidateKafkaTopicLocator(locator string) error    // backend_kafka.go
```

### 14.2 AccountBuilder Methods

```go
// CommAddress setters (direct)
func (b *AccountBuilder) WithStdIn(addr CommAddress) *AccountBuilder
func (b *AccountBuilder) WithStdOut(addr CommAddress) *AccountBuilder
func (b *AccountBuilder) WithStdErr(addr CommAddress) *AccountBuilder

// CommAddress setters (generic backend)
func (b *AccountBuilder) WithStdInBackend(kind, locator string) *AccountBuilder
func (b *AccountBuilder) WithStdOutBackend(kind, locator string) *AccountBuilder
func (b *AccountBuilder) WithStdErrBackend(kind, locator string) *AccountBuilder

// CommAddress setters (Hedera convenience)
func (b *AccountBuilder) WithStdInHedera(topicID string) *AccountBuilder
func (b *AccountBuilder) WithStdOutHedera(topicID string) *AccountBuilder
func (b *AccountBuilder) WithStdErrHedera(topicID string) *AccountBuilder

// CommAddress setters (Kafka convenience)
func (b *AccountBuilder) WithStdInKafka(topicName string) *AccountBuilder
func (b *AccountBuilder) WithStdOutKafka(topicName string) *AccountBuilder
func (b *AccountBuilder) WithStdErrKafka(topicName string) *AccountBuilder

// Bulk CommAddress setters
func (b *AccountBuilder) WithBackendTopics(kind, in, out, err string) *AccountBuilder
func (b *AccountBuilder) WithHederaTopics(stdIn, stdOut, stdErr string) *AccountBuilder
func (b *AccountBuilder) WithKafkaTopics(stdIn, stdOut, stdErr string) *AccountBuilder

// Reachable address setters
func (b *AccountBuilder) WithReachableAddr(multiaddr string) *AccountBuilder
func (b *AccountBuilder) WithReachableAddrs(multiaddrs ...string) *AccountBuilder
func (b *AccountBuilder) WithReachableAddrValidated(multiaddr string) *AccountBuilder

// Build methods
func (b *AccountBuilder) Build() (NeuronAccount, error)
func (b *AccountBuilder) MustBuild() NeuronAccount
func (b *AccountBuilder) HasErrors() bool
func (b *AccountBuilder) Errors() []error
```

### 14.3 NeuronAccount Methods

```go
// Identity accessors
func (a NeuronAccount) PublicKey() keylib.NeuronPublicKey
func (a NeuronAccount) PeerID() keylib.PeerID
func (a NeuronAccount) EVMAddress() keylib.EVMAddress

// Hierarchy accessors
func (a NeuronAccount) AccountType() AccountType
func (a NeuronAccount) IsParent() bool
func (a NeuronAccount) IsChild() bool
func (a NeuronAccount) DID() NeuronDID
func (a NeuronAccount) ParentPublicKey() keylib.NeuronPublicKey

// Communication accessors
func (a NeuronAccount) StdIn() CommAddress
func (a NeuronAccount) StdOut() CommAddress
func (a NeuronAccount) StdErr() CommAddress

// Reachability accessors
func (a NeuronAccount) ReachableAddrs() ReachableAddrs
func (a NeuronAccount) FirstReachableAddr() ReachableAddr

// Validation
func (a NeuronAccount) IsZero() bool
func (a NeuronAccount) Validate() error
func (a NeuronAccount) ValidateAll() *ValidationResult

// Comparison and serialization
func (a NeuronAccount) Equal(other NeuronAccount) bool
func (a NeuronAccount) MarshalJSON() ([]byte, error)
func (a NeuronAccount) String() string
```

### 14.4 DIDKey Methods (didkey package)

```go
func FromPublicKey(pubKey keylib.NeuronPublicKey) (*DIDKey, error)
func Parse(didString string) (*DIDKey, error)

func (d *DIDKey) String() string
func (d *DIDKey) Method() string
func (d *DIDKey) Identifier() string
func (d *DIDKey) Validate() error
func (d *DIDKey) Equal(other account.NeuronDID) bool
func (d *DIDKey) PublicKey() (keylib.NeuronPublicKey, error)
func (d *DIDKey) MatchesKey(pubKey keylib.NeuronPublicKey) bool
func (d *DIDKey) IsZero() bool
func (d *DIDKey) Resolve() (*DIDDocument, error)
func (d *DIDKey) VerifySignature(message, signature []byte) bool
```

---

## Appendix A: Test File Summary

| Test File             | Source           | Test Cases |
| --------------------- | ---------------- | ---------- |
| errors_test.go        | errors.go        | 25         |
| account_type_test.go  | account_type.go  | 20         |
| comm_address_test.go  | comm_address.go  | 40         |
| reachable_test.go     | reachable.go     | 60         |
| didkey/didkey_test.go | didkey/didkey.go | 35         |
| message_test.go       | message.go       | 25         |
| topic_test.go         | topic.go         | 10         |
| validation_test.go    | validation.go    | 30         |
| builder_test.go       | builder.go       | 40         |
| account_test.go       | account.go       | 35         |
| registry_test.go      | registry.go      | 20         |
| integration_test.go   | (all)            | 15         |

Total: 355 test cases

## Appendix B: Code Cleanup History

### v1.3.0 - Full Decoupling

The Full Decoupling refactoring removed all hardcoded backend-specific code from `comm_address.go`, making the CommAddress system truly technology-agnostic. Each backend now owns its constants and convenience functions.

| Change | Location | Details |
|--------|----------|---------|
| Add `Technology()` method | backend.go | Backend interface now includes `Technology() TopicTechnology` |
| Move `HederaTopicKind` | backend_hedera.go | Constant defined in backend file (was hardcoded in comm_address.go) |
| Move `KafkaTopicKind` | backend_kafka.go | Constant defined in backend file (was hardcoded in comm_address.go) |
| Create `backend_custom.go` | account/ | New file for custom backend with `CustomTopicKind` constant |
| Move `NewHederaTopicAddress()` | backend_hedera.go | Convenience constructor moved from comm_address.go |
| Move `NewKafkaTopicAddress()` | backend_kafka.go | Convenience constructor moved from comm_address.go |
| Move `ValidateHederaTopicLocator()` | backend_hedera.go | Validation function moved from comm_address.go |
| Move `ValidateKafkaTopicLocator()` | backend_kafka.go | Validation function moved from comm_address.go |
| Refactor `IsValid()` | comm_address.go | Now uses `IsRegisteredBackend()` instead of switch statement |
| Refactor `TopicTechnologyForKind()` | comm_address.go | Now queries backend registry instead of switch statement |
| Remove hardcoded constants | comm_address.go | `CommAddressKindHederaTopic`, `CommAddressKindKafkaTopic`, `CommAddressKindCustom` removed |
| Update builder methods | builder.go | `WithStdInHedera()`, `WithStdInKafka()`, etc. use backend constants |

**Result:** Adding new backends now requires only creating a new file - no core file modifications needed.

---

### v1.2.0 - Dead Code Removal

The following dead code was removed in v1.2.0 to improve maintainability:

| Removed Item | Previous Location | Reason |
|--------------|-------------------|--------|
| `interfaces.go` (entire file) | account/ | 8 unused interfaces with 0 external references |
| `Topic` interface | topic.go | Not implemented or used |
| `Publisher` interface | topic.go | Not implemented or used |
| `Subscriber` interface | topic.go | Not implemented or used |
| `TopicFactory` interface | topic.go | Not implemented or used |
| `TopicValidator` interface | topic.go | Not implemented or used |
| `TopicInfo` struct | topic.go | Not used |
| `MessageHandler` func | topic.go | Not used |
| `DIDResolver` interface | did.go | No implementations |
| `DIDMethodWeb/Hedera/PKH/Neuron` | did.go | Unused constants |
| `hederaTopicRegex` | comm_address.go | Duplicated in backend_hedera.go |
| `kafkaTopicRegex` | comm_address.go | Duplicated in backend_kafka.go |
| Fallback validation code | comm_address.go | Dead code (backends always registered) |

**Lines removed:** ~310 lines of dead/redundant code
