# Task Completion Checklist

When completing a coding task in this project, verify the following:

## Before Committing

### 1. Build Verification
```bash
# Ensure the project builds without errors
go build ./...
```

### 2. Run Tests
```bash
# Run all tests and verify they pass
go test ./... -v

# For keylib changes, run with coverage
go test ./keylib -v -cover
```

### 3. Code Quality
```bash
# Run go vet
go vet ./...

# Check formatting
gofmt -d ./keylib ./account ./api
```

### 4. Documentation
- Update godoc comments if public API changed
- Update README.md if needed
- Update IMPLEMENTATION_MAPPING.md if implementing spec requirements

## For New Features

### Type Safety
- [ ] New types have unexported fields
- [ ] Factory functions validate all inputs
- [ ] `IsZero()` method implemented
- [ ] Zero-value struct is explicitly invalid

### Error Handling
- [ ] Use `KeyError` with appropriate `ErrorKind`
- [ ] Include operation name in `Op` field
- [ ] Provide descriptive `Details` message
- [ ] Wrap underlying errors

### Security (for crypto operations)
- [ ] Use constant-time comparisons for sensitive data
- [ ] Implement `Zeroize()` for private key types
- [ ] Add `*Safe()` variant if method could fail silently
- [ ] No timing-based vulnerabilities

### Testing
- [ ] Table-driven tests for all cases
- [ ] Test error conditions
- [ ] Test edge cases (zero values, invalid inputs)
- [ ] Add integration test with known vectors if applicable

## For API Changes

### Swagger Documentation
```bash
# Regenerate swagger docs
swag init -g cmd/keylib-api/main.go
```

### Request/Response
- [ ] DTOs defined in `api/dto/`
- [ ] Swagger annotations on handlers
- [ ] Proper error responses

## For Account Package Changes

### Backend Compatibility
- [ ] Test with Hedera backend
- [ ] Test with Kafka backend (if applicable)
- [ ] Ensure interface compliance

### Serialization
- [ ] JSON marshaling/unmarshaling works
- [ ] Backward compatibility maintained

## Final Steps

```bash
# One final verification
go build ./...
go test ./... -v
go vet ./...

# If all passes, commit
git add <files>
git commit -m "descriptive message"
```
