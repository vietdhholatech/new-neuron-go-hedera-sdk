package account

// Backend defines a pluggable communication backend.
//
// Implementations register themselves via init() functions, allowing
// new messaging technologies to be added without modifying the core
// account package.
//
// # Backend Lifecycle
//
// The backend system follows this lifecycle:
//
//  1. Registration (Package Init): Backends call RegisterBackend() in init()
//     - Registry validates metadata completeness
//     - Panics on duplicate registration or incomplete metadata
//  2. Construction (NewCommAddress): Backend parses and validates locator
//     - Calls backend.ParseLocator() for registered backends
//     - Allows unknown backends with non-empty locator (forward compatibility)
//  3. Validation (CommAddress.Validate): Re-validates locator format
//     - Calls backend.ValidateLocator() for registered backends
//  4. Technology Mapping: Maps CommAddressKind to TopicTechnology
//     - Calls backend.Technology() for registered backends
//
// Example implementation:
//
//	func init() {
//	    account.RegisterBackend(&myBackend{})
//	}
//
//	type myBackend struct{}
//
//	func (m *myBackend) Kind() string { return "my-backend" }
//	// ... implement other methods
type Backend interface {
	// Kind returns the unique identifier for this backend.
	// This is used as the CommAddressKind value.
	// Examples: "hedera-topic", "kafka-topic", "pulsar-topic"
	Kind() string

	// ValidateLocator checks if a locator string is valid for this backend.
	// Returns nil if valid, or an error describing the validation failure.
	ValidateLocator(locator string) error

	// ParseLocator parses and normalizes a locator string.
	// This allows backends to canonicalize their locators.
	// For most backends, this simply validates and returns the locator unchanged.
	//
	// Idempotency Contract:
	// ParseLocator must be idempotent - calling it multiple times with the
	// same valid input must return the same result. That is, for any valid
	// locator string s:
	//   ParseLocator(s) == ParseLocator(ParseLocator(s))
	ParseLocator(locator string) (string, error)

	// Metadata returns information about this backend.
	// This is used for discovery, documentation, and introspection.
	Metadata() BackendMetadata

	// Technology returns the TopicTechnology for this backend.
	// This enables dynamic mapping from CommAddressKind to TopicTechnology
	// without requiring hardcoded switch statements in core files.
	Technology() TopicTechnology
}

// BackendMetadata provides descriptive information about a backend.
// This metadata is used for documentation, introspection, and tooling.
type BackendMetadata struct {
	// DisplayName is a human-readable name for the backend.
	// Example: "Apache Kafka", "Hedera Consensus Service"
	DisplayName string

	// Description is a short description of the backend's purpose.
	// Example: "High-throughput distributed streaming platform"
	Description string

	// LocatorFormat describes the expected format of locator strings.
	// Example: "shard.realm.topic" or "topic-name"
	LocatorFormat string

	// LocatorExample provides a sample valid locator.
	// Example: "0.0.12345" or "my-events-topic"
	LocatorExample string

	// RequiresConfig indicates whether this backend needs additional
	// configuration beyond the locator (e.g., connection strings, credentials).
	RequiresConfig bool

	// Version is the semantic version of the backend implementation.
	// Example: "1.0.0"
	// Empty string indicates an unversioned backend (for backward compatibility).
	Version string

	// Properties holds additional backend-specific properties.
	// This is an extension point for backend-specific metadata.
	Properties map[string]string
}
