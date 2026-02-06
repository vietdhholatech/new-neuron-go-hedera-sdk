# Neuron Go Hedera SDK - Project Overview

## Purpose
A comprehensive Go SDK for building decentralized agent identities on the Hedera network with multi-ecosystem interoperability (Hedera, Ethereum, libp2p).

## Tech Stack
- **Go 1.24+** (uses `clear()` builtin for secure memory zeroing)
- **Hiero SDK v2.74.0+** - Hedera network integration
- **go-ethereum** - Ethereum address derivation and EIP-55 checksums
- **libp2p** - P2P networking and PeerID derivation
- **Gin** - Web framework for REST API
- **Swagger (swaggo)** - API documentation
- **secp256k1 (ECDSA)** - Cryptographic signatures
- **Argon2id + AES-256-GCM** - Password-based key encryption

## Core Modules

### keylib (87.1% test coverage)
Cryptographic key management with type-safe ECDSA secp256k1 keys:
- `NeuronPrivateKey` - 32-byte private key with signing
- `NeuronPublicKey` - 33-byte compressed public key
- `EVMAddress` - 20-byte Ethereum address (EIP-55 checksum)
- `PeerID` - libp2p peer identifier
- `Signature` - 65-byte ECDSA signature (R‖S‖V)
- `EncryptedPrivateKey` - Password-encrypted key storage (Argon2id + AES-256-GCM)
- `MultisigKey` - M-of-N threshold multi-signature key with sorted deterministic ordering and protocol identifier (secp256k1-aggregated, hedera-threshold, frost, bls)
- `Mnemonic` - BIP39 mnemonic generation and BIP32/BIP44 key derivation

### account (91.6% test coverage)
Agent identity and communication endpoint descriptors:
- `NeuronAccount` - Agent identity with communication endpoints
- `AccountType` - Parent (with DID), Child (with parent reference), or Shared (with MultisigKey)
- `CommAddress` - Technology-agnostic communication address (kind:locator format)
- `LedgerAttachment` - State machine for on-ledger account attachment (Detached → Attached → Verified) with enforced state transition guards
- `ValidateCurrencySymbol` - Conditional validation: currency symbol required when LedgerAttachment is present
- `LedgerVerifier` - Interface for verifying account-ledger relationships
- Backend registry system for Hedera Consensus Service, Kafka, and custom backends
- Builder pattern with error accumulation and type-specific validation

### account/didkey
W3C-compliant DID:Key support for secp256k1 keys.

### api
REST API exposing keylib functionality via Gin:
- HTTP handlers in `api/handlers/`
- DTOs in `api/dto/`
- Swagger docs in `api/docs/`

## Project Structure
```
.
├── keylib/           # Cryptographic key management (secp256k1)
├── account/          # Agent identity descriptors
│   └── didkey/       # DID:Key support (W3C compliant)
├── api/              # Gin REST API
│   ├── handlers/     # HTTP handlers
│   ├── dto/          # Data transfer objects
│   └── docs/         # Swagger documentation
├── cmd/
│   └── keylib-api/   # API server entrypoint
└── docs/
    └── implementation/ # Technical documentation (NeuronAccount.md, keylib.md)
```

## Key Design Principles
1. **Type safety** - Functions accept and return types, not strings
2. **Validation at boundaries** - All Parse* functions validate immediately
3. **No panics** - All fallible operations return (T, error)
4. **Zero-value safety** - Zero values are invalid; methods return errors
5. **Constant-time operations** - All key comparisons prevent timing attacks
6. **Memory safety** - Zeroize() clears sensitive data from memory
7. **Immutability** - All types are immutable after construction
