# keylib Package

## Section 1: Overview

Package `keylib` provides type-safe cryptographic key management for the Neuron SDK.

**Import path:** `github.com/aspect-build/neuron-go-hedera-sdk/keylib`

The package standardizes on ECDSA secp256k1 keys for maximum interoperability across the Hedera, Ethereum, and libp2p ecosystems. A single private key controls Hedera accounts, Ethereum addresses, and libp2p PeerIDs. Ed25519 keys are explicitly rejected with clear error messages.

### Core Design Principles

| Principle                    | Description                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------- |
| **Type safety**              | Functions accept and return types, not strings.                                 |
| **Validation at boundaries** | All `Parse*` functions validate immediately upon construction.                  |
| **No panics**                | All fallible operations return `(T, error)`.                                    |
| **Zero-value safety**        | Zero values are invalid; methods return errors when invoked on them.            |
| **Constant-time operations** | All key comparisons use `subtle.ConstantTimeCompare` to prevent timing attacks. |
| **Memory safety**            | `Zeroize()` clears sensitive data from memory (best-effort due to GC).          |

All types are immutable after construction and safe for concurrent use by multiple goroutines. The only exception is `NeuronPrivateKey.Zeroize()`, which mutates the receiver and requires external synchronization.

---

## Section 2: Architecture & Cryptographic Model

### Type Hierarchy

Seven core types, all immutable after construction:

| #   | Type                  | Description                                                         | Internal Representation                                     |
| --- | --------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------- |
| 1   | `NeuronPrivateKey`    | ECDSA secp256k1 private key (32 bytes). Implements `crypto.Signer`. | `*secp256k1.PrivateKey`                                     |
| 2   | `NeuronPublicKey`     | ECDSA secp256k1 public key.                                         | `*secp256k1.PublicKey`                                      |
| 3   | `EVMAddress`          | 20-byte Ethereum address. Supports EIP-55 checksum.                 | `[20]byte`                                                  |
| 4   | `PeerID`              | libp2p peer identifier.                                             | `peer.ID` from libp2p                                       |
| 5   | `Signature`           | 65-byte ECDSA signature (R \|\| S \|\| V).                          | `[65]byte`                                                  |
| 6   | `EncryptedPrivateKey` | Password-protected private key. Uses Argon2id + AES-256-GCM.        | Struct with Version, Salt, Nonce, Ciphertext, Argon2 params |
| 7   | `MultisigKey`         | M-of-N threshold configuration. Stores sorted public keys.          | Struct with sorted `[]NeuronPublicKey`, threshold, total    |

### Key Derivation Chain

```
NeuronPrivateKey
  |-- PublicKey() --> NeuronPublicKey
  |     |-- EVMAddress() --> EVMAddress (Keccak256 of uncompressed pubkey, last 20 bytes)
  |     |-- PeerID() --> PeerID (libp2p from compressed bytes)
  |     |-- CompressedBytes() --> [33]byte (SEC1 compressed)
  |     +-- UncompressedBytes() --> [65]byte (SEC1 uncompressed)
  |-- Bytes() --> [32]byte (raw scalar)
  |-- Hex() --> "0x..." (with prefix)
  |-- ToECDSA() --> *ecdsa.PrivateKey
  |-- ToHederaPrivateKey() --> hiero.PrivateKey
  +-- SignMessage(msg) --> Signature
```

### Signature Format

65 bytes: **R** (32 bytes) || **S** (32 bytes) || **V** (1 byte)

- **V** is the recovery ID (0 or 1), used for public key recovery.
- `VEthereum()` returns V + 27 (Ethereum legacy format: 27 or 28).
- `EthereumBytes()` returns R || S || (V+27) as a `[]byte` of length 65.
- **Low-S normalization** is enforced (EIP-2): if S > N/2, it is replaced with N - S and V is flipped.
- `ParseSignature` accepts a hex string (130 or 132 chars with `0x` prefix).
- `SignatureFromBytes` accepts a 65-byte slice.
- `RecoverPublicKey(msg, sig)` recovers the signer's public key (hashes msg with Keccak256 first).
- `RecoverPublicKeyFromDigest(digest, sig)` recovers from a pre-hashed 32-byte digest.

