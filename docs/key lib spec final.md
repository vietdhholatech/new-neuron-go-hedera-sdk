# Feature Specification: Key Library

**Feature Branch**: `002-key-library`  
**Created**: 2026-01-23  
**Status**: Draft  
**Input**: User description: "Key Library High-Level Specification"

## Purpose

The Key Library is a core component of the Neuron SDK that provides type-safe, interoperable cryptographic key management. This library standardizes how keys are handled throughout the Neuron SDK ecosystem, ensuring that all keys conform to this specification and work seamlessly across multiple blockchain networks.

The library replaces brittle, string-based key handling with a robust, type-safe system. All keys used within the Neuron SDK must be created, managed, and converted using this library to ensure consistency, security, and interoperability. Keys that conform to this specification can be used throughout the Neuron SDK with confidence that they will work correctly for signing, verification, address derivation, and cross-ecosystem conversions.

The library primarily focuses on single-key operations (NeuronPrivateKey/NeuronPublicKey) as the fully-supported path. Optional wrapper types (MultisigKey, ThresholdKey) MAY be provided for blockchain-specific multisig and threshold key implementations to enable type-safe SDK interoperability, with clear limitations documented for operations that are not supported.

## Architecture Principle

**The Key Library uses the Adapter Pattern** to provide a unified, type-safe interface over existing blockchain SDK and cryptographic libraries. The library does not reimplement cryptographic primitives or blockchain-specific operations. Instead, it adapts underlying libraries to provide type-safe Neuron key types.

**Implementation Constraint**: Implementations of underlying libraries MUST NOT be visible in the Neuron key API. Only the type-safe interface, constraints, and restrictions are exposed. For example, a NeuronPrivateKey (which is secp256k1) MUST NOT expose methods like `toED25519()` or any functionality incompatible with secp256k1 keys. The key type's cryptographic properties (secp256k1) determine what operations are available—no additional functionality beyond what the underlying key type supports should be exposed.



## Type Hierarchy

**NeuronPrivateKey** is the top-level, primary type in the Key Library. All key operations and conversions originate from or relate to NeuronPrivateKey.

### Primary Type

- **NeuronPrivateKey**: A type-safe type representing a private cryptographic key using secp256k1 curve. This secp256k1 format is compatible with multiple blockchain networks and ecosystems, enabling a single key to work across different networks. This is the entry point for all key operations in the library. The key can be converted to underlying blockchain SDK key types when blockchain-specific functionality is needed.

### Direct Derivatives

From a NeuronPrivateKey, the following can be derived:

- **NeuronPublicKey**: The public key corresponding to the private key, derived deterministically from the NeuronPrivateKey using standard secp256k1 operations
- **EVMAddress**: The EVM-compatible address derived from the NeuronPrivateKey's public key using standard Ethereum address derivation
- **PeerID**: The Libp2p peer identifier derived from the NeuronPrivateKey's public key using Libp2p's standard key-to-peer-ID conversion

### Type Relationships

The library maintains the following relationships:

1. **NeuronPrivateKey → NeuronPublicKey**: A private key can derive its corresponding public key
2. **NeuronPrivateKey → EVMAddress**: A private key can derive its EVM-compatible address
3. **NeuronPrivateKey → PeerID**: A private key can derive its Libp2p PeerID
4. **NeuronPublicKey → EVMAddress**: A public key can derive its EVM-compatible address
5. **NeuronPublicKey → PeerID**: A public key can derive its Libp2p PeerID
6. **NeuronPrivateKey ↔ Blockchain Keys**: A NeuronPrivateKey can be converted to underlying blockchain SDK key types (e.g., blockchain-specific private keys) and vice versa
7. **NeuronPublicKey ↔ Blockchain Keys**: A NeuronPublicKey can be converted to underlying blockchain SDK key types (e.g., blockchain-specific public keys) and vice versa
8. **ThresholdKey ↔ Blockchain Threshold Keys** (optional): A ThresholdKey can be converted to/from blockchain-specific threshold key types (e.g., Hedera threshold keys) for SDK interoperability. ThresholdKey cannot derive EVM addresses or PeerIDs
9. **MultisigKey** (optional): Wraps multiple NeuronPrivateKey instances with m-of-n threshold. Operations limited to what underlying keys support

