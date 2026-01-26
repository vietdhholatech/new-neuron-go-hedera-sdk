# NeuronAccount Specification

## 1. Intent & Philosophy

A **NeuronAccount** is the portable, blockchain-agnostic representation of an agent’s **identity and endpoints**.

- It is **identified by a Neuron Key** (from the Key Library).
- It does **not perform communication** or message passing.
- It **describes where communication happens** and **how the agent can be reached**.
- It is **anchored to one or more blockchains**, but the abstraction itself is not blockchain-specific.

This separation enables:

- reuse across Hedera, Ethereum, or future chains
- interchangeable communication backends
- strong type safety with no ambiguous strings

---

## 2. Identity

A **NeuronAccount** is identified by a **Neuron public key**.

### 2.1 Primary Identifier

- `NeuronPublicKey` (required)
- `DID` (required for Parent accounts)

### 2.2 Derived Identifiers

Derived identifiers are computable from the public key and may be cached:

- `PeerID` (Libp2p)
- `EVMAddress` (Ethereum-compatible, if ECDSA / Secp256k1)

The Neuron key is the **root of identity**.  
All other identifiers are **derived**, never authoritative on their own.

---

### 2.3 Account Hierarchy

A **NeuronAccount** MUST be categorized as either a **Parent** or a **Child**.

- **Strict Two-Level Hierarchy**: Nesting is limited to one level deep.
- **Parent Account**:
  - HAS NO parent account.
  - MAY have one or more Child accounts.
  - MUST include **DID (Decentralized Identifier)** information.
- **Child Account**:
  - HAS NO children (no grandchildren in the hierarchy).
  - HAS EXACTLY ONE parent account.

This structure ensures clear ownership and administrative boundaries while maintaining system simplicity.

---

## 3. Public Communication Addresses

A **NeuronAccount** exposes three public communication addresses:

- `stdIn` – where others send messages to the agent
- `stdOut` – where the agent publishes outputs / heartbeats
- `stdErr` – where the agent publishes errors / diagnostics

These addresses are:

- public
- technology-agnostic
- self-describing

The account **does not send, receive, or subscribe**.  
It only declares **where these channels live**.

---

### 3.1 CommAddress Abstraction

All public communication endpoints are represented by a `CommAddress`.

A `CommAddress`:

- declares which backend technology it uses
- declares how the endpoint is located
- validates itself on construction

Examples of backends:

- Hedera Consensus Service topics
- Kafka topics
- Blockchain log streams
- Other pubsub systems

#### Minimal Properties

- `Kind` – e.g. `hedera-topic`, `kafka-topic`
- `Locator` – backend-specific identifier
- `Validate()` – backend-specific validation rules
- Rading and Writing interfaces that are blockchain-agnostic

---

### 3.2 Example: Hedera Topics

stdIn = HederaTopic(0.0.12345)
stdOut = HederaTopic(0.0.12346)
stdErr = HederaTopic(0.0.12347)

Each topic is a **public rendezvous point**.  
Message semantics are out of scope for `NeuronAccount`.

---

## 4. Reachable Connection Addresses (Multiaddr)

A **NeuronAccount** exposes reachable connection addresses using **libp2p Multiaddr**.

They are the agent’s canonical **“find / dial me” endpoints**.

They are used for:

- direct P2P dialing
- NAT traversal
- relay-based connectivity

---

#### Rationale

A `NeuronAccount` must expose **fully-qualified peer endpoints**, not anonymous host/port pairs.  
The peer identity is **not optional**.

---

### 4.2 PeerID Consistency Rule

- Let `PeerID_expected` be the PeerID derived from the account’s `NeuronPublicKey`
- Every reachable multiaddr MUST:
  - contain exactly one `/p2p/<PeerID>`
  - where `<PeerID> == PeerID_expected`

Any mismatch is a **hard validation error**.

---

### 4.3 ReachableAddrs Field

ReachableAddrs: []Multiaddr

Each entry MUST:

- parse as a valid libp2p multiaddr
- include `/p2p/<PeerID>`
- match the derived PeerID
- use supported transports (TCP, QUIC, etc.)

---

### 4.4 Direct Dial Examples (All Valid)

**IPv4 + TCP**

/ip4/203.0.113.10/tcp/4001/p2p/12D3KooWAbCdEf…

**DNS + TCP**

/dns4/node.example.com/tcp/4001/p2p/12D3KooWAbCdEf…

**QUIC over UDP**

/ip4/203.0.113.10/udp/4001/quic-v1/p2p/12D3KooWAbCdEf…

---

### 4.5 Circuit Relay Addresses

Circuit relay addresses allow reaching an agent **via a relay peer**.

**Minimal Circuit Relay**

/p2p/12D3KooWRelayNode/p2p-circuit/p2p/12D3KooWTargetNode

**Relay with Transport**

