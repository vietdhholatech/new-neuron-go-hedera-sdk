# External Backend Registration & Account Building Examples

This guide demonstrates how external developers can add new messaging backends and build NeuronAccounts using the SDK's extensible architecture.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Creating a Custom Backend](#creating-a-custom-backend)
3. [Registering the Backend](#registering-the-backend)
4. [Building Accounts with Custom Backends](#building-accounts-with-custom-backends)
5. [Complete Examples](#complete-examples)

---

## Architecture Overview

The SDK uses a **plugin architecture** where backends register themselves at import time:

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application                          │
├─────────────────────────────────────────────────────────────┤
│  import (                                                    │
│      "github.com/aspect-build/neuron-go-hedera-sdk/account" │
│      _ "github.com/mycompany/neuron-pulsar"  ← blank import │
│  )                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend Registry                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │  Hedera  │ │  Kafka   │ │  Custom  │ │  Pulsar  │       │
│  │ (built-in)│ │(built-in)│ │(built-in)│ │(external)│       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────────────────────────────────────────┘
```

**Key Point:** No SDK modifications needed to add new backends.

---

## Creating a Custom Backend

### Step 1: Implement the Backend Interface

Create a new Go package for your backend:

```go
// File: github.com/mycompany/neuron-pulsar/pulsar.go
package pulsar

import (
    "fmt"
    "strings"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
)

// PulsarTopicKind is the unique identifier for this backend.
// Convention: use "technology-topic" format.
const PulsarTopicKind = "pulsar-topic"

// pulsarBackend implements account.Backend interface.
type pulsarBackend struct{}

// Kind returns the unique identifier for this backend.
// This is used as the "kind" field in CommAddress.
func (p *pulsarBackend) Kind() string {
    return PulsarTopicKind
}

// ValidateLocator checks if a locator string is valid for Pulsar.
// Pulsar topics have format: persistent://tenant/namespace/topic
// or: non-persistent://tenant/namespace/topic
func (p *pulsarBackend) ValidateLocator(locator string) error {
    if locator == "" {
        return fmt.Errorf("empty Pulsar topic locator")
    }

    // Check prefix
    if !strings.HasPrefix(locator, "persistent://") &&
        !strings.HasPrefix(locator, "non-persistent://") {
        return fmt.Errorf("Pulsar topic must start with 'persistent://' or 'non-persistent://', got: %s", locator)
    }

    // Extract and validate path components
    var path string
    if strings.HasPrefix(locator, "persistent://") {
        path = strings.TrimPrefix(locator, "persistent://")
    } else {
        path = strings.TrimPrefix(locator, "non-persistent://")
    }

    parts := strings.Split(path, "/")
    if len(parts) != 3 {
        return fmt.Errorf("Pulsar topic format: [non-]persistent://tenant/namespace/topic, got: %s", locator)
    }

    // Validate each component is non-empty
    for i, part := range parts {
        if part == "" {
            return fmt.Errorf("Pulsar topic component %d is empty", i)
        }
    }

    return nil
}

// ParseLocator validates and normalizes the locator.
// For Pulsar, we just validate and return as-is (no normalization needed).
func (p *pulsarBackend) ParseLocator(locator string) (string, error) {
    if err := p.ValidateLocator(locator); err != nil {
        return "", err
    }
    return locator, nil
}

// Metadata returns descriptive information about this backend.
func (p *pulsarBackend) Metadata() account.BackendMetadata {
    return account.BackendMetadata{
        DisplayName:    "Apache Pulsar",
        Description:    "Cloud-native distributed messaging and streaming platform",
        LocatorFormat:  "persistent://tenant/namespace/topic",
        LocatorExample: "persistent://public/default/my-topic",
        RequiresConfig: true, // Requires Pulsar broker connection
        Properties: map[string]string{
            "persistence":  "persistent or non-persistent",
            "multiTenancy": "supports tenant/namespace isolation",
            "geoReplication": "supports cross-datacenter replication",
        },
    }
}

// Technology returns the TopicTechnology for this backend.
// Use TopicTechnologyCustom for external backends, or define your own
// if you need specific handling.
func (p *pulsarBackend) Technology() account.TopicTechnology {
    return account.TopicTechnologyCustom
}
```

### Step 2: Add Registration and Convenience Functions

```go
// File: github.com/mycompany/neuron-pulsar/pulsar.go (continued)

// init registers the Pulsar backend at package initialization.
// This runs automatically when the package is imported.
func init() {
    account.RegisterBackend(&pulsarBackend{})
}

// NewPulsarTopicAddress creates a CommAddress for a Pulsar topic.
// This is a convenience function for users of your backend.
func NewPulsarTopicAddress(locator string) (account.CommAddress, error) {
    return account.NewCommAddress(PulsarTopicKind, locator)
}

// MustParsePulsarTopic creates a CommAddress and panics on error.
// Use only in tests or for known-valid locators.
func MustParsePulsarTopic(locator string) account.CommAddress {
    addr, err := NewPulsarTopicAddress(locator)
    if err != nil {
        panic(err)
    }
    return addr
}

// ValidatePulsarLocator validates a Pulsar topic locator.
func ValidatePulsarLocator(locator string) error {
    backend, ok := account.GetBackend(PulsarTopicKind)
    if !ok {
        return fmt.Errorf("pulsar backend not registered")
    }
    return backend.ValidateLocator(locator)
}
```

---

## Registering the Backend

### Method 1: Blank Import (Recommended)

The simplest way - just import the package:

```go
package main

import (
    "fmt"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

    // Blank import activates the Pulsar backend
    _ "github.com/mycompany/neuron-pulsar"
)

func main() {
    // Pulsar is now available!
    addr, err := account.NewCommAddress("pulsar-topic", "persistent://public/default/my-topic")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created Pulsar address: %s\n", addr)
}
```

### Method 2: Explicit Import (When Using Convenience Functions)

If you want to use the backend's convenience functions:

```go
package main

import (
    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

    "github.com/mycompany/neuron-pulsar" // Regular import
)

func main() {
    // Use the convenience function
    addr, err := pulsar.NewPulsarTopicAddress("persistent://public/default/my-topic")
    if err != nil {
        panic(err)
    }

    // Or use the constant
    addr2, _ := account.NewCommAddress(pulsar.PulsarTopicKind, "persistent://tenant/ns/topic")
}
```

---

## Building Accounts with Custom Backends

### Example 1: Parent Account with Pulsar Endpoints

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

    _ "github.com/mycompany/neuron-pulsar"
)

func main() {
    // Step 1: Generate cryptographic identity
    privKey, err := keylib.GeneratePrivateKey()
    if err != nil {
        log.Fatal(err)
    }
    pubKey := privKey.PublicKey()

    // Step 2: Create DID from public key
    did := didkey.FromPublicKey(pubKey)

    // Step 3: Build account with Pulsar endpoints
    acct, err := account.NewParentAccountBuilder(pubKey, did).
        // Use generic WithStdInBackend for custom backends
        WithStdInBackend("pulsar-topic", "persistent://myorg/agents/agent-001-stdin").
        WithStdOutBackend("pulsar-topic", "persistent://myorg/agents/agent-001-stdout").
        WithStdErrBackend("pulsar-topic", "persistent://myorg/agents/agent-001-stderr").
        // Add P2P reachability
        WithReachableAddr("/ip4/203.0.113.50/tcp/4001/p2p/" + pubKey.PeerIDString()).
        Build()

    if err != nil {
        log.Fatal(err)
    }

    // Step 4: Inspect the account
    fmt.Printf("Account Type: %s\n", acct.AccountType())
    fmt.Printf("DID: %s\n", acct.DID().String())
    fmt.Printf("PeerID: %s\n", acct.PeerID())
    fmt.Printf("StdIn: %s\n", acct.StdIn())
    fmt.Printf("StdOut: %s\n", acct.StdOut())
    fmt.Printf("StdErr: %s\n", acct.StdErr())
}
```

**Output:**
```
Account Type: parent
DID: did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK
PeerID: 12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN
StdIn: pulsar-topic:persistent://myorg/agents/agent-001-stdin
StdOut: pulsar-topic:persistent://myorg/agents/agent-001-stdout
StdErr: pulsar-topic:persistent://myorg/agents/agent-001-stderr
```

### Example 2: Child Account with Mixed Backends

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

    _ "github.com/mycompany/neuron-pulsar"
)

func main() {
    // Parent's public key (would come from parent account)
    parentPriv, _ := keylib.GeneratePrivateKey()
    parentPub := parentPriv.PublicKey()

    // Child's own key pair
    childPriv, _ := keylib.GeneratePrivateKey()
    childPub := childPriv.PublicKey()

    // Build child account with MIXED backends:
    // - StdIn: Pulsar (for high-throughput commands)
    // - StdOut: Kafka (for event streaming)
    // - StdErr: Hedera (for immutable error logging)
    acct, err := account.NewChildAccountBuilder(childPub, parentPub).
        WithStdInBackend("pulsar-topic", "persistent://myorg/workers/worker-001-stdin").
        WithStdOutKafka("worker-001-events").
        WithStdErrHedera("0.0.98765").
        Build()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Child Account PeerID: %s\n", acct.PeerID())
    fmt.Printf("Parent PubKey: %s\n", acct.ParentPubKey().String()[:32]+"...")
    fmt.Printf("StdIn (Pulsar): %s\n", acct.StdIn())
    fmt.Printf("StdOut (Kafka): %s\n", acct.StdOut())
    fmt.Printf("StdErr (Hedera): %s\n", acct.StdErr())
}
```

### Example 3: Using WithBackendTopics for Uniform Backend

```go
// When all endpoints use the same backend technology
acct, err := account.NewParentAccountBuilder(pubKey, did).
    WithBackendTopics(
        "pulsar-topic",                              // backend kind
        "persistent://org/ns/agent-stdin",           // stdIn
        "persistent://org/ns/agent-stdout",          // stdOut
        "persistent://org/ns/agent-stderr",          // stdErr
    ).
    Build()
```

---

## Complete Examples

### Example: NATS JetStream Backend

```go
// File: github.com/mycompany/neuron-nats/nats.go
package nats

import (
    "fmt"
    "regexp"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
)

const NatsStreamKind = "nats-stream"

// NATS subject pattern: alphanumeric with dots for hierarchy
var natsSubjectPattern = regexp.MustCompile(`^[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*$`)

type natsBackend struct{}

func init() {
    account.RegisterBackend(&natsBackend{})
}

func (n *natsBackend) Kind() string {
    return NatsStreamKind
}

func (n *natsBackend) ValidateLocator(locator string) error {
    if locator == "" {
        return fmt.Errorf("empty NATS subject")
    }
    if len(locator) > 256 {
        return fmt.Errorf("NATS subject too long (max 256): %d", len(locator))
    }
    if !natsSubjectPattern.MatchString(locator) {
        return fmt.Errorf("invalid NATS subject format: %s", locator)
    }
    return nil
}

func (n *natsBackend) ParseLocator(locator string) (string, error) {
    if err := n.ValidateLocator(locator); err != nil {
        return "", err
    }
    return locator, nil
}

func (n *natsBackend) Metadata() account.BackendMetadata {
    return account.BackendMetadata{
        DisplayName:    "NATS JetStream",
        Description:    "Lightweight, high-performance messaging for cloud-native applications",
        LocatorFormat:  "subject.hierarchy.path",
        LocatorExample: "agents.worker.commands",
        RequiresConfig: true,
        Properties: map[string]string{
            "persistence": "JetStream for persistence",
            "wildcards":   "supports * and > wildcards",
        },
    }
}

func (n *natsBackend) Technology() account.TopicTechnology {
    return account.TopicTechnologyCustom
}

// Convenience functions
func NewNatsStreamAddress(subject string) (account.CommAddress, error) {
    return account.NewCommAddress(NatsStreamKind, subject)
}
```

### Example: Redis Streams Backend

```go
// File: github.com/mycompany/neuron-redis/redis.go
package redis

import (
    "fmt"
    "regexp"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
)

const RedisStreamKind = "redis-stream"

// Redis key pattern: alphanumeric with colons for namespacing
var redisKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9:_-]+$`)

type redisBackend struct{}

func init() {
    account.RegisterBackend(&redisBackend{})
}

func (r *redisBackend) Kind() string {
    return RedisStreamKind
}

func (r *redisBackend) ValidateLocator(locator string) error {
    if locator == "" {
        return fmt.Errorf("empty Redis stream key")
    }
    if len(locator) > 512 {
        return fmt.Errorf("Redis stream key too long (max 512): %d", len(locator))
    }
    if !redisKeyPattern.MatchString(locator) {
        return fmt.Errorf("invalid Redis stream key format: %s", locator)
    }
    return nil
}

func (r *redisBackend) ParseLocator(locator string) (string, error) {
    if err := r.ValidateLocator(locator); err != nil {
        return "", err
    }
    return locator, nil
}

func (r *redisBackend) Metadata() account.BackendMetadata {
    return account.BackendMetadata{
        DisplayName:    "Redis Streams",
        Description:    "In-memory data structure store with stream support",
        LocatorFormat:  "namespace:stream:name",
        LocatorExample: "agents:worker-001:commands",
        RequiresConfig: true,
        Properties: map[string]string{
            "persistence":  "RDB/AOF persistence options",
            "consumerGroups": "supports consumer groups",
        },
    }
}

func (r *redisBackend) Technology() account.TopicTechnology {
    return account.TopicTechnologyCustom
}

// Convenience functions
func NewRedisStreamAddress(key string) (account.CommAddress, error) {
    return account.NewCommAddress(RedisStreamKind, key)
}
```

### Example: Using Multiple External Backends Together

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

    // Import all backends you need
    _ "github.com/mycompany/neuron-pulsar"
    _ "github.com/mycompany/neuron-nats"
    _ "github.com/mycompany/neuron-redis"
)

