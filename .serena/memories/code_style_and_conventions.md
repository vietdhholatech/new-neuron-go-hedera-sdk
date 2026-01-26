# Code Style and Conventions

## Naming Conventions

### Factory Functions
| Prefix | Purpose | Example |
|--------|---------|---------|
| `Parse*` | Construct from string (validates) | `ParsePrivateKeyHex()` |
| `*FromBytes` | Construct from raw bytes | `PrivateKeyFromBytes()` |
| `*FromHedera` | Elevate from Hedera SDK type | `PrivateKeyFromHedera()` |
| `Generate*` | Create new random value | `GeneratePrivateKey()` |

### Conversion Methods
| Prefix | Purpose | Example |
|--------|---------|---------|
| `To*` | Extract/convert to external type | `ToHederaPrivateKey()` |
| `*Bytes()` | Serialize to raw bytes | `Bytes()`, `CompressedBytes()` |
| `*Hex()` | Serialize to hex string | `Hex()`, `HexWithoutPrefix()` |

### Validation Methods
| Pattern | Purpose | Example |
|---------|---------|---------|
| `IsZero()` | Check if zero/invalid value | `key.IsZero()` |
| `Validate()` | Validate with error | `peerID.Validate()` |
| `Matches*` | Constant-time equality check | `MatchesEVMAddress()` |
| `Equal()` | Type-safe equality | `key1.Equal(key2)` |

### Safe Variants
Functions that could silently fail have `*Safe()` variants that return errors:
- `EVMAddressSafe()` - prevents burn address
- `ToHederaPrivateKeySafe()` - returns error on failure
- `ToHederaPublicKeySafe()` - returns error on failure

## Type Design

### Immutable Types
All core types are immutable after construction:
```go
type NeuronPrivateKey struct {
    key *secp256k1.PrivateKey  // unexported, never nil after valid construction
}

type NeuronPublicKey struct {
    key *secp256k1.PublicKey   // unexported
}

type EVMAddress struct {
    addr [20]byte              // fixed-size array
}
```

### Zero-Value Safety
- Zero-value structs are explicitly invalid
- `IsZero()` method indicates invalid state
- Factory functions validate all inputs

## Error Handling

### KeyError Type
```go
type KeyError struct {
    Op      string    // Operation (e.g., "ParsePrivateKeyHex")
    Kind    ErrorKind // Category
    Details string    // Human-readable explanation
    Err     error     // Underlying error
}
```

### ErrorKind Values
- `ErrKindInvalidFormat` - Wrong format
- `ErrKindInvalidLength` - Wrong length
- `ErrKindInvalidHex` - Invalid hex character
- `ErrKindInvalidKey` - Invalid key value
- `ErrKindZeroValue` - Zero/uninitialized value
- `ErrKindKeyMismatch` - Key mismatch
- `ErrKindEncryption` - Encryption/decryption error
- `ErrKindMnemonic` - Mnemonic error
- `ErrKindDerivation` - Key derivation error
- `ErrKindUnsupportedKeyType` - Unsupported key type (Ed25519)

## Documentation

### Godoc Conventions
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

### Package Documentation
- Main package doc in `doc.go`
- Implementation mapping in `IMPLEMENTATION_MAPPING.md`

## Testing Conventions

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Type
        wantErr bool
    }{
        {"valid case", "input", expected, false},
        {"invalid case", "bad", Type{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test File Naming
- Unit tests: `<file>_test.go` (e.g., `private_key_test.go`)
- Integration tests: `integration_test.go`

## Security Conventions

### Constant-Time Operations
- All sensitive comparisons use `crypto/subtle`
- `constantTimeEqual()` for byte slices
- `secureZero()` for memory clearing

### Memory Safety
- `Zeroize()` method to clear sensitive data
- Defer zeroization after use
- No global mutable state

## Import Organization
```go
import (
    // Standard library
    "crypto"
    "encoding/hex"
    
    // Third-party
    "github.com/decred/dcrd/dcrec/secp256k1/v4"
    
    // Internal packages
    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)
```
