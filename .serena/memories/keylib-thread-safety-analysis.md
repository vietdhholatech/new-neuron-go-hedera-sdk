# Keylib Thread Safety & Blocking I/O Analysis

**Analysis Date:** January 2026  
**Analyzed By:** Senior Developer Review  
**Package:** `keylib`

---

## Executive Summary

The keylib package is designed with **thread-safety as a core principle**. All exported types are immutable value types (with one exception), and all functions are **pure or use only thread-safe external resources**. The package contains **no global mutable state** and **no mutex locks**, indicating a functional design pattern.

### Key Findings
- ✅ **All exported functions are thread-safe**
- ✅ **No blocking network I/O**
- ✅ **No blocking file I/O** (except minimal entropy reads)
- ⚠️ **CPU-bound operations exist** (Argon2, BIP32 derivation)
- ⚠️ **One mutating method**: `Zeroize()` on pointer receiver

---

## Thread Safety Classification

### Legend
- 🟢 **Thread-Safe**: Can be called concurrently from multiple goroutines without synchronization
- 🟡 **Thread-Safe with Caveat**: Thread-safe but has performance/blocking considerations
- 🔴 **Not Thread-Safe**: Requires external synchronization

---

## Type Analysis

### NeuronPrivateKey
```
Type: Immutable value type (internal *secp256k1.PrivateKey is never modified after construction)
Thread-Safe: 🟢 YES
```

| Method | Thread-Safe | Blocking | Notes |
|--------|-------------|----------|-------|
| `IsZero()` | 🟢 | No | Pure nil check |
| `PublicKey()` | 🟢 | No | Pure derivation |
| `EVMAddress()` | 🟢 | No | Pure Keccak256 hash |
| `EVMAddressSafe()` | 🟢 | No | Pure derivation with error |
| `PeerID()` | 🟢 | No | Pure multihash derivation |
| `Bytes()` | 🟢 | No | Returns copy |
| `Hex()` | 🟢 | No | Pure hex encoding |
| `HexWithoutPrefix()` | 🟢 | No | Pure hex encoding |
| `ToECDSA()` | 🟢 | No | Pure conversion |
| `ToHederaPrivateKey()` | 🟢 | No | Pure conversion |
| `ToHederaPrivateKeySafe()` | 🟢 | No | Pure conversion |
| `SignMessage(msg)` | 🟢 | No | Pure ECDSA signing |
| `SignDigest(digest)` | 🟢 | No | Pure ECDSA signing |
| `Sign(rand, digest, opts)` | 🟡 | Minimal | Uses crypto/rand (see note) |
| `Public()` | 🟢 | No | Pure conversion |
| `MatchesPublicKey(pub)` | 🟢 | No | Constant-time comparison |
| `MatchesEVMAddress(addr)` | 🟢 | No | Constant-time comparison |
| `Equal(other)` | 🟢 | No | Constant-time comparison |
| `Zeroize()` | 🔴 | No | **MUTATES** receiver - requires pointer |

**⚠️ Warning**: `Zeroize()` is the only non-thread-safe method. It mutates the internal key pointer. If you need to zeroize a key while other goroutines might be using it, you must use external synchronization.

### NeuronPublicKey
```
Type: Immutable value type
Thread-Safe: 🟢 YES (all methods)
```

| Method | Thread-Safe | Blocking | Notes |
|--------|-------------|----------|-------|
| `IsZero()` | 🟢 | No | Pure nil check |
| `CompressedBytes()` | 🟢 | No | Returns copy |
| `UncompressedBytes()` | 🟢 | No | Returns copy |
| `Hex()` | 🟢 | No | Pure hex encoding |
| `HexUncompressed()` | 🟢 | No | Pure hex encoding |
| `EVMAddress()` | 🟢 | No | Pure Keccak256 hash |
| `EVMAddressSafe()` | 🟢 | No | Pure derivation |
| `PeerID()` | 🟢 | No | Pure multihash |
| `ToECDSA()` | 🟢 | No | Pure conversion |
| `ToHederaPublicKey()` | 🟢 | No | Pure conversion |
| `ToHederaPublicKeySafe()` | 🟢 | No | Pure conversion |
| `MatchesEVMAddress(addr)` | 🟢 | No | Constant-time comparison |
| `MatchesPeerID(pid)` | 🟢 | No | Constant-time comparison |
| `Verify(msg, sig)` | 🟢 | No | Pure ECDSA verify |
| `VerifyDigest(digest, sig)` | 🟢 | No | Pure ECDSA verify |
| `Equal(other)` | 🟢 | No | Constant-time comparison |