func main() {
    // List all available backends
    fmt.Println("Available Backends:")
    for kind, meta := range account.ListBackendsWithMetadata() {
        fmt.Printf("  - %s: %s\n", kind, meta.DisplayName)
    }
    fmt.Println()

    // Generate identity
    privKey, _ := keylib.GeneratePrivateKey()
    pubKey := privKey.PublicKey()
    did := didkey.FromPublicKey(pubKey)

    // Build account with any available backend
    acct, err := account.NewParentAccountBuilder(pubKey, did).
        WithStdInBackend("nats-stream", "agents.orchestrator.commands").
        WithStdOutBackend("redis-stream", "agents:orchestrator:events").
        WithStdErrBackend("pulsar-topic", "persistent://logs/errors/orchestrator").
        Build()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Account created with multi-backend endpoints:\n")
    fmt.Printf("  StdIn (NATS): %s\n", acct.StdIn())
    fmt.Printf("  StdOut (Redis): %s\n", acct.StdOut())
    fmt.Printf("  StdErr (Pulsar): %s\n", acct.StdErr())
}
```

**Output:**
```
Available Backends:
  - hedera-topic: Hedera Consensus Service
  - kafka-topic: Apache Kafka
  - custom: Custom Backend
  - pulsar-topic: Apache Pulsar
  - nats-stream: NATS JetStream
  - redis-stream: Redis Streams

