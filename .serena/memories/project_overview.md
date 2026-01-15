# Neuron Go Hedera SDK - Project Overview

## Purpose

The Neuron Go Hedera SDK is a cryptographic key management library that provides type-safe, 
interoperable key handling across Hedera, Ethereum, and libp2p ecosystems.

## Module

```
github.com/aspect-build/neuron-go-hedera-sdk
```

## Tech Stack

- **Language**: Go 1.24.0
- **Cryptography**: 
  - secp256k1 (github.com/decred/dcrd/dcrec/secp256k1/v4)
  - Ethereum crypto (github.com/ethereum/go-ethereum)
  - Argon2id + AES-256-GCM for key encryption
  - BIP39/BIP32 mnemonic support (github.com/tyler-smith/go-bip39)
- **Hedera Integration**: Hiero SDK v2.74.0 (github.com/hiero-ledger/hiero-sdk-go/v2)
- **P2P**: libp2p (github.com/libp2p/go-libp2p)
- **HTTP API**: Gin framework (github.com/gin-gonic/gin)
- **API Documentation**: Swagger (swaggo/swag, swaggo/gin-swagger)

## Project Structure

```
.
├── keylib/                 # Core cryptographic library
│   ├── doc.go             # Package documentation
│   ├── private_key.go     # NeuronPrivateKey type
│   ├── public_key.go      # NeuronPublicKey type
│   ├── signature.go       # Signature type with recovery
│   ├── evm_address.go     # EVMAddress with EIP-55
│   ├── peer_id.go         # PeerID for libp2p
│   ├── factory.go         # Parse*/From*/Generate functions
│   ├── mnemonic.go        # BIP39/BIP32 support
│   ├── encrypted_key.go   # Scramble/Unscramble with Argon2id
│   ├── errors.go          # KeyError type hierarchy
│   ├── validation.go      # Input validation helpers
│   ├── constant_time.go   # Secure comparison utilities
│   └── *_test.go          # Test files (85.4% coverage)
├── api/                    # HTTP REST API
│   ├── server.go          # Gin server setup
│   ├── middleware.go      # HTTP middleware
│   ├── handlers/          # API endpoint handlers
│   ├── dto/               # Request/Response DTOs
│   └── docs/              # Swagger documentation
├── cmd/
│   └── keylib-api/        # API server entry point
│       └── main.go
└── go.mod, go.sum         # Go module files
```

## Core Types

| Type | Purpose |
|------|---------|
| `NeuronPrivateKey` | ECDSA secp256k1 private key (immutable) |
| `NeuronPublicKey` | ECDSA secp256k1 public key (immutable) |
| `EVMAddress` | 20-byte Ethereum address with EIP-55 |
| `PeerID` | libp2p peer identifier |
| `Signature` | 65-byte ECDSA signature (R||S||V) |
| `EncryptedPrivateKey` | Password-protected key storage |

## Key Design Principles

1. **Type safety**: Functions accept/return types, not strings
2. **Validation at boundaries**: All Parse* functions validate immediately
3. **No panics**: All fallible operations return (T, error)
4. **Zero-value safety**: Zero values are invalid; methods return errors
5. **Constant-time operations**: All key comparisons prevent timing attacks
6. **Immutability**: All types immutable after construction (except Zeroize)
7. **Thread-safe**: All operations safe for concurrent use (except Zeroize)

## Ed25519 Rejection

The library standardizes on ECDSA secp256k1 and explicitly rejects Ed25519 keys with:
- DER encoding pattern detection
- Raw bytes length check (64 vs 32)
- secp256k1 scalar validation
