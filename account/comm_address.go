package account

import (
	"fmt"
	"strings"
)

// CommAddressKind identifies the type of communication address.
// Each kind has its own locator format and validation rules.
//
// Kind constants are defined in their respective backend files:
//   - HederaTopicKind in backend_hedera.go
//   - KafkaTopicKind in backend_kafka.go
//   - CustomTopicKind in backend_custom.go
type CommAddressKind string

// String returns the string representation of the address kind.
func (k CommAddressKind) String() string {
	return string(k)
}

// IsValid checks if this is a registered address kind.
// A kind is valid if it has a registered backend in the registry.
func (k CommAddressKind) IsValid() bool {
	return IsRegisteredBackend(string(k))
}

// TopicTechnologyForKind returns the corresponding TopicTechnology for this address kind.
// This queries the registered backend to determine the technology.
// Returns TopicTechnologyCustom if the backend is not registered.
func (k CommAddressKind) TopicTechnologyForKind() TopicTechnology {
	backend, ok := GetBackend(string(k))
	if !ok {
		return TopicTechnologyCustom
	}
	return backend.Technology()
}

// CommAddress represents a public communication endpoint.
// It is the abstraction used for stdIn, stdOut, and stdErr addresses in NeuronAccount.
//
// A CommAddress:
//   - declares which backend technology it uses
//   - declares how the endpoint is located
//   - validates itself on construction
//
// The account does not send/receive messages - it only declares where channels live.
//
// # Extensibility
//
// CommAddress supports any registered backend. Built-in backends include "hedera-topic"
// and "kafka-topic". Additional backends can be registered via RegisterBackend().
type CommAddress struct {
	// kind identifies the type of communication address
	kind CommAddressKind

	// locator is the backend-specific identifier
	// Examples:
	//   - Hedera: "0.0.12345"
	//   - Kafka: "my-topic-name"
	locator string
}

// NewCustomAddress creates a CommAddress with a custom kind.
// This is for extending to other communication backends.
//
// Deprecated: Use NewCommAddress with a registered backend instead.
func NewCustomAddress(kind CommAddressKind, locator string) (CommAddress, error) {
	if locator == "" {
		return CommAddress{}, errInvalidAddress("NewCustomAddress", "locator cannot be empty", nil)
	}
	return CommAddress{
		kind:    kind,
		locator: locator,
	}, nil
}

// NewCommAddress creates a CommAddress with the specified backend kind.
//
// If the kind corresponds to a registered backend, the locator is validated
// and normalized according to that backend's rules via backend.ParseLocator().
//
// # Forward Compatibility
//
// If the kind is not registered, the address is created without validation.
// This design allows applications to work with backend types that may be
// added in future versions without requiring code changes. The only
// requirement for unregistered backends is a non-empty locator string.
//
// This forward compatibility enables:
//   - Deserializing accounts with future backend types
//   - Supporting custom third-party backends
//   - Gradual rollout of new communication technologies
//
// Returns an error if:
//   - The backend is registered and the locator is invalid
//   - The locator is empty for an unregistered backend
func NewCommAddress(kind string, locator string) (CommAddress, error) {
	const op = "NewCommAddress"

	backend, ok := GetBackend(kind)
	if !ok {
		// Allow unregistered backends with no validation (for forward compatibility)
		if locator == "" {
			return CommAddress{}, errInvalidAddress(op,
				fmt.Sprintf("empty locator for unknown backend %q", kind), nil)
		}
		return CommAddress{kind: CommAddressKind(kind), locator: locator}, nil
	}

	// Validate and normalize the locator using the registered backend
	normalized, err := backend.ParseLocator(locator)
	if err != nil {
		return CommAddress{}, err
	}

	return CommAddress{kind: CommAddressKind(kind), locator: normalized}, nil
}

