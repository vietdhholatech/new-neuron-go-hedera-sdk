# Suggested Commands

## Build Commands

```bash
# Build entire project
go build ./...

# Build specific packages
go build ./keylib
go build ./account
go build ./api

# Build API binary
go build -o keylib-api ./cmd/keylib-api
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run tests for specific package
go test ./keylib -v
go test ./account -v

# Run specific test
go test ./keylib -v -run TestPrivateKeyFromHex

# Generate coverage report
go test ./keylib -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test ./keylib -bench=.
```

## Run Commands

```bash
# Run API server (development)
go run ./cmd/keylib-api

# Run built API binary
./keylib-api

# API server starts on :8080
# Swagger UI: http://localhost:8080/swagger/index.html
```

## Documentation Commands

```bash
# View package documentation
go doc ./keylib
go doc -all ./keylib

# View specific type
go doc ./keylib NeuronPrivateKey
go doc ./keylib NeuronPublicKey

# View specific function
go doc ./keylib ParsePrivateKeyHex
go doc ./keylib GenerateMnemonic

# Start local godoc server
godoc -http=:6060
# Then open: http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/keylib/

# Generate Swagger docs (requires swag CLI)
swag init -g cmd/keylib-api/main.go
```

## Code Quality Commands

```bash
# Run go vet
go vet ./...

# Format code
gofmt -w ./keylib
gofmt -w ./account
gofmt -w ./api

# Check formatting (dry run)
gofmt -d ./keylib

# Run staticcheck (if installed)
staticcheck ./...

# Tidy dependencies
go mod tidy
```

## Git Commands

```bash
# Check status
git status

# View diff
git diff

# Add and commit
git add <files>
git commit -m "message"

# Push changes
git push origin <branch>
```

## Utility Commands (macOS/Darwin)

```bash
# List files
ls -la

# Find files
find . -name "*.go" -type f

# Search in files
grep -r "pattern" ./keylib

# View file
cat <file>
head -n 50 <file>
tail -n 50 <file>
```