### Encryption Model (EncryptedPrivateKey)

Two-layer encryption:

1. **Key derivation:** Argon2id (password --> 32-byte AES key)
   - Default params: time=3, memory=64 MiB (64\*1024 KiB), threads=4
   - Configurable via `ScrambleOption`: `WithArgon2Time()`, `WithArgon2Memory()`, `WithArgon2Threads()`
   - Parameter bounds: time 1-100, memory 8-64 MiB, threads 1-32
2. **Encryption:** AES-256-GCM (authenticated encryption)
   - Random 16-byte salt, random 12-byte nonce
   - Ciphertext = 32 (key) + 16 (GCM tag) = 48 bytes
   - `CurrentEncryptionVersion = 2`

`EncryptedPrivateKey` struct fields:

| Field           | Type     | JSON Tag                    | Description                        |
| --------------- | -------- | --------------------------- | ---------------------------------- |
| `Version`       | `int`    | `"version"`                 | Encryption scheme version (1 or 2) |
| `Salt`          | `[]byte` | `"salt"`                    | 16-byte Argon2id salt              |
| `Nonce`         | `[]byte` | `"nonce"`                   | 12-byte AES-GCM nonce              |
| `Ciphertext`    | `[]byte` | `"ciphertext"`              | 48 bytes (32 key + 16 GCM tag)     |
| `Argon2Time`    | `uint32` | `"argon2Time,omitempty"`    | Iterations (Version 2+)            |
| `Argon2Memory`  | `uint32` | `"argon2Memory,omitempty"`  | Memory in KiB (Version 2+)         |
| `Argon2Threads` | `uint8`  | `"argon2Threads,omitempty"` | Parallelism (Version 2+)           |

Version history:

- **Version 1:** Legacy format. Uses hardcoded default Argon2 parameters for decryption.
- **Version 2:** Stores Argon2 parameters in the struct for reliable decryption with custom settings.

`MarshalJSON` / `UnmarshalJSON` encode binary fields as base64 for JSON compatibility.

Operations:

- `(NeuronPrivateKey).Scramble(password string, opts ...ScrambleOption) (EncryptedPrivateKey, error)`
- `UnscramblePrivateKey(encrypted EncryptedPrivateKey, password string) (NeuronPrivateKey, error)`
- Wrong password returns `ErrWrongPassword` sentinel error.

### MultisigKey

```go
type MultisigKey struct {
    publicKeys []NeuronPublicKey // sorted by compressed bytes for determinism
    threshold  int               // M (minimum signers required)
    total      int               // N (total number of signers)
}
```

- `NewMultisigKey(publicKeys []NeuronPublicKey, threshold int) (MultisigKey, error)` -- validates inputs, checks for duplicates, creates a defensive copy, and sorts keys by compressed bytes.
- `MustNewMultisigKey(publicKeys []NeuronPublicKey, threshold int) MultisigKey` -- panics on error (intended for tests).

Validation rules:

- `publicKeys` must not be empty.
- `threshold` must be >= 1.
- `threshold` must be <= `len(publicKeys)`.
- All public keys must be non-zero.
- No duplicate public keys allowed (detected via hex comparison).

Keys are stored in sorted order (lexicographic by compressed bytes), producing deterministic behavior regardless of input order.

Methods: `IsZero()`, `Validate()`, `Threshold()`, `Total()`, `PublicKeys()` (returns defensive copy), `ContainsKey()`, `Equal()`, `String()`.

### BIP39/BIP44 Mnemonic Support

