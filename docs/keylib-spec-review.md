# keylib Specification Review: Documentation vs Implementation Discrepancies

**Document Under Review:** `docs/key lib spec.md`
**Review Date:** 2026-01-23
**Reviewer:** Tech Lead
**Status:** Action Required

---

## Executive Summary

This document identifies discrepancies between the keylib specification draft and the actual source code implementation. Several critical issues must be addressed before the documentation can be published.

| Severity | Count | Description |
|----------|-------|-------------|
| 🔴 Critical | 1 | Completely incorrect technical specification |
| 🟠 Major | 4 | Missing features or significant inaccuracies |
| 🟡 Minor | 6 | Missing details or clarifications needed |

---

## 🔴 Critical Issues

### 1. FR-015: Encryption Standard is WRONG

**Location:** Requirements → Functional Requirements → FR-015

**Documentation States:**
> "The library MUST provide methods to encrypt (scramble) and decrypt (unscramble) private keys using password-based encryption following the **BIP-38 standard (scrypt key derivation function with AES-256-CBC encryption)**"

**Actual Implementation:**
```go
// keylib/encrypted_key.go

// Uses Argon2id for key derivation (NOT scrypt)
encKey := argon2.IDKey(
    []byte(password),
    salt,
    cfg.argon2Time,    // Configurable: 1-100, default 3
    cfg.argon2Memory,  // Configurable: 8-256 MiB, default 64 MiB
    cfg.argon2Threads, // Configurable: 1-32, default 4
    argon2KeyLen,      // 32 bytes for AES-256
)

// Uses AES-256-GCM (NOT AES-256-CBC)
gcm, err := cipher.NewGCM(block)
ciphertext := gcm.Seal(nil, nonce, keyBytes[:], nil)
```

**Why This Matters:**
- BIP-38 uses scrypt + AES-CBC (unauthenticated encryption)
- Implementation uses Argon2id + AES-GCM (authenticated encryption)
- These are fundamentally different cryptographic approaches
- GCM provides integrity protection; CBC does not
- Argon2id is the modern OWASP-recommended KDF

**Recommended Fix:**
```markdown
- **FR-015**: The library MUST provide methods to encrypt (scramble) and
  decrypt (unscramble) private keys using password-based encryption with:
  - **Argon2id** key derivation function (memory-hard, GPU-resistant)
  - **AES-256-GCM** authenticated encryption (provides confidentiality and integrity)
  - Configurable Argon2 parameters (time, memory, threads) via functional options
  - Version-based backward compatibility (v1: fixed defaults, v2: stored parameters)
```

---

## 🟠 Major Issues

### 2. Missing: Functional Options Pattern

**Location:** Not documented anywhere

**What's Missing:**
The library uses the modern Go functional options pattern for extensible APIs, but this is not documented.

**Actual Implementation:**
```go
// keylib/mnemonic.go
type MnemonicOption func(*mnemonicConfig)

func WithPassphrase(passphrase string) MnemonicOption
func WithDerivationPath(path string) MnemonicOption

// Usage:
key, err := PrivateKeyFromMnemonic(mnemonic,
    WithPassphrase("secret"),
    WithDerivationPath("m/44'/0'/0'/0/0"),
)

// keylib/encrypted_key.go
type ScrambleOption func(*scrambleConfig)

func WithArgon2Time(t uint32) ScrambleOption
func WithArgon2Memory(m uint32) ScrambleOption
func WithArgon2Threads(t uint8) ScrambleOption

// Usage:
encrypted, err := key.Scramble(password,
    WithArgon2Time(5),
    WithArgon2Memory(128*1024),
)
```

**Recommended Addition:**
```markdown
- **FR-023**: The library MUST use the functional options pattern for
  configurable operations to enable extensible API design without breaking
  changes. This applies to:
  - Mnemonic derivation: `WithPassphrase()`, `WithDerivationPath()`
  - Key encryption: `WithArgon2Time()`, `WithArgon2Memory()`, `WithArgon2Threads()`
```

---

### 3. Missing: Encryption Version History

**Location:** Not documented in Key Entities or Requirements

**What's Missing:**
The encryption scheme has versioning for backward compatibility.

**Actual Implementation:**
```go
// keylib/encrypted_key.go

// Version History:
//   - Version 1: Original format, uses hardcoded default Argon2 parameters
//   - Version 2: Stores Argon2 parameters for custom encryption settings
type EncryptedPrivateKey struct {
    Version       int    `json:"version"`
    Salt          []byte `json:"salt"`
    Nonce         []byte `json:"nonce"`
    Ciphertext    []byte `json:"ciphertext"`
    // Version 2+ fields:
    Argon2Time    uint32 `json:"argon2Time,omitempty"`
    Argon2Memory  uint32 `json:"argon2Memory,omitempty"`
    Argon2Threads uint8  `json:"argon2Threads,omitempty"`
}

const CurrentEncryptionVersion = 2
```

