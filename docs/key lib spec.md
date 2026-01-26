# Feature Specification: Key Library

**Feature Branch**: `002-key-library`  
**Created**: 2026-01-23  
**Status**: Draft  
**Input**: User description: "Key Library High-Level Specification"

## Purpose

The Key Library is a core component of the Neuron SDK that provides type-safe, interoperable cryptographic key management. This library standardizes how keys are handled throughout the Neuron SDK ecosystem, ensuring that all keys conform to this specification and work seamlessly across multiple blockchain networks.

The library replaces brittle, string-based key handling with a robust, type-safe system. All keys used within the Neuron SDK must be created, managed, and converted using this library to ensure consistency, security, and interoperability. Keys that conform to this specification can be used throughout the Neuron SDK with confidence that they will work correctly for signing, verification, address derivation, and cross-ecosystem conversions.

## Type Hierarchy

**NeuronPrivateKey** is the top-level, primary type in the Key Library. All key operations and conversions originate from or relate to NeuronPrivateKey.

### Primary Type

- **NeuronPrivateKey**: The foundational type that represents a private cryptographic key using secp256k1 curve. This secp256k1 format is compatible with multiple blockchain networks and ecosystems, enabling a single key to work across different networks. This is the entry point for all key operations in the library.

### Direct Derivatives

From a NeuronPrivateKey, the following can be derived:

- **NeuronPublicKey**: The public key corresponding to the private key, derived deterministically from the NeuronPrivateKey
- **EVMAddress**: The EVM-compatible address derived from the NeuronPrivateKey's public key
- **PeerID**: The Libp2p peer identifier derived from the NeuronPrivateKey's public key

### Type Relationships

The library maintains the following relationships:

1. **NeuronPrivateKey → NeuronPublicKey**: A private key can derive its corresponding public key
2. **NeuronPrivateKey → EVMAddress**: A private key can derive its EVM-compatible address
3. **NeuronPrivateKey → PeerID**: A private key can derive its Libp2p PeerID
4. **NeuronPublicKey → EVMAddress**: A public key can derive its EVM-compatible address
5. **NeuronPublicKey → PeerID**: A public key can derive its Libp2p PeerID
6. **NeuronPrivateKey ↔ Blockchain Keys**: A NeuronPrivateKey can be converted to underlying blockchain SDK key types (e.g., blockchain-specific private keys) and vice versa
7. **NeuronPublicKey ↔ Blockchain Keys**: A NeuronPublicKey can be converted to underlying blockchain SDK key types (e.g., blockchain-specific public keys) and vice versa

All conversions and derivations maintain cryptographic relationships that can be verified through type-safe matching functions. The NeuronPrivateKey serves as the single source of truth for all derived types. The library provides bidirectional conversion interfaces to convert between Neuron keys and native blockchain SDK key types, enabling interoperability with blockchain-specific SDKs.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Type-Safe Key Operations (Priority: P1)

Developers need to work with cryptographic keys using strong types instead of error-prone string handling. When converting between different key formats (blockchain keys, EVM addresses, Libp2p PeerIDs), developers should use type-safe functions that prevent common mistakes like mixing up hex strings with base58 encoded values or forgetting to handle 0x prefixes.

**Why this priority**: This is the core value proposition - eliminating string-based key handling errors that cause runtime failures and security issues. Without type safety, the library cannot deliver its primary benefit.

**Independent Test**: Can be fully tested by creating a NeuronPrivateKey from a blockchain key, converting it to an EVM address and PeerID, then verifying the relationships using type-safe matching functions. All operations should use types, not strings, and invalid inputs should return descriptive errors.

**Acceptance Scenarios**:

1. **Given** a blockchain private key, **When** a developer elevates it to a NeuronPrivateKey, **Then** they can use type-safe methods to convert to EVM address and PeerID without string manipulation
2. **Given** an invalid hex string (e.g., contains non-hex characters), **When** parsing a key, **Then** the system returns a descriptive error explaining the exact issue (e.g., "invalid hex character 'g' at position 5")
3. **Given** a NeuronPublicKey, **When** converting to different formats (EVM address, PeerID), **Then** all conversions use type-safe functions that return typed values, not strings
4. **Given** a NeuronPrivateKey, **When** converting to an underlying blockchain key type, **Then** the converted key can be used with blockchain-specific SDKs, and converting back to NeuronPrivateKey produces an identical key

---

### User Story 2 - Cross-Ecosystem Key Conversions (Priority: P2)

Developers building applications that span multiple blockchain ecosystems need seamless conversion between Neuron keys, EVM addresses, and Libp2p PeerIDs. A single private key should control accounts across multiple networks, with deterministic conversions between all formats.

