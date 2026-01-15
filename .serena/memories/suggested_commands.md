# Suggested Commands

## Building

```bash
# Build keylib package
go build ./keylib

# Build API server
go build ./cmd/keylib-api

# Build all packages
go build ./...
```

## Testing

```bash
# Run all tests
go test ./keylib -v

# Run tests with coverage
go test ./keylib -v -cover

# Generate coverage report (HTML)
go test ./keylib -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test ./keylib -v -run TestPrivateKeyFromHex

# Run benchmarks
go test ./keylib -bench=.

# Run tests with race detection
go test ./keylib -race
```

## Linting and Formatting

```bash
# Run go vet
go vet ./keylib
go vet ./...

# Check formatting (shows diff)
gofmt -d ./keylib

# Apply formatting
gofmt -w ./keylib

# Run staticcheck (if installed)
staticcheck ./keylib
staticcheck ./...
```

## Running the API Server

```bash
# Run directly
go run ./cmd/keylib-api

# Build and run
go build -o keylib-api ./cmd/keylib-api
./keylib-api

# Server starts on :8080
# Swagger UI: http://localhost:8080/swagger/index.html
```

## Viewing Documentation

```bash
# View package overview
go doc ./keylib

# View complete package documentation
go doc -all ./keylib

# View specific type
go doc ./keylib NeuronPrivateKey

# View specific function
go doc ./keylib ParsePrivateKeyHex

# View method documentation
go doc ./keylib NeuronPrivateKey.SignMessage

# Start local godoc server
godoc -http=:6060
# Then open: http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/keylib/
```

## Swagger Documentation

```bash
# Regenerate Swagger docs (requires swag CLI)
swag init -g cmd/keylib-api/main.go -o api/docs

# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest
```

## Dependencies

```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph
```

## Git Operations

```bash
# Status
git status

# Add and commit
git add .
git commit -m "message"

# View log
git log --oneline -10
```

## System Commands (Darwin/macOS)

```bash
# List files
ls -la

# Find files
find . -name "*.go" -type f

# Search in files (prefer ripgrep if available)
grep -r "pattern" ./keylib
rg "pattern" ./keylib

# View file
cat keylib/doc.go
less keylib/doc.go
```
