package keylib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/crypto/argon2"
)

// Default Argon2id parameters for key derivation.
// These provide strong security while remaining practical for user-facing operations.
const (
	defaultArgon2Time    = 3         // Number of iterations
	defaultArgon2Memory  = 64 * 1024 // Memory in KiB (64 MB)
	defaultArgon2Threads = 4         // Number of parallel threads
	argon2KeyLen         = 32        // Output key length in bytes (256 bits for AES-256)

	saltLen  = 16 // Salt length in bytes
	nonceLen = 12 // AES-GCM nonce length in bytes

	// Argon2 parameter bounds for validation
	minArgon2Time    = 1
	maxArgon2Time    = 100
	minArgon2Memory  = 8 * 1024  // 8 MiB minimum
	maxArgon2Memory  = 64 * 1024 // 64 MiB maximum (prevents memory exhaustion DoS)
	minArgon2Threads = 1
	maxArgon2Threads = 32
)

// scrambleConfig holds configuration for key encryption.
type scrambleConfig struct {
	argon2Time    uint32
	argon2Memory  uint32
	argon2Threads uint8
}

// validate checks if Argon2 parameters are within acceptable bounds.
func (c *scrambleConfig) validate(op string) error {
	if c.argon2Time < minArgon2Time || c.argon2Time > maxArgon2Time {
		return errEncryption(op, "argon2Time must be between 1 and 100", nil)
	}
	if c.argon2Memory < minArgon2Memory || c.argon2Memory > maxArgon2Memory {
		return errEncryption(op, "argon2Memory must be between 8 MiB and 64 MiB", nil)
	}
	if c.argon2Threads < minArgon2Threads || c.argon2Threads > maxArgon2Threads {
		return errEncryption(op, "argon2Threads must be between 1 and 32", nil)
	}
	return nil
}

// ScrambleOption configures key encryption parameters.
// Use with Scramble to customize Argon2id parameters for different security/performance trade-offs.
type ScrambleOption func(*scrambleConfig)

// WithArgon2Time sets the Argon2id time parameter (number of iterations).
// Higher values increase security but also computation time.
// Default: 3
func WithArgon2Time(t uint32) ScrambleOption {
	return func(c *scrambleConfig) {
		c.argon2Time = t
	}
}

// WithArgon2Memory sets the Argon2id memory parameter in KiB.
// Higher values increase security but also memory usage.
// Default: 64*1024 (64 MiB)
func WithArgon2Memory(m uint32) ScrambleOption {
	return func(c *scrambleConfig) {
		c.argon2Memory = m
	}
}

// WithArgon2Threads sets the Argon2id parallelism (number of threads).
// Should typically match or be less than available CPU cores.
// Default: 4
func WithArgon2Threads(t uint8) ScrambleOption {
	return func(c *scrambleConfig) {
		c.argon2Threads = t
	}
}

// EncryptedPrivateKey holds an encrypted private key with metadata.
// Uses AES-256-GCM with Argon2id key derivation for password-based encryption.
//
// The encryption scheme provides:
// - Authenticated encryption (AES-GCM prevents tampering)
// - Strong key derivation (Argon2id resists brute force)
// - Random salt and nonce for each encryption
// - Stored Argon2 parameters (Version 2+) for reliable decryption
//
// Version History:
//   - Version 1: Original format, uses hardcoded default Argon2 parameters
//   - Version 2: Stores Argon2 parameters for custom encryption settings
type EncryptedPrivateKey struct {
	// Version identifies the encryption scheme for future upgrades.
	// Version 1: Legacy format (assumes default Argon2 parameters)
	// Version 2: Includes stored Argon2 parameters
	Version int `json:"version"`

	// Salt is the random salt used for Argon2id key derivation (16 bytes).
	Salt []byte `json:"salt"`

	// Nonce is the random nonce used for AES-GCM (12 bytes).
	Nonce []byte `json:"nonce"`

	// Ciphertext contains the encrypted key and GCM authentication tag.
	// Length is 32 (key) + 16 (tag) = 48 bytes.
	Ciphertext []byte `json:"ciphertext"`

	// Argon2 parameters (Version 2+)
	// These are zero for Version 1 keys (use defaults for decryption).
	Argon2Time    uint32 `json:"argon2Time,omitempty"`
	Argon2Memory  uint32 `json:"argon2Memory,omitempty"`
	Argon2Threads uint8  `json:"argon2Threads,omitempty"`
}

// CurrentEncryptionVersion is the current encryption scheme version.
// Version 2 stores Argon2 parameters to ensure decryption works with custom settings.
const CurrentEncryptionVersion = 2