### EVMAddress
```
Type: Immutable value type ([20]byte array)
Thread-Safe: 🟢 YES (all methods)
```

All methods are pure functions operating on the internal byte array.

### PeerID
```
Type: Immutable value type (wraps peer.ID string)
Thread-Safe: 🟢 YES (all methods)
```

All methods are pure functions.

### Signature
```
Type: Immutable value type ([65]byte array)
Thread-Safe: 🟢 YES (all methods)
```

All methods are pure functions.

### EncryptedPrivateKey
```
Type: Mutable struct (but typically used as value)
Thread-Safe: 🟢 YES (methods don't mutate)
```

---

## Factory Function Analysis

| Function | Thread-Safe | Blocking | CPU Intensive | Notes |
|----------|-------------|----------|---------------|-------|
| `GeneratePrivateKey()` | 🟡 | Minimal | Low | Reads from crypto/rand |
| `ParsePrivateKeyHex(s)` | 🟢 | No | Low | Pure parsing |
| `PrivateKeyFromBytes(b)` | 🟢 | No | Low | Pure construction |
| `PrivateKeyFromHedera(key)` | 🟢 | No | Low | Pure conversion |
| `ParsePublicKeyHex(s)` | 🟢 | No | Low | Pure parsing |
| `PublicKeyFromBytes(b)` | 🟢 | No | Low | EC point validation |
| `PublicKeyFromHedera(key)` | 🟢 | No | Low | Pure conversion |
| `PublicKeyFromPrivateKey(k)` | 🟢 | No | Low | EC multiplication |
| `ParseEVMAddress(s)` | 🟢 | No | Low | Pure parsing |
| `EVMAddressFromCommon(addr)` | 🟢 | No | Low | Pure copy |
| `ParsePeerID(s)` | 🟢 | No | Low | Base58 decode |
| `PeerIDFromLibp2p(id)` | 🟢 | No | Low | Pure wrap |
| `ParseSignature(s)` | 🟢 | No | Low | Pure parsing |
| `SignatureFromBytes(b)` | 🟢 | No | Low | Pure copy |
| `IsEd25519Key(key)` | 🟢 | No | Low | Pure detection |
| `IsEd25519PublicKey(key)` | 🟢 | No | Low | Pure detection |

---

## Mnemonic Functions

| Function | Thread-Safe | Blocking | CPU Intensive | Notes |
|----------|-------------|----------|---------------|-------|
| `GenerateMnemonic(wordCount)` | 🟡 | Minimal | Low | Reads from crypto/rand |
| `PrivateKeyFromMnemonic(m)` | 🟢 | No | Medium | BIP32 derivation |
| `PrivateKeyFromMnemonicWithPath(m, p)` | 🟢 | No | Medium | BIP32 derivation |
| `PrivateKeyFromMnemonicWithPassphrase(m, p)` | 🟢 | No | Medium | BIP32 derivation |
| `PrivateKeyFromMnemonicWithOptions(m, p, path)` | 🟢 | No | Medium | Multiple HMAC-SHA512 |
| `ValidateMnemonic(m)` | 🟢 | No | Low | Word list lookup |

**Note**: BIP32 derivation involves multiple HMAC-SHA512 operations. For deep paths, this can accumulate.

---

## Signature Functions

| Function | Thread-Safe | Blocking | CPU Intensive | Notes |
|----------|-------------|----------|---------------|-------|
| `RecoverPublicKey(msg, sig)` | 🟢 | No | Medium | EC point recovery |
| `RecoverPublicKeyFromDigest(d, sig)` | 🟢 | No | Medium | EC point recovery |

---

## Encryption Functions (CPU-INTENSIVE)

| Function | Thread-Safe | Blocking | CPU Intensive | Notes |
|----------|-------------|----------|---------------|-------|
| `(k).Scramble(password)` | 🟡 | **Yes** | **HIGH** | Argon2id: 64MB RAM, 3 iterations |
| `UnscramblePrivateKey(enc, pwd)` | 🟡 | **Yes** | **HIGH** | Argon2id: 64MB RAM, 3 iterations |

**⚠️ IMPORTANT**: `Scramble()` and `UnscramblePrivateKey()` are CPU-intensive and memory-bound operations:
- Allocates 64MB of memory per call
- Runs 3 iterations of Argon2id
- Uses 4 parallel threads internally
- Typical execution time: 100-500ms depending on hardware