- `GenerateMnemonic(wordCount int) (string, error)` -- `wordCount` must be 12, 15, 18, 21, or 24.
- `PrivateKeyFromMnemonic(mnemonic string, opts ...MnemonicOption) (NeuronPrivateKey, error)` -- the primary mnemonic derivation function with functional options.
- `PrivateKeyFromMnemonicWithPath(mnemonic, path string) (NeuronPrivateKey, error)` -- deprecated, wraps `PrivateKeyFromMnemonic`.
- `PrivateKeyFromMnemonicWithPassphrase(mnemonic, passphrase string) (NeuronPrivateKey, error)` -- deprecated, wraps `PrivateKeyFromMnemonic`.
- `PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, path string) (NeuronPrivateKey, error)` -- deprecated, wraps `PrivateKeyFromMnemonic`.
- `ValidateMnemonic(mnemonic string) error`
- `DefaultDerivationPath = "m/44'/60'/0'/0/0"` (Ethereum BIP44 path)
- `ValidMnemonicWordCounts = []int{12, 15, 18, 21, 24}`

Mnemonic options:

- `WithPassphrase(passphrase string) MnemonicOption`
- `WithDerivationPath(path string) MnemonicOption`

Mnemonics are normalized (trimmed, lowercased, single-spaced) before processing. BIP32 derivation uses HMAC-SHA512 with the "Bitcoin seed" master key.

---

## Section 3: Public API Reference

### Factory Functions (keylib/factory.go)

```go
func GeneratePrivateKey() (NeuronPrivateKey, error)
func ParsePrivateKeyHex(s string) (NeuronPrivateKey, error)
func PrivateKeyFromBytes(b [32]byte) (NeuronPrivateKey, error)
func PrivateKeyFromHedera(hederaKey hiero.PrivateKey) (NeuronPrivateKey, error)
func ParsePublicKeyHex(s string) (NeuronPublicKey, error)
func PublicKeyFromBytes(b []byte) (NeuronPublicKey, error)
func PublicKeyFromHedera(hederaKey hiero.PublicKey) (NeuronPublicKey, error)
func PublicKeyFromPrivateKey(privKey NeuronPrivateKey) NeuronPublicKey

// Must* variants (panic on error, for tests only)
func MustParsePrivateKeyHex(s string) NeuronPrivateKey
func MustParsePublicKeyHex(s string) NeuronPublicKey
func MustParseEVMAddress(s string) EVMAddress
func MustParsePeerID(s string) PeerID

// Ed25519 detection
func IsEd25519Key(hederaKey hiero.PrivateKey) bool
func IsEd25519PublicKey(hederaKey hiero.PublicKey) bool
```

### NeuronPrivateKey Methods

```go
func (k NeuronPrivateKey) IsZero() bool
func (k NeuronPrivateKey) PublicKey() NeuronPublicKey
func (k NeuronPrivateKey) EVMAddress() EVMAddress           // WARNING: returns burn address if zero
func (k NeuronPrivateKey) EVMAddressSafe() (EVMAddress, error)
func (k NeuronPrivateKey) PeerID() (PeerID, error)
func (k NeuronPrivateKey) Bytes() [32]byte                  // WARNING: secret material
func (k NeuronPrivateKey) Hex() string                      // WARNING: secret material, 0x prefix
func (k NeuronPrivateKey) HexWithoutPrefix() string
func (k NeuronPrivateKey) ToECDSA() *ecdsa.PrivateKey
func (k NeuronPrivateKey) ToHederaPrivateKey() hiero.PrivateKey
func (k NeuronPrivateKey) ToHederaPrivateKeySafe() (hiero.PrivateKey, error)
func (k NeuronPrivateKey) SignMessage(msg []byte) (Signature, error)
func (k NeuronPrivateKey) SignDigest(digest [32]byte) (Signature, error)
func (k NeuronPrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
func (k NeuronPrivateKey) Public() crypto.PublicKey
func (k NeuronPrivateKey) MatchesPublicKey(pub NeuronPublicKey) bool
func (k NeuronPrivateKey) MatchesEVMAddress(addr EVMAddress) bool
func (k *NeuronPrivateKey) Zeroize()                        // NOT concurrent-safe
func (k NeuronPrivateKey) Equal(other NeuronPrivateKey) bool
func (k NeuronPrivateKey) Scramble(password string, opts ...ScrambleOption) (EncryptedPrivateKey, error)
```

