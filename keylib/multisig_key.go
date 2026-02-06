package keylib

import (
	"fmt"
	"sort"
)

// Multisig protocol identifiers. These constants identify the threshold signature
// protocol/standard used by a MultisigKey, enabling protocol-aware operations.
const (
	// MultisigProtocolSecp256k1Aggregated is the default protocol for standard secp256k1 key aggregation.
	MultisigProtocolSecp256k1Aggregated = "secp256k1-aggregated"
	// MultisigProtocolHederaThreshold represents Hedera-specific threshold keys.
	MultisigProtocolHederaThreshold = "hedera-threshold"
	// MultisigProtocolFROST represents FROST threshold signature scheme.
	MultisigProtocolFROST = "frost"
	// MultisigProtocolBLS represents BLS threshold signature scheme.
	MultisigProtocolBLS = "bls"
)

// MultisigKey represents an M-of-N threshold signing configuration for shared accounts.
// This type is immutable after construction. The zero value is invalid;
// methods return errors on zero value.
//
// A MultisigKey specifies:
//   - A set of public keys that can participate in signing
//   - A threshold (M) of signatures required from the total (N) keys
//
// Example: A 2-of-3 MultisigKey requires any 2 of 3 keys to sign.
//
// # Concurrency
//
// MultisigKey is safe for concurrent use by multiple goroutines.
// All methods are thread-safe and non-blocking.
type MultisigKey struct {
	publicKeys []NeuronPublicKey // participating public keys (stored sorted for determinism)
	threshold  int               // M in M-of-N (minimum signatures required)
	total      int               // N in M-of-N (total number of keys)
	protocol   string            // threshold signature protocol/standard identifier
}

// NewMultisigKey creates a new MultisigKey with the given public keys and threshold.
//
// Validation rules:
//   - publicKeys must not be empty
//   - threshold must be >= 1
//   - threshold must be <= len(publicKeys)
//   - All publicKeys must be non-zero
//   - No duplicate publicKeys allowed
//
// The keys are stored in a deterministic order (sorted by compressed bytes)
// to ensure consistent behavior regardless of input order.
func NewMultisigKey(publicKeys []NeuronPublicKey, threshold int) (MultisigKey, error) {
	const op = "NewMultisigKey"

	if len(publicKeys) == 0 {
		return MultisigKey{}, errInvalidThreshold(op, threshold, 0, "public keys list cannot be empty")
	}

	if threshold < 1 {
		return MultisigKey{}, errInvalidThreshold(op, threshold, len(publicKeys), "threshold must be at least 1")
	}

	if threshold > len(publicKeys) {
		return MultisigKey{}, errInvalidThreshold(op, threshold, len(publicKeys), "threshold cannot exceed total number of keys")
	}

	// Validate all keys are non-zero
	for i, key := range publicKeys {
		if key.IsZero() {
			return MultisigKey{}, errZeroValue(op, fmt.Sprintf("NeuronPublicKey at index %d", i))
		}
	}

	// Check for duplicates using a map of hex strings
	seen := make(map[string]bool, len(publicKeys))
	for i, key := range publicKeys {
		hex := key.Hex()
		if seen[hex] {
			return MultisigKey{}, errDuplicateKey(op, i)
		}
		seen[hex] = true
	}

	// Create defensive copy and sort for determinism
	keysCopy := make([]NeuronPublicKey, len(publicKeys))
	copy(keysCopy, publicKeys)
	sortPublicKeys(keysCopy)

	return MultisigKey{
		publicKeys: keysCopy,
		threshold:  threshold,
		total:      len(keysCopy),
		protocol:   MultisigProtocolSecp256k1Aggregated,
	}, nil
}

// MustNewMultisigKey creates a new MultisigKey and panics on error.
// This is intended for use in tests and initialization of known-valid configurations.
func MustNewMultisigKey(publicKeys []NeuronPublicKey, threshold int) MultisigKey {
	key, err := NewMultisigKey(publicKeys, threshold)
	if err != nil {
		panic(err)
	}
	return key
}