// ParseCommAddress parses a string representation of a CommAddress.
// Format: "kind:locator" (e.g., "hedera-topic:0.0.12345")
//
// If the kind corresponds to a registered backend, the locator is validated
// and normalized according to that backend's rules.
func ParseCommAddress(s string) (CommAddress, error) {
	const op = "ParseCommAddress"

	if s == "" {
		return CommAddress{}, errInvalidAddress(op, "empty address string", nil)
	}

	// Split on first colon
	idx := strings.Index(s, ":")
	if idx == -1 {
		return CommAddress{}, errInvalidAddress(op,
			fmt.Sprintf("invalid format, expected 'kind:locator', got: %s", s), nil)
	}

	kindStr := s[:idx]
	locator := s[idx+1:]

	if locator == "" {
		return CommAddress{}, errInvalidAddress(op, "empty locator in address", nil)
	}

	// Use the generic NewCommAddress which handles registry lookup
	return NewCommAddress(kindStr, locator)
}

// MustParseCommAddress parses a CommAddress string and panics on error.
// This is intended for use in tests and initialization of known-valid addresses.
func MustParseCommAddress(s string) CommAddress {
	addr, err := ParseCommAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}

// Kind returns the type of this communication address.
func (a CommAddress) Kind() CommAddressKind {
	return a.kind
}

// Locator returns the backend-specific identifier.
func (a CommAddress) Locator() string {
	return a.locator
}

// Technology returns the TopicTechnology corresponding to this address kind.
func (a CommAddress) Technology() TopicTechnology {
	return a.kind.TopicTechnologyForKind()
}

// String returns the string representation of this address.
// Format: "kind:locator"
func (a CommAddress) String() string {
	if a.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s:%s", a.kind, a.locator)
}

// IsZero returns true if this is a zero-value CommAddress.
func (a CommAddress) IsZero() bool {
	return a.kind == "" && a.locator == ""
}

// Validate checks if this address is well-formed according to its kind's rules.
//
// If the kind corresponds to a registered backend, validation is delegated
// to that backend's ValidateLocator method.
//
// For unregistered backends, validation only ensures the locator is non-empty.
func (a CommAddress) Validate() error {
	const op = "CommAddress.Validate"

	if a.IsZero() {
		return errZeroValue(op, "CommAddress")
	}

	// Try to use registered backend for validation
	backend, ok := GetBackend(string(a.kind))
	if ok {
		return backend.ValidateLocator(a.locator)
	}

	// Fallback for unregistered backends (including custom)
	if a.locator == "" {
		return errInvalidAddress(op,
			fmt.Sprintf("address has empty locator for kind %q", a.kind), nil)
	}
	return nil
}

// Equal compares this address with another for equality.
func (a CommAddress) Equal(other CommAddress) bool {
	return a.kind == other.kind && a.locator == other.locator
}

// CommAddressSet represents a collection of CommAddresses.
// This is used for cases where multiple addresses may be needed.
type CommAddressSet struct {
	addresses []CommAddress
}

// NewCommAddressSet creates a new set from the given addresses.
func NewCommAddressSet(addrs ...CommAddress) CommAddressSet {
	// Filter out zero values
	filtered := make([]CommAddress, 0, len(addrs))
	for _, addr := range addrs {
		if !addr.IsZero() {
			filtered = append(filtered, addr)
		}
	}
	return CommAddressSet{addresses: filtered}
}

// Addresses returns a copy of the addresses in this set.
func (s CommAddressSet) Addresses() []CommAddress {
	if len(s.addresses) == 0 {
		return nil
	}
	result := make([]CommAddress, len(s.addresses))
	copy(result, s.addresses)
	return result
}

// IsEmpty returns true if the set contains no addresses.
func (s CommAddressSet) IsEmpty() bool {
	return len(s.addresses) == 0
}

// Len returns the number of addresses in the set.
func (s CommAddressSet) Len() int {
	return len(s.addresses)
}

// First returns the first address in the set, or zero value if empty.
func (s CommAddressSet) First() CommAddress {
	if len(s.addresses) == 0 {
		return CommAddress{}
	}
	return s.addresses[0]
}