**Recommendation**: When processing multiple keys concurrently, consider using a semaphore or worker pool to limit concurrent Scramble/Unscramble operations.

---

## Blocking I/O Analysis

### Sources of Potential Blocking

1. **crypto/rand.Reader** (used in 4 places):
   - `GeneratePrivateKey()` - ECDSA key generation
   - `GenerateMnemonic()` - Entropy generation via bip39
   - `Scramble()` - Salt and nonce generation
   - `Sign()` (crypto.Signer interface) - ECDSA signing
   
   **Reality**: `crypto/rand` uses `/dev/urandom` on Unix and `CryptGenRandom` on Windows. These are non-blocking in normal operation. Blocking only occurs on system boot before entropy pool initialization (extremely rare in practice).

2. **File I/O**: None

3. **Network I/O**: None

4. **System Calls**: Only crypto/rand (see above)

---

## Goroutine Usage Recommendations

### Safe for Unlimited Concurrency
```go
// These can be called from any number of goroutines simultaneously
pubKey := privKey.PublicKey()
addr := pubKey.EVMAddress()
valid := pubKey.Verify(msg, sig)
parsed, _ := keylib.ParsePrivateKeyHex(hex)
```

### Recommend Limiting Concurrency
```go
// Limit concurrent Scramble/Unscramble to prevent memory pressure
sem := make(chan struct{}, runtime.NumCPU())
for _, key := range keys {
    sem <- struct{}{}
    go func(k NeuronPrivateKey) {
        defer func() { <-sem }()
        encrypted, _ := k.Scramble(password)
        // ... use encrypted
    }(key)
}
```

### Requires External Synchronization
```go
// Zeroize() mutates - ensure no concurrent access
var mu sync.Mutex
mu.Lock()
key.Zeroize()
mu.Unlock()
```

---

## Global State Analysis

| Variable | Type | Mutated | Thread-Safe |
|----------|------|---------|-------------|
| `secp256k1N` | `*big.Int` | Never (set once at init) | 🟢 |
| `ZeroEVMAddress` | `EVMAddress` | Never | 🟢 |
| `ZeroSignature` | `Signature` | Never | 🟢 |
| `ValidMnemonicWordCounts` | `[]int` | Never | 🟢 |
| `DefaultDerivationPath` | `string` | Never (const) | 🟢 |
| `ErrWrongPassword` | `error` | Never | 🟢 |

**Conclusion**: No global mutable state exists in the package.

---

## Summary Table: All Exported Functions

| Function/Method | Thread-Safe | Blocking I/O | CPU Cost |
|-----------------|-------------|--------------|----------|
| `GeneratePrivateKey` | 🟢 | Minimal | Low |
| `ParsePrivateKeyHex` | 🟢 | None | Low |
| `PrivateKeyFromBytes` | 🟢 | None | Low |
| `PrivateKeyFromHedera` | 🟢 | None | Low |
| `ParsePublicKeyHex` | 🟢 | None | Low |
| `PublicKeyFromBytes` | 🟢 | None | Low |
| `PublicKeyFromHedera` | 🟢 | None | Low |
| `ParseEVMAddress` | 🟢 | None | Low |
| `ParsePeerID` | 🟢 | None | Low |
| `ParseSignature` | 🟢 | None | Low |
| `SignatureFromBytes` | 🟢 | None | Low |
| `GenerateMnemonic` | 🟢 | Minimal | Low |
| `PrivateKeyFromMnemonic*` | 🟢 | None | Medium |
| `ValidateMnemonic` | 🟢 | None | Low |
| `RecoverPublicKey*` | 🟢 | None | Medium |
| `Scramble` | 🟢 | Minimal | **HIGH** |
| `UnscramblePrivateKey` | 🟢 | Minimal | **HIGH** |
| All type methods | 🟢* | None | Low |

*Except `Zeroize()` which mutates

---

## Recommendations

1. **For HTTP/API handlers**: All keylib functions are safe to use directly in concurrent HTTP handlers without synchronization.

2. **For Scramble/Unscramble in bulk operations**: Implement a worker pool with limited concurrency (e.g., `runtime.NumCPU()`) to prevent memory exhaustion.

3. **For long-lived key objects**: If using `Zeroize()`, ensure the key is no longer referenced by other goroutines before calling.

4. **For high-throughput signing**: The signing operations are CPU-bound but fast. No special handling needed unless signing millions of messages per second.