// Scramble encrypts a private key with a password.
// Uses Argon2id for key derivation and AES-256-GCM for encryption.
//
// Use functional options to customize Argon2id parameters:
//
//	// Default parameters (recommended for most use cases)
//	encrypted, err := key.Scramble(password)
//
//	// Higher security (slower)
//	encrypted, err := key.Scramble(password, WithArgon2Time(5), WithArgon2Memory(128*1024))
//
//	// Lower resource usage (faster but less secure)
//	encrypted, err := key.Scramble(password, WithArgon2Time(1), WithArgon2Memory(32*1024))
//
// The password should be at least 8 characters for reasonable security,
// though this is not enforced to allow flexibility.
//
// # Concurrency
//
// This function is safe for concurrent use but is CPU and memory intensive.
// Each call allocates 64MB of memory (default) and runs Argon2id for ~100-500ms.
// In high-throughput scenarios, consider limiting concurrent calls using
// a semaphore or worker pool to prevent memory exhaustion.
//
// # Blocking
//
// Reads 28 bytes from crypto/rand for salt and nonce (non-blocking).
// The Argon2id computation is CPU-bound, not I/O-bound.
func (k NeuronPrivateKey) Scramble(password string, opts ...ScrambleOption) (EncryptedPrivateKey, error) {
	const op = "NeuronPrivateKey.Scramble"

	// Apply options with defaults
	cfg := &scrambleConfig{
		argon2Time:    defaultArgon2Time,
		argon2Memory:  defaultArgon2Memory,
		argon2Threads: defaultArgon2Threads,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate Argon2 parameters
	if err := cfg.validate(op); err != nil {
		return EncryptedPrivateKey{}, err
	}

	if err := k.checkValid(op); err != nil {
		return EncryptedPrivateKey{}, err
	}

	if password == "" {
		return EncryptedPrivateKey{}, errEncryption(op, "password cannot be empty", nil)
	}

	// Generate random salt
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return EncryptedPrivateKey{}, errEncryption(op, "failed to generate random salt", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return EncryptedPrivateKey{}, errEncryption(op, "failed to generate random nonce", err)
	}

	// Derive encryption key using Argon2id
	encKey := argon2.IDKey(
		[]byte(password),
		salt,
		cfg.argon2Time,
		cfg.argon2Memory,
		cfg.argon2Threads,
		argon2KeyLen,
	)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return EncryptedPrivateKey{}, errEncryption(op, "failed to create AES cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return EncryptedPrivateKey{}, errEncryption(op, "failed to create GCM", err)
	}

	// Encrypt the private key
	keyBytes := k.Bytes()
	ciphertext := gcm.Seal(nil, nonce, keyBytes[:], nil)

	// Clear sensitive data from memory
	secureZeroArray32(&keyBytes)
	secureZero(encKey)

	return EncryptedPrivateKey{
		Version:       CurrentEncryptionVersion,
		Salt:          salt,
		Nonce:         nonce,
		Ciphertext:    ciphertext,
		Argon2Time:    cfg.argon2Time,
		Argon2Memory:  cfg.argon2Memory,
		Argon2Threads: cfg.argon2Threads,
	}, nil
}

// UnscramblePrivateKey decrypts an encrypted private key with a password.
// Returns an error if the password is wrong or the data is corrupted.
//
// Supports both encryption versions:
//   - Version 1: Uses default Argon2 parameters (backward compatibility)
//   - Version 2: Uses stored Argon2 parameters
//
// # Concurrency
//
// This function is safe for concurrent use but is CPU and memory intensive.
// Each call allocates 64MB of memory (default) and runs Argon2id for ~100-500ms.
// In high-throughput scenarios, consider limiting concurrent calls.
//
// # Blocking
//
// No I/O operations. The Argon2id computation is CPU-bound.
func UnscramblePrivateKey(encrypted EncryptedPrivateKey, password string) (NeuronPrivateKey, error) {
	const op = "UnscramblePrivateKey"

	// Validate version (support both v1 and v2)
	if encrypted.Version != 1 && encrypted.Version != 2 {
		return NeuronPrivateKey{}, errEncryption(op,
			"unsupported encryption version; expected version 1 or 2", nil)
	}

	// Validate lengths
	if len(encrypted.Salt) != saltLen {
		return NeuronPrivateKey{}, errEncryption(op, "invalid salt length", nil)
	}
	if len(encrypted.Nonce) != nonceLen {
		return NeuronPrivateKey{}, errEncryption(op, "invalid nonce length", nil)
	}
	// Ciphertext should be 32 (key) + 16 (GCM tag) = 48 bytes
	if len(encrypted.Ciphertext) != 48 {
		return NeuronPrivateKey{}, errEncryption(op, "invalid ciphertext length", nil)
	}

	if password == "" {
		return NeuronPrivateKey{}, errEncryption(op, "password cannot be empty", nil)
	}

	// Determine Argon2 parameters based on version
	var argon2Time uint32 = defaultArgon2Time
	var argon2Memory uint32 = defaultArgon2Memory
	var argon2Threads uint8 = defaultArgon2Threads

	if encrypted.Version == 2 {
		// Version 2: Use stored parameters
		argon2Time = encrypted.Argon2Time
		argon2Memory = encrypted.Argon2Memory
		argon2Threads = encrypted.Argon2Threads

		// Validate stored parameters
		if argon2Time < minArgon2Time || argon2Time > maxArgon2Time {
			return NeuronPrivateKey{}, errEncryption(op, "stored argon2Time is invalid", nil)
		}
		if argon2Memory < minArgon2Memory || argon2Memory > maxArgon2Memory {
			return NeuronPrivateKey{}, errEncryption(op, "stored argon2Memory is invalid", nil)
		}
		if argon2Threads < minArgon2Threads || argon2Threads > maxArgon2Threads {
			return NeuronPrivateKey{}, errEncryption(op, "stored argon2Threads is invalid", nil)
		}
	}
	// Version 1: Use defaults (already set above)

	// Derive encryption key using Argon2id
	encKey := argon2.IDKey(
		[]byte(password),
		encrypted.Salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(encKey)
	if err != nil {
		secureZero(encKey)
		return NeuronPrivateKey{}, errEncryption(op, "failed to create AES cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		secureZero(encKey)
		return NeuronPrivateKey{}, errEncryption(op, "failed to create GCM", err)
	}

	// Decrypt the private key
	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)

	// Clear derived key from memory immediately after use
	secureZero(encKey)

	if err != nil {
		// GCM authentication failed - wrong password or corrupted data
		return NeuronPrivateKey{}, ErrWrongPassword
	}

	// Create the private key
	if len(plaintext) != 32 {
		secureZero(plaintext)
		return NeuronPrivateKey{}, errEncryption(op, "decrypted data has invalid length", nil)
	}

	privKey, err := PrivateKeyFromBytes(*(*[32]byte)(plaintext))

	// Clear plaintext from memory
	secureZero(plaintext)

	if err != nil {
		return NeuronPrivateKey{}, errEncryption(op, "decrypted key is invalid", err)
	}

	return privKey, nil
}

// IsZero returns true if this is a zero-value EncryptedPrivateKey.
func (e EncryptedPrivateKey) IsZero() bool {
	return e.Version == 0 && len(e.Salt) == 0 && len(e.Nonce) == 0 && len(e.Ciphertext) == 0
}

// MarshalJSON implements json.Marshaler.
// Encodes binary fields as base64 for JSON compatibility.
func (e EncryptedPrivateKey) MarshalJSON() ([]byte, error) {
	type jsonEncryptedKey struct {
		Version       int    `json:"version"`
		Salt          string `json:"salt"`
		Nonce         string `json:"nonce"`
		Ciphertext    string `json:"ciphertext"`
		Argon2Time    uint32 `json:"argon2Time,omitempty"`
		Argon2Memory  uint32 `json:"argon2Memory,omitempty"`
		Argon2Threads uint8  `json:"argon2Threads,omitempty"`
	}

	return json.Marshal(jsonEncryptedKey{
		Version:       e.Version,
		Salt:          base64.StdEncoding.EncodeToString(e.Salt),
		Nonce:         base64.StdEncoding.EncodeToString(e.Nonce),
		Ciphertext:    base64.StdEncoding.EncodeToString(e.Ciphertext),
		Argon2Time:    e.Argon2Time,
		Argon2Memory:  e.Argon2Memory,
		Argon2Threads: e.Argon2Threads,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
// Decodes base64-encoded binary fields.
func (e *EncryptedPrivateKey) UnmarshalJSON(data []byte) error {
	type jsonEncryptedKey struct {
		Version       int    `json:"version"`
		Salt          string `json:"salt"`
		Nonce         string `json:"nonce"`
		Ciphertext    string `json:"ciphertext"`
		Argon2Time    uint32 `json:"argon2Time,omitempty"`
		Argon2Memory  uint32 `json:"argon2Memory,omitempty"`
		Argon2Threads uint8  `json:"argon2Threads,omitempty"`
	}

	var j jsonEncryptedKey
	if err := json.Unmarshal(data, &j); err != nil {
		return errEncryption("EncryptedPrivateKey.UnmarshalJSON", "invalid JSON", err)
	}

	salt, err := base64.StdEncoding.DecodeString(j.Salt)
	if err != nil {
		return errEncryption("EncryptedPrivateKey.UnmarshalJSON", "invalid base64 salt", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(j.Nonce)
	if err != nil {
		return errEncryption("EncryptedPrivateKey.UnmarshalJSON", "invalid base64 nonce", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(j.Ciphertext)
	if err != nil {
		return errEncryption("EncryptedPrivateKey.UnmarshalJSON", "invalid base64 ciphertext", err)
	}

	e.Version = j.Version
	e.Salt = salt
	e.Nonce = nonce
	e.Ciphertext = ciphertext
	e.Argon2Time = j.Argon2Time
	e.Argon2Memory = j.Argon2Memory
	e.Argon2Threads = j.Argon2Threads

	return nil
}

// String returns a safe string representation (does not expose encrypted data).
func (e EncryptedPrivateKey) String() string {
	if e.IsZero() {
		return "EncryptedPrivateKey{zero}"
	}
	if e.Version == 1 {
		return "EncryptedPrivateKey{version=1, encrypted}"
	}
	return "EncryptedPrivateKey{version=2, encrypted}"
}