**Why This Matters:**
- Version 1 keys encrypted with defaults can still be decrypted
- Version 2 keys store their Argon2 parameters for reliable decryption
- Without this documentation, users won't understand backward compatibility

**Recommended Addition to Key Entities:**
```markdown
- **EncryptedPrivateKey**: A JSON-serializable structure containing:
  - `Version`: Encryption scheme version (1 or 2)
  - `Salt`: 16-byte random salt for Argon2id
  - `Nonce`: 12-byte random nonce for AES-GCM
  - `Ciphertext`: 48-byte encrypted key + GCM authentication tag
  - `Argon2Time`, `Argon2Memory`, `Argon2Threads`: KDF parameters (Version 2+)

  **Version History:**
  - Version 1: Legacy format with hardcoded Argon2 defaults
  - Version 2: Stores Argon2 parameters for custom settings
```

---

### 4. Missing: Signature Ethereum Format Support

**Location:** FR-014 and Key Entities → Signature

**What's Missing:**
The `Signature` type provides Ethereum-specific format methods.

**Actual Implementation:**
```go
// keylib/signature.go

// Standard format (V = 0 or 1)
func (s Signature) V() byte           // Returns 0 or 1
func (s Signature) RSV() (r, s *big.Int, v byte)

// Ethereum legacy format (V = 27 or 28)
func (s Signature) VEthereum() byte   // Returns 27 or 28
func (s Signature) EthereumBytes() []byte  // Returns 65 bytes with V+27
```

**Recommended Update to FR-014:**
```markdown
- **FR-014**: The library MUST provide a standard signing interface that
  NeuronPrivateKey implements, producing ECDSA signatures with recovery ID
  in R||S||V format (65 bytes total). The library MUST support both:
  - Standard format: V = 0 or 1 (for non-Ethereum use)
  - Ethereum legacy format: V = 27 or 28 (for Ethereum compatibility)
```

---

### 5. Missing: crypto.Signer Interface Implementation

**Location:** FR-014 (partial mention)

**What's Missing:**
`NeuronPrivateKey` implements Go's `crypto.Signer` interface.

**Actual Implementation:**
```go
// keylib/private_key.go

// NeuronPrivateKey implements crypto.Signer
func (k NeuronPrivateKey) Public() crypto.PublicKey
func (k NeuronPrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
```

**Why This Matters:**
- Enables use with any Go library that accepts `crypto.Signer`
- Standard TLS libraries, certificate signing, etc.
- Important for interoperability

**Recommended Addition:**
```markdown
- **FR-024**: NeuronPrivateKey MUST implement the `crypto.Signer` interface
  from Go's standard library (`crypto.Signer`), enabling interoperability
  with any cryptographic library that accepts this interface (TLS, X.509, etc.)
```

---

## 🟡 Minor Issues

### 6. FR-001: Public Key Format Clarification

**Location:** Requirements → Functional Requirements → FR-001

**Documentation States:**
> "NeuronPublicKey MUST use compressed format (33 bytes) as the primary representation in the public API, with uncompressed format (65 bytes) handled internally"

**Actual Implementation:**
Both formats are exposed in the public API:
```go
func (k NeuronPublicKey) CompressedBytes() [33]byte
func (k NeuronPublicKey) UncompressedBytes() [65]byte
func (k NeuronPublicKey) Hex() string          // Compressed
func (k NeuronPublicKey) HexUncompressed() string  // Uncompressed
```

**Recommended Clarification:**
```markdown
NeuronPublicKey MUST provide both compressed (33 bytes) and uncompressed
(65 bytes) representations via explicit methods. The default serialization
(`Hex()`) SHOULD use compressed format for efficiency.
```

---

### 7. Missing: Must-Parse Functions

**Location:** Not documented

**Actual Implementation:**
```go
func MustParsePrivateKeyHex(s string) NeuronPrivateKey  // Panics on error
func MustParsePublicKeyHex(s string) NeuronPublicKey    // Panics on error
func MustParseEVMAddress(s string) EVMAddress           // Panics on error
func MustParsePeerID(s string) PeerID                   // Panics on error
```

**Recommended Addition:**
```markdown
The library SHOULD provide `Must*` variants of parsing functions for use in
initialization code where invalid input represents a programming error.
These functions panic on invalid input and should only be used with
compile-time constants or validated configuration.
```

---

### 8. Incomplete: Error Kinds List

**Location:** FR-008 mentions error kinds but doesn't list all of them

**Documentation Mentions:**
> "specific error kinds (e.g., InvalidFormat, InvalidLength, InvalidKey)"