### NeuronPublicKey Methods

```go
func (k NeuronPublicKey) IsZero() bool
func (k NeuronPublicKey) CompressedBytes() [33]byte
func (k NeuronPublicKey) UncompressedBytes() [65]byte
func (k NeuronPublicKey) Hex() string                        // compressed, 0x prefix
func (k NeuronPublicKey) HexUncompressed() string
func (k NeuronPublicKey) EVMAddress() EVMAddress             // WARNING: burn address if zero
func (k NeuronPublicKey) EVMAddressSafe() (EVMAddress, error)
func (k NeuronPublicKey) PeerID() (PeerID, error)
func (k NeuronPublicKey) ToECDSA() *ecdsa.PublicKey
func (k NeuronPublicKey) ToHederaPublicKey() hiero.PublicKey
func (k NeuronPublicKey) ToHederaPublicKeySafe() (hiero.PublicKey, error)
func (k NeuronPublicKey) MatchesEVMAddress(addr EVMAddress) bool
func (k NeuronPublicKey) MatchesPeerID(pid PeerID) bool
func (k NeuronPublicKey) Verify(msg []byte, sig Signature) bool
func (k NeuronPublicKey) VerifyDigest(digest [32]byte, sig Signature) bool
func (k NeuronPublicKey) Equal(other NeuronPublicKey) bool   // constant-time
```

### EVMAddress

```go
var ZeroEVMAddress EVMAddress

func ParseEVMAddress(s string) (EVMAddress, error)
func EVMAddressFromCommon(addr common.Address) EVMAddress

func (a EVMAddress) IsZero() bool
func (a EVMAddress) Bytes() [20]byte
func (a EVMAddress) Hex() string          // lowercase, 0x prefix
func (a EVMAddress) ChecksumHex() string  // EIP-55
func (a EVMAddress) String() string       // same as ChecksumHex
func (a EVMAddress) Equal(other EVMAddress) bool
func (a EVMAddress) ToCommon() common.Address
```

### PeerID

```go
func ParsePeerID(s string) (PeerID, error)
func PeerIDFromLibp2p(id peer.ID) PeerID

func (p PeerID) IsZero() bool
func (p PeerID) String() string
func (p PeerID) Bytes() []byte
func (p PeerID) ToLibp2p() peer.ID
func (p PeerID) Equal(other PeerID) bool
func (p PeerID) Validate() error
```

### Signature

```go
var ZeroSignature Signature

func ParseSignature(s string) (Signature, error)
func SignatureFromBytes(b []byte) (Signature, error)
func RecoverPublicKey(msg []byte, sig Signature) (NeuronPublicKey, error)
func RecoverPublicKeyFromDigest(digest [32]byte, sig Signature) (NeuronPublicKey, error)

func (s Signature) IsZero() bool
func (s Signature) Bytes() []byte           // returns 65-byte slice (defensive copy)
func (s Signature) Hex() string
func (s Signature) RS() (r, s *big.Int)
func (s Signature) RSV() (r, s *big.Int, v byte)
func (s Signature) V() byte
func (s Signature) VEthereum() byte          // V + 27
func (s Signature) EthereumBytes() []byte    // R||S||(V+27), 65-byte slice
func (s Signature) Equal(other Signature) bool
```

### EncryptedPrivateKey

