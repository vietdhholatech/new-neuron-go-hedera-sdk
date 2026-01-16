# Suggested Commands

## Build Commands

```bash
# Build the keylib package
go build ./keylib

# Build the API server
go build ./cmd/keylib-api

# Build all packages
go build ./...
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for keylib only
go test ./keylib -v

# Run tests with coverage
go test ./keylib -v -cover

# Generate coverage report
go test ./keylib -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test ./keylib -v -run TestPrivateKeyFromHex

# Run benchmarks
go test ./keylib -bench=.
```

## Linting and Formatting

```bash
# Check for issues
go vet ./...

# Check formatting
gofmt -d ./keylib

# Format code
gofmt -w ./keylib

# Run staticcheck (if installed)
staticcheck ./...
```

## Running the API Server

```bash
# Run the API server
go run ./cmd/keylib-api

# The server starts on :8080
# Swagger UI: http://localhost:8080/swagger/index.html
# Health check: http://localhost:8080/health
```

## Documentation Commands

```bash
# View package documentation
go doc ./keylib

# View full documentation with all symbols
go doc -all ./keylib

# View specific type
go doc ./keylib NeuronPrivateKey

# View specific function
go doc ./keylib GeneratePrivateKey

# Start local godoc server
godoc -http=:6060
# Then open http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/keylib/
```

## Swagger Documentation

```bash
# Regenerate swagger docs (if swag is installed)
swag init -g cmd/keylib-api/main.go -o api/docs
```

## Dependency Management

```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Update dependencies
go get -u ./...
```

## Utility Commands (Darwin/macOS)

```bash
# List files
ls -la

# Find files
find . -name "*.go" -type f

# Search in files
grep -r "pattern" --include="*.go"

# Git operations
git status
git diff
git log --oneline -10
```
