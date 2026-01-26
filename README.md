# Neuron Go Hedera SDK

A comprehensive Go SDK for building decentralized agent identities on the Hedera network with multi-ecosystem interoperability.

## Overview

The Neuron Go Hedera SDK provides two core modules:

| Module | Purpose | Coverage |
|--------|---------|----------|
| **keylib** | Cryptographic key management with type-safe ECDSA secp256k1 keys | 87.1% |
| **account** | Agent identity and communication endpoint descriptors | 91.6% |

### Key Features

- **Type-Safe Cryptography** - Immutable key types prevent accidental misuse
- **Multi-Ecosystem Support** - Seamless conversion between Hedera, Ethereum, and libp2p identities
- **Extensible Backend System** - Plugin architecture for any messaging technology
- **DID:Key Support** - W3C-compliant decentralized identifiers
- **Production Ready** - Comprehensive test coverage with race detection

---

## Installation

```bash
go get github.com/aspect-build/neuron-go-hedera-sdk
```

### Requirements

- Go 1.21+ (uses `clear()` builtin for secure memory zeroing)
- Hiero SDK v2.74.0+ (Hedera network interoperability)

---

## Quick Start

### Keylib: Key Management

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

func main() {
    // Generate a new private key
    privateKey, err := keylib.GeneratePrivateKey()
    if err != nil {
        log.Fatal(err)
    }
    defer privateKey.Zeroize() // Secure cleanup

    // Derive identities for multiple ecosystems
    publicKey := privateKey.PublicKey()
    evmAddress := publicKey.EVMAddress()
    peerID, _ := publicKey.PeerID()

    fmt.Printf("EVM Address: %s\n", evmAddress.ChecksumHex())
    fmt.Printf("Peer ID: %s\n", peerID.String())

    // Sign and verify messages
    message := []byte("Hello, Neuron!")
    signature, _ := privateKey.SignMessage(message)

    if publicKey.Verify(message, signature) {
        fmt.Println("Signature verified!")
    }
}
```

### Account: Agent Identity

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

func main() {
    // Generate key and create DID
    privateKey, _ := keylib.GeneratePrivateKey()
    publicKey := privateKey.PublicKey()
    did, _ := didkey.FromPublicKey(publicKey)

    // Build a parent account with communication channels
    neuronAccount, err := account.NewParentAccountBuilder(publicKey, did).
        WithHederaTopics("0.0.111", "0.0.222", "0.0.333"). // stdIn, stdOut, stdErr
        WithReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + publicKey.MustPeerID().String()).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Account Type: %s\n", neuronAccount.AccountType())
    fmt.Printf("DID: %s\n", neuronAccount.DID().String())
    fmt.Printf("StdIn: %s\n", neuronAccount.StdIn().String())
}
```

---

## Module: keylib

The keylib module provides type-safe cryptographic key management for ECDSA secp256k1 keys.

### Core Types

| Type | Description |
|------|-------------|
| `NeuronPrivateKey` | Immutable 32-byte secp256k1 private key |
| `NeuronPublicKey` | Immutable 33-byte compressed public key |
| `EVMAddress` | 20-byte Ethereum address |
| `PeerID` | libp2p peer identifier |
| `Signature` | 65-byte ECDSA signature (R‖S‖V) |
| `EncryptedPrivateKey` | Password-encrypted key with Argon2id + AES-256-GCM |

### Key Operations

```go
// Generation
privateKey, _ := keylib.GeneratePrivateKey()
mnemonic, _ := keylib.GenerateMnemonic(24)  // BIP39

// Parsing
privateKey, _ := keylib.ParsePrivateKeyHex("0x4c0883a69102937d...")
publicKey, _ := keylib.ParsePublicKeyHex("0x02759b048e7ccf6ba...")

// Mnemonic restoration
privateKey, _ := keylib.PrivateKeyFromMnemonic(mnemonic)
privateKey, _ := keylib.PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'/0/1")

// Hedera interoperability
neuronKey, _ := keylib.PrivateKeyFromHedera(hederaPrivateKey)
hederaKey, _ := neuronKey.ToHederaPrivateKeySafe()

// Encryption for storage
encrypted, _ := privateKey.Scramble("password")
restored, _ := keylib.UnscramblePrivateKey(encrypted, "password")
```

### Identity Derivation