```go
const CurrentEncryptionVersion = 2

func UnscramblePrivateKey(encrypted EncryptedPrivateKey, password string) (NeuronPrivateKey, error)

// ScrambleOption functions
func WithArgon2Time(t uint32) ScrambleOption
func WithArgon2Memory(m uint32) ScrambleOption
func WithArgon2Threads(t uint8) ScrambleOption

func (e EncryptedPrivateKey) IsZero() bool
func (e EncryptedPrivateKey) MarshalJSON() ([]byte, error)
func (e *EncryptedPrivateKey) UnmarshalJSON(data []byte) error
func (e EncryptedPrivateKey) String() string
```

### MultisigKey

```go
func NewMultisigKey(publicKeys []NeuronPublicKey, threshold int) (MultisigKey, error)
func MustNewMultisigKey(publicKeys []NeuronPublicKey, threshold int) MultisigKey

func (m MultisigKey) IsZero() bool
func (m MultisigKey) Validate() error
func (m MultisigKey) Threshold() int
func (m MultisigKey) Total() int
func (m MultisigKey) PublicKeys() []NeuronPublicKey  // returns defensive copy
func (m MultisigKey) ContainsKey(key NeuronPublicKey) bool
func (m MultisigKey) Equal(other MultisigKey) bool
func (m MultisigKey) String() string
```

### Mnemonic

```go
const DefaultDerivationPath = "m/44'/60'/0'/0/0"
var ValidMnemonicWordCounts = []int{12, 15, 18, 21, 24}

func GenerateMnemonic(wordCount int) (string, error)
func PrivateKeyFromMnemonic(mnemonic string, opts ...MnemonicOption) (NeuronPrivateKey, error)
func PrivateKeyFromMnemonicWithPath(mnemonic, path string) (NeuronPrivateKey, error)               // deprecated
func PrivateKeyFromMnemonicWithPassphrase(mnemonic, passphrase string) (NeuronPrivateKey, error)    // deprecated
func PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, path string) (NeuronPrivateKey, error) // deprecated
func ValidateMnemonic(mnemonic string) error

func WithPassphrase(passphrase string) MnemonicOption
func WithDerivationPath(path string) MnemonicOption
```

### Error Types

```go
type ErrorKind int

const (
    ErrKindInvalidFormat    ErrorKind = iota + 1
    ErrKindInvalidLength
    ErrKindInvalidHex
    ErrKindInvalidKey
    ErrKindZeroValue
    ErrKindKeyMismatch
    ErrKindEncryption
    ErrKindMnemonic
    ErrKindDerivation
    ErrKindUnsupportedKeyType
    ErrKindInvalidThreshold
)

type KeyError struct {
    Op      string
    Kind    ErrorKind
    Details string
    Err     error
}

func (e *KeyError) Error() string    // "keylib.{Op}: {Kind}: {Details}: {Err}"
func (e *KeyError) Unwrap() error
func (e *KeyError) Is(target error) bool  // matches by Kind

var ErrWrongPassword = &KeyError{Kind: ErrKindEncryption, Details: "wrong password or corrupted data"}
```

---

## Section 4: Usage Guide

### Key Generation

```go
privKey, err := keylib.GeneratePrivateKey()
if err != nil {
    return err
}
pubKey := privKey.PublicKey()
```

### Parsing Keys

```go
// From hex (with or without 0x prefix)
privKey, err := keylib.ParsePrivateKeyHex("0xabc123...")
pubKey, err := keylib.ParsePublicKeyHex("0x02abc123...")

// From raw bytes
privKey, err := keylib.PrivateKeyFromBytes(bytes32)
pubKey, err := keylib.PublicKeyFromBytes(secBytes)

// From Hedera SDK
privKey, err := keylib.PrivateKeyFromHedera(hederaPrivKey)
pubKey, err := keylib.PublicKeyFromHedera(hederaPubKey)
```

### Deriving Addresses

```go
evmAddr := pubKey.EVMAddress()               // WARNING: burn address if zero
evmAddr, err := pubKey.EVMAddressSafe()      // safe version with error on zero
peerID, err := pubKey.PeerID()
```

