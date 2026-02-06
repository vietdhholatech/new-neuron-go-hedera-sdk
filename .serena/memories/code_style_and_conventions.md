# Code Style and Conventions

## General Go Style
- Follow standard Go conventions (effective Go, Go code review comments)
- Use `gofmt` for formatting
- Package names are lowercase, single-word (no underscores or mixedCaps)

## Documentation
- Each package has a `doc.go` file with comprehensive package-level documentation
- Document all exported types, functions, and methods
- Use godoc-style comments with examples where appropriate

## Error Handling
- Use structured error types with rich context:
  ```go
  type KeyError struct {
      Op      string    // Operation that failed
      Kind    ErrorKind // Error category for programmatic handling
      Details string    // Human-readable explanation
      Err     error     // Underlying error
  }
  ```
- Never panic in library code; always return errors
- Use `errors.As()` for error type assertions
- ErrorKind enum has 12 kinds including `ErrKindSDKError` for wrapping blockchain SDK failures

## Type Design
- Types are immutable after construction
- Zero values are invalid; use `IsZero()` methods
- Provide both safe (return error) and unsafe (return zero value) variants:
  - `EVMAddress()` returns zero on error
  - `EVMAddressSafe()` returns error

## Naming Conventions
- Factory functions: `NewXxx()`, `ParseXxx()`, `GenerateXxx()`
- Conversion methods: `ToXxx()`, `ToXxxSafe()`
- Check methods: `IsZero()`, `IsXxx()`
- Match methods (constant-time): `MatchesXxx()`, `Equal()`
- Secure cleanup: `Zeroize()`

## Testing Style
- Table-driven tests with subtests using `t.Run()`
- Section separators in test files:
  ```go
  // =============================================================================
  // TestSubjectName Tests
  // =============================================================================
  ```
- Descriptive test names: `TestNeuronPrivateKey_PublicKey`
- Test file naming: `xxx_test.go` in the same package

## Security Patterns
- Constant-time comparisons for all cryptographic data
- `Zeroize()` method to clear sensitive data from memory
- Use Go 1.21+ `clear()` builtin for secure zeroing
- Password-based encryption uses Argon2id + AES-256-GCM

## Builder Pattern (account module)
Three account type constructors with type-specific constraints:
```go
// Parent: requires PublicKey + DID, prohibits comm channels
parent, err := account.NewParentAccountBuilder(publicKey, did).
    WithHederaTopics("0.0.111", "0.0.222", "0.0.333").
    WithReachableAddr("/ip4/...").
    Build()

// Child: requires PublicKey + Parent, must have all 3 comm channels
child, err := account.NewChildAccountBuilder(publicKey, parentDID).
    WithCommAddress(commAddr).
    Build()

// Shared: requires MultisigKey, prohibits DID/comm/parent/single public key
shared, err := account.NewSharedAccountBuilder(multisigKey).
    Build()
```
Builder uses error accumulation (collects all errors, returns on Build()).

## Interface Design
- Keep interfaces small and focused
- Define interfaces where they're used, not where they're implemented
- Backend system uses registry pattern for extensibility

## Comments
- Avoid redundant comments that repeat the code
- Comment on "why", not "what"
- Use `// TODO:` for incomplete items
