# Task Completion Checklist

## Before Committing Changes

### 1. Code Quality
- [ ] Code follows project naming conventions (Parse*, From*, To*, etc.)
- [ ] All exported functions/types have godoc comments
- [ ] No panics - all fallible operations return errors
- [ ] Zero-value structs handled appropriately (IsZero() checks)
- [ ] Sensitive comparisons use constant-time operations

### 2. Formatting
```bash
# Format all code
gofmt -w .

# Or just check
gofmt -d .
```

### 3. Static Analysis
```bash
# Run go vet
go vet ./...
```

### 4. Testing
```bash
# Run all tests
go test ./... -v

# With coverage (should be 85%+)
go test ./keylib -cover
```

### 5. Build Verification
```bash
# Ensure it builds
go build ./...
```

## After Making API Changes

If you modified the HTTP API:

### 1. Update Swagger Documentation
```bash
# Regenerate swagger docs
swag init -g cmd/keylib-api/main.go -o api/docs
```

### 2. Test API Endpoints
```bash
# Start the server
go run ./cmd/keylib-api

# Test endpoints via Swagger UI or curl
curl http://localhost:8080/health
```

## For keylib Changes

### 1. Update Documentation
- Update godoc comments for new/changed functions
- Update `doc.go` if adding new major features
- Update `README.md` or `IMPLEMENTATION_MAPPING.md` if needed

### 2. Thread Safety Considerations
- New types should be immutable after construction
- Document any non-thread-safe methods
- CPU-intensive operations should be noted

### 3. Error Handling
- Use appropriate ErrorKind
- Include operation name in errors
- Provide helpful error details

## Quick Verification Commands

```bash
# One-liner to verify before commit
gofmt -w . && go vet ./... && go test ./... && go build ./...
```
