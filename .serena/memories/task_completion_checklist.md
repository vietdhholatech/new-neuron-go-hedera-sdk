# Task Completion Checklist

When completing a task in this project, run through the following checklist:

## 1. Code Quality

```bash
# Format code
gofmt -w ./keylib ./api ./cmd

# Run vet
go vet ./...
```

## 2. Build Verification

```bash
# Ensure code compiles
go build ./...
```

## 3. Test Execution

```bash
# Run all tests
go test ./keylib -v

# Run with coverage (target: >85%)
go test ./keylib -v -cover

# Run with race detection for concurrency changes
go test ./keylib -race
```

## 4. If API Changes Were Made

```bash
# Regenerate Swagger docs
swag init -g cmd/keylib-api/main.go -o api/docs

# Verify API server starts
go run ./cmd/keylib-api &
curl http://localhost:8080/swagger/index.html
```

## 5. Documentation Updates

If public API was modified:
- Update doc comments on affected functions/types
- Update `keylib/doc.go` if package-level behavior changed
- Update `keylib/IMPLEMENTATION_MAPPING.md` if specification compliance changed

## 6. Security Considerations

For crypto-related changes:
- Ensure constant-time comparison for sensitive data
- Verify Zeroize() is called where appropriate
- Check for timing attack vulnerabilities
- Never log private key values

## 7. Thread Safety

For concurrency-related changes:
- Document thread safety in godoc
- Run tests with `-race` flag
- Update concurrency section in IMPLEMENTATION_MAPPING.md if needed

## Quick Validation Commands

```bash
# All-in-one validation
gofmt -w ./keylib && go vet ./keylib && go build ./keylib && go test ./keylib -v -cover

# Full project validation
gofmt -w . && go vet ./... && go build ./... && go test ./keylib -v -cover
```
