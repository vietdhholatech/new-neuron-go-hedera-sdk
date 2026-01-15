# Keylib Implementation Mapping

This document maps the implemented keylib functionality to the exact technical requirements in the **Key Library High-Level Specification**.

---

## Specification Compliance Summary

| Requirement Category            | Status   | Coverage | Technical Implementation                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------- | -------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A. The "Neuron Key" Elevation   | Complete | 100%     | Immutable wrapper types (`NeuronPrivateKey`, `NeuronPublicKey`) with unexported `*secp256k1.PrivateKey`/`*secp256k1.PublicKey` fields. Factory functions (`ParsePrivateKeyHex`, `PrivateKeyFromBytes`, `PrivateKeyFromHedera`) validate inputs and return `(T, error)`. Zero-value structs are explicitly invalid (`IsZero()` returns true). All constructors enforce secp256k1 curve constraints.                                                                                                                                                              |
| B. Primary Key Strategy (ECDSA) | Complete | 100%     | All key operations use `github.com/decred/dcrd/dcrec/secp256k1/v4` curve. Ed25519 rejection via 3-tier detection: (1) DER encoding pattern matching for OID sequences `302e`, `2b6570`, `302a`; (2) raw bytes length check (Ed25519=64 bytes vs ECDSA=32 bytes); (3) secp256k1 scalar validation (0 < scalar < N where N=0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141).                                                                                                                                                                   |
| C. Robust Conversion Matrix     | Complete | 100%     | **EVM Address**: Keccak256(uncompressed_pubkey[1:]) → last 20 bytes. **PeerID**: `libp2p/go-libp2p/core/crypto.UnmarshalSecp256k1PublicKey()` → `peer.IDFromPublicKey()` → base58 multihash encoding. **Hedera interop**: `ToHederaPrivateKey()` via `hiero.PrivateKeyFromBytesECDSA()`, `ToHederaPublicKey()` via `hiero.PublicKeyFromBytesECDSA()` (using Hiero SDK v2.74.0). Safe variants (`*Safe()`) return errors instead of zero values.                                                                                                                 |
| D. Input Handling & Validation  | Complete | 100%     | **Hex parsing**: `normalizeHex()` strips `0x`/`0X` prefix, lowercases; `validateHexString()` returns exact position of invalid character. **Private key validation**: `isValidSecp256k1Scalar()` checks 0 < scalar < curve order N. **Public key validation**: SEC1 format check (33 bytes compressed with 0x02/0x03 prefix, or 65 bytes uncompressed with 0x04 prefix); point-on-curve verification via `secp256k1.ParsePubKey()`. **Error types**: `KeyError{Op, Kind, Details, Err}` with 10 distinct `ErrorKind` values.                                    |
| E. Type Safety                  | Complete | 100%     | Distinct struct types prevent confusion: `NeuronPrivateKey{key *secp256k1.PrivateKey}`, `NeuronPublicKey{key *secp256k1.PublicKey}`, `EVMAddress{addr [20]byte}`, `PeerID{id peer.ID}`, `Signature{data [65]byte}`, `EncryptedPrivateKey{Version, Salt, Nonce, Ciphertext}`. All fields unexported; access only via validated constructors. Method receivers prevent nil-pointer panics with `IsZero()` guards.                                                                                                                                                 |
| Key Management Functions        | Complete | 100%     | **Generation**: `crypto/rand.Read()` → secp256k1 scalar validation → `NeuronPrivateKey`. **Mnemonic**: BIP39 via `go-bip39`; BIP32 derivation via custom HMAC-SHA512 implementation; default path `m/44'/60'/0'/0/0`. **Signing**: `SignMessage()` applies Keccak256 then ECDSA sign with recovery ID; `SignDigest()` for pre-hashed data. **Encryption**: Argon2id KDF (time=3, memory=64MB, threads=4, keyLen=32) → AES-256-GCM (12-byte random nonce per encryption). **Matching**: `subtle.ConstantTimeCompare()` for all `Matches*` and `Equal()` methods. |
| Inter-Ecosystem Conversions     | Complete | 100%     | **Neuron→Ethereum**: `EVMAddress()` computes address deterministically; `ChecksumHex()` implements EIP-55 mixed-case encoding. **Neuron→Libp2p**: `PeerID()` converts secp256k1 public key to libp2p crypto format, then derives multihash-encoded peer ID. **Reverse mappings**: `EVMAddressFromCommon()` from go-ethereum `common.Address`; `PeerIDFromLibp2p()` from `peer.ID`. **Cross-validation**: `MatchesEVMAddress()`, `MatchesPeerID()` verify cryptographic relationships.                                                                           |
| Concurrency & Thread Safety     | Complete | 99%      | All types immutable after construction; no global mutable state; no mutex locks required. All functions thread-safe except `Zeroize()` which mutates receiver. CPU-intensive functions (`Scramble`, `UnscramblePrivateKey`) allocate 64MB each—limit concurrency in high-throughput scenarios. Minimal blocking I/O via `crypto/rand.Reader` (non-blocking on modern systems). See [Concurrency & Thread Safety](#concurrency--thread-safety) section for detailed goroutine usage guidelines and warnings.                                                     |

**Test Coverage**: 85.4%

---

## Section 2.A: The "Neuron Key" Elevation

### Requirement

> Any primitive key (Hedera Public Key, raw bytes) must be "elevated" into a `NeuronPublicKey` or `NeuronPrivateKey`.

### Implementation

| Spec Requirement            | Implementation                                    | File:Line                                                                                          |
| --------------------------- | ------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Elevation from Hedera types | `PrivateKeyFromHedera()`, `PublicKeyFromHedera()` | [factory.go:75-111](keylib/factory.go#L75-L111), [factory.go:161-179](keylib/factory.go#L161-L179) |
| Elevation from raw bytes    | `PrivateKeyFromBytes()`, `PublicKeyFromBytes()`   | [factory.go:54-66](keylib/factory.go#L54-L66), [factory.go:142-157](keylib/factory.go#L142-L157)   |
| Elevation from hex strings  | `ParsePrivateKeyHex()`, `ParsePublicKeyHex()`     | [factory.go:33-50](keylib/factory.go#L33-L50), [factory.go:121-138](keylib/factory.go#L121-L138)   |

### Core Types

```go
// NeuronPrivateKey - Immutable, validated ECDSA secp256k1 private key
type NeuronPrivateKey struct {
    key *secp256k1.PrivateKey  // unexported, never nil after valid construction
}

// NeuronPublicKey - Immutable, validated ECDSA secp256k1 public key
type NeuronPublicKey struct {
    key *secp256k1.PublicKey   // unexported, never nil after valid construction
}
```

**Files**: [private_key.go](keylib/private_key.go), [public_key.go](keylib/public_key.go)

---

## Section 2.B: Primary Key Strategy - ECDSA & Ethereum Compatibility

### Requirement

> We will standardize on ECDSA (Secp256k1) keys to ensure seamless interoperability between Hedera and Ethereum ecosystems.

### Implementation

| Spec Requirement           | Implementation                                       | Status |
| -------------------------- | ---------------------------------------------------- | ------ |
| Default to ECDSA secp256k1 | All key operations use secp256k1 curve               | Done   |
| Reject Ed25519 keys        | `PrivateKeyFromHedera()` detects and rejects Ed25519 | Done   |
| Ed25519 detection helpers  | `IsEd25519Key()`, `IsEd25519PublicKey()`             | Done   |

### Ed25519 Detection (QA/QC Enhancement)

**File**: [factory.go:231-257](keylib/factory.go#L231-L257)

Detection uses three methods:

1. **DER encoding pattern detection** - Checks for Ed25519 OID (`302e`, `2b6570`, `302a`)
2. **Raw bytes length check** - Ed25519 = 64 bytes, ECDSA = 32 bytes
3. **secp256k1 scalar validation** - Final safety check

```go
// IsEd25519Key detects if a Hedera private key is Ed25519 type.
// Note: Uses hiero.PrivateKey from the Hiero SDK (formerly Hedera SDK)
func IsEd25519Key(hederaKey hiero.PrivateKey) bool {
    // Method 1: Check DER encoding for Ed25519 OID
    derStr := hederaKey.String()
    if strings.Contains(derStr, "302e") ||
        strings.Contains(derStr, "2b6570") ||
        strings.HasPrefix(derStr, "302a") {
        return true
    }
    // Method 2: Ed25519 raw bytes are 64 bytes
    if len(hederaKey.BytesRaw()) == 64 {
        return true
    }
    return false
}
```

---

## Section 2.C: Robust Conversion Matrix

### Requirement

> The library must provide a "Rosetta Stone" for keys, allowing conversions between Neuron Public Keys, Ethereum Addresses, and Libp2p PeerIDs.

### Implementation Matrix

| From              | To                 | Method                   | File:Line                                                 |
| ----------------- | ------------------ | ------------------------ | --------------------------------------------------------- |
| NeuronPublicKey   | EVMAddress         | `EVMAddress()`           | [public_key.go:84-93](keylib/public_key.go#L84-L93)       |
| NeuronPublicKey   | PeerID             | `PeerID()`               | [public_key.go:110-118](keylib/public_key.go#L110-L118)   |
| NeuronPublicKey   | Hedera PublicKey   | `ToHederaPublicKey()`    | [public_key.go:143-156](keylib/public_key.go#L143-L156)   |
| NeuronPublicKey   | \*ecdsa.PublicKey  | `ToECDSA()`              | [public_key.go:123-136](keylib/public_key.go#L123-L136)   |
| NeuronPrivateKey  | EVMAddress         | `EVMAddress()`           | [private_key.go:53-55](keylib/private_key.go#L53-L55)     |
| NeuronPrivateKey  | PeerID             | `PeerID()`               | [private_key.go:70-76](keylib/private_key.go#L70-L76)     |
| NeuronPrivateKey  | NeuronPublicKey    | `PublicKey()`            | [private_key.go:41-46](keylib/private_key.go#L41-L46)     |
| NeuronPrivateKey  | Hedera PrivateKey  | `ToHederaPrivateKey()`   | [private_key.go:128-140](keylib/private_key.go#L128-L140) |
| NeuronPrivateKey  | \*ecdsa.PrivateKey | `ToECDSA()`              | [private_key.go:116-121](keylib/private_key.go#L116-L121) |
| Hedera PrivateKey | NeuronPrivateKey   | `PrivateKeyFromHedera()` | [factory.go:75-111](keylib/factory.go#L75-L111)           |
| Hedera PublicKey  | NeuronPublicKey    | `PublicKeyFromHedera()`  | [factory.go:161-179](keylib/factory.go#L161-L179)         |

### Safe Variants (QA/QC Enhancement)

For safety-critical operations, safe variants return errors instead of zero values:

| Method                 | Safe Variant               | Purpose                                    |
| ---------------------- | -------------------------- | ------------------------------------------ |
| `EVMAddress()`         | `EVMAddressSafe()`         | Prevents silent derivation to burn address |
| `ToHederaPrivateKey()` | `ToHederaPrivateKeySafe()` | Returns error instead of empty key         |
| `ToHederaPublicKey()`  | `ToHederaPublicKeySafe()`  | Returns error instead of empty key         |

**Files**: [public_key.go:100-106](keylib/public_key.go#L100-L106), [private_key.go:60-66](keylib/private_key.go#L60-L66), [private_key.go:144-157](keylib/private_key.go#L144-L157), [public_key.go:160-173](keylib/public_key.go#L160-L173)

---

## Section 2.D: Input Handling & Validation

### Requirement

> "Don't accept stupid strings." Functions accepting strings must validate all preconditions.

### Implementation

#### Validation Functions

**File**: [validation.go](keylib/validation.go)

| Function                    | Purpose                                       |
| --------------------------- | --------------------------------------------- |
| `normalizeHex()`            | Strips `0x`/`0X` prefix, lowercases           |
| `isValidHexChar()`          | Validates hex character                       |
| `validateHexString()`       | Returns position of invalid character         |
| `decodeHexStrict()`         | Strict hex decoding with validation           |
| `isValidSecp256k1Scalar()`  | Validates private key scalar (0 < scalar < N) |
| `validatePrivateKeyBytes()` | Full private key validation                   |
| `validatePublicKeyBytes()`  | SEC1 format validation                        |

#### Error Handling

**File**: [errors.go](keylib/errors.go)

Rich error types with context:

```go
type KeyError struct {
    Op      string    // Operation that failed (e.g., "ParsePrivateKeyHex")
    Kind    ErrorKind // Category (e.g., ErrKindInvalidHex)
    Details string    // Human-readable explanation
    Err     error     // Underlying error
}

type ErrorKind int
const (
    ErrKindInvalidFormat ErrorKind = iota + 1
    ErrKindInvalidLength
    ErrKindInvalidHex
    ErrKindInvalidKey
    ErrKindZeroValue
    ErrKindKeyMismatch
    ErrKindEncryption
    ErrKindMnemonic
    ErrKindDerivation
    ErrKindUnsupportedKeyType
)
```

#### Descriptive Error Examples

| Input Issue                        | Error Message                                                                                  |
| ---------------------------------- | ---------------------------------------------------------------------------------------------- |
| Invalid hex char 'g' at position 5 | `keylib.ParsePrivateKeyHex: InvalidHex: invalid hex character 'g' at position 5`               |
| Wrong length (63 chars)            | `keylib.ParsePrivateKeyHex: InvalidLength: expected 64 hex characters for private key, got 63` |
| Zero scalar                        | `keylib.PrivateKeyFromBytes: InvalidKey: zero scalar is not valid on secp256k1`                |
| Ed25519 key provided               | `keylib.PrivateKeyFromHedera: UnsupportedKeyType: Ed25519 (detected via DER encoding)`         |

---

## Section 2.E: Type Safety

### Requirement

> Functions should accept and return **Types**, not strings.

### Implementation

All functions use strong types:

```go
// Correct: Type-safe function signatures
func ParsePrivateKeyHex(s string) (NeuronPrivateKey, error)
func (k NeuronPublicKey) EVMAddress() EVMAddress
func (k NeuronPrivateKey) SignMessage(msg []byte) (Signature, error)
func (k NeuronPublicKey) Verify(msg []byte, sig Signature) bool
```

### Type Definitions

| Type                  | Purpose                         | File                                        |
| --------------------- | ------------------------------- | ------------------------------------------- |
| `NeuronPrivateKey`    | ECDSA secp256k1 private key     | [private_key.go](keylib/private_key.go)     |
| `NeuronPublicKey`     | ECDSA secp256k1 public key      | [public_key.go](keylib/public_key.go)       |
| `EVMAddress`          | 20-byte Ethereum address        | [evm_address.go](keylib/evm_address.go)     |
| `PeerID`              | libp2p peer identifier          | [peer_id.go](keylib/peer_id.go)             |
| `Signature`           | 65-byte ECDSA signature (R‖S‖V) | [signature.go](keylib/signature.go)         |
| `EncryptedPrivateKey` | Encrypted key with metadata     | [encrypted_key.go](keylib/encrypted_key.go) |

---

## Section 3: Scope of Functionality - Key Management

### A. Elevation (Factory Methods)

| Spec Requirement      | Implementation                                    | File:Line                                                                                          |
| --------------------- | ------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| From hex strings      | `ParsePrivateKeyHex()`, `ParsePublicKeyHex()`     | [factory.go:33-50](keylib/factory.go#L33-L50), [factory.go:121-138](keylib/factory.go#L121-L138)   |
| From raw bytes        | `PrivateKeyFromBytes()`, `PublicKeyFromBytes()`   | [factory.go:54-66](keylib/factory.go#L54-L66), [factory.go:142-157](keylib/factory.go#L142-L157)   |
| From Hedera SDK types | `PrivateKeyFromHedera()`, `PublicKeyFromHedera()` | [factory.go:75-111](keylib/factory.go#L75-L111), [factory.go:161-179](keylib/factory.go#L161-L179) |

### B. Extraction (To External Types)

| Spec Requirement  | Implementation                                        | File:Line                                                                                                          |
| ----------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| To Hedera types   | `ToHederaPrivateKey()`, `ToHederaPublicKey()`         | [private_key.go:128-140](keylib/private_key.go#L128-L140), [public_key.go:143-156](keylib/public_key.go#L143-L156) |
| To standard ECDSA | `ToECDSA()`                                           | [private_key.go:116-121](keylib/private_key.go#L116-L121), [public_key.go:123-136](keylib/public_key.go#L123-L136) |
| To raw bytes      | `Bytes()`, `CompressedBytes()`, `UncompressedBytes()` | Various                                                                                                            |
| To hex string     | `Hex()`, `HexWithoutPrefix()`                         | Various                                                                                                            |

### C. Generation

| Spec Requirement         | Implementation                           | File:Line                                     |
| ------------------------ | ---------------------------------------- | --------------------------------------------- |
| Generate new private key | `GeneratePrivateKey()`                   | [factory.go:15-24](keylib/factory.go#L15-L24) |
| Default to ECDSA         | Uses `crypto/ecdsa` with secp256k1 curve | Done                                          |

### D. Restoration from Mnemonic

**File**: [mnemonic.go](keylib/mnemonic.go)

| Spec Requirement      | Implementation                                                         |
| --------------------- | ---------------------------------------------------------------------- |
| Generate mnemonic     | `GenerateMnemonic(wordCount int)`                                      |
| Restore from mnemonic | `PrivateKeyFromMnemonic(mnemonic string)`                              |
| With custom path      | `PrivateKeyFromMnemonicWithPath(mnemonic, path string)`                |
| With passphrase       | `PrivateKeyFromMnemonicWithPassphrase(mnemonic, passphrase string)`    |
| Full options          | `PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, path string)` |
| Validate mnemonic     | `ValidateMnemonic(mnemonic string)`                                    |

Default derivation path: `m/44'/60'/0'/0/0` (Ethereum standard)

### E. Signer Interface

**File**: [private_key.go:198-230](keylib/private_key.go#L198-L230)

```go
// NeuronPrivateKey implements crypto.Signer
var _ crypto.Signer = NeuronPrivateKey{}

func (k NeuronPrivateKey) Public() crypto.PublicKey
func (k NeuronPrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
```

### F. Key Safety (Scramble/Unscramble)

**File**: [encrypted_key.go](keylib/encrypted_key.go)

| Spec Requirement    | Implementation                                   |
| ------------------- | ------------------------------------------------ |
| Encrypt private key | `(k NeuronPrivateKey) Scramble(password string)` |
| Decrypt private key | `UnscramblePrivateKey(encrypted, password)`      |
| JSON serialization  | `MarshalJSON()`, `UnmarshalJSON()`               |

**Encryption Scheme**:

- **KDF**: Argon2id (time=3, memory=64MB, threads=4)
- **Cipher**: AES-256-GCM
- **Salt**: 16 bytes random
- **Nonce**: 12 bytes random per encryption

```go
type EncryptedPrivateKey struct {
    Version    int    `json:"version"`    // Currently 1
    Salt       []byte `json:"salt"`       // 16 bytes
    Nonce      []byte `json:"nonce"`      // 12 bytes
    Ciphertext []byte `json:"ciphertext"` // 32 + 16 (GCM tag) = 48 bytes
}
```

### G. Key Matching Verification

| Spec Requirement                 | Implementation                          | File:Line                                                 |
| -------------------------------- | --------------------------------------- | --------------------------------------------------------- |
| `PrivateKey.Matches(PublicKey)`  | `MatchesPublicKey(pub NeuronPublicKey)` | [private_key.go:234-240](keylib/private_key.go#L234-L240) |
| `PrivateKey.Matches(EVMAddress)` | `MatchesEVMAddress(addr EVMAddress)`    | [private_key.go:244-249](keylib/private_key.go#L244-L249) |
| `PublicKey.Matches(PeerID)`      | `MatchesPeerID(pid PeerID)`             | [public_key.go:187-196](keylib/public_key.go#L187-L196)   |
| `PublicKey.Matches(EVMAddress)`  | `MatchesEVMAddress(addr EVMAddress)`    | [public_key.go:177-183](keylib/public_key.go#L177-L183)   |

All matching functions use **constant-time comparison** to prevent timing attacks.

### H. Signing & Verification

| Spec Requirement         | Implementation                                 | File:Line                                                 |
| ------------------------ | ---------------------------------------------- | --------------------------------------------------------- |
| Sign message             | `SignMessage(msg []byte)`                      | [private_key.go:162-172](keylib/private_key.go#L162-L172) |
| Sign pre-hashed digest   | `SignDigest(digest [32]byte)`                  | [private_key.go:176-191](keylib/private_key.go#L176-L191) |
| Verify signature         | `Verify(msg []byte, sig Signature)`            | [public_key.go:201-209](keylib/public_key.go#L201-L209)   |
| Verify pre-hashed digest | `VerifyDigest(digest [32]byte, sig Signature)` | [public_key.go:213-231](keylib/public_key.go#L213-L231)   |
| Recover public key       | `RecoverPublicKey(msg []byte, sig Signature)`  | [signature.go](keylib/signature.go)                       |

**Signature Format**: 65 bytes (R‖S‖V) where V is recovery ID (0 or 1)

---

## Section 3: Inter-Ecosystem Conversions

### Neuron <-> Ethereum

| Conversion             | Implementation      | Algorithm                                   |
| ---------------------- | ------------------- | ------------------------------------------- |
| PublicKey → EVMAddress | `EVMAddress()`      | Keccak256(uncompressed[1:]) → last 20 bytes |
| EVMAddress parsing     | `ParseEVMAddress()` | Validates 40 hex chars, optional 0x prefix  |
| EIP-55 checksum        | `ChecksumHex()`     | Mixed-case encoding per EIP-55              |

### Neuron <-> Libp2p

| Conversion         | Implementation  | Algorithm                             |
| ------------------ | --------------- | ------------------------------------- |
| PublicKey → PeerID | `PeerID()`      | libp2p secp256k1 → multihash encoding |
| PeerID parsing     | `ParsePeerID()` | Base58 multibase decoding             |
| PeerID validation  | `Validate()`    | Checks non-empty and valid format     |

---

## Section 4: Developer Experience

### A. Predictability

> If a function returns a `NeuronKey`, it is valid.

**Implementation**: All constructors validate input. Zero-value keys are explicitly invalid:

```go
func (k NeuronPrivateKey) IsZero() bool { return k.key == nil }
func (k NeuronPublicKey) IsZero() bool { return k.key == nil }
```

### B. Safety

> Impossible to mix up an Ethereum address with a Neuron Key due to distinct types.

**Implementation**: Separate types prevent confusion:

```go
type NeuronPrivateKey struct { key *secp256k1.PrivateKey }
type NeuronPublicKey struct { key *secp256k1.PublicKey }
type EVMAddress struct { addr [20]byte }
type PeerID struct { id peer.ID }
type Signature struct { data [65]byte }
```

### C. Clarity

> API names should clearly indicate if they are performing a computation or transformation.

**Naming Conventions**:

| Prefix        | Meaning                           | Example                  |
| ------------- | --------------------------------- | ------------------------ |
| `Parse*`      | Construct from string (validates) | `ParsePrivateKeyHex()`   |
| `*FromBytes`  | Construct from raw bytes          | `PrivateKeyFromBytes()`  |
| `*FromHedera` | Elevate from Hedera SDK type      | `PrivateKeyFromHedera()` |
| `Generate*`   | Create new random value           | `GeneratePrivateKey()`   |
| `To*`         | Extract/convert to external type  | `ToHederaPrivateKey()`   |
| `*Bytes()`    | Serialize to raw bytes            | `Bytes()`                |
| `*Hex()`      | Serialize to hex string           | `Hex()`                  |
| `Matches*`    | Constant-time equality check      | `MatchesEVMAddress()`    |

---

## API Usage Guide

This section provides detailed operational instructions for all implemented APIs.

---

### 1. Key Generation and Parsing

#### Generate a New Private Key

```go
import "github.com/aspect-build/neuron-go-hedera-sdk/keylib"

// Generate a cryptographically secure random private key
privateKey, err := keylib.GeneratePrivateKey()
if err != nil {
    log.Fatalf("failed to generate key: %v", err)
}

// The key is immediately usable
pubKey := privateKey.PublicKey()
evmAddr := privateKey.EVMAddress()
```

#### Parse Private Key from Hex String

```go
// Accepts with or without "0x" prefix
hexKey := "0x4c0883a69102937d6231471b5dbb6204fe512961708279238e2e62c9f2fa3f1a"

privateKey, err := keylib.ParsePrivateKeyHex(hexKey)
if err != nil {
    // Error includes specific details:
    // - InvalidHex: "invalid hex character 'g' at position 5"
    // - InvalidLength: "expected 64 hex characters, got 63"
    // - InvalidKey: "zero scalar is not valid on secp256k1"
    log.Fatalf("invalid key: %v", err)
}
```

#### Parse Public Key from Hex String

```go
// Compressed format (33 bytes / 66 hex chars)
compressed := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"

// Uncompressed format (65 bytes / 130 hex chars)
uncompressed := "0x04759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae..."

pubKey, err := keylib.ParsePublicKeyHex(compressed)
if err != nil {
    log.Fatalf("invalid public key: %v", err)
}
```

#### Construct from Raw Bytes

```go
// Private key from 32-byte array
var rawPrivate [32]byte
copy(rawPrivate[:], someBytes)
privateKey, err := keylib.PrivateKeyFromBytes(rawPrivate)

// Public key from slice (33 or 65 bytes)
pubKey, err := keylib.PublicKeyFromBytes(compressedBytes)
```

---

### 2. Hedera SDK Interoperability

#### Elevate Hedera Key to Neuron Key

```go
import hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

// From Hedera ECDSA private key (using Hiero SDK)
hederaPrivate, _ := hiero.PrivateKeyGenerateEcdsa()
neuronPrivate, err := keylib.PrivateKeyFromHedera(hederaPrivate)
if err != nil {
    // Possible errors:
    // - UnsupportedKeyType: "Ed25519 (detected via DER encoding)"
    // - InvalidLength: "expected 32 bytes for ECDSA private key"
    log.Fatalf("cannot elevate key: %v", err)
}

// From Hedera public key
hederaPub := hederaPrivate.PublicKey()
neuronPub, err := keylib.PublicKeyFromHedera(hederaPub)
```

#### Check Key Type Before Elevation

```go
// Recommended pattern when key type is unknown
hederaKey := getKeyFromSomewhere()

if keylib.IsEd25519Key(hederaKey) {
    log.Fatal("Ed25519 keys are not supported; use ECDSA secp256k1")
}

neuronKey, err := keylib.PrivateKeyFromHedera(hederaKey)
```

#### Convert Neuron Key Back to Hedera SDK

```go
// Standard method (returns empty key on failure)
hederaKey := neuronPrivate.ToHederaPrivateKey()

// Safe method (returns error on failure) - RECOMMENDED
hederaKey, err := neuronPrivate.ToHederaPrivateKeySafe()
if err != nil {
    log.Fatalf("conversion failed: %v", err)
}

// Use with Hiero SDK (Hedera network)
client, _ := hiero.ClientForTestnet()
client.SetOperator(accountID, hederaKey)
```

---

### 3. Ethereum Address Derivation

#### Derive EVM Address from Key

```go
// From public key
evmAddr := pubKey.EVMAddress()
fmt.Println(evmAddr.Hex())          // "0xe364f2f1e5f4f03d1df682322500b9c68c997ec3"
fmt.Println(evmAddr.ChecksumHex())  // "0xe364f2f1e5F4F03d1df682322500b9c68C997ec3" (EIP-55)

// From private key (convenience method)
evmAddr := privateKey.EVMAddress()
```

#### Safe Address Derivation (Prevents Burn Address)

```go
// WARNING: EVMAddress() returns ZeroEVMAddress for zero-value keys
// The zero address (0x0000...0000) is the Ethereum burn address

// UNSAFE - may silently return burn address
var maybeZeroKey NeuronPrivateKey
addr := maybeZeroKey.EVMAddress() // Returns 0x0000...0000

// SAFE - returns error for zero-value keys
addr, err := maybeZeroKey.EVMAddressSafe()
if err != nil {
    log.Fatalf("cannot derive address from invalid key: %v", err)
}
```

#### Parse and Validate EVM Address

```go
// Parse from hex string
addr, err := keylib.ParseEVMAddress("0xe364f2f1e5F4F03d1df682322500b9c68C997ec3")
if err != nil {
    log.Fatalf("invalid address: %v", err)
}

// Check if address is zero (burn address)
if addr.IsZero() {
    log.Warn("this is the burn address")
}

// Verify key matches address
if privateKey.MatchesEVMAddress(addr) {
    fmt.Println("key controls this address")
}
```

---

### 4. Libp2p PeerID Derivation

#### Derive PeerID from Key

```go
// From public key
peerID, err := pubKey.PeerID()
if err != nil {
    log.Fatalf("failed to derive peer ID: %v", err)
}
fmt.Println(peerID.String()) // "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

// From private key
peerID, err := privateKey.PeerID()
```

#### Parse and Validate PeerID

```go
// Parse from string
peerID, err := keylib.ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")
if err != nil {
    log.Fatalf("invalid peer ID: %v", err)
}

// Validate peer ID
if err := peerID.Validate(); err != nil {
    log.Fatalf("peer ID validation failed: %v", err)
}

// Verify key matches peer ID
if pubKey.MatchesPeerID(peerID) {
    fmt.Println("this key owns this peer ID")
}
```

#### Interop with go-libp2p

```go
import "github.com/libp2p/go-libp2p/core/peer"

// Convert to libp2p peer.ID
libp2pID := peerID.ToLibp2p()

// Create from libp2p peer.ID
peerID := keylib.PeerIDFromLibp2p(libp2pID)
```

---

### 5. Message Signing and Verification

#### Sign a Message

```go
message := []byte("Hello, Neuron!")

// SignMessage hashes the message with Keccak256 before signing
signature, err := privateKey.SignMessage(message)
if err != nil {
    log.Fatalf("signing failed: %v", err)
}

// Signature is 65 bytes: R (32) || S (32) || V (1)
fmt.Printf("Signature: %s\n", signature.Hex())
```

#### Sign a Pre-Hashed Digest

```go
import "github.com/ethereum/go-ethereum/crypto"

// For pre-hashed data (e.g., Ethereum transaction hash)
digest := crypto.Keccak256Hash(message)
var digestArray [32]byte
copy(digestArray[:], digest[:])

signature, err := privateKey.SignDigest(digestArray)
```

#### Verify a Signature

```go
// Verify with public key
isValid := pubKey.Verify(message, signature)
if !isValid {
    log.Fatal("signature verification failed")
}

// Verify pre-hashed digest
isValid := pubKey.VerifyDigest(digestArray, signature)
```

#### Recover Public Key from Signature

```go
// Recover the signer's public key from message and signature
recoveredPub, err := keylib.RecoverPublicKey(message, signature)
if err != nil {
    log.Fatalf("recovery failed: %v", err)
}

// Verify it matches expected key
if recoveredPub.Equal(expectedPubKey) {
    fmt.Println("signature is from expected signer")
}
```

---

### 6. Mnemonic Operations (BIP39/BIP32)

#### Generate New Mnemonic

```go
// Generate 12-word mnemonic (128 bits entropy)
mnemonic, err := keylib.GenerateMnemonic(12)
if err != nil {
    log.Fatalf("mnemonic generation failed: %v", err)
}
fmt.Println(mnemonic) // "abandon ability able about above absent..."

// Supported word counts: 12, 15, 18, 21, 24
mnemonic24, _ := keylib.GenerateMnemonic(24) // 256 bits entropy
```

#### Restore Key from Mnemonic

```go
// Default path: m/44'/60'/0'/0/0 (Ethereum standard)
privateKey, err := keylib.PrivateKeyFromMnemonic(mnemonic)
if err != nil {
    log.Fatalf("restoration failed: %v", err)
}

// Custom derivation path
privateKey, err := keylib.PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'/0/1")

// With passphrase (25th word)
privateKey, err := keylib.PrivateKeyFromMnemonicWithPassphrase(mnemonic, "my-secret-passphrase")

// Full options
privateKey, err := keylib.PrivateKeyFromMnemonicWithOptions(mnemonic, "passphrase", "m/44'/60'/0'/0/0")
```

#### Validate Mnemonic

```go
err := keylib.ValidateMnemonic(mnemonic)
if err != nil {
    // Possible errors:
    // - Invalid word count
    // - Unknown word in mnemonic
    // - Invalid checksum
    log.Fatalf("invalid mnemonic: %v", err)
}
```

---

### 7. Key Encryption (Scramble/Unscramble)

#### Encrypt Private Key for Storage

```go
password := "strong-password-here"

// Scramble encrypts the key with Argon2id + AES-256-GCM
encrypted, err := privateKey.Scramble(password)
if err != nil {
    log.Fatalf("encryption failed: %v", err)
}

// Serialize to JSON for storage
jsonBytes, _ := json.Marshal(encrypted)
// Store jsonBytes securely
```

#### Decrypt Private Key

```go
// Load from storage
var encrypted keylib.EncryptedPrivateKey
json.Unmarshal(storedBytes, &encrypted)

// Unscramble with password
privateKey, err := keylib.UnscramblePrivateKey(encrypted, password)
if err != nil {
    // Possible errors:
    // - Wrong password (GCM authentication fails)
    // - Corrupted data
    // - Invalid format
    log.Fatalf("decryption failed: %v", err)
}
```

---

### 8. Key Serialization

#### Export Key to Various Formats

```go
// Private key to bytes and hex
rawBytes := privateKey.Bytes()              // [32]byte
hexWithPrefix := privateKey.Hex()           // "0x4c0883a69102937d..."
hexNoPrefix := privateKey.HexWithoutPrefix() // "4c0883a69102937d..."

// Public key formats
compressed := pubKey.CompressedBytes()      // [33]byte, prefix 0x02 or 0x03
uncompressed := pubKey.UncompressedBytes()  // [65]byte, prefix 0x04
hexCompressed := pubKey.Hex()               // "0x02759b048e7ccf..."
hexUncompressed := pubKey.HexUncompressed() // "0x04759b048e7ccf..."

// EVM Address formats
addrBytes := evmAddr.Bytes()                // [20]byte
addrHex := evmAddr.Hex()                    // "0xe364f2f1e5f4f03d..."
addrChecksum := evmAddr.ChecksumHex()       // "0xe364f2f1e5F4F03d..." (EIP-55)
```

---

### 9. Key Comparison and Matching

#### Check Key Relationships

```go
// Private key matches public key?
if privateKey.MatchesPublicKey(pubKey) {
    fmt.Println("keys are a pair")
}

// Key matches EVM address?
if privateKey.MatchesEVMAddress(evmAddr) {
    fmt.Println("key controls this address")
}

// Public key matches peer ID?
if pubKey.MatchesPeerID(peerID) {
    fmt.Println("key owns this peer ID")
}

// Direct equality check
if key1.Equal(key2) {
    fmt.Println("keys are identical")
}
```

#### Zero-Value Checks

```go
// Check if key is valid (not zero-value)
if privateKey.IsZero() {
    log.Fatal("key is invalid/uninitialized")
}

if evmAddr.IsZero() {
    log.Warn("this is the burn address")
}

if peerID.IsZero() {
    log.Fatal("peer ID is invalid")
}
```

---

### 10. Secure Memory Handling

#### Zeroize Sensitive Data

```go
// After using a private key, clear it from memory
defer privateKey.Zeroize()

// Use the key...
signature, _ := privateKey.SignMessage(message)

// When function returns, key memory is cleared
```

#### crypto.Signer Interface

```go
import "crypto"

// NeuronPrivateKey implements crypto.Signer
var signer crypto.Signer = privateKey

// Use with standard library TLS, x509, etc.
pubKey := signer.Public()
signature, err := signer.Sign(rand.Reader, digest, crypto.SHA256)
```

---

### 11. Error Handling Patterns

#### Check Error Kind

```go
privateKey, err := keylib.ParsePrivateKeyHex(input)
if err != nil {
    keyErr, ok := err.(*keylib.KeyError)
    if ok {
        switch keyErr.Kind {
        case keylib.ErrKindInvalidHex:
            fmt.Println("input contains non-hex characters")
        case keylib.ErrKindInvalidLength:
            fmt.Println("wrong number of characters")
        case keylib.ErrKindInvalidKey:
            fmt.Println("key value is invalid for secp256k1")
        case keylib.ErrKindUnsupportedKeyType:
            fmt.Println("Ed25519 keys are not supported")
        default:
            fmt.Printf("key error: %v\n", err)
        }
    }
}
```

#### Use Safe Methods for Critical Operations

```go
// Pattern for safety-critical code paths
func transferFunds(key NeuronPrivateKey, recipient EVMAddress) error {
    // Use safe methods to catch zero-value errors
    senderAddr, err := key.EVMAddressSafe()
    if err != nil {
        return fmt.Errorf("invalid sender key: %w", err)
    }

    if senderAddr.Equal(recipient) {
        return errors.New("cannot transfer to self")
    }

    if recipient.IsZero() {
        return errors.New("cannot transfer to burn address")
    }

    // Proceed with transaction...
    return nil
}
```

---

## QA/QC Enhancements (Beyond Specification)

The following enhancements were added during QA/QC review:

### 1. Ed25519 Detection Hardening

**Problem**: Ed25519 keys could slip through if their scalar was < secp256k1 N.

**Solution**: Three-tier detection in `PrivateKeyFromHedera()`:

1. DER encoding pattern detection
2. Raw bytes length check (64 vs 32)
3. secp256k1 scalar validation

### 2. Safe Method Variants

**Problem**: Zero keys silently returned dangerous values (burn address, empty keys).

**Solution**: Added `*Safe()` variants that return errors:

- `EVMAddressSafe()` - Prevents silent derivation to burn address
- `ToHederaPrivateKeySafe()` - Returns error instead of empty key
- `ToHederaPublicKeySafe()` - Returns error instead of empty key

### 3. Constant-Time Operations

**Implementation**: All sensitive comparisons use `crypto/subtle`:

- `constantTimeEqual()` - Byte slice comparison
- `secureZero()` - Memory clearing
- All `Matches*` and `Equal()` methods

**File**: [constant_time.go](keylib/constant_time.go)

---

## Concurrency & Thread Safety

All types and functions in the keylib package are designed to be **safe for concurrent use** by multiple goroutines, with one exception noted below.

### Thread Safety Summary

| Category                                                                               | Thread-Safe | Notes                                           |
| -------------------------------------------------------------------------------------- | ----------- | ----------------------------------------------- |
| All types (`NeuronPrivateKey`, `NeuronPublicKey`, `EVMAddress`, `PeerID`, `Signature`) | Yes         | Immutable after construction                    |
| All `Parse*` functions                                                                 | Yes         | Pure parsing, no shared state                   |
| All `*FromBytes` functions                                                             | Yes         | Pure construction                               |
| All `*FromHedera` functions                                                            | Yes         | Pure conversion                                 |
| All `To*` methods                                                                      | Yes         | Pure extraction                                 |
| All `Matches*` methods                                                                 | Yes         | Constant-time comparison                        |
| `GeneratePrivateKey()`                                                                 | Yes         | Uses thread-safe `crypto/rand`                  |
| `GenerateMnemonic()`                                                                   | Yes         | Uses thread-safe `crypto/rand`                  |
| `Scramble()`                                                                           | Yes         | Thread-safe but CPU-intensive                   |
| `UnscramblePrivateKey()`                                                               | Yes         | Thread-safe but CPU-intensive                   |
| **`Zeroize()`**                                                                        | **NO**      | **Mutates receiver - requires synchronization** |

### WARNING: Non-Thread-Safe Method

**`NeuronPrivateKey.Zeroize()`** is the **only non-thread-safe method** in the entire package.

```go
// WARNING: NOT safe for concurrent use!
func (k *NeuronPrivateKey) Zeroize()
```

**Why it's unsafe**: `Zeroize()` mutates the internal key pointer, setting it to `nil` after clearing the key bytes. If other goroutines are reading the key while `Zeroize()` is called, a data race occurs.

**Safe usage pattern**:

```go
// Option 1: Use defer in single-goroutine context
func processKey(key NeuronPrivateKey) {
    defer key.Zeroize() // Safe - no concurrent access
    // ... use key
}

// Option 2: Use mutex for shared key
type SafeKey struct {
    mu  sync.Mutex
    key keylib.NeuronPrivateKey
}

func (sk *SafeKey) Zeroize() {
    sk.mu.Lock()
    defer sk.mu.Unlock()
    sk.key.Zeroize()
}
```

### CPU-Intensive Operations (Goroutine Warnings)

The following functions are **thread-safe** but **CPU and memory intensive**. When using them in goroutines, consider limiting concurrency to prevent resource exhaustion.

| Function                    | Memory    | CPU Time  | Recommendation            |
| --------------------------- | --------- | --------- | ------------------------- |
| `Scramble()`                | **64 MB** | 100-500ms | Limit concurrent calls    |
| `UnscramblePrivateKey()`    | **64 MB** | 100-500ms | Limit concurrent calls    |
| `PrivateKeyFromMnemonic*()` | Low       | <10ms     | Safe for high concurrency |
| `RecoverPublicKey*()`       | Low       | <1ms      | Safe for high concurrency |

**WARNING: Concurrent Scramble/Unscramble**

Each call to `Scramble()` or `UnscramblePrivateKey()` allocates **64 MB of memory** for Argon2id key derivation. Running many concurrent calls can cause:

- Memory exhaustion (OOM)
- System slowdown
- Increased latency

**Recommended pattern for bulk operations**:

```go
// Use a semaphore to limit concurrent Scramble/Unscramble calls
sem := make(chan struct{}, runtime.NumCPU())

for _, key := range keysToEncrypt {
    sem <- struct{}{} // Acquire
    go func(k NeuronPrivateKey) {
        defer func() { <-sem }() // Release
        encrypted, err := k.Scramble(password)
        // ... handle result
    }(key)
}
```

### Blocking I/O Analysis

| Function                 | Blocking I/O | Source                                                |
| ------------------------ | ------------ | ----------------------------------------------------- |
| `GeneratePrivateKey()`   | Minimal      | `crypto/rand.Reader` (non-blocking on modern systems) |
| `GenerateMnemonic()`     | Minimal      | `crypto/rand.Reader` via bip39                        |
| `Scramble()`             | Minimal      | `crypto/rand.Reader` for salt/nonce (28 bytes)        |
| `Sign()` (crypto.Signer) | Minimal      | `crypto/rand.Reader`                                  |
| All other functions      | **None**     | Pure CPU computation                                  |

**Note**: `crypto/rand.Reader` uses `/dev/urandom` on Unix and `CryptGenRandom` on Windows, which are non-blocking in practice. Blocking only occurs on system boot before entropy pool initialization (extremely rare).

### Global State

The keylib package has **no global mutable state**:

| Global Variable           | Type         | Mutated           | Thread-Safe |
| ------------------------- | ------------ | ----------------- | ----------- |
| `secp256k1N`              | `*big.Int`   | Never (init only) | Yes         |
| `ZeroEVMAddress`          | `EVMAddress` | Never             | Yes         |
| `ZeroSignature`           | `Signature`  | Never             | Yes         |
| `ValidMnemonicWordCounts` | `[]int`      | Never             | Yes         |
| `DefaultDerivationPath`   | `string`     | Never (const)     | Yes         |
| `ErrWrongPassword`        | `error`      | Never             | Yes         |

### Goroutine Usage Examples

**Safe - All these can run concurrently without synchronization**:

```go
// Multiple goroutines can safely:
go func() { pubKey := privKey.PublicKey() }()           // Safe
go func() { addr := pubKey.EVMAddress() }()            // Safe
go func() { valid := pubKey.Verify(msg, sig) }()       // Safe
go func() { key, _ := keylib.ParsePrivateKeyHex(h) }() // Safe
go func() { sig, _ := privKey.SignMessage(msg) }()     // Safe
```

**Unsafe - Requires external synchronization**:

```go
// DON'T DO THIS - data race!
go func() { privKey.Zeroize() }() // UNSAFE: Writing
go func() { _ = privKey.Hex() }() // UNSAFE: Reading concurrently

// DO THIS INSTEAD
var mu sync.Mutex
go func() {
    mu.Lock()
    defer mu.Unlock()
    privKey.Zeroize()
}()
```

---

## File Index

| File                                        | Purpose                                   |
| ------------------------------------------- | ----------------------------------------- |
| [private_key.go](keylib/private_key.go)     | `NeuronPrivateKey` type and methods       |
| [public_key.go](keylib/public_key.go)       | `NeuronPublicKey` type and methods        |
| [signature.go](keylib/signature.go)         | `Signature` type and recovery             |
| [evm_address.go](keylib/evm_address.go)     | `EVMAddress` type with EIP-55             |
| [peer_id.go](keylib/peer_id.go)             | `PeerID` wrapper for libp2p               |
| [factory.go](keylib/factory.go)             | All Parse*/From*/Generate functions       |
| [mnemonic.go](keylib/mnemonic.go)           | BIP39/BIP32 mnemonic support              |
| [encrypted_key.go](keylib/encrypted_key.go) | Scramble/Unscramble with Argon2id+AES-GCM |
| [errors.go](keylib/errors.go)               | `KeyError` type hierarchy                 |
| [validation.go](keylib/validation.go)       | Internal validation helpers               |
| [constant_time.go](keylib/constant_time.go) | Secure comparison utilities               |

---

## Test Coverage

**Current Coverage**: 85.4%

| Test File                                             | Tests              |
| ----------------------------------------------------- | ------------------ |
| [private_key_test.go](keylib/private_key_test.go)     | 45+ tests          |
| [public_key_test.go](keylib/public_key_test.go)       | 50+ tests          |
| [factory_test.go](keylib/factory_test.go)             | 40+ tests          |
| [signature_test.go](keylib/signature_test.go)         | 25+ tests          |
| [mnemonic_test.go](keylib/mnemonic_test.go)           | 15+ tests          |
| [encrypted_key_test.go](keylib/encrypted_key_test.go) | 15+ tests          |
| [evm_address_test.go](keylib/evm_address_test.go)     | 20+ tests          |
| [peer_id_test.go](keylib/peer_id_test.go)             | 15+ tests          |
| [validation_test.go](keylib/validation_test.go)       | 25+ tests          |
| [constant_time_test.go](keylib/constant_time_test.go) | 30+ tests          |
| [integration_test.go](keylib/integration_test.go)     | Known test vectors |

---

## Documentation Commands

### Viewing Package Documentation

The keylib package includes comprehensive godoc documentation in [doc.go](keylib/doc.go). Use these commands to view documentation:

#### View Full Package Documentation

```bash
# View complete package documentation
go doc -all ./keylib

# View package overview only
go doc ./keylib

# Pipe to less for easier reading
go doc -all ./keylib | less
```

#### View Specific Type Documentation

```bash
# View NeuronPrivateKey type and all its methods
go doc ./keylib NeuronPrivateKey

# View NeuronPublicKey type
go doc ./keylib NeuronPublicKey

# View EVMAddress type
go doc ./keylib EVMAddress

# View Signature type
go doc ./keylib Signature

# View EncryptedPrivateKey type
go doc ./keylib EncryptedPrivateKey

# View PeerID type
go doc ./keylib PeerID
```

#### View Specific Function Documentation

```bash
# Key generation
go doc ./keylib GeneratePrivateKey
go doc ./keylib GenerateMnemonic

# Parsing functions
go doc ./keylib ParsePrivateKeyHex
go doc ./keylib ParsePublicKeyHex
go doc ./keylib ParseEVMAddress
go doc ./keylib ParsePeerID

# Hedera interop
go doc ./keylib PrivateKeyFromHedera
go doc ./keylib PublicKeyFromHedera

# Mnemonic functions
go doc ./keylib PrivateKeyFromMnemonic
go doc ./keylib PrivateKeyFromMnemonicWithOptions
go doc ./keylib ValidateMnemonic

# Encryption functions
go doc ./keylib UnscramblePrivateKey

# Signature recovery
go doc ./keylib RecoverPublicKey
```

#### View Method Documentation

```bash
# Private key methods
go doc ./keylib NeuronPrivateKey.SignMessage
go doc ./keylib NeuronPrivateKey.SignDigest
go doc ./keylib NeuronPrivateKey.Scramble
go doc ./keylib NeuronPrivateKey.Zeroize
go doc ./keylib NeuronPrivateKey.ToHederaPrivateKey

# Public key methods
go doc ./keylib NeuronPublicKey.Verify
go doc ./keylib NeuronPublicKey.EVMAddress
go doc ./keylib NeuronPublicKey.PeerID
go doc ./keylib NeuronPublicKey.ToHederaPublicKey
```

#### Start Local Documentation Server

```bash
# Start godoc server on port 6060
godoc -http=:6060

# Then open in browser:
# http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/keylib/
```

### Building and Testing

```bash
# Build the keylib package
go build ./keylib

# Run all tests
go test ./keylib -v

# Run tests with coverage
go test ./keylib -v -cover

# Generate coverage report
go test ./keylib -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test ./keylib -v -run TestPrivateKeyFromHex

# Run benchmarks
go test ./keylib -bench=.
```

### Linting and Verification

```bash
# Run go vet
go vet ./keylib

# Check formatting
gofmt -d ./keylib

# Format code
gofmt -w ./keylib

# Run staticcheck (if installed)
staticcheck ./keylib
```

---

## Quick Start Guide

### Installation

```bash
# Add keylib to your project
go get github.com/aspect-build/neuron-go-hedera-sdk/keylib
```

### Import Statement

```go
import "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
```

### Minimal Working Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

func main() {
    // 1. Generate a new private key
    privateKey, err := keylib.GeneratePrivateKey()
    if err != nil {
        log.Fatal(err)
    }
    defer privateKey.Zeroize() // Clean up when done

    // 2. Derive public key and addresses
    publicKey := privateKey.PublicKey()
    evmAddress := publicKey.EVMAddress()
    peerID, _ := publicKey.PeerID()

    fmt.Printf("EVM Address: %s\n", evmAddress.ChecksumHex())
    fmt.Printf("Peer ID: %s\n", peerID.String())

    // 3. Sign a message
    message := []byte("Hello, Neuron!")
    signature, err := privateKey.SignMessage(message)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Verify the signature
    if publicKey.Verify(message, signature) {
        fmt.Println("Signature verified successfully")
    }

    // 5. Encrypt key for storage
    encrypted, err := privateKey.Scramble("secure-password")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Encrypted key version: %d\n", encrypted.Version)
}
```

### Common Workflows

#### Workflow 1: Import Existing Key

```go
// From hex string
privateKey, err := keylib.ParsePrivateKeyHex("0x4c0883a69102937d...")
if err != nil {
    log.Fatal(err)
}

// From raw bytes
var rawBytes [32]byte
copy(rawBytes[:], someSource)
privateKey, err := keylib.PrivateKeyFromBytes(rawBytes)
```

#### Workflow 2: Restore from Mnemonic

```go
// Generate new mnemonic
mnemonic, err := keylib.GenerateMnemonic(12)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Save this mnemonic:", mnemonic)

// Restore key later
privateKey, err := keylib.PrivateKeyFromMnemonic(mnemonic)
if err != nil {
    log.Fatal(err)
}
```

#### Workflow 3: Hedera Integration

```go
import hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

// Convert Neuron key to Hedera key
hederaKey, err := neuronPrivateKey.ToHederaPrivateKeySafe()
if err != nil {
    log.Fatal(err)
}

// Use with Hedera client
client, _ := hiero.ClientForTestnet()
client.SetOperator(accountID, hederaKey)
```

#### Workflow 4: Secure Key Storage

```go
// Encrypt for storage
encrypted, err := privateKey.Scramble("user-password")
if err != nil {
    log.Fatal(err)
}

// Serialize to JSON
jsonData, _ := json.Marshal(encrypted)
// Store jsonData to database or file

// Later: restore from storage
var loaded keylib.EncryptedPrivateKey
json.Unmarshal(jsonData, &loaded)

restored, err := keylib.UnscramblePrivateKey(loaded, "user-password")
if err != nil {
    log.Fatal(err)
}
```

#### Workflow 5: Signature Verification

```go
// Verify signature from known public key
isValid := publicKey.Verify(message, signature)

// Recover public key from signature (unknown signer)
recoveredKey, err := keylib.RecoverPublicKey(message, signature)
if err != nil {
    log.Fatal(err)
}

// Check if recovered key matches expected
if recoveredKey.Equal(expectedPublicKey) {
    fmt.Println("Signature is from expected signer")
}
```

---

## doc.go Reference

The [doc.go](keylib/doc.go) file contains the package-level documentation that appears when running `go doc ./keylib`. It documents:

### Package Overview

- Purpose: Type-safe cryptographic key management for Neuron SDK
- Core types: `NeuronPrivateKey`, `NeuronPublicKey`, `EVMAddress`, `PeerID`, `Signature`, `EncryptedPrivateKey`
- Design philosophy: Type safety, validation at boundaries, no panics, zero-value safety

### Documented Sections in doc.go

| Section                 | Description                                             |
| ----------------------- | ------------------------------------------------------- |
| Overview                | Package purpose and core capabilities                   |
| Core Types              | List of main types with brief descriptions              |
| Key Strategy            | ECDSA secp256k1 standardization rationale               |
| Design Philosophy       | Principles: type safety, validation, no panics          |
| Basic Usage             | Code examples for generation, parsing, derivation       |
| Mnemonic Support        | BIP39 mnemonic generation and restoration               |
| Key Protection          | Scramble/Unscramble for secure storage                  |
| Hedera Integration      | Conversion to/from Hedera SDK types                     |
| Error Handling          | KeyError type with Op, Kind, Details fields             |
| Security Considerations | Zeroize, Scramble, logging warnings                     |
| Concurrency             | Thread safety guarantees and exceptions                 |
| Blocking Operations     | I/O characteristics and CPU-intensive function warnings |

### Viewing doc.go Content

```bash
# View the raw doc.go file
cat keylib/doc.go

# View formatted package documentation
go doc ./keylib

# View with all exported symbols
go doc -all ./keylib | head -200
```

### Key Documentation Annotations

The following godoc annotations are used throughout the codebase:

```go
// # Section Header
// Used to create sections in godoc output

// [TypeName] or [FunctionName]
// Creates clickable links in godoc

// Deprecated: reason
// Marks deprecated functionality

// BUG(author): description
// Documents known bugs
```

### Example Documentation Pattern

```go
// FunctionName does X and returns Y.
// Detailed description of behavior.
//
// # Concurrency
//
// Thread safety information.
//
// # Blocking
//
// I/O characteristics.
func FunctionName(param Type) (Result, error) {
    // implementation
}
```
