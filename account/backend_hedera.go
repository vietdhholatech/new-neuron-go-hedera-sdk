package account

import (
	"fmt"
	"regexp"
)

// HederaTopicKind is the unique identifier for the Hedera Consensus Service backend.
// Use this constant when creating CommAddresses for Hedera topics.
const HederaTopicKind = "hedera-topic"

// init registers the Hedera backend at package initialization.
func init() {
	RegisterBackend(&hederaBackend{})
}

// hederaTopicPattern validates Hedera topic ID format: shard.realm.topic
// Each component is a non-negative integer.
var hederaTopicPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// hederaBackend implements the Backend interface for Hedera Consensus Service topics.
type hederaBackend struct{}

// Kind returns the unique identifier for the Hedera backend.
func (h *hederaBackend) Kind() string {
	return HederaTopicKind
}

// ValidateLocator checks if a locator is a valid Hedera topic ID.
// Valid format: "shard.realm.topic" where each component is a non-negative integer.
// Example: "0.0.12345"
func (h *hederaBackend) ValidateLocator(locator string) error {
	if locator == "" {
		return errInvalidAddress("hederaBackend.ValidateLocator", "empty Hedera topic locator", nil)
	}

	if !hederaTopicPattern.MatchString(locator) {
		return errInvalidAddress("hederaBackend.ValidateLocator",
			fmt.Sprintf("invalid Hedera topic format, expected 'shard.realm.topic', got: %s", locator), nil)
	}

	return nil
}

// ParseLocator validates and returns the locator unchanged.
// Hedera topic IDs are already in canonical form.
func (h *hederaBackend) ParseLocator(locator string) (string, error) {
	if err := h.ValidateLocator(locator); err != nil {
		return "", err
	}
	return locator, nil
}

// Metadata returns descriptive information about the Hedera backend.
func (h *hederaBackend) Metadata() BackendMetadata {
	return BackendMetadata{
		DisplayName:    "Hedera Consensus Service",
		Description:    "Hedera HCS topic for consensus-based messaging with ordering guarantees",
		LocatorFormat:  "shard.realm.topic",
		LocatorExample: "0.0.12345",
		RequiresConfig: true, // Requires Hedera network credentials
		Version:        "1.0.0",
		Properties: map[string]string{
			"network":      "mainnet, testnet, or previewnet",
			"orderingType": "consensus-time",
		},
	}
}

// Technology returns the TopicTechnology for the Hedera backend.
func (h *hederaBackend) Technology() TopicTechnology {
	return TopicTechnologyHedera
}

// NewHederaTopicAddress creates a CommAddress for a Hedera Consensus Service topic.
// The locator must be in format "shard.realm.topic" (e.g., "0.0.12345").
func NewHederaTopicAddress(locator string) (CommAddress, error) {
	return NewCommAddress(HederaTopicKind, locator)
}

// ValidateHederaTopicLocator validates a Hedera topic ID format.
// Valid format: "shard.realm.topic" where each is a non-negative integer.
//
// This function delegates to the registered Hedera backend for validation.
func ValidateHederaTopicLocator(locator string) error {
	backend, ok := GetBackend(HederaTopicKind)
	if !ok {
		return errInvalidAddress("ValidateHederaTopicLocator", "hedera backend not registered", nil)
	}
	return backend.ValidateLocator(locator)
}
