# Suggested Commands

## Testing
```bash
# Run all tests
go test ./keylib/... ./account/... -v

# Run tests with race detector
go test ./keylib/... ./account/... -race

# Run tests for a specific package
go test ./keylib/... -v

# Run a specific test
go test ./keylib/... -run TestNeuronPrivateKey_PublicKey -v

# Run integration tests only
go test ./keylib/... -run TestIntegration -v
go test ./account/... -run TestIntegration -v

# Run edge case tests only
go test ./account/... -run TestEdgeCase -v
```

## Coverage
```bash
# Check coverage for keylib
go test ./keylib/... -coverprofile=keylib.out && go tool cover -func=keylib.out

# Check coverage for account
go test ./account/... -coverprofile=account.out && go tool cover -func=account.out

# Generate HTML coverage report for all
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

## Running the API Server
```bash
# Run the API server (starts on :8080)
go run ./cmd/keylib-api

# After starting, Swagger UI is available at:
# http://localhost:8080/swagger/index.html
```

## Documentation
```bash
# View keylib documentation
go doc -all ./keylib

# View account documentation
go doc -all ./account

# Start local documentation server
godoc -http=:6060
# Then open: http://localhost:6060/pkg/github.com/aspect-build/neuron-go-hedera-sdk/
```

## Building
```bash
# Build the API server
go build -o keylib-api ./cmd/keylib-api

# Run the built binary
./keylib-api
```

## Formatting and Linting
```bash
# Format code
go fmt ./...

# Run Go vet
go vet ./...
```

## Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Regenerating Swagger Docs
```bash
# Install swag CLI if needed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init -g api/server.go -o api/docs
```

## System Utilities (macOS/Darwin)
- `git` - Version control
- `ls` - List files
- `grep` - Search text (or use Go tools)
- `find` - Find files (or use Go tools)
