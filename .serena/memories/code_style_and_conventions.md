# Code Style and Conventions

## Go Conventions

This project follows standard Go conventions and idioms.

### File Organization

- One main type per file (e.g., `private_key.go` for `NeuronPrivateKey`)
- Test files named `*_test.go` alongside source files
- Package documentation in `doc.go`

### Naming Conventions

| Pattern | Usage | Example |
|---------|-------|---------|
| `Parse*` | Construct from string (validates) | `ParsePrivateKeyHex()` |
| `*FromBytes` | Construct from raw bytes | `PrivateKeyFromBytes()` |
| `*FromHedera` | Elevate from Hedera SDK type | `PrivateKeyFromHedera()` |
| `Generate*` | Create new random value | `GeneratePrivateKey()` |
| `To*` | Extract/convert to external type | `ToHederaPrivateKey()` |
| `*Bytes()` | Serialize to raw bytes | `Bytes()` |
| `*Hex()` | Serialize to hex string | `Hex()` |
| `Matches*` | Constant-time equality check | `MatchesEVMAddress()` |
| `Is*` | Boolean check | `IsZero()`, `IsEd25519Key()` |
| `*Safe()` | Returns error instead of zero value | `EVMAddressSafe()` |

### Type Design

- **Immutable types**: Fields are unexported, modified only via constructors
- **Zero-value invalid**: `IsZero()` method to check validity
- **No panics**: Always return `(T, error)` for fallible operations
- **Pointer receivers**: Only for mutating methods like `Zeroize()`

### Error Handling

Use custom `KeyError` type with structured fields:

```go
type KeyError struct {
    Op      string    // Operation that failed
    Kind    ErrorKind // Error category
    Details string    // Human-readable explanation
    Err     error     // Underlying error (optional)
}
```

### Documentation

- Package doc in `doc.go` using godoc format
- Each exported function/type has a doc comment
- Use `# Section` headers in doc comments for godoc sections
- Use `[TypeName]` syntax for cross-references

Example:
```go
// SignMessage signs a message by first hashing it with Keccak256.
// Returns a 65-byte signature in R||S||V format.
//
// # Concurrency
//
// This method is safe for concurrent use.
func (k NeuronPrivateKey) SignMessage(msg []byte) (Signature, error) {
```

### Imports

Standard Go import grouping:
1. Standard library
2. External packages
3. Internal packages

### Security Practices

- Constant-time comparison for all sensitive data (`crypto/subtle`)
- `Zeroize()` pattern for clearing sensitive memory
- No logging of private key values
- Input validation at all boundaries

### Testing

- Table-driven tests preferred
- Test file next to source file
- Test function naming: `TestTypeName_MethodName`
- Use `t.Parallel()` where safe

### Thread Safety

- All types thread-safe after construction
- Exception: `Zeroize()` requires external synchronization
- Document thread safety in godoc