// NewMultisigKeyWithProtocol creates a new MultisigKey with the given public keys, threshold,
// and protocol identifier. Use this for blockchain-specific threshold implementations
// (e.g., "hedera-threshold", "frost", "bls"). For standard secp256k1 aggregation,
// use NewMultisigKey which defaults to "secp256k1-aggregated".
//
// The protocol parameter identifies the threshold signature scheme and determines
// which operations are available. Non-"secp256k1-aggregated" protocols cannot derive
// EVM addresses or PeerIDs.
func NewMultisigKeyWithProtocol(publicKeys []NeuronPublicKey, threshold int, protocol string) (MultisigKey, error) {
	const op = "NewMultisigKeyWithProtocol"

	if protocol == "" {
		return MultisigKey{}, newKeyError(op, ErrKindInvalidFormat, "protocol must not be empty", nil)
	}

	mk, err := NewMultisigKey(publicKeys, threshold)
	if err != nil {
		return MultisigKey{}, err
	}
	mk.protocol = protocol
	return mk, nil
}

// IsZero returns true if this is a zero-value (invalid) MultisigKey.
func (k MultisigKey) IsZero() bool {
	return len(k.publicKeys) == 0
}

// Validate checks if the MultisigKey is valid.
// Returns nil if valid, or an error describing the validation failure.
func (k MultisigKey) Validate() error {
	const op = "MultisigKey.Validate"

	if k.IsZero() {
		return errZeroValue(op, "MultisigKey")
	}

	if k.threshold < 1 {
		return errInvalidThreshold(op, k.threshold, k.total, "threshold must be at least 1")
	}

	if k.threshold > k.total {
		return errInvalidThreshold(op, k.threshold, k.total, "threshold cannot exceed total")
	}

	if k.total != len(k.publicKeys) {
		return errInvalidThreshold(op, k.threshold, k.total, "total does not match public keys count")
	}

	if k.protocol == "" {
		return newKeyError(op, ErrKindInvalidFormat, "MultisigKey protocol must not be empty", nil)
	}

	// Validate all keys are non-zero
	for i, key := range k.publicKeys {
		if key.IsZero() {
			return errZeroValue(op, fmt.Sprintf("NeuronPublicKey at index %d", i))
		}
	}

	return nil
}

// Threshold returns M (the minimum number of signatures required).
func (k MultisigKey) Threshold() int {
	return k.threshold
}

// Total returns N (the total number of participating keys).
func (k MultisigKey) Total() int {
	return k.total
}

// Protocol returns the threshold signature protocol/standard identifier.
// Examples: "secp256k1-aggregated", "hedera-threshold", "frost", "bls".
func (k MultisigKey) Protocol() string {
	return k.protocol
}

// PublicKeys returns a defensive copy of the participating public keys.
// The returned slice is sorted deterministically by compressed bytes.
func (k MultisigKey) PublicKeys() []NeuronPublicKey {
	if k.IsZero() {
		return nil
	}
	result := make([]NeuronPublicKey, len(k.publicKeys))
	copy(result, k.publicKeys)
	return result
}

// ContainsKey returns true if the given public key is part of this MultisigKey.
func (k MultisigKey) ContainsKey(pubKey NeuronPublicKey) bool {
	if k.IsZero() || pubKey.IsZero() {
		return false
	}
	for _, key := range k.publicKeys {
		if key.Equal(pubKey) {
			return true
		}
	}
	return false
}

// Equal returns true if this MultisigKey equals another.
// Two MultisigKeys are equal if they have the same threshold and the same set of keys.
// Uses constant-time comparison for the key comparison.
func (k MultisigKey) Equal(other MultisigKey) bool {
	if k.IsZero() && other.IsZero() {
		return true
	}
	if k.IsZero() || other.IsZero() {
		return false
	}

	if k.threshold != other.threshold {
		return false
	}
	if k.total != other.total {
		return false
	}
	if k.protocol != other.protocol {
		return false
	}

	// Since keys are stored sorted, we can compare directly
	for i := range k.publicKeys {
		if !k.publicKeys[i].Equal(other.publicKeys[i]) {
			return false
		}
	}
	return true
}

// String returns a human-readable representation of the MultisigKey.
func (k MultisigKey) String() string {
	if k.IsZero() {
		return "MultisigKey{zero-value}"
	}
	return fmt.Sprintf("MultisigKey{%d-of-%d, protocol=%s}", k.threshold, k.total, k.protocol)
}

// sortPublicKeys sorts a slice of public keys by their compressed bytes.
// This ensures deterministic ordering regardless of input order.
func sortPublicKeys(keys []NeuronPublicKey) {
	sort.Slice(keys, func(i, j int) bool {
		bytesI := keys[i].CompressedBytes()
		bytesJ := keys[j].CompressedBytes()
		for k := 0; k < 33; k++ {
			if bytesI[k] != bytesJ[k] {
				return bytesI[k] < bytesJ[k]
			}
		}
		return false
	})
}
