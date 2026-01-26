# Neuron Go Hedera SDK - Project Overview

## Purpose
A Go SDK for Neuron that provides type-safe cryptographic key management with seamless interoperability between the Hedera/Hiero, Ethereum, and libp2p ecosystems. The library implements a "Rosetta Stone" for keys, allowing conversions between different key formats and address types.

## Tech Stack
- **Go version**: 1.24.0
- **Module path**: `github.com/aspect-build/neuron-go-hedera-sdk`

### Key Dependencies
- `github.com/hiero-ledger/hiero-sdk-go/v2` (v2.74.0) - Hiero/Hedera SDK
- `github.com/decred/dcrd/dcrec/secp256k1/v4` - secp256k1 curve implementation
- `github.com/ethereum/go-ethereum` - Ethereum types (common.Address, crypto)
- `github.com/libp2p/go-libp2p` - libp2p peer identity
- `github.com/tyler-smith/go-bip39` - BIP39 mnemonic support
- `github.com/gin-gonic/gin` - HTTP web framework for API
- `github.com/swaggo/swag` - Swagger documentation generation
- `golang.org/x/crypto` - Argon2id KDF

## Project Structure
```
new-neuron-go-hedera-sdk/
├── keylib/                    # Core cryptographic key library
│   ├── private_key.go         # NeuronPrivateKey type
│   ├── public_key.go          # NeuronPublicKey type
│   ├── signature.go           # Signature type with recovery
│   ├── evm_address.go         # EVMAddress type (EIP-55)
│   ├── peer_id.go             # PeerID wrapper for libp2p
│   ├── factory.go             # Parse/From/Generate functions
│   ├── mnemonic.go            # BIP39/BIP32 mnemonic support
│   ├── encrypted_key.go       # Argon2id + AES-256-GCM encryption
│   ├── validation.go          # Input validation helpers
│   ├── constant_time.go       # Secure comparison utilities
│   ├── errors.go              # KeyError type hierarchy
│   ├── doc.go                 # Package documentation
│   └── *_test.go              # Test files
├── account/                   # NeuronAccount management
│   ├── account.go             # NeuronAccount struct
│   ├── account_type.go        # Account types (Parent/Child)
│   ├── did.go                 # DID (Decentralized Identifier)
│   ├── registry.go            # Account registry
│   ├── backend*.go            # Backend implementations (Hedera, Kafka, Custom)
│   ├── message.go             # Messaging primitives
│   ├── topic.go               # Topic management
│   ├── reachable.go           # Reachability addresses
│   ├── builder.go             # Account builder pattern
│   └── *_test.go              # Test files
├── api/                       # REST API server
│   ├── server.go              # Gin server setup
│   ├── middleware.go          # HTTP middleware
│   ├── handlers/              # API endpoint handlers
│   │   ├── keys.go            # Key generation/parsing
│   │   ├── signing.go         # Signing endpoints
│   │   ├── mnemonic.go        # Mnemonic endpoints
│   │   ├── encryption.go      # Encryption endpoints
│   │   ├── hedera.go          # Hedera conversion endpoints
│   │   └── ...
│   ├── dto/                   # Data transfer objects
│   └── docs/                  # Swagger documentation
├── cmd/
│   └── keylib-api/
│       └── main.go            # API server entry point
├── docs/                      # Documentation and specs
│   ├── Key Library High-Level Specification.md
│   ├── NeuronAccount Specification.md
│   └── ...
├── go.mod
├── go.sum
└── README.md
```

## Core Capabilities

### keylib Package
1. **Key Management**
   - Generate ECDSA secp256k1 keys
   - Parse keys from hex strings, bytes, Hedera SDK types
   - BIP39 mnemonic generation and restoration
   - BIP32 hierarchical derivation (default path: m/44'/60'/0'/0/0)

2. **Key Encryption**
   - Argon2id KDF (64MB memory, 3 iterations, 4 threads)
   - AES-256-GCM authenticated encryption
   - Scramble/Unscramble API

3. **Conversions**
   - NeuronKey ↔ Hedera SDK types
   - PublicKey → EVMAddress (EIP-55 checksum)
   - PublicKey → PeerID (libp2p)
   - NeuronKey → standard ECDSA types

4. **Signing**
   - SignMessage (Keccak256 + ECDSA)
   - SignDigest (pre-hashed data)
   - Signature recovery (ecrecover)
   - Verification

### account Package
- NeuronAccount identity management
- Parent/Child account types
- DID (Decentralized Identifier) support
- Multiple backend support (Hedera, Kafka, Custom)
- Account registry
- Communication addresses and topics

## Test Coverage
- keylib: ~85.4%
- Tests use table-driven pattern
- Integration tests with known vectors