### Signing and Verification

```go
// Sign a message (Keccak256-hashed internally)
sig, err := privKey.SignMessage([]byte("hello"))

// Sign a pre-hashed digest
sig, err := privKey.SignDigest(hash32)

// Verify a signature
valid := pubKey.Verify([]byte("hello"), sig)
valid := pubKey.VerifyDigest(hash32, sig)

// Recover the signer's public key
recovered, err := keylib.RecoverPublicKey([]byte("hello"), sig)
```

### Mnemonic Operations

```go
// Generate new mnemonic
mnemonic, err := keylib.GenerateMnemonic(12)

// Derive key from mnemonic (default path, no passphrase)
privKey, err := keylib.PrivateKeyFromMnemonic(mnemonic)

// With custom options
privKey, err := keylib.PrivateKeyFromMnemonic(mnemonic,
    keylib.WithPassphrase("my-passphrase"),
    keylib.WithDerivationPath("m/44'/60'/0'/0/1"),
)

// Validate mnemonic
err := keylib.ValidateMnemonic(mnemonic)
```

### Encryption / Decryption

```go
// Encrypt with default Argon2 params
encrypted, err := privKey.Scramble("password")
jsonData, _ := json.Marshal(encrypted)

// With custom Argon2 params
encrypted, err := privKey.Scramble("password",
    keylib.WithArgon2Time(4),
    keylib.WithArgon2Memory(32*1024),
)

// Decrypt
restored, err := keylib.UnscramblePrivateKey(encrypted, "password")
if errors.Is(err, keylib.ErrWrongPassword) {
    // wrong password
}
```

### Hedera Interop

```go
// To Hedera
hederaPrivKey := privKey.ToHederaPrivateKey()
hederaPubKey := pubKey.ToHederaPublicKey()

// Safe versions with error handling
hederaPrivKey, err := privKey.ToHederaPrivateKeySafe()
hederaPubKey, err := pubKey.ToHederaPublicKeySafe()

// From Hedera
neuronPriv, err := keylib.PrivateKeyFromHedera(hederaPrivKey)
neuronPub, err := keylib.PublicKeyFromHedera(hederaPubKey)

// Detect Ed25519 (rejected by all conversion functions)
if keylib.IsEd25519Key(hederaPrivKey) {
    // handle unsupported key type
}
```

### MultisigKey

```go
keys := make([]keylib.NeuronPublicKey, 3)
for i := range keys {
    priv, _ := keylib.GeneratePrivateKey()
    keys[i] = priv.PublicKey()
}

// Create 2-of-3 multisig
multisig, err := keylib.NewMultisigKey(keys, 2)

// Query
fmt.Println(multisig.Threshold()) // 2
fmt.Println(multisig.Total())     // 3
fmt.Println(multisig.ContainsKey(keys[0])) // true

// Keys are sorted deterministically by compressed bytes
sortedKeys := multisig.PublicKeys() // defensive copy
```

---

## Section 5: Error Handling & Security Notes

### Error Handling

All errors are `*KeyError` with structured fields:

| Field     | Description                                                                  |
| --------- | ---------------------------------------------------------------------------- |
| `Op`      | Operation name (e.g., `"ParsePrivateKeyHex"`, `"NeuronPrivateKey.Scramble"`) |
| `Kind`    | Error category (`ErrorKind`) for programmatic handling                       |
| `Details` | Human-readable explanation of what went wrong                                |
| `Err`     | Underlying error (optional, supports `errors.Unwrap`)                        |

Error format: `"keylib.{Op}: {Kind}: {Details}: {Err}"`

If `Err` is nil: `"keylib.{Op}: {Kind}: {Details}"`

If `Op` is empty: `"keylib: {Kind}: {Details}"`

Error matching by kind:

