package keylib

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"math/big"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// DefaultDerivationPath is the Ethereum standard derivation path.
// m/44'/60'/0'/0/0 - BIP44 path for Ethereum
const DefaultDerivationPath = "m/44'/60'/0'/0/0"

// ValidMnemonicWordCounts are the valid word counts for BIP39 mnemonics.
var ValidMnemonicWordCounts = []int{12, 15, 18, 21, 24}

// GenerateMnemonic generates a new BIP39 mnemonic phrase.
// wordCount must be 12, 15, 18, 21, or 24.
//
// # Concurrency
//
// This function is safe for concurrent use. It reads from crypto/rand
// for entropy generation (non-blocking on modern systems).
func GenerateMnemonic(wordCount int) (string, error) {
	const op = "GenerateMnemonic"

	// Validate word count
	bitSize, err := wordCountToBitSize(wordCount)
	if err != nil {
		return "", err
	}

	// Generate entropy
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", wrapError(op, ErrKindDerivation, "failed to generate entropy", err)
	}

	// Generate mnemonic from entropy
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", wrapError(op, ErrKindDerivation, "failed to generate mnemonic", err)
	}

	return mnemonic, nil
}

// PrivateKeyFromMnemonic derives a private key from a BIP39 mnemonic.
// Uses the default derivation path (m/44'/60'/0'/0/0) and empty passphrase.
func PrivateKeyFromMnemonic(mnemonic string) (NeuronPrivateKey, error) {
	return PrivateKeyFromMnemonicWithOptions(mnemonic, "", DefaultDerivationPath)
}

// PrivateKeyFromMnemonicWithPath derives a private key using a custom derivation path.
// Uses empty passphrase.
func PrivateKeyFromMnemonicWithPath(mnemonic, path string) (NeuronPrivateKey, error) {
	return PrivateKeyFromMnemonicWithOptions(mnemonic, "", path)
}

// PrivateKeyFromMnemonicWithPassphrase derives a private key using a BIP39 passphrase.
// Uses the default derivation path.
func PrivateKeyFromMnemonicWithPassphrase(mnemonic, passphrase string) (NeuronPrivateKey, error) {
	return PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, DefaultDerivationPath)
}

