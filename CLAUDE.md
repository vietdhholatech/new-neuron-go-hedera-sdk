# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK for building decentralized agent identities on the Hedera network with multi-ecosystem interoperability (Hedera, Ethereum, libp2p). Module path: `github.com/aspect-build/neuron-go-hedera-sdk`. Requires **Go 1.24+**.

## Common Commands

```bash
# Run all tests with race detector
go test ./keylib/... ./account/... -race -v

# Run tests for a single package
go test ./keylib/... -v
go test ./account/... -v

# Run a specific test
go test ./account/... -run TestNeuronAccount_Validate -v

# Coverage (keylib target: >87%, account target: >91%)
go test ./keylib/... -coverprofile=keylib.out && go tool cover -func=keylib.out
go test ./account/... -coverprofile=account.out && go tool cover -func=account.out

# Format and vet
go fmt ./...
go vet ./...

# Build API server
go build -o keylib-api ./cmd/keylib-api

# Regenerate swagger docs
swag init -g api/server.go -o api/docs
```

## Architecture

### Two core library packages

**`keylib/`** — Cryptographic key management (secp256k1 only, Ed25519 rejected)
- `NeuronPrivateKey` (32-byte) / `NeuronPublicKey` (33-byte compressed) — immutable, thread-safe
- `EVMAddress` (20-byte, EIP-55 checksum), `PeerID` (libp2p), `Signature` (65-byte R‖S‖V)
- `EncryptedPrivateKey` — Argon2id (64MB memory) + AES-256-GCM encryption
- `MultisigKey` — M-of-N threshold with sorted key ordering
- Factory functions: `GeneratePrivateKey()`, `ParsePrivateKeyHex()`, `ParsePublicKeyHex()`, etc.

**`account/`** — Agent identity descriptors with three account types:
- **Parent**: root identity with DID, NO communication channels
- **Child**: operational endpoint, ALL 3 channels required (stdIn/Out/Err), references parent
- **Shared**: M-of-N multisig via `MultisigKey`, NO DID/channels/parent
- `CommAddress` — technology-agnostic address in `kind:locator` format
- `Backend` interface + thread-safe registry — extensible plugin system for messaging (Hedera, Kafka, Custom)

### Supporting packages

- **`api/`** — Gin REST API with Swagger docs, handlers in `api/handlers/`, DTOs in `api/dto/`
- **`cmd/keylib-api/`** — API server entry point (port 8080, Swagger at `/swagger/index.html`)
- **`account/didkey/`** — W3C `did:key` implementation

### Dependency flow

`account/` depends on `keylib/` (for key types, PeerID, EVMAddress). `api/` depends on both.

## Key Design Patterns

### Immutable types with validation at boundaries
All `Parse*` functions validate immediately and return `(T, error)`. Zero values are always invalid — use `IsZero()` to check. Types are immutable after construction; no setters.

### Structured errors with `ErrorKind`
Both packages use typed errors (`KeyError`, `AccountError`) with an `Op` field (operation name), `Kind` field (programmatic category), and `Details` field. Check with `errors.As()` and switch on `Kind`.

### Builder pattern for accounts
Three type-specific constructors enforce different constraints: `NewParentAccountBuilder()`, `NewChildAccountBuilder()`, `NewSharedAccountBuilder()`. Errors accumulate and return on `Build()`.

### Safe vs unsafe method variants
Unsafe methods return zero values on error (e.g., `EVMAddress()`). Safe variants return explicit errors (e.g., `EVMAddressSafe()`).

### Constant-time comparisons
All key comparisons (`Equal()`, `MatchesPublicKey()`, `MatchesEVMAddress()`) use `crypto/subtle` to prevent timing attacks.

### Backend registry
Backends register via `init()` functions. Thread-safe with `sync.RWMutex`. Built-in: `hedera`, `kafka`, `custom`.

## Testing Conventions

- Table-driven tests with `t.Run()` subtests
- Test helpers use `t.Helper()`
- Special test files: `*_integration_test.go`, `e2e_lifecycle_test.go`, `concurrent_test.go`, `edge_cases_test.go`
- Always run with `-race` flag
- `Scramble()`/`UnscramblePrivateKey()` use 64MB each — limit concurrency in tests

## Security Notes

- `Zeroize()` clears private key memory (uses Go `clear()` builtin) — only mutation allowed on key types
- `Scramble()` is CPU/memory intensive (Argon2id: 64MB, 3 iterations, 4 threads)
- Only secp256k1 keys are supported; Ed25519 keys are explicitly rejected with `ErrKindUnsupportedKeyType`