All conversions and derivations maintain cryptographic relationships that can be verified through type-safe matching functions. The NeuronPrivateKey serves as the single source of truth for all derived types. The library provides bidirectional conversion interfaces to convert between Neuron key types and native blockchain SDK key types, enabling interoperability with blockchain-specific SDKs.

### Optional Multisig/Threshold Key Types

The library MAY provide optional types for multisig and threshold keys to enable type-safe interoperability with blockchain-specific multisig implementations:

- **MultisigKey**: A type representing multiple NeuronPrivateKey instances with m-of-n threshold signing configuration. This enables type-safe representation of multisig configurations using individual secp256k1 keys. Supports bidirectional conversion to/from blockchain SDK multisig types.

- **ThresholdKey**: A type representing blockchain-specific threshold keys (e.g., Hedera threshold keys). This provides type safety for blockchain-specific multisig implementations that cannot be represented as standard secp256k1 keys. Tracks the threshold signature protocol/standard used (e.g., Hedera's ThresholdKey protocol, FROST, BLS) to enable protocol-aware operations.

**Important Limitations**:
- Blockchain-specific threshold keys (e.g., Hedera threshold keys) CANNOT derive EVM addresses or PeerIDs, as they are not standard secp256k1 keys
- MultisigKey supports conversion to/from blockchain SDK types only; signing/verification operations are delegated to the blockchain SDK
- These wrapper types are primarily for blockchain SDK interoperability, not full feature parity with single-key operations
- Single-key operations (NeuronPrivateKey) remain the primary, fully-supported path

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

### User Story 4 - Multisig/Threshold Key Interoperability (Priority: P4, Optional)

Developers working with blockchain-specific multisig and threshold key implementations (e.g., Hedera threshold keys) need type-safe types to integrate these keys with the Neuron SDK ecosystem, even when they cannot perform all standard operations like EVM address derivation.

**Why this priority**: Enables interoperability with blockchain-specific multisig implementations, but is optional and has limited functionality compared to single-key operations.

**Independent Test**: Can be fully tested by creating a ThresholdKey from a blockchain-specific threshold key (e.g., Hedera), converting it to the blockchain SDK type and back, and verifying that operations not supported (like EVM address derivation) return clear errors.

**Acceptance Scenarios**:

1. **Given** a blockchain-specific threshold key (e.g., Hedera threshold key), **When** converting it to a ThresholdKey, **Then** it can be converted to/from the blockchain SDK type, but attempts to derive EVM address or PeerID return clear errors
2. **Given** multiple NeuronPrivateKey instances, **When** creating a MultisigKey with m-of-n threshold, **Then** the multisig key can perform operations supported by all underlying keys, with clear documentation of limitations
3. **Given** a ThresholdKey from a blockchain-specific implementation, **When** attempting to derive EVM address, **Then** the system returns a clear error indicating the operation is not supported for blockchain-specific threshold keys

---

### Edge Cases

- What happens when a key string has invalid hex characters? (Return descriptive error with exact position of invalid character)
- How does the system handle keys with missing or incorrect 0x prefixes? (Normalize input by stripping/adding prefix as needed, document behavior)
- What if an Ed25519 key is provided when converting to Neuron key? (Always reject with clear error message indicating Ed25519 keys cannot be converted to secp256k1 Neuron keys, as they use different cryptographic curves)
- How are zero-value or invalid keys handled? (Return error or invalid key type that cannot be used for operations)
- What happens when converting between formats fails (e.g., invalid key format)? (Return typed error explaining the failure reason)
- How does the system handle mnemonic phrases with invalid checksums? (Reject with error indicating checksum validation failure)
- What if encryption/decryption fails due to wrong password? (Return error without revealing whether key or password was wrong, to prevent timing attacks)
- What happens when attempting to derive EVM address or PeerID from a blockchain-specific threshold key? (Return clear error indicating the operation is not supported for blockchain-specific threshold keys)
- How are multisig operations handled when individual keys have different capabilities? (Operations are limited to what all keys support)
- What happens when attempting to use functionality incompatible with the key type (e.g., calling `toED25519()` on a secp256k1 NeuronPrivateKey)? (The method MUST NOT exist in the API—incompatible functionality is prevented at the type level, not through runtime errors)
- What happens when an underlying blockchain SDK operation fails (e.g., Hedera SDK throws an exception)? (Return a typed SDKError that wraps the underlying SDK error with Neuron SDK context, preserving original error information while providing operation context)
- What happens when attempting to modify properties of a sealed Neuron key? (Keys are sealed at construction—no setters or mutators exist in the API. Attempts to modify keys are prevented at the type level, not through runtime errors)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The library MUST provide NeuronPrivateKey and NeuronPublicKey types that ensure validity and provide type-safe interfaces. Key validation MUST occur at construction/elevation time only—once a NeuronKey instance exists, it is guaranteed to be valid and no further validation is required. NeuronPublicKey MUST provide both compressed (33 bytes) and uncompressed (65 bytes) format representations via explicit methods. The default serialization SHOULD use compressed format for efficiency, with uncompressed format available when needed for operations such as EVM address derivation
- **FR-002**: The library MUST support elevating primitive keys (blockchain SDK keys, raw bytes, hex strings) into NeuronPrivateKey or NeuronPublicKey types. Validation MUST occur during elevation, and invalid keys MUST be rejected with descriptive errors (FR-008). Once elevated, keys are sealed and immutable
- **FR-003**: The library MUST default to ECDSA secp256k1 keys for primary operations to ensure compatibility with multiple blockchain networks
- **FR-004**: The library MUST detect and reject Ed25519 keys from external sources with clear error messages. Ed25519 keys cannot be converted to secp256k1 Neuron keys as they use different cryptographic curves. Neuron keys are always secp256k1, and Ed25519 keys are not supported for conversion
- **FR-004a**: Neuron key types MUST NOT expose methods or functionality that is incompatible with their underlying cryptographic properties. For example, a NeuronPrivateKey (secp256k1) MUST NOT have methods like `toED25519()` or any operations that would require a different curve. The API surface MUST be constrained to operations that are valid for the underlying key type
- **FR-005**: The library MUST provide deterministic conversion from NeuronPublicKey to EVM-compatible address
- **FR-006**: The library MUST provide deterministic conversion from NeuronPublicKey to Libp2p PeerID
- **FR-007**: The library MUST provide bidirectional conversion between EVM addresses and PeerIDs when the underlying NeuronPublicKey is known
- **FR-008**: The library MUST validate all string inputs (hex validity, length checks, prefix handling) and return structured error types with specific error kinds and descriptive messages for invalid inputs. The library MUST support the following error kinds: InvalidFormat, InvalidLength, InvalidHex, InvalidKey, ZeroValue, KeyMismatch, Encryption, Mnemonic, Derivation, UnsupportedKeyType, and SDKError
- **FR-008a**: When underlying blockchain SDK operations fail, the library MUST return typed errors (SDKError kind) that wrap the underlying SDK error with Neuron SDK context. The wrapped error MUST preserve the original SDK error information while providing context about which Neuron operation failed. The library MUST NOT let underlying SDK exceptions propagate directly to users
- **FR-009**: The library MUST use type-safe function signatures (accepting and returning types, not strings) for all key operations
- **FR-010**: The library MUST provide factory methods to create NeuronPrivateKey and NeuronPublicKey from various sources. Serialization formats for raw keys MUST include hex strings (with or without 0x prefix) and raw bytes as required formats. Base58 encoding MAY be supported optionally for PeerID compatibility, but is not required for core key operations. JSON serialization is ONLY supported for EncryptedPrivateKey (FR-015); raw NeuronPrivateKey and NeuronPublicKey MUST NOT be JSON-serializable
- **FR-011**: The library MUST provide bidirectional conversion interfaces between Neuron key types and underlying blockchain SDK key types (e.g., NeuronPrivateKey ↔ blockchain private key, NeuronPublicKey ↔ blockchain public key). Converting a Neuron key to a blockchain key and back MUST produce an identical Neuron key
- **FR-012**: The library MUST support generating new NeuronPrivateKeys with cryptographically secure randomness
- **FR-013**: The library MUST support restoring NeuronPrivateKeys from mnemonic phrases following BIP-39 standard for mnemonic generation and BIP-44 standard for derivation paths. The library MUST use a default derivation path of m/44'/60'/0'/0/0 (Ethereum standard) but MUST also support custom derivation paths specified by users
- **FR-014**: The library MUST provide a standard signing interface that NeuronPrivateKey implements, producing ECDSA signatures with recovery ID in R||S||V format (65 bytes total). The library MUST support both standard format (V = 0 or 1) and Ethereum legacy format (V = 27 or 28) for compatibility with Ethereum ecosystem
- **FR-015**: The library MUST provide methods to encrypt (scramble) and decrypt (unscramble) private keys using password-based encryption with Argon2id key derivation function and AES-256-GCM authenticated encryption. The encryption scheme MUST support versioning for backward compatibility, with version 1 using hardcoded default Argon2 parameters and version 2 storing custom Argon2 parameters (time, memory, threads) for reliable decryption
- **FR-016**: The library MUST provide type-safe matching functions to verify relationships: PrivateKey.Matches(PublicKey), PrivateKey.Matches(EVMAddress), PublicKey.Matches(PeerID)
- **FR-017**: The library MUST support signing messages and verifying signatures using Neuron keys. Message signing MUST use Keccak256 hashing before ECDSA signing. Signature verification MUST support both: (1) direct verification of a signature against a known public key (message + signature + public key → boolean), and (2) recovery of the public key from signatures (message + signature → public key) using the R||S||V format
- **FR-018**: The library MUST ensure that if a function returns a NeuronKey type, the key is guaranteed to be valid and usable
- **FR-019**: The library MUST use distinct types for different key representations (NeuronKey, EVMAddress, PeerID) to prevent mixing them up
- **FR-020**: The library MUST use clear API naming that indicates whether operations are computations (deriving values) or transformations (elevating types)
- **FR-021**: The library MUST provide methods to zeroize private key material from memory after use or when explicitly cleared to prevent key material from persisting in memory
- **FR-022**: Neuron key types (NeuronPrivateKey, NeuronPublicKey) MUST be immutable, sealed, and thread-safe for concurrent read operations. Keys MUST be sealed at construction/elevation time—no properties, fields, or internal state can be modified after the key is created. Keys allow safe sharing across multiple threads without synchronization for read operations. The `Zeroize()` method is an exception: it mutates the receiver and requires external synchronization if other threads may access the same key instance
- **FR-023**: The library MAY provide optional MultisigKey and ThresholdKey types for blockchain-specific multisig and threshold key implementations. These types MUST support bidirectional conversion to/from blockchain SDK multisig/threshold key types. ThresholdKey MUST track and expose the threshold signature protocol/standard identifier (e.g., "hedera-threshold", "frost", "bls") to enable developers to understand which protocol is being used. Blockchain-specific threshold keys MUST NOT support EVM address or PeerID derivation, and attempts to perform these operations MUST return clear errors indicating the limitation
- **FR-024**: If MultisigKey is provided, it MUST represent multiple NeuronPrivateKey instances with m-of-n threshold configuration. MultisigKey MUST support bidirectional conversion to/from blockchain SDK multisig key types only

### Security Requirements

Following the Adapter Pattern, security requirements and threat assumptions are inherited from the underlying blockchain SDK and cryptographic libraries. The library MUST NOT reimplement security measures but MUST ensure that security properties of underlying libraries are preserved and not weakened.

- **SEC-001**: The library MUST inherit security properties and threat assumptions from underlying blockchain SDKs. For example, when adapting Hedera SDK keys, the library MUST follow Hedera's security principles and threat model. When adapting Ethereum/EVM keys, the library MUST follow Ethereum ecosystem security practices
- **SEC-002**: The library MUST preserve security guarantees provided by underlying cryptographic libraries. All cryptographic operations (signing, verification, key derivation) MUST delegate to underlying libraries without weakening their security properties
- **SEC-003**: The library MUST implement secure memory handling for private key material. Private keys MUST be zeroized from memory when explicitly cleared (FR-021). The library MUST NOT expose private key material through logging, error messages, or debugging interfaces
- **SEC-004**: The library MUST protect against timing attacks in operations that compare keys or verify relationships. Matching functions (FR-016) MUST use constant-time comparisons when comparing cryptographic material
- **SEC-005**: The library MUST validate all inputs before processing to prevent injection attacks, buffer overflows, or other input-based vulnerabilities. Invalid inputs MUST be rejected with descriptive errors (FR-008) without revealing sensitive information
- **SEC-006**: Encryption and decryption operations (FR-015) MUST use secure, industry-standard algorithms (Argon2id, AES-256-GCM) as specified. The library MUST NOT implement custom encryption schemes
- **SEC-007**: When converting between Neuron keys and blockchain SDK keys, the library MUST ensure that security properties are maintained. Converting a secure blockchain SDK key to a Neuron key and back MUST not weaken security guarantees
- **SEC-008**: The library MUST NOT introduce new attack vectors beyond those present in underlying libraries. All security considerations (side-channel attacks, memory attacks, etc.) are inherited from the underlying blockchain SDK being adapted

### Observability Requirements

- **OBS-001**: The library MUST log errors that occur during operations. Error logs MUST include sufficient context to diagnose issues (operation type, error kind, relevant identifiers) but MUST NOT include private key material or sensitive cryptographic data
- **OBS-002**: The library MUST NOT emit metrics or distributed tracing signals. Observability is limited to error logging only
- **OBS-003**: Error logging MUST use structured logging format when available, but the library MUST NOT require a specific logging framework or implementation

### Key Entities *(include if feature involves data)*

- **NeuronPrivateKey**: An immutable, type-safe type representing a private cryptographic key using secp256k1 curve that ensures validity. Thread-safe for concurrent read operations. Can be converted bidirectionally to/from underlying blockchain SDK key types (Ed25519 keys from external sources are rejected, not converted). MUST NOT expose functionality incompatible with secp256k1 (e.g., no `toED25519()` method)
- **NeuronPublicKey**: An immutable, type-safe type representing a public cryptographic key that ensures validity and enables conversions to various formats. Thread-safe for concurrent read operations. Public keys MUST provide both compressed (33 bytes) and uncompressed (65 bytes) format representations via explicit methods, with compressed format as the default for efficiency. Can be converted bidirectionally to/from underlying blockchain SDK key types
- **EVMAddress**: A distinct type representing an EVM-compatible address (20-byte value) that cannot be confused with key types. The library MUST support EIP-55 checksummed addresses for Ethereum ecosystem compatibility, providing both lowercase hex and mixed-case checksummed hex representations
- **PeerID**: A distinct type representing a Libp2p peer identifier that cannot be confused with other key types
- **Signature**: A type representing an ECDSA cryptographic signature in R||S||V format (65 bytes total: 32 bytes R, 32 bytes S, 1 byte recovery ID V) that can be verified against a public key and enables public key recovery. The signature MUST support both standard format (V = 0 or 1) and Ethereum legacy format (V = 27 or 28) for compatibility
- **EncryptedPrivateKey**: A JSON-serializable structure containing an encrypted private key with the following fields: Version (encryption scheme version: 1 or 2), Salt (16-byte random salt for Argon2id), Nonce (12-byte random nonce for AES-GCM), Ciphertext (48 bytes: 32-byte encrypted key + 16-byte GCM authentication tag), and optional Argon2 parameters (Time, Memory, Threads) for Version 2+. Version 1 uses hardcoded Argon2 defaults, while Version 2 stores custom parameters for reliable decryption
- **MultisigKey** (optional): A type representing multiple NeuronPrivateKey instances with m-of-n threshold signing configuration. Enables type-safe representation of multisig using individual secp256k1 keys. Supports bidirectional conversion to/from blockchain SDK multisig key types only. Cannot derive EVM addresses or PeerIDs unless all keys support the operation and aggregation is implemented
- **ThresholdKey** (optional): A type representing blockchain-specific threshold keys (e.g., Hedera threshold keys). Provides type safety for blockchain-specific multisig implementations. MUST track and expose the threshold signature protocol/standard identifier (e.g., "hedera-threshold", "frost", "bls") to indicate which protocol is used. Supports bidirectional conversion to/from blockchain SDK threshold key types only. CANNOT derive EVM addresses or PeerIDs as these are not standard secp256k1 keys. Primary purpose is blockchain SDK interoperability

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
