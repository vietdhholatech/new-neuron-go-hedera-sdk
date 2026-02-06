# Task Completion Checklist

When completing a task in this codebase, ensure the following:

## Before Submitting

### 1. Tests Pass
```bash
# Run all tests with race detector
go test ./keylib/... ./account/... -race -v
```

### 2. Coverage is Maintained
- keylib: Target >85% coverage (currently 87.1%)
- account: Target >85% coverage (currently 91.6%)
```bash
go test ./keylib/... -coverprofile=keylib.out && go tool cover -func=keylib.out
go test ./account/... -coverprofile=account.out && go tool cover -func=account.out
```

### 3. Code is Formatted
```bash
go fmt ./...
```

### 4. No Vet Errors
```bash
go vet ./...
```

### 5. Dependencies are Tidy
```bash
go mod tidy
```

## Code Quality Checks

### Security
- [ ] No sensitive data logged or printed
- [ ] Constant-time comparisons used for cryptographic data
- [ ] `Zeroize()` called when done with private keys
- [ ] Password-encrypted storage used for persisted keys

### Error Handling
- [ ] All errors returned, not panicked
- [ ] Errors use structured `KeyError` or `AccountError` types
- [ ] Error messages are descriptive and actionable

### Type Safety
- [ ] No string-based key handling (use typed wrappers)
- [ ] Zero values handled appropriately
- [ ] Safe variants provided for fallible operations

### Testing
- [ ] Unit tests added/updated for new code
- [ ] Table-driven tests with descriptive subtests
- [ ] Edge cases covered (zero values, invalid inputs)
- [ ] Race conditions tested (`-race` flag)

### Documentation
- [ ] Exported symbols documented
- [ ] doc.go updated if package API changed
- [ ] README.md updated if user-facing changes

## API Changes (if applicable)

### Swagger
```bash
# Regenerate swagger docs if API changed
swag init -g api/server.go -o api/docs
```

### Verify API Works
```bash
# Start server and test manually
go run ./cmd/keylib-api
# Open http://localhost:8080/swagger/index.html
```
