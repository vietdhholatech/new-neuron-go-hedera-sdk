# Project Overview: neuron-go-hedera-sdk

## Purpose
This is a Go library (`keylib`) that provides type-safe cryptographic key management for the Neuron SDK. It replaces brittle string-based key handling with a robust, type-safe system enabling seamless interoperability between Hedera, Ethereum, and libp2p ecosystems.

## Module
- Module: `github.com/aspect-build/neuron-go-hedera-sdk`
- Go version: 1.24.0

## Core Components

### 1. keylib Package (`keylib/`)
The main library providing cryptographic key management:

**Core Types:**
- `NeuronPrivateKey` - ECDSA secp256k1 private key with signing capabilities
- `NeuronPublicKey` - ECDSA secp256k1 public key with verification
- `EVMAddress` - 20-byte Ethereum address with EIP-55 checksum
- `PeerID` - libp2p peer identifier for P2P networking
- `Signature` - 65-byte ECDSA signature (R‖S‖V)
- `EncryptedPrivateKey` - Password-protected private key storage

**Key Files:**
- `private_key.go` - NeuronPrivateKey type and methods
- `public_key.go` - NeuronPublicKey type and methods
- `factory.go` - Parse*/From*/Generate factory functions
- `mnemonic.go` - BIP39/BIP32 mnemonic support
- `encrypted_key.go` - Scramble/Unscramble with Argon2id+AES-GCM
- `signature.go` - Signature type and recovery
- `evm_address.go` - EVMAddress with EIP-55 checksum
- `peer_id.go` - PeerID wrapper for libp2p
- `errors.go` - KeyError type hierarchy
- `validation.go` - Input validation helpers
- `constant_time.go` - Secure comparison utilities
- `doc.go` - Package documentation

### 2. API Server (`api/` and `cmd/keylib-api/`)
HTTP REST API for testing keylib functionality:
- Swagger UI at `/swagger/index.html`
- Runs on port 8080
- Routes organized by functionality (keys, derive, signing, signatures, mnemonic, encryption, matching, identifiers, hedera, bytes)

### 3. API Handlers (`api/handlers/`)
Handler implementations for all API endpoints

### 4. DTOs (`api/dto/`)
Data Transfer Objects for API request/response

## Key Design Decisions

### ECDSA secp256k1 Only
- Standardizes on ECDSA secp256k1 for maximum interoperability
- Same key works for Hedera, Ethereum, and libp2p
- Ed25519 keys are explicitly rejected

### Type Safety
- Functions accept and return types, not strings
- All types are immutable after construction
- Zero-value structs are explicitly invalid (IsZero() returns true)

### Validation
- All input validation happens at Parse* boundaries
- Rich KeyError type with Op, Kind, Details, Err fields
- 10 distinct ErrorKind values

### Security
- Constant-time comparisons for all key operations
- Argon2id (64MB, time=3) + AES-256-GCM for encryption
- Zeroize() method for clearing sensitive memory

## Dependencies
Key external dependencies:
- `github.com/decred/dcrd/dcrec/secp256k1/v4` - secp256k1 curve
- `github.com/ethereum/go-ethereum` - Ethereum utilities
- `github.com/hiero-ledger/hiero-sdk-go/v2` - Hedera SDK
- `github.com/libp2p/go-libp2p` - libp2p for PeerID
- `github.com/tyler-smith/go-bip39` - BIP39 mnemonic
- `github.com/gin-gonic/gin` - HTTP framework for API
- `github.com/swaggo/swag` - Swagger documentation

## Test Coverage
Current coverage: 85.4%
Tests are co-located with source files (*_test.go)
