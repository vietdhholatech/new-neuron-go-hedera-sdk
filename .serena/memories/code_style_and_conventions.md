# Code Style and Conventions

## Go Style
- Standard Go formatting (gofmt)
- Package-level documentation in `doc.go`
- All exported functions/types have godoc comments

## Naming Conventions

### Factory Functions
| Prefix | Meaning | Example |
|--------|---------|---------|
| `Parse*` | Construct from string (validates) | `ParsePrivateKeyHex()` |
| `*FromBytes` | Construct from raw bytes | `PrivateKeyFromBytes()` |
| `*FromHedera` | Elevate from Hedera SDK type | `PrivateKeyFromHedera()` |
| `Generate*` | Create new random value | `GeneratePrivateKey()` |

### Extraction Methods
| Prefix | Meaning | Example |
|--------|---------|---------|
| `To*` | Extract/convert to external type | `ToHederaPrivateKey()` |
| `*Bytes()` | Serialize to raw bytes | `Bytes()`, `CompressedBytes()` |
| `*Hex()` | Serialize to hex string | `Hex()`, `HexWithoutPrefix()` |

### Verification Methods
| Prefix | Meaning | Example |
|--------|---------|---------|
| `Matches*` | Constant-time equality check | `MatchesEVMAddress()` |
| `Is*` | Boolean type check | `IsZero()`, `IsEd25519Key()` |
| `Validate*` | Return error if invalid | `ValidateMnemonic()` |

### Safe Variants
Functions returning errors instead of zero values use `*Safe` suffix:
- `EVMAddressSafe()` - Returns error for zero-value keys
- `ToHederaPrivateKeySafe()` - Returns error instead of empty key

## Type Definitions

### Core Types (immutable structs)
```go
type NeuronPrivateKey struct { key *secp256k1.PrivateKey }  // unexported field
type NeuronPublicKey struct { key *secp256k1.PublicKey }
type EVMAddress struct { addr [20]byte }
type PeerID struct { id peer.ID }
type Signature struct { data [65]byte }
```

### Error Types
```go
type KeyError struct {
    Op      string    // Operation that failed (e.g., "ParsePrivateKeyHex")
    Kind    ErrorKind // Category (e.g., ErrKindInvalidHex)
    Details string    // Human-readable explanation
    Err     error     // Underlying error
}
```

## Function Signatures
- All fallible operations return `(T, error)`, never panic
- Factory functions validate all input
- Zero-value structs are invalid (check with `IsZero()`)

## Error Handling
- Use descriptive errors with context
- Errors include operation name, error kind, details, and wrapped error
- 10 distinct ErrorKind values:
  - `ErrKindInvalidFormat`
  - `ErrKindInvalidLength`
  - `ErrKindInvalidHex`
  - `ErrKindInvalidKey`
  - `ErrKindZeroValue`
  - `ErrKindKeyMismatch`
  - `ErrKindEncryption`
  - `ErrKindMnemonic`
  - `ErrKindDerivation`
  - `ErrKindUnsupportedKeyType`

## Security Patterns
- All sensitive comparisons use `crypto/subtle` (constant-time)
- Use `Zeroize()` to clear private key memory
- Use `*Safe()` variants in critical paths
- Never log or print private key values

## Test Conventions
- Test files are co-located: `*_test.go`
- Table-driven tests where appropriate
- Test coverage target: 85%+
- Test file naming matches source file

## Documentation
- Package docs in `doc.go`
- Godoc sections using `# Header` syntax
- Link types with `[TypeName]` syntax
- All exported symbols documented

## API Conventions (for HTTP API)
- RESTful endpoints under `/api/v1`
- JSON request/response bodies
- Swagger annotations for documentation
- CORS middleware enabled
- Error handler middleware