/ip4/198.51.100.7/tcp/4001/p2p/12D3KooWRelayNode/p2p-circuit/p2p/12D3KooWTargetNode

**DNS-based Relay**

/dns4/relay.example.com/tcp/4001/p2p/12D3KooWRelayNode/p2p-circuit/p2p/12D3KooWTargetNode

In all cases, the **final `/p2p/<TargetPeerID>` MUST match the NeuronAccount identity**.

---

### 4.6 Canonicalization & Validation

When constructing a `NeuronAccount`:

1. Parse each address as a multiaddr
2. Reject if `/p2p/<PeerID>` is missing
3. Derive `PeerID_expected` from `NeuronPublicKey`
4. Reject if any address uses a different PeerID
5. Store the canonical multiaddr representation

Strings are allowed **only at serialization boundaries**.

---

## 5. Blockchain Anchoring & Yellow Pages

A `NeuronAccount` must declare **where it is anchored on-chain**.

This answers:

- which blockchain recognizes the identity
- where the account can be discovered

---

### 5.1 Chain Anchors

A `NeuronAccount` may have one or more chain anchors.

Each anchor specifies:

- blockchain + network
- on-chain account reference (if any)
- registry (“yellow pages”) reference (if any)

**Examples**

Chain: hedera:testnet
AccountRef: 0.0.98765
Registry: Contract(0.0.55555, 0xAbC123…)

Chain: ethereum:mainnet
AccountRef: 0xAbC123…
Registry: Contract(0xDef456…)

---

### 5.2 Yellow Pages Registry

A yellow pages registry maps:

NeuronPublicKey -> NeuronAccount metadata

Typical metadata:

- stdIn / stdOut / stdErr addresses
- reachable multiaddrs
- service capabilities

The NeuronAccount library:

- defines how registries are referenced
- does **not** perform chain writes

---

## 6. Responsibilities

### NeuronAccount IS responsible for

- identity via Neuron keys
- declaring public communication addresses
- declaring reachable multiaddrs
- declaring blockchain anchors and registry locations

### NeuronAccount IS NOT responsible for

- sending or receiving messages
- subscribing to topics
- dialing peers
- creating blockchain accounts
- updating registries

---

## 7. Required Interfaces (Blockchain-Agnostic)

The library **MUST define interfaces** for interacting with blockchains,  
but **MUST NOT implement them here**.

These interfaces allow other packages to provide concrete implementations
(e.g. Hedera, Ethereum) without coupling the core model to any chain.

### 7.1 Account Creation Interface

Responsible for creating or associating an on-chain account identity.

Examples of responsibilities:

- creating a blockchain account
- associating a public key with an existing account
- returning a `ChainAnchor`

No blockchain-specific logic is implemented in this library.

---

### 7.2 Account Verification Interface

Responsible for verifying that:

- a `NeuronAccount` exists on a given blockchain
- the declared public key is correctly associated
- the declared registry entry exists (if applicable)

This interface performs **read-only verification**, not mutation.

---

### 7.3 Endpoint Verification Interface

Responsible for verifying that declared endpoints exist and are well-formed,
for example:

- Hedera topics exist and are readable
- registry entries are present
- multiaddrs are syntactically valid

This library defines **the interface only**.  
Concrete implementations live elsewhere.

---

### 7.4 Summary of Interfaces

Typical interfaces include (names illustrative):

- `AccountCreator`
- `AccountVerifier`
- `EndpointVerifier`
- `RegistryResolver`

All are **pure interfaces**.  
No blockchain SDKs are linked from this package.

---

## 8. Concrete Examples

### 8.1 Example: Parent NeuronAccount

NeuronAccount:
Type: Parent
PublicKey: NeuronPublicKey(secp256k1)
DID: did:neuron:z6MkpTHR8VNsBxYbgRGr..

stdIn: HederaTopic(0.0.1001)
stdOut: HederaTopic(0.0.1002)
stdErr: HederaTopic(0.0.1003)

ReachableAddrs:

- /ip4/203.0.113.10/tcp/4001/p2p/12D3KooWTarget
- /p2p/12D3KooWRelay/p2p-circuit/p2p/12D3KooWTarget

Anchors:

- Chain: hedera:testnet
  AccountRef: 0.0.9001
  Registry: Contract(0.0.7777)

---

### 8.2 Example: Child NeuronAccount

NeuronAccount:
Type: Child
Parent: NeuronPublicKey(ParentKey)
PublicKey: NeuronPublicKey(ChildKey)

stdIn: HederaTopic(0.0.2001)
stdOut: HederaTopic(0.0.2002)
stdErr: HederaTopic(0.0.2003)

ReachableAddrs:

- /ip4/203.0.113.11/tcp/4001/p2p/12D3KooWChildTarget

---

## 9. Design Guarantees

- Single-root identity (Neuron key)
- Mandatory peer-qualified multiaddrs
- Blockchain independence
- Transport independence
- No ambiguous strings
- Clean separation of concerns