// PrivateKeyFromMnemonicWithOptions derives a private key with full control over options.
// This is the most flexible mnemonic derivation function.
//
// # Concurrency
//
// This function is safe for concurrent use. No blocking I/O.
// Note: BIP32 derivation involves multiple HMAC-SHA512 operations,
// which is CPU-bound but typically fast (<10ms for standard paths).
func PrivateKeyFromMnemonicWithOptions(mnemonic, passphrase, path string) (NeuronPrivateKey, error) {
	const op = "PrivateKeyFromMnemonic"

	// Validate mnemonic
	mnemonic = normalizeMnemonic(mnemonic)
	if !bip39.IsMnemonicValid(mnemonic) {
		return NeuronPrivateKey{}, errMnemonic(op, "invalid mnemonic phrase", nil)
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Parse derivation path
	pathComponents, err := parseDerivationPath(path)
	if err != nil {
		return NeuronPrivateKey{}, err
	}

	// Derive key using BIP32
	key, err := deriveKeyFromSeed(seed, pathComponents)
	if err != nil {
		return NeuronPrivateKey{}, err
	}

	return PrivateKeyFromBytes(key)
}

// ValidateMnemonic checks if a mnemonic phrase is valid BIP39.
// Returns nil if valid, error with details if invalid.
func ValidateMnemonic(mnemonic string) error {
	const op = "ValidateMnemonic"

	mnemonic = normalizeMnemonic(mnemonic)

	// Check word count
	words := strings.Fields(mnemonic)
	wordCount := len(words)

	validCount := false
	for _, valid := range ValidMnemonicWordCounts {
		if wordCount == valid {
			validCount = true
			break
		}
	}

	if !validCount {
		return errMnemonic(op, "invalid word count; must be 12, 15, 18, 21, or 24", nil)
	}

	// Validate using bip39 library
	if !bip39.IsMnemonicValid(mnemonic) {
		return errMnemonic(op, "invalid mnemonic phrase (bad checksum or unknown words)", nil)
	}

	return nil
}

// normalizeMnemonic normalizes a mnemonic by trimming whitespace and
// converting to lowercase with single spaces between words.
func normalizeMnemonic(mnemonic string) string {
	words := strings.Fields(strings.TrimSpace(mnemonic))
	return strings.ToLower(strings.Join(words, " "))
}

// wordCountToBitSize converts word count to entropy bit size.
func wordCountToBitSize(wordCount int) (int, error) {
	const op = "GenerateMnemonic"

	switch wordCount {
	case 12:
		return 128, nil
	case 15:
		return 160, nil
	case 18:
		return 192, nil
	case 21:
		return 224, nil
	case 24:
		return 256, nil
	default:
		return 0, errMnemonic(op, "invalid word count; must be 12, 15, 18, 21, or 24", nil)
	}
}

// DerivationPathComponent represents a single component in a BIP32 path.
type DerivationPathComponent struct {
	Index    uint32
	Hardened bool
}

// parseDerivationPath parses a BIP32 derivation path string.
// Format: m/44'/60'/0'/0/0 (apostrophe indicates hardened derivation)
func parseDerivationPath(path string) ([]DerivationPathComponent, error) {
	const op = "parseDerivationPath"

	// Normalize path
	path = strings.TrimSpace(path)

	// Must start with "m/" or "M/"
	if !strings.HasPrefix(strings.ToLower(path), "m/") {
		return nil, errInvalidFormat(op, "derivation path must start with 'm/'")
	}

	// Remove "m/" prefix
	path = path[2:]

	if path == "" {
		return []DerivationPathComponent{}, nil
	}

	// Split by "/"
	parts := strings.Split(path, "/")
	components := make([]DerivationPathComponent, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for hardened indicator
		hardened := false
		if strings.HasSuffix(part, "'") || strings.HasSuffix(part, "h") || strings.HasSuffix(part, "H") {
			hardened = true
			part = part[:len(part)-1]
		}

		// Parse index
		var index uint32
		for _, c := range part {
			if c < '0' || c > '9' {
				return nil, errInvalidFormat(op, "invalid derivation path component: "+part)
			}
			index = index*10 + uint32(c-'0')
		}

		components = append(components, DerivationPathComponent{
			Index:    index,
			Hardened: hardened,
		})
	}

	return components, nil
}

// deriveKeyFromSeed derives a private key from a seed using BIP32.
func deriveKeyFromSeed(seed []byte, path []DerivationPathComponent) ([32]byte, error) {
	const op = "deriveKeyFromSeed"

	// Generate master key using HMAC-SHA512
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	I := mac.Sum(nil)

	// Split into key and chain code
	key := I[:32]
	chainCode := I[32:]

	// Derive child keys along the path
	for _, component := range path {
		var err error
		key, chainCode, err = deriveChildKey(key, chainCode, component.Index, component.Hardened)
		if err != nil {
			return [32]byte{}, err
		}
	}

	// Validate the final key
	if err := validatePrivateKeyBytes(op, key); err != nil {
		return [32]byte{}, errDerivation(op, "derived key is invalid", err)
	}

	var result [32]byte
	copy(result[:], key)
	return result, nil
}

// deriveChildKey derives a child key using BIP32.
func deriveChildKey(key, chainCode []byte, index uint32, hardened bool) ([]byte, []byte, error) {
	// Prepare data for HMAC
	var data []byte

	if hardened {
		// Hardened derivation: 0x00 || key || index
		index += 0x80000000 // Set hardened bit
		data = make([]byte, 37)
		data[0] = 0x00
		copy(data[1:33], key)
		binary.BigEndian.PutUint32(data[33:], index)
	} else {
		// Normal derivation: publicKey || index
		// We need to compute the public key from the private key
		privKey, err := PrivateKeyFromBytes(*(*[32]byte)(key))
		if err != nil {
			return nil, nil, err
		}
		pubKey := privKey.PublicKey().CompressedBytes()

		data = make([]byte, 37)
		copy(data[0:33], pubKey[:])
		binary.BigEndian.PutUint32(data[33:], index)
	}

	// HMAC-SHA512
	mac := hmac.New(sha512.New, chainCode)
	mac.Write(data)
	I := mac.Sum(nil)

	// Split result
	IL := I[:32]
	IR := I[32:]

	// Add IL to parent key (mod n)
	childKey := addPrivateKeys(key, IL)

	return childKey, IR, nil
}

// addPrivateKeys adds two private keys modulo the curve order.
func addPrivateKeys(key1, key2 []byte) []byte {
	// Convert to big integers and add
	k1 := new(big.Int).SetBytes(key1)
	k2 := new(big.Int).SetBytes(key2)

	result := new(big.Int).Add(k1, k2)
	result.Mod(result, secp256k1N)

	// Pad to 32 bytes
	return padLeftZeros(result.Bytes(), 32)
}