**Why this priority**: Enables interoperability between ecosystems, but depends on type-safe key operations (P1) being established first.

**Independent Test**: Can be fully tested by generating a NeuronPrivateKey, deriving its EVM address and PeerID, then verifying that the private key matches both the address and PeerID using type-safe matching functions. Conversions should be bidirectional where applicable.

**Acceptance Scenarios**:

1. **Given** a NeuronPrivateKey, **When** deriving its EVM address and PeerID, **Then** both are deterministically derived and the private key matches both using type-safe verification
2. **Given** an EVM address and PeerID, **When** both are derived from the same NeuronPublicKey, **Then** the system can verify they belong to the same key using type-safe matching
3. **Given** a NeuronPublicKey, **When** converting to EVM address format, **Then** the conversion is deterministic and can be verified to match the original key

---

### User Story 3 - Key Generation and Recovery (Priority: P3)

Developers need to generate new keys, restore keys from mnemonic phrases, and securely store/retrieve encrypted keys. This enables wallet functionality and key management workflows.

**Why this priority**: Essential for complete key management, but builds on the core type-safe operations (P1) and conversions (P2).

**Independent Test**: Can be fully tested by generating a new NeuronPrivateKey, creating a mnemonic phrase, restoring the key from the mnemonic, encrypting it with a password, then decrypting and verifying it matches the original. All operations should use type-safe interfaces.

**Acceptance Scenarios**:

1. **Given** a need for a new key, **When** generating a NeuronPrivateKey, **Then** the key is cryptographically secure, defaults to ECDSA secp256k1, and is immediately usable for all operations
2. **Given** a mnemonic phrase, **When** restoring a NeuronPrivateKey, **Then** the restored key is identical to the original key that generated the mnemonic
3. **Given** a NeuronPrivateKey and password, **When** encrypting and then decrypting the key, **Then** the decrypted key matches the original and can be used for signing operations

---

### Edge Cases

- What happens when a key string has invalid hex characters? (Return descriptive error with exact position of invalid character)
- How does the system handle keys with missing or incorrect 0x prefixes? (Normalize input by stripping/adding prefix as needed, document behavior)
- What if an Ed25519 key is provided when converting to Neuron key? (Always reject with clear error message indicating Ed25519 keys cannot be converted to secp256k1 Neuron keys, as they use different cryptographic curves)
- How are zero-value or invalid keys handled? (Return error or invalid key type that cannot be used for operations)
- What happens when converting between formats fails (e.g., invalid key format)? (Return typed error explaining the failure reason)
- How does the system handle mnemonic phrases with invalid checksums? (Reject with error indicating checksum validation failure)
- What if encryption/decryption fails due to wrong password? (Return error without revealing whether key or password was wrong, to prevent timing attacks)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The library MUST provide NeuronPrivateKey and NeuronPublicKey types that wrap underlying cryptographic keys and ensure validity. NeuronPublicKey MUST use compressed format (33 bytes) as the primary representation in the public API, with uncompressed format (65 bytes) handled internally when required for operations such as EVM address derivation
- **FR-002**: The library MUST support elevating primitive keys (blockchain keys, raw bytes, hex strings) into NeuronPrivateKey or NeuronPublicKey types
- **FR-003**: The library MUST default to ECDSA secp256k1 keys for primary operations to ensure compatibility with multiple blockchain networks
- **FR-004**: The library MUST detect and reject Ed25519 keys from external sources with clear error messages. Ed25519 keys cannot be converted to secp256k1 Neuron keys as they use different cryptographic curves. Neuron keys are always secp256k1, and Ed25519 keys are not supported for conversion
- **FR-005**: The library MUST provide deterministic conversion from NeuronPublicKey to EVM-compatible address
- **FR-006**: The library MUST provide deterministic conversion from NeuronPublicKey to Libp2p PeerID
- **FR-007**: The library MUST provide bidirectional conversion between EVM addresses and PeerIDs when the underlying NeuronPublicKey is known
- **FR-008**: The library MUST validate all string inputs (hex validity, length checks, prefix handling) and return structured error types with specific error kinds (e.g., InvalidFormat, InvalidLength, InvalidKey) and descriptive messages for invalid inputs
- **FR-009**: The library MUST use type-safe function signatures (accepting and returning types, not strings) for all key operations
- **FR-010**: The library MUST provide factory methods to create NeuronPrivateKey and NeuronPublicKey from various sources. Serialization formats MUST include hex strings (with or without 0x prefix) and raw bytes as required formats. Base58 encoding MAY be supported optionally for PeerID compatibility, but is not required for core key operations
- **FR-011**: The library MUST provide bidirectional conversion interfaces between Neuron keys and underlying blockchain SDK key types (e.g., NeuronPrivateKey ↔ blockchain private key, NeuronPublicKey ↔ blockchain public key), ensuring that converting a Neuron key to a blockchain key and back produces an identical key
- **FR-012**: The library MUST support generating new NeuronPrivateKeys with cryptographically secure randomness
- **FR-013**: The library MUST support restoring NeuronPrivateKeys from mnemonic phrases following BIP-39 standard for mnemonic generation and BIP-44 standard for derivation paths. The library MUST use a default derivation path of m/44'/60'/0'/0/0 (Ethereum standard) but MUST also support custom derivation paths specified by users
- **FR-014**: The library MUST provide a standard signing interface that NeuronPrivateKey implements, producing ECDSA signatures with recovery ID in R||S||V format (65 bytes total)
- **FR-015**: The library MUST provide methods to encrypt (scramble) and decrypt (unscramble) private keys using password-based encryption following the BIP-38 standard (scrypt key derivation function with AES-256-CBC encryption)
- **FR-016**: The library MUST provide type-safe matching functions to verify relationships: PrivateKey.Matches(PublicKey), PrivateKey.Matches(EVMAddress), PublicKey.Matches(PeerID)
- **FR-017**: The library MUST support signing messages and verifying signatures using Neuron keys. Message signing MUST use Keccak256 hashing before ECDSA signing. Signature verification MUST support both: (1) direct verification of a signature against a known public key (message + signature + public key → boolean), and (2) recovery of the public key from signatures (message + signature → public key) using the R||S||V format
- **FR-018**: The library MUST ensure that if a function returns a NeuronKey type, the key is guaranteed to be valid and usable
- **FR-019**: The library MUST use distinct types for different key representations (NeuronKey, EVMAddress, PeerID) to prevent mixing them up
- **FR-020**: The library MUST use clear API naming that indicates whether operations are computations (deriving values) or transformations (elevating types)
- **FR-021**: The library MUST provide methods to zeroize private key material from memory after use or when explicitly cleared to prevent key material from persisting in memory
- **FR-022**: Neuron key types (NeuronPrivateKey, NeuronPublicKey) MUST be immutable and thread-safe for concurrent read operations, allowing safe sharing across multiple threads without synchronization