```go
var keyErr *keylib.KeyError
if errors.As(err, &keyErr) {
    switch keyErr.Kind {
    case keylib.ErrKindInvalidHex:
        // invalid hex input
    case keylib.ErrKindUnsupportedKeyType:
        // Ed25519 key rejected
    case keylib.ErrKindEncryption:
        // decryption failed
    case keylib.ErrKindInvalidThreshold:
        // invalid M-of-N configuration
    case keylib.ErrKindZeroValue:
        // operation on zero-value type
    }
}

// Sentinel matching
if errors.Is(err, keylib.ErrWrongPassword) {
    // wrong password or corrupted data
}
```

The `Is` method on `KeyError` matches by `Kind`, so `errors.Is(err, ErrWrongPassword)` succeeds for any `*KeyError` with `Kind == ErrKindEncryption`.

### Security Notes

1. **Constant-time comparisons.** All `Equal()` methods on `NeuronPrivateKey`, `NeuronPublicKey`, `EVMAddress`, `PeerID`, and `Signature` use `subtle.ConstantTimeCompare`. Never use `==` or `bytes.Equal` for key comparison in security-sensitive contexts.

2. **Zeroize private keys.** Call `privKey.Zeroize()` when done with a private key. This zeros the internal key bytes and sets the internal pointer to nil. Due to Go's garbage collector, complete removal from memory is not guaranteed; this is a best-effort approach. `Zeroize()` is NOT concurrent-safe -- ensure no other goroutines are accessing the key when calling it.

3. **EVMAddress from zero key.** `EVMAddress()` on a zero-value key returns `ZeroEVMAddress` (`0x0000...0000`), which is the Ethereum burn address. Funds sent to this address are lost forever. Use `EVMAddressSafe()` to get an error instead.

4. **ToHederaPrivateKey silent failure.** `ToHederaPrivateKey()` silently returns an empty Hedera key on zero-value or conversion failure. Use `ToHederaPrivateKeySafe()` for explicit error handling. The same pattern applies to `ToHederaPublicKey()` / `ToHederaPublicKeySafe()`.

5. **Ed25519 rejection.** All Hedera conversion functions (`PrivateKeyFromHedera`, `PublicKeyFromHedera`) reject Ed25519 keys with `ErrKindUnsupportedKeyType`. Detection uses DER encoding pattern matching (Ed25519 OID `2b6570`) and raw byte length checks (Ed25519 private = 64 bytes, public = 32 bytes). Use `IsEd25519Key()` / `IsEd25519PublicKey()` for pre-checking.

6. **Hex input length limit.** `maxHexInputLen = 256` characters. The `decodeHexStrict` function rejects oversized input before any processing to prevent DoS via extremely long strings.

7. **Argon2id parameters.** Defaults: time=3, memory=64 MiB, threads=4. Parameter bounds enforced: time 1-100, memory 8-64 MiB, threads 1-32. `Scramble` / `UnscramblePrivateKey` are CPU and memory intensive (~100-500ms, 64 MiB allocation per call). Limit concurrent calls in high-throughput scenarios using a semaphore or worker pool.

8. **MultisigKey determinism.** Keys are sorted by compressed bytes (lexicographic byte comparison). Input order does not matter; the same set of keys always produces the same `MultisigKey`. Duplicate keys are rejected during construction.

9. **Signature low-S normalization.** All signatures are normalized to low-S form (EIP-2) during construction via the `normalizeToLowS()` method. If S > N/2, S is replaced with N - S and V is flipped. This is transparent to callers and applies to `ParseSignature`, `SignatureFromBytes`, `signatureFromRSV`, and all signing operations.

10. **Blocking operations.** `GeneratePrivateKey` and `GenerateMnemonic` read from `crypto/rand` (non-blocking on modern systems). `Scramble` / `UnscramblePrivateKey` use Argon2id (CPU-bound, allocates 64 MiB). `Sign` / `SignMessage` / `SignDigest` are CPU-bound but fast (<1ms). `RecoverPublicKey` involves EC point recovery, also typically fast (<1ms).