Account created with multi-backend endpoints:
  StdIn (NATS): nats-stream:agents.orchestrator.commands
  StdOut (Redis): redis-stream:agents:orchestrator:events
  StdErr (Pulsar): pulsar-topic:persistent://logs/errors/orchestrator
```

---

## Testing Your Backend

```go
// File: github.com/mycompany/neuron-pulsar/pulsar_test.go
package pulsar

import (
    "testing"

    "github.com/aspect-build/neuron-go-hedera-sdk/account"
)

func TestPulsarBackendRegistration(t *testing.T) {
    // Verify backend is registered (init() ran)
    if !account.IsRegisteredBackend(PulsarTopicKind) {
        t.Fatal("Pulsar backend not registered")
    }
}

func TestPulsarBackendKind(t *testing.T) {
    backend, ok := account.GetBackend(PulsarTopicKind)
    if !ok {
        t.Fatal("Could not get Pulsar backend")
    }
    if backend.Kind() != PulsarTopicKind {
        t.Errorf("Kind() = %s, want %s", backend.Kind(), PulsarTopicKind)
    }
}

func TestValidateLocator_Valid(t *testing.T) {
    validLocators := []string{
        "persistent://public/default/my-topic",
        "non-persistent://tenant/namespace/topic",
        "persistent://my-org/production/events",
    }

    for _, locator := range validLocators {
        t.Run(locator, func(t *testing.T) {
            _, err := NewPulsarTopicAddress(locator)
            if err != nil {
                t.Errorf("NewPulsarTopicAddress(%s) returned error: %v", locator, err)
            }
        })
    }
}