### Key Entities *(include if feature involves data)*

- **NeuronPrivateKey**: An immutable, type-safe wrapper around a private cryptographic key using secp256k1 curve that ensures validity. Thread-safe for concurrent read operations. Can be converted bidirectionally to/from underlying blockchain SDK key types (Ed25519 keys from external sources are rejected, not converted)
- **NeuronPublicKey**: An immutable, type-safe wrapper around a public cryptographic key that ensures validity and enables conversions to various formats. Thread-safe for concurrent read operations. Public keys are represented in compressed format (33 bytes) in the public API, with uncompressed format (65 bytes) handled internally when required for operations like EVM address derivation. Can be converted bidirectionally to/from underlying blockchain SDK key types
- **EVMAddress**: A distinct type representing an EVM-compatible address (20-byte value) that cannot be confused with key types
- **PeerID**: A distinct type representing a Libp2p peer identifier that cannot be confused with other key types
- **Signature**: A type representing an ECDSA cryptographic signature in R||S||V format (65 bytes total: 32 bytes R, 32 bytes S, 1 byte recovery ID V) that can be verified against a public key and enables public key recovery

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can convert between Neuron keys, EVM addresses, and PeerIDs using only type-safe functions (no string manipulation required) in 100% of use cases
- **SC-002**: Invalid key inputs are rejected with descriptive error messages that identify the exact issue (invalid character position, wrong length, etc.) in 100% of validation failures
- **SC-003**: A single NeuronPrivateKey can control accounts across multiple blockchain networks with deterministic conversions verified through type-safe matching functions
- **SC-004**: Key generation, mnemonic restoration, and encryption/decryption operations complete successfully for 99.9% of valid inputs
- **SC-005**: Type-safe matching functions (MatchesPublicKey, MatchesEVMAddress, MatchesPeerID) correctly verify key relationships in 100% of valid cases
- **SC-006**: Developers can perform all key operations (elevation, conversion, signing, verification) without using string-based functions in 100% of standard workflows
- **SC-007**: The library prevents type confusion errors (mixing EVM addresses with keys, etc.) through distinct types that cannot be accidentally interchanged
- **SC-008**: Converting a NeuronPrivateKey to an underlying blockchain SDK key type and back produces an identical NeuronPrivateKey in 100% of valid conversions
- **SC-009**: Public keys can be recovered from signatures (message + signature → public key) in 100% of valid signatures using the R||S||V format
- **SC-010**: Signatures can be verified directly against a known public key (message + signature + public key → boolean) with 100% accuracy for valid signatures and correct rejection of invalid signatures
