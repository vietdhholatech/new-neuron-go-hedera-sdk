# Neuron SDK: On-Chain Interaction and Peer Discovery Architecture

**Version:** 1.1
**Status:** Technical Specification
**Last Updated:** 2026-01-23
**Go Version:** 1.23+ recommended

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Architecture Overview](#2-architecture-overview)
3. [Package Structure](#3-package-structure)
4. [Transport Layer Interfaces](#4-transport-layer-interfaces)
5. [Hedera Provider Implementation](#5-hedera-provider-implementation)
6. [Presence and Heartbeat Services](#6-presence-and-heartbeat-services)
7. [Peer Discovery Service](#7-peer-discovery-service)
8. [Data Flow Diagrams](#8-data-flow-diagrams)
9. [Error Handling Strategy](#9-error-handling-strategy)
10. [Integration Points](#10-integration-points)
11. [Configuration Management](#11-configuration-management)
12. [Comparison with Legacy SDK](#12-comparison-with-legacy-sdk)
13. [Implementation Roadmap](#13-implementation-roadmap)

---

## 1. Executive Summary

This document defines the technical architecture for Hedera on-chain interaction and peer discovery in the Neuron Go SDK. The design introduces a layered transport abstraction that builds upon the existing `keylib` (cryptographic primitives) and `account` (identity descriptors) packages.

### 1.1 Design Principles

| Principle                | Description                                                                         |
| ------------------------ | ----------------------------------------------------------------------------------- |
| **Type Safety**          | Use `keylib.EVMAddress`, `keylib.PeerID` throughout; no raw strings for identifiers |
| **Testability**          | Interface-driven design allows mocking at every layer                               |
| **Proper Lifecycle**     | Single long-lived client instead of per-operation creation                          |
| **Configuration-Driven** | Explicit config structs; no `os.Getenv()` calls inside SDK                          |
| **Clean Separation**     | Transport, discovery, and presence are independent services                         |
| **Structured Errors**    | Consistent error types following `keylib.KeyError` pattern                          |
| **Structured Logging**   | Use `log/slog` (Go 1.21+) for structured, contextual logging                        |
| **Go 1.23 Iterators**    | Use `iter.Seq`/`iter.Seq2` for memory-efficient streaming                           |
| **Functional Options**   | Extensible constructors via `...Option` pattern                                     |
| **Graceful Shutdown**    | Context-aware shutdown with drain periods                                           |

### 1.2 Go Version Requirements

| Feature                 | Minimum Go Version | Usage                         |
| ----------------------- | ------------------ | ----------------------------- |
| `log/slog`              | Go 1.21            | Structured logging throughout |
| `iter.Seq`, `iter.Seq2` | Go 1.23            | Memory-efficient iterators    |
| `errors.Join`           | Go 1.20            | Multi-error aggregation       |
| Range-over-func         | Go 1.23            | Iterator consumption          |

**Recommended: Go 1.23+** for full feature support.

### 1.3 Scope

**In Scope:**

- Hedera Consensus Service (HCS) topic publishing and subscription
- Smart contract interaction for peer registry (read and write)
- Heartbeat publishing with cryptographic signatures
- Presence tracking with signature verification
- Peer discovery with LRU caching and circuit breaker
- Graceful shutdown with operation draining
- Structured error handling with retry semantics

**Explicitly Out of Scope:**

| Item                      | Rationale                                                    |
| ------------------------- | ------------------------------------------------------------ |
| P2P networking (libp2p)   | Handled by separate p2p package; transport is msg-layer only |
| Payment/escrow mechanisms | Business logic, not transport concern                        |
| Message encryption        | Application-layer concern; transport handles signed bytes    |
| Hedera account creation   | Accounts assumed pre-existing; use Hedera Portal/SDK         |
| Scheduled transactions    | Not needed for current messaging use cases                   |
| Token operations (HTS)    | Focus is HCS messaging, not token transfers                  |
| File service operations   | Not needed for messaging architecture                        |
| Multi-signature workflows | Single operator key assumed per client                       |

---

## 2. Architecture Overview

### 2.1 Layer Diagram

```
+===========================================================================+
|                           APPLICATION LAYER                                |
|  +---------------------------------------------------------------------+  |
|  |                         User Application                             |  |
|  |    Uses NeuronAccount, HeartbeatService, PeerDiscoveryService       |  |
|  +---------------------------------------------------------------------+  |
+===========================================================================+
                                    |
                                    v
+===========================================================================+
|                            SERVICE LAYER                                   |
|  +-------------------+  +-------------------+  +-----------------------+  |
|  | HeartbeatService  |  | PresenceService   |  | PeerDiscoveryService  |  |
|  |                   |  |                   |  |                       |  |
|  | - Periodic beat   |  | - Track liveness  |  | - Cache peer info     |  |
|  | - Configurable    |  | - Active peers    |  | - Registry lookup     |  |
|  +-------------------+  +-------------------+  +-----------------------+  |
+===========================================================================+
                                    |
                                    v
+===========================================================================+
|                           TRANSPORT LAYER                                  |
|  +---------------------------------------------------------------------+  |
|  |                        INTERFACES                                    |  |
|  |  TopicPublisher | TopicSubscriber | RegistryReader | RegistryWriter |  |
|  +---------------------------------------------------------------------+  |
|                                    |                                       |
|  +---------------------------------------------------------------------+  |
|  |                   HEDERA IMPLEMENTATION                              |  |
|  |  +---------------+  +----------------+  +------------------------+  |  |
|  |  | HederaClient  |  | HederaPublisher|  | HederaSubscriber       |  |  |
|  |  | (lifecycle)   |  | (HCS publish)  |  | (HCS subscribe)        |  |  |
|  |  +---------------+  +----------------+  +------------------------+  |  |
|  |  +------------------------+                                         |  |
|  |  | HederaRegistry         |                                         |  |
|  |  | (contract reads)       |                                         |  |
|  |  +------------------------+                                         |  |
|  +---------------------------------------------------------------------+  |
+===========================================================================+
                                    |
                                    v
+===========================================================================+
|                          FOUNDATION LAYER                                  |
|  +--------------------------------+  +--------------------------------+   |
|  |            keylib/             |  |            account/            |   |
|  |                                |  |                                |   |
|  | - NeuronPrivateKey             |  | - NeuronAccount                |   |
|  | - NeuronPublicKey              |  | - CommAddress                  |   |
|  | - EVMAddress                   |  | - TopicMessage                 |   |
|  | - PeerID                       |  | - Backend interface            |   |
|  | - ToHederaPrivateKeySafe()     |  | - HederaTopicBackend           |   |
|  +--------------------------------+  +--------------------------------+   |
+===========================================================================+
                                    |
                                    v
+===========================================================================+
|                          EXTERNAL SYSTEMS                                  |
|  +-------------------+  +-------------------+  +-----------------------+  |
|  | Hedera Network    |  | Mirror Node       |  | JSON-RPC Endpoint     |  |
|  | (HCS topics)      |  | (subscriptions)   |  | (contract reads)      |  |
|  +-------------------+  +-------------------+  +-----------------------+  |
+===========================================================================+
```

### 2.2 Component Relationships

```
                    +------------------+
                    |  HederaConfig    |
                    +--------+---------+
                             |
                             v
                    +------------------+
                    |  HederaClient    |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
+----------------+  +----------------+  +----------------+
|HederaPublisher |  |HederaSubscriber|  |HederaRegistry  |
+-------+--------+  +-------+--------+  +-------+--------+
        |                   |                   |
        v                   v                   v
+----------------+  +----------------+  +----------------+
|TopicPublisher  |  |TopicSubscriber |  |RegistryReader  |
|(interface)     |  |(interface)     |  |(interface)     |
+----------------+  +----------------+  +----------------+
        |                   |                   |
        v                   v                   v
+----------------+  +----------------+  +----------------+
|HeartbeatService|  |PresenceService |  |PeerDiscovery   |
|                |  |                |  |Service         |
+----------------+  +----------------+  +----------------+
```

---

## 3. Package Structure

```
neuron-go-hedera-sdk/
|
|-- keylib/                          [EXISTS] Cryptographic primitives
|   |-- private_key.go               NeuronPrivateKey, ToHederaPrivateKeySafe()
|   |-- public_key.go                NeuronPublicKey, ToHederaPublicKeySafe()
|   |-- evm_address.go               EVMAddress type
|   |-- peer_id.go                   PeerID type
|
|-- account/                         [EXISTS] Identity descriptors
|   |-- account.go                   NeuronAccount
|   |-- comm_address.go              CommAddress, CommAddressKind
|   |-- message.go                   TopicMessage, TopicMessageBuilder
|   |-- backend.go                   Backend interface
|   |-- backend_hedera.go            HederaTopicBackend
|
|-- transport/                       [NEW] Transport layer abstractions
|   |-- doc.go                       Package documentation
|   |-- interfaces.go                TopicPublisher, TopicSubscriber, RegistryReader
|   |-- types.go                     PeerInfo, Message envelope
|   |-- errors.go                    TransportError, TransportErrorKind
|
|-- hedera/                          [NEW] Hedera-specific implementation
|   |-- doc.go                       Package documentation
|   |-- client.go                    HederaClient with lifecycle management
|   |-- config.go                    HederaConfig struct with validation
|   |-- publisher.go                 HederaPublisher (TopicPublisher impl)
|   |-- subscriber.go                HederaSubscriber (TopicSubscriber impl)
|   |-- registry.go                  HederaRegistry (RegistryReader impl)
|   |-- topic_id.go                  Type-safe TopicID wrapper
|   |-- errors.go                    Hedera-specific error types
|
|-- presence/                        [NEW] Heartbeat and presence tracking
|   |-- doc.go                       Package documentation
|   |-- heartbeat.go                 HeartbeatService
|   |-- presence.go                  PresenceService
|   |-- types.go                     HeartbeatPayload, PresenceInfo, GeoLocation
|   |-- config.go                    HeartbeatConfig, PresenceConfig
|
|-- discovery/                       [NEW] Peer discovery with caching
|   |-- doc.go                       Package documentation
|   |-- service.go                   PeerDiscoveryService
|   |-- cache.go                     LRU cache implementation
|   |-- types.go                     Extended PeerInfo with cache metadata
|   |-- config.go                    DiscoveryConfig
```

---

## 4. Transport Layer Interfaces

**Required Imports:**

```go
import (
    "context"
    "iter"      // Go 1.23+
    "log/slog"  // Go 1.21+
    "time"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)
```

### 4.1 TopicPublisher Interface

```go
// TopicPublisher publishes messages to topics.
// Implementations handle backend-specific details such as
// transaction creation, signing, and submission.
type TopicPublisher interface {
    // Publish sends a message to the specified topic.
    // The context controls cancellation and timeouts.
    //
    // Returns:
    //   - ErrKindClosed: Client has been shut down
    //   - ErrKindInvalidTopic: Topic address is malformed
    //   - ErrKindPayloadTooLarge: Payload exceeds HCS limit (currently 1024 bytes)
    //   - ErrKindConnection: Network error (retried internally)
    //   - ErrKindTimeout: Operation timed out after retries
    //   - ErrKindRateLimit: Rate limit exceeded (retried internally)
    Publish(ctx context.Context, topic account.CommAddress, payload []byte) error

    // PublishSigned sends a message with an attached signature.
    // The signature is prepended to the payload (65 bytes + payload).
    // Receivers can verify the sender using signature recovery via ExtractSignedMessage().
    //
    // Wire format: [Signature (65 bytes)][Payload (variable)]
    //   - Signature: R (32 bytes) || S (32 bytes) || V (1 byte, recovery ID)
    //   - Payload: Original payload bytes
    //
    // Returns same errors as Publish(), plus:
    //   - ErrKindSignature: Failed to sign payload (invalid key)
    PublishSigned(
        ctx context.Context,
        topic account.CommAddress,
        payload []byte,
        signer keylib.NeuronPrivateKey,
    ) error

    // Close releases resources associated with this publisher.
    Close() error
}
```

### 4.1.1 Signed Message Verification

When receiving messages sent via `PublishSigned()`, use these types and functions to verify the sender:

```go
// SignedMessage represents a message with its cryptographic signature.
// Use ExtractSignedMessage() to parse a signed message from raw bytes.
type SignedMessage struct {
    // Signature is the 65-byte ECDSA signature (R || S || V format).
    Signature keylib.Signature

    // Payload is the original message content (after signature).
    Payload []byte
}

// ExtractSignedMessage parses a signed message from raw bytes.
// The input must be at least 65 bytes (signature) + 1 byte (minimum payload).
//
// Returns:
//   - ErrKindInvalidFormat: Message too short (< 65 bytes)
//   - ErrKindSignature: Signature bytes are malformed
func ExtractSignedMessage(data []byte) (SignedMessage, error)

// RecoverSender derives the sender's EVM address from the signed message.
// Uses ECDSA public key recovery from the signature and payload hash.
//
// Returns:
//   - ErrKindSignature: Recovery failed (corrupted signature or payload)
func (m SignedMessage) RecoverSender() (keylib.EVMAddress, error)

// VerifySender checks if the message was signed by the expected address.
// This is a convenience method equivalent to:
//   recovered, err := m.RecoverSender()
//   return err == nil && recovered.Equal(expected)
//
// Returns false if recovery fails or addresses don't match.
func (m SignedMessage) VerifySender(expected keylib.EVMAddress) bool

// RecoverPublicKey derives the sender's full public key from the signed message.
// Use this when you need the public key for additional operations (e.g., PeerID derivation).
//
// Returns:
//   - ErrKindSignature: Recovery failed
func (m SignedMessage) RecoverPublicKey() (keylib.NeuronPublicKey, error)
```

**Usage Example:**

```go
// In a message handler, verify the sender before processing
handler := func(ctx context.Context, msg account.TopicMessage) error {
    // Extract and verify signed message
    signed, err := transport.ExtractSignedMessage(msg.Payload)
    if err != nil {
        return fmt.Errorf("unsigned or malformed message: %w", err)
    }

    // Verify sender matches expected peer
    if !signed.VerifySender(expectedPeerAddr) {
        return fmt.Errorf("message not from expected sender")
    }

    // Process the verified payload
    actualPayload := signed.Payload
    // ... handle actualPayload ...
    return nil
}
```

### 4.2 TopicSubscriber Interface

```go
// MessageHandler processes incoming messages.
//
// Context Semantics:
//   - The context is derived from the subscription context with an optional per-message timeout.
//   - When the subscription context is canceled, the handler context is also canceled.
//   - Handlers SHOULD check ctx.Done() for long-running operations.
//   - If a handler blocks indefinitely ignoring the context, the subscription continues
//     but logs a warning. The blocked handler goroutine may leak.
//
// Return Values:
//   - Return nil to acknowledge successful processing.
//   - Return error to indicate processing failure. The error is logged at WARN level
//     but NOT propagated - the subscription continues processing subsequent messages.
//   - Messages are NOT redelivered on handler errors (at-most-once delivery).
//
// Panics:
//   - If the handler panics, the panic is recovered, logged at ERROR level,
//     and the subscription continues. The panic does NOT crash the application.
type MessageHandler func(ctx context.Context, msg account.TopicMessage) error

// Subscription represents an active topic subscription.
//
// Thread Safety:
//   - All methods are safe to call from any goroutine.
//   - Unsubscribe() can be called from within the handler itself.
//   - Done() channel can be selected on from any goroutine.
//   - Err() should only be called after Done() is closed; calling before
//     may return nil even if an error will occur.
type Subscription interface {
    // Unsubscribe stops the subscription gracefully.
    // Waits for any in-flight handler to complete (with timeout).
    // Safe to call multiple times - subsequent calls are no-ops.
    // Safe to call from within the message handler.
    Unsubscribe()

    // Done returns a channel that closes when the subscription ends.
    // The channel closes when:
    //   - Unsubscribe() is called
    //   - The subscription context is canceled
    //   - A fatal error occurs (check Err() for details)
    // Use this to detect subscription termination in select statements.
    Done() <-chan struct{}

    // Err returns the error that caused the subscription to end, if any.
    // Returns nil if:
    //   - Subscription is still active (Done() not yet closed)
    //   - Unsubscribe() was called explicitly
    //   - The subscription context was canceled (returns nil, not ctx.Err())
    // Returns non-nil error for fatal failures (connection loss, etc.).
    //
    // IMPORTANT: Only call after Done() is closed to get accurate results.
    Err() error
}

// TopicSubscriber receives messages from topics.
type TopicSubscriber interface {
    // Subscribe begins listening to the topic from the current time.
    // Messages are delivered to the handler asynchronously.
    // Returns a Subscription for lifecycle management.
    Subscribe(
        ctx context.Context,
        topic account.CommAddress,
        handler MessageHandler,
    ) (Subscription, error)

    // SubscribeFrom begins listening from a specific timestamp.
    // Use this to resume from a known position or replay history.
    SubscribeFrom(
        ctx context.Context,
        topic account.CommAddress,
        from time.Time,
        handler MessageHandler,
    ) (Subscription, error)

    // Close releases resources associated with this subscriber.
    Close() error
}
```

### 4.3 RegistryReader Interface

```go
// PeerInfo contains peer discovery information from the registry.
// All fields use type-safe keylib types.
type PeerInfo struct {
    Available   bool                  // Whether peer is currently available
    PeerID      keylib.PeerID         // Libp2p peer identifier
    EVMAddress  keylib.EVMAddress     // Ethereum-compatible address
    StdOut      account.CommAddress   // Peer's output topic (heartbeats)
    StdIn       account.CommAddress   // Peer's input topic (requests)
    StdErr      account.CommAddress   // Peer's error topic (diagnostics)
}

// RegistryReader reads peer information from an on-chain registry.
// Implementations interact with the rendezvous smart contract.
type RegistryReader interface {
    // GetPeerInfo retrieves info for a specific peer by EVM address.
    // Returns ErrPeerNotFound if the address is not registered.
    GetPeerInfo(ctx context.Context, addr keylib.EVMAddress) (PeerInfo, error)

    // GetPeerByIndex retrieves peer address at a given index.
    // Use with GetPeerCount() to iterate all peers.
    GetPeerByIndex(ctx context.Context, index uint64) (keylib.EVMAddress, error)

    // GetPeerCount returns the total number of registered peers.
    GetPeerCount(ctx context.Context) (uint64, error)

    // ListAvailablePeers returns addresses of all available peers.
    // Filters out unavailable peers and internal Hedera addresses.
    ListAvailablePeers(ctx context.Context) ([]keylib.EVMAddress, error)

    // IterAvailablePeers returns a Go 1.23 iterator over available peers.
    // Streams results without loading all into memory.
    // Use with range-over-func: for addr, err := range registry.IterAvailablePeers(ctx) { ... }
    // Stops iteration early if context is canceled.
    IterAvailablePeers(ctx context.Context) iter.Seq2[keylib.EVMAddress, error]
}

// RegistryWriter writes peer information to an on-chain registry.
// Use HederaClient.RegistryWriter() to obtain an implementation.
//
// All write operations require the operator key configured in HederaConfig.
// The operator's EVM address (derived from the key) is used as the peer identifier.
type RegistryWriter interface {
    // RegisterSelf registers or updates the caller's peer info in the registry.
    // The caller is identified by the operator key's EVM address.
    //
    // This calls the smart contract method:
    //   PutPeerAvailableSelf(stdOut, stdIn, stdErr, peerID, serviceIDs, prices)
    //
    // The peer is marked as available upon successful registration.
    //
    // Returns:
    //   - ErrKindConfiguration: ServiceIDs and Prices have different lengths
    //   - ErrKindConfiguration: Required fields (StdOut, StdIn, StdErr, PeerID) are zero
    //   - ErrKindConnection: Network error during transaction
    //   - ErrKindTimeout: Transaction timed out
    //   - ErrKindRegistry: Contract rejected the registration
    RegisterSelf(ctx context.Context, info SelfRegistrationInfo) error

    // SetAvailability updates the caller's availability status.
    // Use this to mark a peer as temporarily unavailable without de-registering.
    //
    // The caller is identified by the operator key's EVM address.
    //
    // Returns:
    //   - ErrKindNotFound: Peer is not registered (call RegisterSelf first)
    //   - ErrKindConnection: Network error
    //   - ErrKindRegistry: Contract rejected the update
    SetAvailability(ctx context.Context, available bool) error
}

// SelfRegistrationInfo contains data for peer self-registration.
// All topic fields must be Hedera HCS topics (CommAddressKind = HederaTopicKind).
type SelfRegistrationInfo struct {
    // StdOut is the topic where this peer publishes heartbeats and announcements.
    // Required. Must be a valid Hedera topic address.
    StdOut account.CommAddress

    // StdIn is the topic where this peer receives incoming requests.
    // Required. Must be a valid Hedera topic address.
    StdIn account.CommAddress

    // StdErr is the topic where this peer publishes error/diagnostic messages.
    // Required. Must be a valid Hedera topic address.
    StdErr account.CommAddress

    // PeerID is the libp2p peer identifier for P2P connections.
    // Required. Must not be zero-value.
    PeerID keylib.PeerID

    // ServiceIDs lists the services this peer offers.
    // Optional. If provided, must have same length as Prices.
    // Service ID meanings are application-defined.
    ServiceIDs []uint8

    // Prices lists the price for each service in ServiceIDs.
    // Optional. If provided, must have same length as ServiceIDs.
    // Price units are application-defined.
    Prices []uint8
}

// Validate checks that all required fields are set and arrays match.
// Called automatically by RegisterSelf; can be called manually for early validation.
func (s SelfRegistrationInfo) Validate() error
```

---

## 5. Hedera Provider Implementation

### 5.1 HederaConfig

```go
// NetworkName identifies the Hedera network.
type NetworkName string

const (
    NetworkMainnet    NetworkName = "mainnet"
    NetworkTestnet    NetworkName = "testnet"
    NetworkPreviewnet NetworkName = "previewnet"
)

// HederaConfig contains all configuration for Hedera connections.
// All values must be explicitly provided; no environment variable lookups.
type HederaConfig struct {
    // Network identifies which Hedera network to connect to.
    // Must be one of: "mainnet", "testnet", "previewnet"
    // Required.
    Network NetworkName

    // OperatorAccountID is the Hedera account ID in "shard.realm.account" format.
    // Must match regex: ^\d+\.\d+\.\d+$
    // Example: "0.0.12345"
    // Required.
    OperatorAccountID string

    // OperatorKey is the private key for signing transactions.
    // Must be ECDSA secp256k1 (Ed25519 is NOT supported).
    // The corresponding public key must be authorized for the OperatorAccountID.
    // Required.
    OperatorKey keylib.NeuronPrivateKey

    // MirrorNodeURL is the mirror node gRPC endpoint for topic subscriptions.
    // If empty, uses the default for the selected network (see defaults below).
    // Must be a valid URL with host:port format.
    // Optional.
    MirrorNodeURL string

    // EthRPCURL is the JSON-RPC endpoint for smart contract reads.
    // Must be a valid URL starting with "http://" or "https://".
    // Example: "https://testnet.hashio.io/api"
    // Required.
    EthRPCURL string

    // RegistryContractAddress is the peer registry smart contract address.
    // Must be a valid, deployed EVM contract address (not zero).
    // Required.
    RegistryContractAddress keylib.EVMAddress

    // SubscriptionTimeout is the maximum time without messages before reconnecting.
    // Set to 0 to disable timeout-based reconnection (not recommended).
    // For quiet topics, consider increasing to 10+ minutes.
    // Default: 3 minutes.
    SubscriptionTimeout time.Duration

    // MaxRetries is the maximum number of retry attempts for operations.
    // Set to 0 to disable retries (not recommended for production).
    // Default: 5.
    MaxRetries int

    // BaseRetryDelay is the initial delay for exponential backoff.
    // Actual delay: min(BaseRetryDelay * 2^attempt, MaxRetryDelay)
    // Default: 1 second.
    BaseRetryDelay time.Duration

    // MaxRetryDelay caps the maximum delay between retries.
    // Prevents exponential backoff from growing unbounded.
    // Default: 30 seconds.
    MaxRetryDelay time.Duration
}

// Default mirror node gRPC endpoints by network.
// These are the official Hedera-operated mirror nodes.
const (
    DefaultMirrorNodeMainnet    = "mainnet-public.mirrornode.hedera.com:443"
    DefaultMirrorNodeTestnet    = "testnet.mirrornode.hedera.com:443"
    DefaultMirrorNodePreviewnet = "previewnet.mirrornode.hedera.com:443"
)

// Validate checks all required fields and returns error if invalid.
//
// Validation rules:
//   - Network: Must be "mainnet", "testnet", or "previewnet"
//   - OperatorAccountID: Must match pattern "^\d+\.\d+\.\d+$"
//   - OperatorKey: Must not be zero-value; must be ECDSA secp256k1
//   - EthRPCURL: Must be valid URL starting with "http://" or "https://"
//   - RegistryContractAddress: Must not be zero-value
//   - SubscriptionTimeout: If > 0, must be >= 30 seconds
//   - MaxRetries: If set, must be >= 0
//   - BaseRetryDelay: If set, must be >= 100ms
//   - MaxRetryDelay: If set, must be >= BaseRetryDelay
//
// Returns *TransportError with Kind=ErrKindConfiguration on failure.
func (c *HederaConfig) Validate() error {
    const op = "HederaConfig.Validate"

    // Network validation
    switch c.Network {
    case NetworkMainnet, NetworkTestnet, NetworkPreviewnet:
        // OK
    default:
        return NewTransportError(op, ErrKindConfiguration,
            fmt.Sprintf("invalid network %q, must be mainnet/testnet/previewnet", c.Network), nil)
    }

    // Account ID validation
    accountIDPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
    if !accountIDPattern.MatchString(c.OperatorAccountID) {
        return NewTransportError(op, ErrKindConfiguration,
            fmt.Sprintf("invalid account ID %q, must match X.Y.Z format", c.OperatorAccountID), nil)
    }

    // Operator key validation
    if c.OperatorKey.IsZero() {
        return NewTransportError(op, ErrKindConfiguration, "operator key is required", nil)
    }

    // EthRPCURL validation
    if !strings.HasPrefix(c.EthRPCURL, "http://") && !strings.HasPrefix(c.EthRPCURL, "https://") {
        return NewTransportError(op, ErrKindConfiguration,
            fmt.Sprintf("invalid EthRPCURL %q, must start with http:// or https://", c.EthRPCURL), nil)
    }

    // Registry contract validation
    if c.RegistryContractAddress.IsZero() {
        return NewTransportError(op, ErrKindConfiguration, "registry contract address is required", nil)
    }

    // Optional field validation
    if c.SubscriptionTimeout > 0 && c.SubscriptionTimeout < 30*time.Second {
        return NewTransportError(op, ErrKindConfiguration,
            "subscription timeout must be >= 30 seconds or 0 (disabled)", nil)
    }
    if c.BaseRetryDelay > 0 && c.BaseRetryDelay < 100*time.Millisecond {
        return NewTransportError(op, ErrKindConfiguration,
            "base retry delay must be >= 100ms", nil)
    }

    return nil
}

// WithDefaults returns a copy with default values applied to unset fields.
func (c HederaConfig) WithDefaults() HederaConfig {
    if c.MirrorNodeURL == "" {
        switch c.Network {
        case NetworkMainnet:
            c.MirrorNodeURL = DefaultMirrorNodeMainnet
        case NetworkTestnet:
            c.MirrorNodeURL = DefaultMirrorNodeTestnet
        case NetworkPreviewnet:
            c.MirrorNodeURL = DefaultMirrorNodePreviewnet
        }
    }
    if c.SubscriptionTimeout == 0 {
        c.SubscriptionTimeout = 3 * time.Minute
    }
    if c.MaxRetries == 0 {
        c.MaxRetries = 5
    }
    if c.BaseRetryDelay == 0 {
        c.BaseRetryDelay = 1 * time.Second
    }
    if c.MaxRetryDelay == 0 {
        c.MaxRetryDelay = 30 * time.Second
    }
    return c
}
```

### 5.2 HederaClient

```go
// HederaClient provides a long-lived, managed connection to Hedera.
// It owns both the Hedera SDK client and the Ethereum RPC client,
// managing their lifecycle together.
//
// Thread-safe for concurrent use.
type HederaClient struct {
    config     HederaConfig
    hclient    *hiero.Client       // Hedera SDK client for HCS operations
    ethClient  *ethclient.Client   // Ethereum client for contract reads
    logger     *slog.Logger        // Structured logger (Go 1.21+)
    mu         sync.RWMutex
    closed     bool
    closeCh    chan struct{}
    wg         sync.WaitGroup      // Tracks in-flight operations
}

// Option configures a HederaClient.
// Use functional options for extensible configuration.
type Option func(*HederaClient)

// WithLogger sets a custom structured logger.
// If not provided, uses slog.Default().
func WithLogger(logger *slog.Logger) Option {
    return func(c *HederaClient) {
        c.logger = logger
    }
}

// WithMirrorNodeURL overrides the default mirror node URL.
func WithMirrorNodeURL(url string) Option {
    return func(c *HederaClient) {
        c.config.MirrorNodeURL = url
    }
}

// NewHederaClient creates a new managed Hedera client.
//
// The client:
//   - Validates configuration
//   - Creates Hedera SDK client with operator credentials
//   - Connects to JSON-RPC endpoint for contract reads
//   - Maintains connections until Close() is called
//
// Uses functional options for extensibility:
//   client, err := hedera.NewHederaClient(config,
//       hedera.WithLogger(myLogger),
//       hedera.WithMirrorNodeURL("custom-url"),
//   )
//
// Returns error if configuration is invalid or connection fails.
func NewHederaClient(config HederaConfig, opts ...Option) (*HederaClient, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }
    config = config.WithDefaults()

    c := &HederaClient{
        config:  config,
        logger:  slog.Default(),
        closeCh: make(chan struct{}),
    }

    // Apply options
    for _, opt := range opts {
        opt(c)
    }

    // Initialize connections...
    return c, nil
}

// Publisher returns a TopicPublisher backed by this client.
// The returned publisher shares the client's connection.
//
// Returns an interface type for testability. Use type assertion if you need
// Hedera-specific methods: publisher.(*HederaPublisher).SomeHederaMethod()
func (c *HederaClient) Publisher() transport.TopicPublisher

// Subscriber returns a TopicSubscriber backed by this client.
// The returned subscriber shares the client's connection.
//
// Returns an interface type for testability. Use type assertion if you need
// Hedera-specific methods: subscriber.(*HederaSubscriber).SomeHederaMethod()
func (c *HederaClient) Subscriber() transport.TopicSubscriber

// Registry returns a RegistryReader backed by this client.
// The returned registry uses the ethClient for contract calls.
//
// Returns an interface type for testability. Use type assertion if you need
// Hedera-specific methods: registry.(*HederaRegistry).SomeHederaMethod()
func (c *HederaClient) Registry() transport.RegistryReader

// RegistryWriter returns a RegistryWriter for self-registration.
// Returns nil if OperatorKey is not configured in HederaConfig.
//
// Usage:
//
//   writer := client.RegistryWriter()
//   if writer == nil {
//       return errors.New("operator key required for registration")
//   }
//   err := writer.RegisterSelf(ctx, info)
func (c *HederaClient) RegistryWriter() transport.RegistryWriter

// Shutdown gracefully shuts down the client.
// Waits for in-flight operations to complete (up to ctx deadline).
// This is the preferred shutdown method for production use.
//
// Implementation:
//   1. Signal shutdown (close closeCh)
//   2. Wait for in-flight ops with timeout
//   3. Close Hedera SDK client
//   4. Close Ethereum RPC client
//   5. Return any errors
func (c *HederaClient) Shutdown(ctx context.Context) error {
    c.mu.Lock()
    if c.closed {
        c.mu.Unlock()
        return nil
    }
    c.closed = true
    close(c.closeCh)
    c.mu.Unlock()

    // Wait for in-flight operations with timeout
    done := make(chan struct{})
    go func() {
        c.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        c.logger.Info("all in-flight operations completed")
    case <-ctx.Done():
        c.logger.Warn("shutdown deadline exceeded, forcing close",
            slog.Int("pending_ops", c.pendingOps()))
    }

    // Close underlying clients
    var errs []error
    if err := c.hclient.Close(); err != nil {
        errs = append(errs, fmt.Errorf("hedera client: %w", err))
    }
    c.ethClient.Close()

    return errors.Join(errs...)
}

// Close releases all resources immediately.
// For graceful shutdown, prefer Shutdown(ctx).
// Safe to call multiple times.
func (c *HederaClient) Close() error {
    return c.Shutdown(context.Background())
}

// IsClosed returns true if the client has been closed or is shutting down.
// Thread-safe.
func (c *HederaClient) IsClosed() bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.closed
}

// pendingOps returns the current count of in-flight operations.
// Used during shutdown to monitor graceful drain progress.
//
// Implementation note: The client tracks operations using an atomic counter
// incremented at operation start and decremented at completion. This is
// separate from the WaitGroup which only tracks goroutine completion.
func (c *HederaClient) pendingOps() int {
    return int(atomic.LoadInt64(&c.opCount))
}

// trackOp increments the operation counter and adds to the WaitGroup.
// Call doneOp() when the operation completes.
// Returns false if the client is closed (operation should not proceed).
func (c *HederaClient) trackOp() bool {
    c.mu.RLock()
    if c.closed {
        c.mu.RUnlock()
        return false
    }
    c.mu.RUnlock()

    atomic.AddInt64(&c.opCount, 1)
    c.wg.Add(1)
    return true
}

// doneOp decrements the operation counter and signals WaitGroup.
// Must be called exactly once for each successful trackOp() call.
func (c *HederaClient) doneOp() {
    atomic.AddInt64(&c.opCount, -1)
    c.wg.Done()
}
```

**Structured Logging Example:**

```go
// Logging with context in operations
func (p *HederaPublisher) Publish(ctx context.Context, topic account.CommAddress, payload []byte) error {
    logger := p.client.logger.With(
        slog.String("op", "publish"),
        slog.String("topic", topic.String()),
        slog.Int("payload_size", len(payload)),
    )

    logger.DebugContext(ctx, "publishing message")

    // ... perform publish ...

    if err != nil {
        logger.ErrorContext(ctx, "publish failed",
            slog.String("error", err.Error()),
            slog.Int("attempt", attempt),
        )
        return err
    }

    logger.InfoContext(ctx, "message published successfully",
        slog.Duration("latency", time.Since(start)),
    )
    return nil
}
```

### 5.3 TopicID Type

```go
// TopicID is a type-safe wrapper around Hedera topic identifiers.
// It provides parsing, validation, and conversion utilities.
type TopicID struct {
    Shard uint64
    Realm uint64
    Topic uint64
}

// ParseTopicID parses a topic ID from "shard.realm.topic" format.
// Returns error if format is invalid.
func ParseTopicID(s string) (TopicID, error)

// TopicIDFromCommAddress extracts a TopicID from a CommAddress.
// Returns error if the address is not a Hedera topic.
func TopicIDFromCommAddress(addr account.CommAddress) (TopicID, error)

// TopicIDFromUint64 creates a TopicID from just the topic number.
// Assumes shard=0, realm=0 (standard for mainnet/testnet).
func TopicIDFromUint64(topic uint64) TopicID

// String returns the canonical "shard.realm.topic" representation.
func (t TopicID) String() string

// ToHedera converts to the Hedera SDK TopicID type.
func (t TopicID) ToHedera() hiero.TopicID

// ToCommAddress converts to an account.CommAddress.
func (t TopicID) ToCommAddress() (account.CommAddress, error)

// IsZero returns true if this is a zero-value TopicID (0.0.0).
func (t TopicID) IsZero() bool
```

### 5.4 HederaPublisher Implementation

```go
// HederaPublisher implements transport.TopicPublisher for Hedera HCS.
type HederaPublisher struct {
    client *HederaClient
}

// Publish sends a message to a Hedera topic.
//
// Implementation:
//   1. Validates client is not closed
//   2. Converts CommAddress to TopicID
//   3. Creates TopicMessageSubmitTransaction
//   4. Executes with retry on transient failures
//   5. Returns on success or max retries exceeded
func (p *HederaPublisher) Publish(
    ctx context.Context,
    topic account.CommAddress,
    payload []byte,
) error

// PublishSigned sends a signed message.
//
// Implementation:
//   1. Signs payload with Keccak256 + ECDSA
//   2. Prepends 65-byte signature to payload
//   3. Calls Publish with combined data
func (p *HederaPublisher) PublishSigned(
    ctx context.Context,
    topic account.CommAddress,
    payload []byte,
    signer keylib.NeuronPrivateKey,
) error
```

**Retry Logic:**

```
Attempt 1: Execute immediately
Attempt 2: Wait BaseRetryDelay (1s default)
Attempt 3: Wait BaseRetryDelay * 2 (2s)
Attempt 4: Wait BaseRetryDelay * 4 (4s)
Attempt 5: Wait BaseRetryDelay * 8 (8s)
-> Return error after MaxRetries exceeded
```

### 5.5 HederaSubscriber Implementation

```go
// HederaSubscriber implements transport.TopicSubscriber for Hedera HCS.
type HederaSubscriber struct {
    client *HederaClient
}

// Subscribe begins listening from the current time.
func (s *HederaSubscriber) Subscribe(
    ctx context.Context,
    topic account.CommAddress,
    handler MessageHandler,
) (Subscription, error)

// SubscribeFrom begins listening from a specific timestamp.
//
// Implementation:
//   1. Creates hederaSubscription with configuration
//   2. Starts background goroutine for subscription loop
//   3. Returns immediately with Subscription handle
//
// The subscription loop:
//   1. Creates TopicMessageQuery
//   2. Subscribes to mirror node
//   3. Delivers messages to handler with context (for cancellation)
//   4. Tracks last timestamp for reconnection
//   5. Reconnects on SubscriptionTimeout (default 3 min)
//
// The handler receives context for each message:
//   handler := func(ctx context.Context, msg account.TopicMessage) error {
//       select {
//       case <-ctx.Done():
//           return ctx.Err()
//       default:
//           // process message
//       }
//   }
func (s *HederaSubscriber) SubscribeFrom(
    ctx context.Context,
    topic account.CommAddress,
    from time.Time,
    handler MessageHandler,
) (Subscription, error)
```

**Subscription Lifecycle:**

```
                         Start
                           |
                           v
                   +---------------+
                   | Create Query  |
                   +-------+-------+
                           |
                           v
                   +---------------+
              +--->| Subscribe     |
              |    +-------+-------+
              |            |
              |            v
              |    +---------------+
              |    | Wait for      |
              |    | Message       |
              |    +-------+-------+
              |            |
              |    +-------+-------+
              |    |               |
              |    v               v
              | Message       Timeout (3min)
              | Received      No Messages
              |    |               |
              |    v               |
              | handler(msg)       |
              |    |               |
              |    v               v
              | Reset Timer   Unsubscribe
              |    |               |
              +----+               |
                                   v
                           +---------------+
                           | Reconnect     |
                           +-------+-------+
                                   |
                                   v
                           (back to Subscribe)
```

### 5.6 HederaRegistry Implementation

```go
// HederaRegistry implements transport.RegistryReader using
// the rendezvous smart contract via JSON-RPC.
type HederaRegistry struct {
    client *HederaClient
}

// GetPeerInfo retrieves info for a specific peer.
//
// Implementation:
//   1. Creates contract caller from ethClient
//   2. Calls HederaAddressToPeer(address) on contract
//   3. Converts response to transport.PeerInfo
//   4. Returns ErrPeerNotFound if not registered
//
// Uses exponential backoff retry on transient failures.
func (r *HederaRegistry) GetPeerInfo(
    ctx context.Context,
    addr keylib.EVMAddress,
) (PeerInfo, error)

// GetPeerCount returns total registered peers.
//
// Calls GetPeerArraySize() on contract.
func (r *HederaRegistry) GetPeerCount(ctx context.Context) (uint64, error)

// GetPeerByIndex retrieves peer address at index.
//
// Calls PeerList(index) on contract.
func (r *HederaRegistry) GetPeerByIndex(
    ctx context.Context,
    index uint64,
) (keylib.EVMAddress, error)

// ListAvailablePeers returns all available peers.
//
// Implementation:
//   1. Get total peer count
//   2. Iterate all peers by index
//   3. Filter out Hedera-internal addresses (12+ leading zero bytes)
//   4. Filter out unavailable peers
//   5. Return list of available addresses
func (r *HederaRegistry) ListAvailablePeers(
    ctx context.Context,
) ([]keylib.EVMAddress, error)

// IterAvailablePeers returns a Go 1.23 iterator for memory-efficient streaming.
// This is preferred over ListAvailablePeers for large peer lists.
//
// Error Handling Semantics:
//   - Fatal errors (context canceled, GetPeerCount failed): Yields error and stops iteration
//   - Transient errors (single peer lookup failed): Yields error, caller decides whether to continue
//   - Silent skips (internal addresses, unavailable peers): No error yielded, continues automatically
//
// The caller MUST check the error on each iteration. Errors are wrapped with
// TransportError to indicate whether they are fatal (ErrKindCanceled, ErrKindConnection)
// or transient (ErrKindRegistry).
//
// Usage pattern:
//
//   for addr, err := range registry.IterAvailablePeers(ctx) {
//       if err != nil {
//           var tErr *TransportError
//           if errors.As(err, &tErr) {
//               switch tErr.Kind {
//               case ErrKindCanceled:
//                   return err // Fatal: context was canceled
//               case ErrKindConnection:
//                   return err // Fatal: lost connection to registry
//               default:
//                   slog.Warn("skipping peer", "error", err)
//                   continue // Transient: skip this peer and continue
//               }
//           }
//           continue // Unknown error type, skip
//       }
//       // Successfully retrieved available peer address
//       processPeer(addr)
//   }
func (r *HederaRegistry) IterAvailablePeers(ctx context.Context) iter.Seq2[keylib.EVMAddress, error] {
    return func(yield func(keylib.EVMAddress, error) bool) {
        count, err := r.GetPeerCount(ctx)
        if err != nil {
            // Fatal: cannot determine peer count
            yield(keylib.EVMAddress{}, &TransportError{
                Kind:    ErrKindConnection,
                Op:      "GetPeerCount",
                Message: "failed to get peer count from registry",
                Err:     err,
            })
            return
        }

        for i := uint64(0); i < count; i++ {
            // Check for context cancellation
            select {
            case <-ctx.Done():
                yield(keylib.EVMAddress{}, &TransportError{
                    Kind:    ErrKindCanceled,
                    Op:      "IterAvailablePeers",
                    Message: "iteration canceled",
                    Err:     ctx.Err(),
                })
                return
            default:
            }

            addr, err := r.GetPeerByIndex(ctx, i)
            if err != nil {
                // Transient: single index lookup failed, yield error and continue
                if !yield(keylib.EVMAddress{}, &TransportError{
                    Kind:    ErrKindRegistry,
                    Op:      "GetPeerByIndex",
                    Message: fmt.Sprintf("failed to get peer at index %d", i),
                    Err:     err,
                }) {
                    return // Caller stopped iteration
                }
                continue
            }

            // Skip Hedera-internal addresses (see isInternalAddress below)
            if isInternalAddress(addr) {
                continue
            }

            info, err := r.GetPeerInfo(ctx, addr)
            if err != nil {
                // Transient: peer info lookup failed, yield error and continue
                if !yield(keylib.EVMAddress{}, &TransportError{
                    Kind:    ErrKindRegistry,
                    Op:      "GetPeerInfo",
                    Message: fmt.Sprintf("failed to get info for peer %s", addr.Hex()),
                    Err:     err,
                }) {
                    return // Caller stopped iteration
                }
                continue
            }

            if info.Available {
                if !yield(addr, nil) {
                    return // Caller stopped iteration
                }
            }
        }
    }
}

// isInternalAddress returns true for Hedera-internal addresses.
//
// Hedera encodes account IDs as EVM addresses using the format:
//
//   0x000000000000000000000000{8-byte account number}
//
// For example, account 0.0.12345 becomes 0x0000000000000000000000000000000000003039.
// These addresses have the first 12 bytes set to zero.
//
// We filter these out because:
//   1. They represent Hedera system accounts, not Neuron peers
//   2. They cannot sign Ethereum-style transactions
//   3. Including them would pollute peer lists with non-peer entries
func isInternalAddress(addr keylib.EVMAddress) bool {
    bytes := addr.Bytes() // Returns [20]byte
    for i := 0; i < 12; i++ {
        if bytes[i] != 0 {
            return false // Found non-zero byte in first 12, not internal
        }
    }
    return true // First 12 bytes are all zero, this is internal
}
```

**Contract Method Mappings:**

| Registry Method | Contract Method              | Return Type     |
| --------------- | ---------------------------- | --------------- |
| GetPeerInfo     | HederaAddressToPeer(address) | PeerInfo struct |
| GetPeerCount    | GetPeerArraySize()           | uint256         |
| GetPeerByIndex  | PeerList(index)              | address         |
| RegisterSelf    | PutPeerAvailableSelf(...)    | -               |

---

## 6. Presence and Heartbeat Services

### 6.1 HeartbeatPayload

```go
// HeartbeatPayload is the structured data sent in heartbeat messages.
// JSON-serialized before publishing.
type HeartbeatPayload struct {
    // MessageType identifies this as a heartbeat.
    // Always "heartbeat".
    MessageType string `json:"messageType"`

    // Timestamp when this heartbeat was generated (UTC).
    Timestamp time.Time `json:"timestamp"`

    // Location is optional GPS coordinates.
    Location *GeoLocation `json:"location,omitempty"`

    // NATType describes the NAT device type (if detected).
    NATType string `json:"natDeviceType,omitempty"`

    // NATReachable indicates if the node is publicly reachable.
    NATReachable bool `json:"natReachability"`

    // ConnectedPeers lists abbreviated PeerIDs of connected peers.
    // Typically first 8 characters of each PeerID.
    ConnectedPeers []string `json:"connectedPeers,omitempty"`

    // Role identifies the node's function (buyer, seller, relay).
    Role string `json:"role,omitempty"`

    // Version is the SDK/protocol version string.
    Version string `json:"version"`
}

// GeoLocation represents geographic coordinates.
//
// PRIVACY WARNING: Location data published to HCS is permanent and public.
// Once published, it cannot be deleted or modified. Consider:
//   - Setting Location to nil for maximum privacy
//   - Using coarse coordinates (city-level only, e.g., 2 decimal places)
//   - Implementing location obfuscation before publishing
//   - Documenting location sharing in your user privacy policy
//
// Example of coarse vs precise coordinates:
//   Precise: 37.7749295, -122.4194155 (GPS-level, identifies a building)
//   Coarse:  37.77, -122.42 (city block level, acceptable for most uses)
type GeoLocation struct {
    Latitude  float64 `json:"lat"`
    Longitude float64 `json:"lon"`
    Altitude  float64 `json:"alt,omitempty"`
    Fix       string  `json:"gpsfix,omitempty"`
}
```

### 6.2 HeartbeatService

```go
// HeartbeatConfig configures the heartbeat service.
type HeartbeatConfig struct {
    // Interval between heartbeats.
    // Default: 30 seconds.
    Interval time.Duration

    // StdOutTopic where heartbeats are published.
    // Required.
    StdOutTopic account.CommAddress

    // PayloadProvider returns the heartbeat payload.
    // Called immediately before each heartbeat.
    // Allows dynamic data (connected peers, NAT status).
    // If nil, uses NewHeartbeatPayload(Version).
    PayloadProvider func() HeartbeatPayload

    // Version string included in heartbeats.
    // Required.
    Version string

    // SignHeartbeats enables cryptographic signing of heartbeat messages.
    // When true, heartbeats use PublishSigned() instead of Publish().
    // This allows receivers to verify the heartbeat sender's identity.
    // Default: true (STRONGLY RECOMMENDED for production).
    //
    // SECURITY WARNING: Unsigned heartbeats can be spoofed by anyone with
    // topic access. An attacker could fake heartbeats to make a peer appear
    // online when it's not, or inject false location/NAT data.
    SignHeartbeats bool

    // Signer is the private key used to sign heartbeat messages.
    // Required if SignHeartbeats is true.
    // Must correspond to the EVM address registered in the peer registry.
    Signer keylib.NeuronPrivateKey

    // MaxRetries is the number of retry attempts on publish failure.
    // Set to 0 to disable retries.
    // Default: 3.
    MaxRetries int

    // RetryDelay is the base delay between retry attempts.
    // Uses exponential backoff: delay * 2^attempt.
    // Default: 1 second.
    RetryDelay time.Duration

    // OnPublishError is called when a heartbeat publish fails after all retries.
    // If nil, errors are logged at WARN level.
    // The error will be a *TransportError with Kind=ErrKindPublish.
    OnPublishError func(error)
}

// HeartbeatService periodically publishes heartbeat messages.
// Decoupled from P2P through the TopicPublisher interface.
type HeartbeatService struct {
    config    HeartbeatConfig
    publisher transport.TopicPublisher
    // internal fields...
}

// NewHeartbeatService creates a new heartbeat service.
// Does not start the service; call Start() separately.
//
// Panics if publisher is nil or config.StdOutTopic is zero.
func NewHeartbeatService(
    publisher transport.TopicPublisher,
    config HeartbeatConfig,
) *HeartbeatService

// Start begins the heartbeat loop.
// Sends first heartbeat immediately, then at configured interval.
// Non-blocking; runs in background goroutine.
// Returns error if already running.
//
// The heartbeat loop:
//   1. Calls PayloadProvider (or uses default payload)
//   2. JSON-encodes the payload
//   3. If SignHeartbeats=true, uses PublishSigned; otherwise Publish
//   4. On failure, retries up to MaxRetries times
//   5. On final failure, calls OnPublishError (or logs at WARN)
//   6. Waits for Interval before next heartbeat
//   7. Repeats until Stop() or context canceled
func (s *HeartbeatService) Start(ctx context.Context) error {
    s.mu.Lock()
    if s.running {
        s.mu.Unlock()
        return errors.New("heartbeat service already running")
    }
    s.running = true
    s.stopCh = make(chan struct{})
    s.mu.Unlock()

    go s.loop(ctx)
    return nil
}

// Stop halts the heartbeat loop.
// Blocks until the loop has exited.
// Safe to call multiple times.
func (s *HeartbeatService) Stop() {
    s.mu.Lock()
    if !s.running {
        s.mu.Unlock()
        return
    }
    close(s.stopCh)
    s.mu.Unlock()
    s.wg.Wait()
}

// IsRunning returns true if the heartbeat loop is active.
// Thread-safe.
func (s *HeartbeatService) IsRunning() bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.running
}
```

**Heartbeat Timing:**

```
Start()
   |
   v
[Send immediately]
   |
   v
[Wait Interval]
   |
   v
[Send heartbeat]
   |
   v
[Wait Interval]
   |
   v
...continues until Stop() or context canceled...
```

### 6.3 PresenceService

```go
// PresenceConfig configures presence tracking.
type PresenceConfig struct {
    // ActiveThreshold is how long without heartbeat before peer is inactive.
    // Should be at least 2x the expected heartbeat interval to tolerate
    // occasional network delays or missed heartbeats.
    // Default: 5 minutes (assumes 30s-2min heartbeat intervals).
    ActiveThreshold time.Duration

    // CleanupInterval is how often to remove stale entries from memory.
    // Set to 0 to disable automatic cleanup (manual cleanup via RemovePeer).
    // Default: 1 minute.
    CleanupInterval time.Duration

    // MaxTrackedPeers limits memory usage by capping tracked peers.
    // When limit is reached, oldest inactive peers are evicted first.
    // Set to 0 for unlimited (not recommended in production).
    // Default: 1000.
    MaxTrackedPeers int

    // VerifySignatures enables cryptographic verification of heartbeat senders.
    // When true, only heartbeats signed by the expected peer address are accepted.
    // Heartbeats failing verification are silently dropped and logged at DEBUG level.
    // Default: true (STRONGLY RECOMMENDED for production).
    //
    // SECURITY WARNING: Without signature verification, any party with topic
    // access can spoof heartbeats for arbitrary peer addresses. This could allow:
    //   - Making offline peers appear online (availability spoofing)
    //   - Injecting false location/NAT/peer data (data poisoning)
    //   - Impersonating trusted peers (identity spoofing)
    VerifySignatures bool

    // OnVerificationFailure is called when a heartbeat fails signature verification.
    // Useful for security monitoring and alerting on potential spoofing attempts.
    // If nil, verification failures are logged at DEBUG level only.
    // Parameters: claimedAddr (who the heartbeat claims to be from), actualAddr (recovered signer)
    OnVerificationFailure func(claimedAddr, actualAddr keylib.EVMAddress)

    // OnSubscribeError is called when subscribing to a peer's topic fails.
    // If nil, errors are logged at WARN level.
    OnSubscribeError func(peerAddr keylib.EVMAddress, topic account.CommAddress, err error)
}

// PresenceInfo tracks observed presence of a peer.
type PresenceInfo struct {
    PeerID                keylib.PeerID
    EVMAddress            keylib.EVMAddress
    LastSeen              time.Time
    LastPayload           HeartbeatPayload
    ConsecutiveHeartbeats int
    FailedPings           int
}

// PresenceService tracks peer liveness based on heartbeat messages.
type PresenceService struct {
    config     PresenceConfig
    subscriber transport.TopicSubscriber
    peers      map[keylib.EVMAddress]PresenceInfo
    // internal fields...
}

// NewPresenceService creates a presence tracking service.
func NewPresenceService(
    subscriber transport.TopicSubscriber,
    config PresenceConfig,
) *PresenceService

// TrackPeer starts watching a peer's StdOut topic for heartbeats.
// Updates internal state when heartbeats are received.
//
// When VerifySignatures is enabled (default), heartbeat processing follows:
//   1. Extract signed message using transport.ExtractSignedMessage()
//   2. Recover signer address using SignedMessage.RecoverSender()
//   3. Verify signer matches peerAddr
//   4. If verification fails, drop message and invoke OnVerificationFailure
//   5. If verification succeeds, parse payload and update PresenceInfo
//
// Example internal handler (simplified):
//
//   func (s *PresenceService) handleHeartbeat(peerAddr keylib.EVMAddress, msg transport.TopicMessage) {
//       if s.config.VerifySignatures {
//           signed, err := transport.ExtractSignedMessage(msg.Payload)
//           if err != nil {
//               slog.Debug("invalid signed message", "peer", peerAddr, "error", err)
//               return
//           }
//           actualSender, err := signed.RecoverSender()
//           if err != nil || actualSender != peerAddr {
//               if s.config.OnVerificationFailure != nil {
//                   s.config.OnVerificationFailure(peerAddr, actualSender)
//               }
//               return
//           }
//       }
//       // Parse and process verified heartbeat...
//   }
//
// Returns error if subscription to the topic fails.
func (s *PresenceService) TrackPeer(
    ctx context.Context,
    stdOutTopic account.CommAddress,
    peerAddr keylib.EVMAddress,
) error

// GetPresence returns presence info for a tracked peer.
//
// Return values:
//   - (info, true): Peer is being tracked; info contains latest presence data
//   - (PresenceInfo{}, false): Peer is NOT being tracked (call TrackPeer first)
//
// Note: This returns tracking status, not liveness status. A tracked peer may
// still be inactive (no recent heartbeat). Use IsActive() to check liveness.
//
// Example:
//
//   info, tracked := svc.GetPresence(addr)
//   if !tracked {
//       // Not tracking this peer yet
//       return
//   }
//   if time.Since(info.LastSeen) > threshold {
//       // Peer is tracked but inactive
//   }
func (s *PresenceService) GetPresence(addr keylib.EVMAddress) (PresenceInfo, bool)

// IsActive returns true if:
//   1. The peer is being tracked (TrackPeer was called), AND
//   2. A heartbeat was received within ActiveThreshold
//
// Returns false if peer is not tracked or no recent heartbeat.
// This is a convenience method equivalent to:
//
//   info, tracked := svc.GetPresence(addr)
//   return tracked && time.Since(info.LastSeen) <= cfg.ActiveThreshold
func (s *PresenceService) IsActive(addr keylib.EVMAddress) bool

// ListActivePeers returns addresses of all peers that are:
//   1. Currently being tracked, AND
//   2. Have sent a heartbeat within ActiveThreshold
//
// Returns empty slice if no peers are active.
// Thread-safe; takes a snapshot of current state.
func (s *PresenceService) ListActivePeers() []keylib.EVMAddress

// RemovePeer stops tracking a peer and removes all associated state.
// Safe to call for peers that are not being tracked (no-op).
func (s *PresenceService) RemovePeer(addr keylib.EVMAddress)
```

---

## 7. Peer Discovery Service

### 7.1 DiscoveryConfig

```go
// DiscoveryConfig configures the peer discovery service.
type DiscoveryConfig struct {
    // CacheTTL is how long to cache peer info before considering it stale.
    // Stale entries are refreshed on next access.
    // Set to 0 to disable caching (always fetch from registry).
    // Default: 5 minutes.
    CacheTTL time.Duration

    // MaxCacheSize is the maximum number of cached peers.
    // When exceeded, least-recently-used entries are evicted.
    // Set to 0 for unlimited (not recommended for large networks).
    // Default: 1000.
    MaxCacheSize int

    // RefreshInterval for background refresh of frequently accessed entries.
    // Only refreshes entries accessed within the last refresh interval.
    // Set to 0 to disable background refresh (on-demand only).
    // Default: 2 minutes.
    RefreshInterval time.Duration

    // MaxConcurrentRequests limits parallel registry calls during batch operations.
    // Prevents overwhelming the registry with concurrent requests.
    // Default: 10.
    MaxConcurrentRequests int

    // CircuitBreakerEnabled activates the circuit breaker for registry failures.
    // When enabled, after ConsecutiveFailures, requests fail fast for BreakDuration.
    // Default: true.
    CircuitBreakerEnabled bool

    // ConsecutiveFailures before the circuit breaker trips.
    // Only used if CircuitBreakerEnabled is true.
    // Default: 5.
    ConsecutiveFailures int

    // BreakDuration is how long the circuit stays open after tripping.
    // During this time, requests fail immediately without calling registry.
    // Default: 30 seconds.
    BreakDuration time.Duration

    // OnCacheEviction is called when a peer is evicted from cache.
    // Useful for metrics and debugging cache behavior.
    // Optional.
    OnCacheEviction func(addr keylib.EVMAddress, reason string)
}

// WithDefaults returns a DiscoveryConfig with default values applied.
func (c DiscoveryConfig) WithDefaults() DiscoveryConfig {
    if c.CacheTTL == 0 {
        c.CacheTTL = 5 * time.Minute
    }
    if c.MaxCacheSize == 0 {
        c.MaxCacheSize = 1000
    }
    if c.RefreshInterval == 0 {
        c.RefreshInterval = 2 * time.Minute
    }
    if c.MaxConcurrentRequests == 0 {
        c.MaxConcurrentRequests = 10
    }
    if c.ConsecutiveFailures == 0 {
        c.ConsecutiveFailures = 5
    }
    if c.BreakDuration == 0 {
        c.BreakDuration = 30 * time.Second
    }
    // CircuitBreakerEnabled defaults to true (zero value is false)
    // Callers should explicitly set to true in production
    return c
}
```

### 7.2 PeerDiscoveryService

```go
// CachedPeerInfo extends transport.PeerInfo with cache metadata.
// Named differently from transport.PeerInfo to avoid confusion.
// Use transport.PeerInfo for registry operations, CachedPeerInfo for cached data.
type CachedPeerInfo struct {
    // Embedded transport.PeerInfo with all peer data
    transport.PeerInfo

    // FetchedAt is when this data was retrieved from the registry.
    // Useful for debugging stale data issues.
    FetchedAt time.Time

    // ExpiresAt is when this cache entry becomes stale.
    // Computed as FetchedAt + CacheTTL.
    ExpiresAt time.Time
}

// IsExpired returns true if cache entry is stale.
func (p CachedPeerInfo) IsExpired() bool {
    return time.Now().After(p.ExpiresAt)
}

// TTLRemaining returns how long until this entry expires.
// Returns zero if already expired.
func (p CachedPeerInfo) TTLRemaining() time.Duration {
    remaining := time.Until(p.ExpiresAt)
    if remaining < 0 {
        return 0
    }
    return remaining
}

// LRUCache is a generic least-recently-used cache.
// Thread-safe for concurrent access.
type LRUCache[K comparable, V any] struct {
    maxSize int
    // internal fields (mutex, list, map)...
}

// NewLRUCache creates a cache with the specified maximum size.
// When full, least-recently-used entries are evicted.
func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V]

// Get retrieves a value by key. Returns (value, true) if found, (zero, false) if not.
func (c *LRUCache[K, V]) Get(key K) (V, bool)

// Put stores a value, evicting LRU entry if at capacity.
func (c *LRUCache[K, V]) Put(key K, value V)

// Delete removes a key from the cache.
func (c *LRUCache[K, V]) Delete(key K)

// Len returns the current number of entries.
func (c *LRUCache[K, V]) Len() int

// PeerDiscoveryService provides cached access to peer registry data.
// Reduces registry queries by caching peer info with configurable TTL.
//
// Thread Safety: All methods are safe for concurrent access.
type PeerDiscoveryService struct {
    config   DiscoveryConfig
    registry transport.RegistryReader
    cache    *LRUCache[keylib.EVMAddress, CachedPeerInfo]
    mu       sync.RWMutex // Protects concurrent cache operations
    // internal fields...
}

// NewPeerDiscoveryService creates a discovery service with caching.
func NewPeerDiscoveryService(
    registry transport.RegistryReader,
    config DiscoveryConfig,
) *PeerDiscoveryService

// GetPeer retrieves peer info, using cache if available and fresh.
//
// Implementation:
//   1. Check cache for unexpired entry
//   2. Return cached if valid (cache hit)
//   3. Fetch from registry on miss/expiry
//   4. Update cache with new entry (sets ExpiresAt = now + CacheTTL)
//   5. Return result
//
// Returns ErrKindNotFound if peer doesn't exist in registry.
func (s *PeerDiscoveryService) GetPeer(
    ctx context.Context,
    addr keylib.EVMAddress,
) (CachedPeerInfo, error)

// ListAvailablePeers returns all available peers from the registry.
// Does NOT use cache - always fetches fresh data for completeness.
// Consider using GetPeer for individual lookups with caching.
func (s *PeerDiscoveryService) ListAvailablePeers(
    ctx context.Context,
) ([]CachedPeerInfo, error)

// Invalidate removes a peer from cache.
// Forces fresh fetch on next GetPeer call.
func (s *PeerDiscoveryService) Invalidate(addr keylib.EVMAddress)

// Refresh forces a cache refresh for a specific peer.
// Always fetches from registry, ignoring any cached value.
func (s *PeerDiscoveryService) Refresh(
    ctx context.Context,
    addr keylib.EVMAddress,
) (CachedPeerInfo, error)
```

**Cache Lookup Flow:**

```
GetPeer(addr)
     |
     v
+----------+     Yes    +----------+
| In Cache?|----------->| Expired? |
+----+-----+            +----+-----+
     |                       |
     | No                    | No
     |                       |
     |                       v
     |               [Return cached]
     |
     v
+--------------+
| Fetch from   |
| Registry     |
+------+-------+
       |
       v
+--------------+
| Update Cache |
| (TTL=5min)   |
+------+-------+
       |
       v
  [Return info]
```

---

## 8. Data Flow Diagrams

### 8.1 Heartbeat Publishing Flow

```
+-------------+     +------------------+     +------------------+     +---------------+
| Application |     | HeartbeatService |     | HederaPublisher  |     | Hedera        |
+------+------+     +--------+---------+     +--------+---------+     | Network       |
       |                     |                        |               +-------+-------+
       | Start(ctx)          |                        |                       |
       |-------------------->|                        |                       |
       |                     |                        |                       |
       |                [ticker fires]                |                       |
       |                     |                        |                       |
       |                     | PayloadProvider()      |                       |
       |                     |<-----------------------|                       |
       |                     |                        |                       |
       |                     | json.Marshal(payload)  |                       |
       |                     |----------------------->|                       |
       |                     |                        |                       |
       |                     |                        | TopicMessageSubmit    |
       |                     |                        | Transaction           |
       |                     |                        |---------------------->|
       |                     |                        |                       |
       |                     |                        |<----------------------|
       |                     |                        | Receipt               |
       |                     |<-----------------------|                       |
       |                     |                        |                       |
       |                [wait interval]               |                       |
       |                     |                        |                       |
       |                [repeat...]                   |                       |
```

### 8.2 Peer Discovery Flow

```
+-------------+     +---------------------+     +----------------+     +----------+
| Application |     | PeerDiscoveryService|     | HederaRegistry |     | Contract |
+------+------+     +---------+-----------+     +-------+--------+     +----+-----+
       |                      |                         |                    |
       | GetPeer(addr)        |                         |                    |
       |--------------------->|                         |                    |
       |                      |                         |                    |
       |                [check cache]                   |                    |
       |                      |                         |                    |
       |           [cache miss or expired]              |                    |
       |                      |                         |                    |
       |                      | GetPeerInfo(addr)       |                    |
       |                      |------------------------>|                    |
       |                      |                         |                    |
       |                      |                         | eth_call           |
       |                      |                         | HederaAddressToPeer|
       |                      |                         |------------------->|
       |                      |                         |                    |
       |                      |                         |<-------------------|
       |                      |                         | PeerInfo struct    |
       |                      |<------------------------|                    |
       |                      |                         |                    |
       |                [update cache]                  |                    |
       |                      |                         |                    |
       |<---------------------|                         |                    |
       | PeerInfo             |                         |                    |
```

### 8.3 Topic Subscription Flow

```
+-------------+     +------------------+     +-------------+
| Application |     | HederaSubscriber |     | Mirror Node |
+------+------+     +--------+---------+     +------+------+
       |                     |                      |
       | Subscribe(topic, h) |                      |
       |-------------------->|                      |
       |                     |                      |
       |    Subscription     | TopicMessageQuery    |
       |<--------------------|--------------------->|
       |                     |    [gRPC stream]     |
       |                     |                      |
       |                     |<---------------------|
       |                     |    TopicMessage      |
       |                     |                      |
       |                [convert message]           |
       |                     |                      |
       |    handler(msg)     |                      |
       |<--------------------|                      |
       |                     |                      |
       |                [reset timeout]             |
       |                     |                      |
       |                     |<---------------------|
       |                     |    TopicMessage      |
       |    handler(msg)     |                      |
       |<--------------------|                      |
       |                     |                      |
       |            [3-min timeout, no messages]    |
       |                     |                      |
       |                     | Unsubscribe          |
       |                     |--------------------->|
       |                     |                      |
       |                     | Resubscribe          |
       |                     |--------------------->|
       |                     |    [new stream]      |
       |                     |                      |
```

---

## 9. Error Handling Strategy

### 9.1 Error Type Hierarchy

```go
// TransportErrorKind categorizes transport errors.
type TransportErrorKind int

const (
    ErrKindConnection     TransportErrorKind = iota + 1 // Network/connection failure
    ErrKindPublish                                      // Failed to publish message
    ErrKindSubscribe                                    // Failed to subscribe
    ErrKindRegistry                                     // Contract read/write failure
    ErrKindTimeout                                      // Operation timed out
    ErrKindCanceled                                     // Context was canceled
    ErrKindRateLimit                                    // Rate limit exceeded
    ErrKindNotFound                                     // Resource not found
    ErrKindConfiguration                                // Invalid configuration
    ErrKindClosed                                       // Operation on closed client
    ErrKindInvalidTopic                                 // Malformed topic address
    ErrKindPayloadTooLarge                              // Message exceeds HCS limit
    ErrKindSignature                                    // Signature operation failed
    ErrKindInvalidFormat                                // Malformed message/data
)

// TransportError is the base error type for all transport errors.
type TransportError struct {
    Op      string             // Operation name (e.g., "HederaPublisher.Publish")
    Kind    TransportErrorKind // Error category
    Message string             // Human-readable error message
    Err     error              // Underlying error (may be nil)
}

func (e *TransportError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Op, e.Message)
}
func (e *TransportError) Unwrap() error { return e.Err }
func (e *TransportError) Is(target error) bool // Match by Kind

// IsRetryable returns true if this error may succeed on retry.
// Use this to determine whether to retry an operation.
func (e *TransportError) IsRetryable() bool {
    switch e.Kind {
    case ErrKindConnection, ErrKindTimeout, ErrKindRateLimit:
        return true
    default:
        return false
    }
}
```

**Retry Semantics Table:**

| Error Kind             | Retried? | Rationale                                         |
| ---------------------- | -------- | ------------------------------------------------- |
| ErrKindConnection      | Yes      | Transient network issue, may recover              |
| ErrKindTimeout         | Yes      | Transient, may succeed on retry                   |
| ErrKindRateLimit       | Yes      | Back off and retry with exponential delay         |
| ErrKindPublish         | Maybe    | Depends on underlying cause (check wrapped error) |
| ErrKindSubscribe       | Maybe    | Depends on underlying cause                       |
| ErrKindRegistry        | Maybe    | Network issues yes, contract errors no            |
| ErrKindNotFound        | No       | Resource doesn't exist, permanent                 |
| ErrKindCanceled        | No       | User/caller requested cancellation                |
| ErrKindClosed          | No       | Client is shut down, cannot recover               |
| ErrKindConfiguration   | No       | Invalid config, won't change without code fix     |
| ErrKindInvalidTopic    | No       | Bad topic format, permanent                       |
| ErrKindPayloadTooLarge | No       | Must reduce payload size                          |
| ErrKindSignature       | No       | Cryptographic failure, permanent                  |
| ErrKindInvalidFormat   | No       | Malformed data, permanent                         |

### 9.2 Sentinel Errors

```go
var (
    // Connection errors
    ErrNotConnected = &TransportError{Kind: ErrKindConnection, Message: "client is not connected"}
    ErrClientClosed = &TransportError{Kind: ErrKindClosed, Message: "client has been closed"}

    // Resource errors
    ErrPeerNotFound  = &TransportError{Kind: ErrKindNotFound, Message: "peer not found in registry"}
    ErrTopicNotFound = &TransportError{Kind: ErrKindNotFound, Message: "topic does not exist"}

    // Validation errors
    ErrInvalidTopic      = &TransportError{Kind: ErrKindInvalidTopic, Message: "invalid topic address format"}
    ErrPayloadTooLarge   = &TransportError{Kind: ErrKindPayloadTooLarge, Message: "payload exceeds maximum size (1024 bytes)"}
    ErrInvalidSignature  = &TransportError{Kind: ErrKindSignature, Message: "invalid or corrupted signature"}
    ErrSignatureRequired = &TransportError{Kind: ErrKindSignature, Message: "signed message required but not provided"}

    // Operation errors
    ErrOperationCanceled = &TransportError{Kind: ErrKindCanceled, Message: "operation was canceled"}
    ErrOperationTimeout  = &TransportError{Kind: ErrKindTimeout, Message: "operation timed out"}
    ErrRateLimitExceeded = &TransportError{Kind: ErrKindRateLimit, Message: "rate limit exceeded, retry later"}
)

// NewTransportError creates a new TransportError with the given parameters.
func NewTransportError(op string, kind TransportErrorKind, msg string, err error) *TransportError {
    return &TransportError{
        Op:      op,
        Kind:    kind,
        Message: msg,
        Err:     err,
    }
}

// WrapError wraps an existing error with transport context.
// Preserves the original error for unwrapping.
func WrapError(op string, kind TransportErrorKind, err error) *TransportError {
    return &TransportError{
        Op:      op,
        Kind:    kind,
        Message: err.Error(),
        Err:     err,
    }
}
```

### 9.3 Error Handling Patterns

**Checking Error Kind:**

```go
info, err := registry.GetPeerInfo(ctx, addr)
if err != nil {
    var te *transport.TransportError
    if errors.As(err, &te) {
        switch te.Kind {
        case transport.ErrKindNotFound:
            // Peer not registered
        case transport.ErrKindTimeout:
            // Retry later
        case transport.ErrKindRateLimit:
            // Back off
        }
    }
}
```

**Using Sentinel Errors:**

```go
if errors.Is(err, transport.ErrPeerNotFound) {
    // Handle missing peer
}
```

---

## 10. Integration Points

### 10.1 keylib Integration

| Transport Usage     | keylib Type                | Purpose                    |
| ------------------- | -------------------------- | -------------------------- |
| Peer identification | `keylib.EVMAddress`        | Unique peer address        |
| P2P identity        | `keylib.PeerID`            | Libp2p peer ID             |
| Transaction signing | `keylib.NeuronPrivateKey`  | Sign HCS messages          |
| Operator setup      | `ToHederaPrivateKeySafe()` | Create Hedera operator     |
| Message signing     | `SignMessage(payload)`     | Create verifiable messages |

### 10.2 account Integration

| Transport Usage    | account Type                 | Purpose                       |
| ------------------ | ---------------------------- | ----------------------------- |
| Topic addressing   | `account.CommAddress`        | Abstract topic reference      |
| Message format     | `account.TopicMessage`       | Standardized message envelope |
| Backend validation | `account.HederaTopicBackend` | Validate topic format         |
| Topic technology   | `account.TopicTechnology`    | Identify backend type         |

### 10.3 Usage Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/hedera"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
    "github.com/aspect-build/neuron-go-hedera-sdk/presence"
    "github.com/aspect-build/neuron-go-hedera-sdk/discovery"
)

func main() {
    // Setup structured logging (Go 1.21+)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Load credentials (application responsibility, not SDK)
    privKey, err := keylib.ParsePrivateKeyHex(os.Getenv("PRIVATE_KEY"))
    if err != nil {
        logger.Error("failed to parse private key", slog.String("error", err.Error()))
        os.Exit(1)
    }
    contractAddr, _ := keylib.ParseEVMAddress(os.Getenv("CONTRACT_ADDRESS"))

    // Create configuration (explicit, no env vars in SDK)
    config := hedera.HederaConfig{
        Network:                 hedera.NetworkTestnet,
        OperatorAccountID:       os.Getenv("HEDERA_ACCOUNT_ID"),
        OperatorKey:             privKey,
        EthRPCURL:               os.Getenv("ETH_RPC_URL"),
        RegistryContractAddress: contractAddr,
    }

    // Create client with functional options
    client, err := hedera.NewHederaClient(config,
        hedera.WithLogger(logger.With(slog.String("component", "hedera"))),
    )
    if err != nil {
        logger.Error("failed to create client", slog.String("error", err.Error()))
        os.Exit(1)
    }

    // Get my peer info from registry
    myAddr := privKey.EVMAddress()
    registry := client.Registry()

    myInfo, err := registry.GetPeerInfo(ctx, myAddr)
    if err != nil {
        logger.Error("failed to get peer info", slog.String("error", err.Error()))
        os.Exit(1)
    }

    // Create discovery service with caching
    discoverySvc := discovery.NewPeerDiscoveryService(
        registry,
        discovery.DiscoveryConfig{
            CacheTTL:     5 * time.Minute,
            MaxCacheSize: 1000,
        },
    )

    // Start heartbeat service
    heartbeatSvc := presence.NewHeartbeatService(
        client.Publisher(),
        presence.HeartbeatConfig{
            Interval:    30 * time.Second,
            StdOutTopic: myInfo.StdOut,
            Version:     "1.0.0",
            PayloadProvider: func() presence.HeartbeatPayload {
                return presence.HeartbeatPayload{
                    MessageType:  "heartbeat",
                    Timestamp:    time.Now().UTC(),
                    NATReachable: true,
                    Version:      "1.0.0",
                    Role:         "seller",
                }
            },
        },
    )

    if err := heartbeatSvc.Start(ctx); err != nil {
        logger.Error("failed to start heartbeat", slog.String("error", err.Error()))
    }

    // Subscribe to incoming messages with context-aware handler
    subscriber := client.Subscriber()
    sub, err := subscriber.Subscribe(ctx, myInfo.StdIn,
        func(ctx context.Context, msg account.TopicMessage) error {
            logger.InfoContext(ctx, "received message",
                slog.String("payload", string(msg.Payload)),
                slog.Time("timestamp", msg.Timestamp),
            )
            return nil
        },
    )
    if err != nil {
        logger.Error("failed to subscribe", slog.String("error", err.Error()))
    }

    // Use Go 1.23 iterator for memory-efficient peer listing
    logger.Info("listing available peers using iterator")
    for addr, err := range registry.IterAvailablePeers(ctx) {
        if err != nil {
            logger.Warn("error iterating peers", slog.String("error", err.Error()))
            continue
        }
        logger.Info("found peer", slog.String("address", addr.String()))
    }

    // Discover a specific peer (uses cache)
    peerAddr, _ := keylib.ParseEVMAddress("0x...")
    peerInfo, err := discoverySvc.GetPeer(ctx, peerAddr)
    if err != nil {
        logger.Warn("peer not found", slog.String("error", err.Error()))
    } else {
        logger.Info("found peer", slog.String("peer_id", peerInfo.PeerID.String()))
    }

    // Wait for shutdown signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    logger.Info("shutting down...")

    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    // Stop services in reverse order
    heartbeatSvc.Stop()
    sub.Unsubscribe()

    // Graceful client shutdown (waits for in-flight operations)
    if err := client.Shutdown(shutdownCtx); err != nil {
        logger.Error("shutdown error", slog.String("error", err.Error()))
    }

    logger.Info("shutdown complete")
}
```

---

## 11. Configuration Management

### 11.1 Configuration Hierarchy

```
+---------------------------+
|      Application Code     |
|                           |
|  config := hedera.Config{ |
|    Network: Testnet,      |
|    OperatorKey: key,      |
|    ...                    |
|  }                        |
+-------------+-------------+
              |
              v
+---------------------------+
|      HederaConfig         |
|                           |
|  - Explicit fields        |
|  - Validate() method      |
|  - WithDefaults() method  |
+-------------+-------------+
              |
              v
+---------------------------+
|      HederaClient         |
|                           |
|  - Uses validated config  |
|  - No env var lookups     |
+---------------------------+
```

### 11.2 Default Values

| Config Field        | Default Value | Rationale                        |
| ------------------- | ------------- | -------------------------------- |
| SubscriptionTimeout | 3 minutes     | Match legacy SDK behavior        |
| MaxRetries          | 5             | Balance reliability vs latency   |
| BaseRetryDelay      | 1 second      | Standard exponential backoff     |
| HeartbeatInterval   | 30 seconds    | Reasonable liveness detection    |
| CacheTTL            | 5 minutes     | Balance freshness vs performance |
| ActiveThreshold     | 5 minutes     | ~10 missed heartbeats            |

### 11.3 Environment Variable Loading (Application Responsibility)

```go
// SDK does NOT do this internally
// Application code example:
func loadConfig() hedera.HederaConfig {
    privKey, _ := keylib.ParsePrivateKeyHex(os.Getenv("PRIVATE_KEY"))
    contractAddr, _ := keylib.ParseEVMAddress(os.Getenv("CONTRACT_ADDRESS"))

    return hedera.HederaConfig{
        Network:                 hedera.NetworkName(os.Getenv("HEDERA_NETWORK")),
        OperatorAccountID:       os.Getenv("HEDERA_ACCOUNT_ID"),
        OperatorKey:             privKey,
        EthRPCURL:               os.Getenv("ETH_RPC_URL"),
        RegistryContractAddress: contractAddr,
    }.WithDefaults()
}
```

---

## 12. Comparison with Legacy SDK

### 12.1 Architectural Differences

| Aspect                   | Legacy SDK                                           | New SDK                                    |
| ------------------------ | ---------------------------------------------------- | ------------------------------------------ |
| **Client Lifecycle**     | Per-operation creation (`GetHederaClientUsingEnv()`) | Single long-lived `HederaClient`           |
| **Configuration**        | `os.Getenv()` scattered throughout                   | Explicit `HederaConfig` struct             |
| **Type Safety**          | String-based (`"0.0.12345"`)                         | Type-safe (`keylib.EVMAddress`, `TopicID`) |
| **Error Handling**       | `log.Fatal`/`log.Panic` mixed                        | Structured `TransportError` returns        |
| **Testability**          | Difficult (global state, env vars)                   | Easy (interface injection, mocks)          |
| **Retry Logic**          | Inline per function                                  | Centralized with config                    |
| **Subscription Timeout** | Hard-coded 3 minutes                                 | Configurable                               |
| **Logging**              | `log.Println` unstructured                           | `slog` structured logging                  |
| **Constructors**         | Single signature                                     | Functional options pattern                 |
| **Shutdown**             | `Close()` immediate                                  | `Shutdown(ctx)` graceful drain             |
| **Iteration**            | Return full slice                                    | `iter.Seq2` streaming                      |
| **Handler Context**      | No context                                           | Context-aware handlers                     |

### 12.2 Code Comparison

**Legacy SDK - Client Creation:**

```go
// Called every operation
func GetHederaClientUsingEnv() *hedera.Client {
    c, _ := hedera.ClientForName(hedera.NetworkNameTestnet.String())
    op, _ := hedera.AccountIDFromString(os.Getenv("hedera_id"))
    pkString := os.Getenv("private_key")
    // ...
    return c
}
```

**New SDK - Client Creation:**

```go
// Created once, reused
config := hedera.HederaConfig{
    Network:           hedera.NetworkTestnet,
    OperatorAccountID: "0.0.12345",
    OperatorKey:       privKey,
    // ...
}
client, _ := hedera.NewHederaClient(config)
defer client.Close()
```

**Legacy SDK - Peer Lookup:**

```go
func GetPeerInfo(hederaAccEvmAddress string) (PeerInfo, error) {
    // String-based, 25 retries hard-coded
}
```

**New SDK - Peer Lookup:**

```go
func (r *HederaRegistry) GetPeerInfo(ctx context.Context, addr keylib.EVMAddress) (PeerInfo, error) {
    // Type-safe, configurable retries
}
```

### 12.3 Migration Path

1. Replace `GetHederaClientUsingEnv()` calls with `HederaClient`
2. Replace string addresses with `keylib.EVMAddress`
3. Replace string topic IDs with `account.CommAddress`
4. Replace `log.Fatal` with error returns
5. Move env var loading to application startup
6. Inject interfaces for testing

---

## 13. Implementation Roadmap

### Phase 1: Transport Interfaces (Week 1)

| Task | File                      | Description                                            |
| ---- | ------------------------- | ------------------------------------------------------ |
| 1.1  | `transport/interfaces.go` | Define TopicPublisher, TopicSubscriber, RegistryReader |
| 1.2  | `transport/types.go`      | Define PeerInfo, SelfRegistrationInfo                  |
| 1.3  | `transport/errors.go`     | Define TransportError, error kinds, sentinels          |
| 1.4  | `transport/doc.go`        | Package documentation                                  |

### Phase 2: Hedera Implementation (Week 2-3)

| Task | File                   | Description                          |
| ---- | ---------------------- | ------------------------------------ |
| 2.1  | `hedera/config.go`     | HederaConfig with validation         |
| 2.2  | `hedera/topic_id.go`   | TopicID type with conversions        |
| 2.3  | `hedera/client.go`     | HederaClient lifecycle management    |
| 2.4  | `hedera/publisher.go`  | HederaPublisher implementation       |
| 2.5  | `hedera/subscriber.go` | HederaSubscriber with auto-reconnect |
| 2.6  | `hedera/registry.go`   | HederaRegistry for contract reads    |
| 2.7  | `hedera/errors.go`     | Hedera-specific errors               |
| 2.8  | `hedera/doc.go`        | Package documentation                |

### Phase 3: Services (Week 4)

| Task | File                    | Description                     |
| ---- | ----------------------- | ------------------------------- |
| 3.1  | `presence/types.go`     | HeartbeatPayload, PresenceInfo  |
| 3.2  | `presence/config.go`    | HeartbeatConfig, PresenceConfig |
| 3.3  | `presence/heartbeat.go` | HeartbeatService                |
| 3.4  | `presence/presence.go`  | PresenceService                 |
| 3.5  | `discovery/cache.go`    | LRU cache implementation        |
| 3.6  | `discovery/types.go`    | Extended PeerInfo               |
| 3.7  | `discovery/service.go`  | PeerDiscoveryService            |
| 3.8  | `discovery/config.go`   | DiscoveryConfig                 |

### Phase 4: Testing (Week 5)

| Task | Description                                 |
| ---- | ------------------------------------------- |
| 4.1  | Unit tests for transport interfaces (mocks) |
| 4.2  | Unit tests for hedera package               |
| 4.3  | Unit tests for presence package             |
| 4.4  | Unit tests for discovery package            |
| 4.5  | Integration tests with Hedera testnet       |
| 4.6  | Example application                         |

### Verification Checklist

**Code Quality:**

- [ ] All interfaces have mock implementations
- [ ] Unit test coverage >80% per package
- [ ] Integration tests pass with testnet credentials
- [ ] Error conditions tested (timeouts, retries, cancellation)
- [ ] Documentation complete (godoc, examples)
- [ ] No `os.Getenv()` calls inside SDK packages
- [ ] All public types use keylib/account types (no raw strings)

**Go 2026 Best Practices:**

- [ ] Uses `log/slog` for structured logging (no `log.Printf`)
- [ ] Uses `iter.Seq2` for streaming list operations
- [ ] Uses functional options pattern for constructors
- [ ] All handlers receive context parameter
- [ ] Graceful shutdown via `Shutdown(ctx)` method
- [ ] Uses `errors.Join` for multi-error aggregation
- [ ] go.mod specifies `go 1.23` or higher

**Hiero SDK Compliance:**

- [ ] Uses `hiero-sdk-go/v2` imports
- [ ] Proper operator setup with ECDSA keys
- [ ] Mirror node subscription with auto-reconnect
- [ ] Respects network rate limits with exponential backoff

---

## Appendix A: Contract ABI Reference

**Rendezvous Contract Methods:**

```solidity
// Read methods (via JSON-RPC)
function HederaAddressToPeer(address) view returns (
    bool available,
    string peerID,
    uint64 stdOutTopic,
    uint64 stdInTopic,
    uint64 stdErrTopic
)

function GetPeerArraySize() view returns (uint256)

function PeerList(uint256 index) view returns (address)

function AllowedServices(uint8 serviceId) view returns (bool)

// Write methods (via Hedera SDK)
function PutPeerAvailableSelf(
    uint64 stdOutTopic,
    uint64 stdInTopic,
    uint64 stdErrTopic,
    string peerID,
    uint8[] serviceIDs,
    uint8[] prices
)
```

---

## Appendix B: Environment Variables Reference

| Variable            | Description             | Example                         |
| ------------------- | ----------------------- | ------------------------------- |
| `HEDERA_NETWORK`    | Network name            | `testnet`                       |
| `HEDERA_ACCOUNT_ID` | Operator account        | `0.0.12345`                     |
| `PRIVATE_KEY`       | Hex-encoded private key | `4c0883a69...`                  |
| `ETH_RPC_URL`       | JSON-RPC endpoint       | `https://testnet.hashio.io/api` |
| `CONTRACT_ADDRESS`  | Registry contract       | `0x1234...abcd`                 |

Note: These are loaded by the application, not the SDK.

---

## Appendix C: Message Format Reference

**Heartbeat Message:**

```json
{
  "messageType": "heartbeat",
  "timestamp": "2026-01-23T10:30:00Z",
  "location": {
    "lat": 37.7749,
    "lon": -122.4194,
    "alt": 10.0,
    "gpsfix": "3D"
  },
  "natDeviceType": "symmetric",
  "natReachability": true,
  "connectedPeers": ["16Uiu2HA", "16Uiu2HB"],
  "role": "seller",
  "version": "1.0.0"
}
```

**Signed Message Format:**

```
+--------------------+--------------------+
|  Signature (65B)   |    Payload (var)   |
+--------------------+--------------------+
|  R (32) S (32) V(1)|    JSON bytes      |
+--------------------+--------------------+
```