func TestValidateLocator_Invalid(t *testing.T) {
    invalidLocators := []struct {
        locator string
        desc    string
    }{
        {"", "empty"},
        {"my-topic", "missing prefix"},
        {"persistent://", "missing path"},
        {"persistent://tenant", "missing namespace and topic"},
        {"persistent://tenant/namespace", "missing topic"},
        {"http://example.com/topic", "wrong prefix"},
    }

    for _, tc := range invalidLocators {
        t.Run(tc.desc, func(t *testing.T) {
            _, err := NewPulsarTopicAddress(tc.locator)
            if err == nil {
                t.Errorf("NewPulsarTopicAddress(%s) should have returned error", tc.locator)
            }
        })
    }
}

func TestCommAddressKind_IsValid(t *testing.T) {
    kind := account.CommAddressKind(PulsarTopicKind)
    if !kind.IsValid() {
        t.Error("Pulsar kind should be valid after registration")
    }
}

func TestAccountBuilderIntegration(t *testing.T) {
    // This tests that the backend works with the account builder
    pubKey := generateTestKey(t)
    did := createTestDID(t, pubKey)

    acct, err := account.NewParentAccountBuilder(pubKey, did).
        WithStdInBackend(PulsarTopicKind, "persistent://test/ns/stdin").
        Build()

    if err != nil {
        t.Fatalf("Build() returned error: %v", err)
    }

    if acct.StdIn().Kind() != account.CommAddressKind(PulsarTopicKind) {
        t.Errorf("StdIn kind = %s, want %s", acct.StdIn().Kind(), PulsarTopicKind)
    }
}
```

---

## Summary

| Task | Code |
|------|------|
| Create backend | Implement `account.Backend` interface (5 methods) |
| Register backend | Add `func init() { account.RegisterBackend(&myBackend{}) }` |
| Use backend | Blank import: `_ "pkg/path"` |
| Create address | `account.NewCommAddress("my-kind", "locator")` |
| Build account | `builder.WithStdInBackend("my-kind", "locator")` |
| List backends | `account.ListBackends()` or `account.ListBackendsWithMetadata()` |
| Check registration | `account.IsRegisteredBackend("my-kind")` |

**The key benefit:** External developers can add new backends WITHOUT modifying any SDK core files.