**Actual Implementation Has 10 Error Kinds:**
```go
// keylib/errors.go
const (
    ErrKindInvalidFormat      // Wrong format or encoding
    ErrKindInvalidLength      // Wrong byte/character length
    ErrKindInvalidHex         // Invalid hexadecimal characters
    ErrKindInvalidKey         // Key validation failed
    ErrKindZeroValue          // Operation on zero-value type
    ErrKindKeyMismatch        // Keys don't match
    ErrKindEncryption         // Encryption/decryption failure
    ErrKindMnemonic           // Invalid mnemonic phrase
    ErrKindDerivation         // Key derivation failure
    ErrKindUnsupportedKeyType // Ed25519 or other non-secp256k1
)
```

**Recommended:** Document all 10 error kinds in the specification.

---

### 9. Missing: Deprecated Functions

**Location:** Not documented

**Actual Implementation:**
```go
// Deprecated: Use PrivateKeyFromMnemonic with WithDerivationPath option
func PrivateKeyFromMnemonicWithPath(mnemonic, path string) (NeuronPrivateKey, error)

// Deprecated: Use PrivateKeyFromMnemonic with WithPassphrase option
func PrivateKeyFromMnemonicWithPassphrase(mnemonic, passphrase string) (NeuronPrivateKey, error)

// Deprecated: Use PrivateKeyFromMnemonic with options
func PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, path string) (NeuronPrivateKey, error)
```

**Recommended:** Add a deprecation notice section explaining migration paths.

---

### 10. Missing: EIP-55 Checksum Support

**Location:** Not documented in EVMAddress entity

**Actual Implementation:**
```go
func (a EVMAddress) Hex() string         // Lowercase: 0xabcd...
func (a EVMAddress) ChecksumHex() string // EIP-55 mixed-case: 0xAbCd...
func (a EVMAddress) String() string      // Alias for ChecksumHex()
```

**Recommended Addition to Key Entities:**
```markdown
- **EVMAddress**: [...] The library MUST support EIP-55 checksummed addresses
  via `ChecksumHex()` method for Ethereum ecosystem compatibility.
```

---

### 11. FR-022: Zeroize() Concurrency Warning

**Location:** FR-022

**Documentation States:**
> "Neuron key types [...] MUST be immutable and thread-safe for concurrent read operations"

**Missing Caveat:**
`Zeroize()` is a **mutating operation** that is NOT thread-safe.

**Actual Implementation:**
```go
// keylib/private_key.go

// Zeroize clears the private key from memory.
// WARNING: This mutates the receiver and is NOT concurrent-safe.
// After calling Zeroize(), the key becomes invalid for all operations.
func (k *NeuronPrivateKey) Zeroize()
```

**Recommended Clarification:**
```markdown
FR-022: [...] thread-safe for concurrent read operations. The `Zeroize()`
method is an exception: it mutates the receiver and requires external
synchronization if other goroutines may access the same key instance.
```

---

## Summary of Required Changes

| Issue | Section | Action |
|-------|---------|--------|
| Wrong encryption standard | FR-015 | **Rewrite completely** |
| Missing functional options | New FR-023 | Add new requirement |
| Missing encryption versions | Key Entities | Add EncryptedPrivateKey details |
| Missing Ethereum signature format | FR-014 | Update existing |
| Missing crypto.Signer | New FR-024 | Add new requirement |
| Public key format | FR-001 | Clarify wording |
| Must-parse functions | New section | Add documentation |
| Error kinds incomplete | FR-008 | List all 10 kinds |
| Deprecated functions | New section | Add deprecation notice |
| EIP-55 checksum | Key Entities | Add to EVMAddress |
| Zeroize() caveat | FR-022 | Add warning |

---

## Appendix: Quick Reference - Actual Encryption Scheme

```
┌─────────────────────────────────────────────────────────────┐
│                 ACTUAL ENCRYPTION SCHEME                    │
├─────────────────────────────────────────────────────────────┤
│ Key Derivation Function: Argon2id (NOT scrypt/BIP-38)       │
│   - Time:    1-100 iterations (default: 3)                  │
│   - Memory:  8-256 MiB (default: 64 MiB)                    │
│   - Threads: 1-32 (default: 4)                              │
│   - Output:  32 bytes (256-bit AES key)                     │
├─────────────────────────────────────────────────────────────┤
│ Encryption: AES-256-GCM (NOT AES-256-CBC)                   │
│   - Nonce:      12 bytes (random)                           │
│   - Ciphertext: 32 bytes (encrypted key)                    │
│   - Auth Tag:   16 bytes (GCM integrity)                    │
│   - Total:      48 bytes                                    │
├─────────────────────────────────────────────────────────────┤
│ Salt: 16 bytes (random, stored with ciphertext)             │
├─────────────────────────────────────────────────────────────┤
│ Version Support:                                            │
│   - v1: Legacy (uses hardcoded Argon2 defaults)             │
│   - v2: Current (stores Argon2 parameters in JSON)          │
└─────────────────────────────────────────────────────────────┘
```
