package keylib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters for key derivation.
// These provide strong security while remaining practical for user-facing operations.
const (
	argon2Time    = 3         // Number of iterations
	argon2Memory  = 64 * 1024 // Memory in KiB (64 MB)
	argon2Threads = 4         // Number of parallel threads
	argon2KeyLen  = 32        // Output key length in bytes (256 bits for AES-256)

	saltLen  = 16 // Salt length in bytes
	nonceLen = 12 // AES-GCM nonce length in bytes
)

// EncryptedPrivateKey holds an encrypted private key with metadata.
// Uses AES-256-GCM with Argon2id key derivation for password-based encryption.
//
// The encryption scheme provides:
// - Authenticated encryption (AES-GCM prevents tampering)
// - Strong key derivation (Argon2id resists brute force)
// - Random salt and nonce for each encryption
type EncryptedPrivateKey struct {
	// Version identifies the encryption scheme for future upgrades.
	Version int `json:"version"`

	// Salt is the random salt used for Argon2id key derivation (16 bytes).
	Salt []byte `json:"salt"`

	// Nonce is the random nonce used for AES-GCM (12 bytes).
	Nonce []byte `json:"nonce"`

	// Ciphertext contains the encrypted key and GCM authentication tag.
	// Length is 32 (key) + 16 (tag) = 48 bytes.
	Ciphertext []byte `json:"ciphertext"`
}

// CurrentEncryptionVersion is the current encryption scheme version.
const CurrentEncryptionVersion = 1

// Scramble encrypts a private key with a password.
// Uses Argon2id for key derivation and AES-256-GCM for encryption.
//
// The password should be at least 8 characters for reasonable security,
// though this is not enforced to allow flexibility.
//
// # Concurrency
//
// This function is safe for concurrent use but is CPU and memory intensive.
// Each call allocates 64MB of memory and runs Argon2id for ~100-500ms.
// In high-throughput scenarios, consider limiting concurrent calls using
// a semaphore or worker pool to prevent memory exhaustion.
//
// # Blocking
//
// Reads 28 bytes from crypto/rand for salt and nonce (non-blocking).
// The Argon2id computation is CPU-bound, not I/O-bound.
func (k NeuronPrivateKey) Scramble(password string) (EncryptedPrivateKey, error) {
	const op = "NeuronPrivateKey.Scramble"

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
		argon2Time,
		argon2Memory,
		argon2Threads,
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

	// Clear the key bytes from memory
	secureZeroArray32(&keyBytes)

	return EncryptedPrivateKey{
		Version:    CurrentEncryptionVersion,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

// UnscramblePrivateKey decrypts an encrypted private key with a password.
// Returns an error if the password is wrong or the data is corrupted.
//
// # Concurrency
//
// This function is safe for concurrent use but is CPU and memory intensive.
// Each call allocates 64MB of memory and runs Argon2id for ~100-500ms.
// In high-throughput scenarios, consider limiting concurrent calls.
//
// # Blocking
//
// No I/O operations. The Argon2id computation is CPU-bound.
func UnscramblePrivateKey(encrypted EncryptedPrivateKey, password string) (NeuronPrivateKey, error) {
	const op = "UnscramblePrivateKey"

	// Validate version
	if encrypted.Version != CurrentEncryptionVersion {
		return NeuronPrivateKey{}, errEncryption(op,
			"unsupported encryption version; expected version 1", nil)
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
		return NeuronPrivateKey{}, errEncryption(op, "failed to create AES cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return NeuronPrivateKey{}, errEncryption(op, "failed to create GCM", err)
	}

	// Decrypt the private key
	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)
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
		Version    int    `json:"version"`
		Salt       string `json:"salt"`
		Nonce      string `json:"nonce"`
		Ciphertext string `json:"ciphertext"`
	}

	return json.Marshal(jsonEncryptedKey{
		Version:    e.Version,
		Salt:       base64.StdEncoding.EncodeToString(e.Salt),
		Nonce:      base64.StdEncoding.EncodeToString(e.Nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(e.Ciphertext),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
// Decodes base64-encoded binary fields.
func (e *EncryptedPrivateKey) UnmarshalJSON(data []byte) error {
	type jsonEncryptedKey struct {
		Version    int    `json:"version"`
		Salt       string `json:"salt"`
		Nonce      string `json:"nonce"`
		Ciphertext string `json:"ciphertext"`
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

	return nil
}

// String returns a safe string representation (does not expose encrypted data).
func (e EncryptedPrivateKey) String() string {
	if e.IsZero() {
		return "EncryptedPrivateKey{zero}"
	}
	return "EncryptedPrivateKey{version=1, encrypted}"
}