```go
publicKey := privateKey.PublicKey()

// Ethereum
evmAddr := publicKey.EVMAddress()
fmt.Println(evmAddr.ChecksumHex())  // EIP-55 checksum

// libp2p
peerID, _ := publicKey.PeerID()
fmt.Println(peerID.String())

// Hedera
hederaPubKey, _ := publicKey.ToHederaPublicKeySafe()
```

### Signing & Verification

```go
// Sign message (auto-hashes with Keccak256)
signature, _ := privateKey.SignMessage(message)

// Sign pre-hashed digest
signature, _ := privateKey.SignDigest(digest)

// Verify
valid := publicKey.Verify(message, signature)

// Recover signer from signature
recoveredPubKey, _ := keylib.RecoverPublicKey(message, signature)
```

### Security Features

- **Constant-time comparisons** - All `Matches*` and `Equal()` methods
- **Secure memory zeroing** - `Zeroize()` clears key material (Go 1.21+ `clear()`)
- **Safe variants** - `EVMAddressSafe()`, `ToHederaPrivateKeySafe()` return errors instead of zero values
- **Ed25519 rejection** - Three-tier detection prevents key type confusion

For detailed API documentation, see [keylib/doc.go](keylib/doc.go).

---

## Module: account

The account module implements blockchain-agnostic agent identity descriptors with hierarchical accounts and extensible communication backends.

### Core Types

| Type | Description |
|------|-------------|
| `NeuronAccount` | Agent identity with communication endpoints |
| `AccountType` | Parent (with DID) or Child (with parent reference) |
| `CommAddress` | Technology-agnostic communication address |
| `ReachableAddrs` | P2P connection endpoints with PeerID validation |
| `NeuronDID` | Interface for decentralized identifiers |

### Account Types

```go
// Parent Account - Root identity with DID
parentAccount, _ := account.NewParentAccountBuilder(publicKey, did).
    WithHederaTopics("0.0.111", "0.0.222", "0.0.333").
    Build()

// Child Account - Derived identity with parent reference
childAccount, _ := account.NewChildAccountBuilder(childPubKey, parentPubKey).
    WithKafkaTopics("child-in", "child-out", "child-err").
    Build()
```

### Communication Backends

The SDK supports multiple messaging technologies through an extensible backend system:

| Backend | Kind | Example Locator |
|---------|------|-----------------|
| Hedera Consensus Service | `hedera-topic` | `0.0.12345` |
| Apache Kafka | `kafka-topic` | `my-events-topic` |
| Custom | `custom` | implementation-defined |

```go
// Using convenience methods
builder.WithStdInHedera("0.0.12345")
builder.WithStdInKafka("events-topic")

// Using generic backend method
builder.WithStdInBackend("hedera-topic", "0.0.12345")
builder.WithStdInBackend("kafka-topic", "my-topic")

// Direct CommAddress
addr, _ := account.NewCommAddress("hedera-topic", "0.0.12345")
builder.WithStdIn(addr)
```

### Adding Custom Backends

```go
// backends/pulsar/pulsar.go
package pulsar

import "github.com/aspect-build/neuron-go-hedera-sdk/account"

const PulsarTopicKind = "pulsar-topic"

func init() {
    account.RegisterBackend(&pulsarBackend{})
}

type pulsarBackend struct{}

func (p *pulsarBackend) Kind() string { return PulsarTopicKind }
func (p *pulsarBackend) Technology() account.TopicTechnology { return account.TopicTechnologyCustom }
func (p *pulsarBackend) ValidateLocator(locator string) error { /* ... */ }
func (p *pulsarBackend) ParseLocator(locator string) (string, error) { /* ... */ }
func (p *pulsarBackend) Metadata() account.BackendMetadata { /* ... */ }
```

### DID:Key Support

```go
import "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"

// Create DID from public key
did, _ := didkey.FromPublicKey(publicKey)
fmt.Println(did.String())  // did:key:zQ3sh...

// Parse existing DID
did, _ := didkey.Parse("did:key:zQ3shVKsZC7kz5fp9beVi83k9jGifqF939bCYiBDJJDnN3Tk9")

// Verify DID matches public key
if did.MatchesKey(publicKey) {
    fmt.Println("DID verified")
}

// Extract public key from DID
extractedKey, _ := did.PublicKey()
```

### Reachable Addresses

All reachable addresses must contain a PeerID matching the account's derived PeerID:

```go
peerID := publicKey.MustPeerID()

builder.WithReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String())
builder.WithReachableAddr("/dns/example.com/tcp/4001/p2p/" + peerID.String())

// Circuit relay (target PeerID is extracted)
builder.WithReachableAddr("/ip4/relay.example.com/tcp/4001/p2p/RELAY_PEER/p2p-circuit/p2p/" + peerID.String())
```

### Validation

```go
// Validate single error
err := neuronAccount.Validate()

// Validate all errors
result := neuronAccount.ValidateAll()
for _, err := range result.Errors {
    fmt.Println(err)
}

// Fluent validator
v := account.NewAccountValidator().
    ValidatePublicKey(pubKey).
    ValidateDID(did).
    ValidateDIDMatchesKey(did, pubKey)

if v.IsValid() {
    fmt.Println("All validations passed")
}
```

### JSON Serialization

```go
// Marshal
jsonBytes, _ := json.Marshal(neuronAccount)

// Unmarshal
var account account.NeuronAccount
json.Unmarshal(jsonBytes, &account)
```

Example JSON output:
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

For detailed specification, see [docs/neuron-account-specification.md](docs/neuron-account-specification.md).

---

## Error Handling

Both modules use structured error types with operation context:

### Keylib Errors

```go
privateKey, err := keylib.ParsePrivateKeyHex(input)
if err != nil {
    if keyErr, ok := err.(*keylib.KeyError); ok {
        switch keyErr.Kind {
        case keylib.ErrKindInvalidHex:
            // Handle invalid hex character
        case keylib.ErrKindInvalidLength:
            // Handle wrong length
        case keylib.ErrKindUnsupportedKeyType:
            // Handle Ed25519 key
        }
    }
}
```

### Account Errors

```go
account, err := builder.Build()
if err != nil {
    if accErr, ok := err.(*account.AccountError); ok {
        switch accErr.Kind {
        case account.ErrKindPeerIDMismatch:
            // Handle PeerID validation failure
        case account.ErrKindMissingRequired:
            // Handle missing required field
        case account.ErrKindInvalidDID:
            // Handle DID validation failure
        }
    }
}
```

---

## Thread Safety

### Keylib

All types are immutable and thread-safe after construction, except:

- **`Zeroize()`** - Mutates internal state; requires external synchronization

CPU-intensive operations (`Scramble`, `UnscramblePrivateKey`) allocate 64MB each; limit concurrency in high-throughput scenarios.

### Account

All types and registry operations are thread-safe. The backend registry uses `sync.RWMutex` for concurrent access.

---

## Documentation

| Document | Description |
|----------|-------------|
| [keylib/doc.go](keylib/doc.go) | Package-level API documentation |
| [docs/neuron-account-specification.md](docs/neuron-account-specification.md) | NeuronAccount technical specification |
| [docs/examples/](docs/examples/) | Usage examples and patterns |

### View Documentation

```bash
# View keylib documentation
go doc -all ./keylib

# View account documentation
go doc -all ./account

# Start local documentation server
godoc -http=:6060
# Then open: http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/
```

---

## Testing

```bash
# Run all tests
go test ./keylib/... ./account/... -v

# Run with race detector
go test ./keylib/... ./account/... -race

# Check coverage
go test ./keylib/... -coverprofile=keylib.out && go tool cover -func=keylib.out
go test ./account/... -coverprofile=account.out && go tool cover -func=account.out

# Generate HTML coverage report
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

---

## Version History

### v1.3.0 (Current)

- **Full Decoupling** - Backend kind constants moved to respective backend files
- **Technology() Method** - Backend interface extended for dynamic technology mapping
- **backend_custom.go** - New custom backend implementation
- **2026 Go Best Practices** - `clear()` builtin, `strings.CutPrefix()`, functional options

### v1.2.0

- **Dead Code Removal** - ~310 lines of unused code removed
- **Backend Interface Pattern** - Extensible plugin system for messaging backends
- **Registry Pattern** - Thread-safe backend registration

### v1.1.0

- **NeuronAccount Module** - Initial account implementation
- **DID:Key Support** - W3C-compliant secp256k1 DID encoding

### v1.0.0

- **Keylib Module** - Initial key management implementation
- **Hedera Interoperability** - Hiero SDK integration

---

## License

[MIT License](LICENSE)

---

## Contributing

Contributions are welcome. Please ensure:

1. All tests pass (`go test ./... -race`)
2. Coverage is maintained (>85%)
3. Code follows existing patterns and style
4. Documentation is updated for API changes
