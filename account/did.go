package account

import (
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// NeuronDID represents a Decentralized Identifier for a NeuronAccount.
// It is a technology-agnostic interface that can be implemented by various DID methods
// such as did:key, did:web, did:hedera, etc.
//
// A DID follows the W3C DID specification format: did:<method>:<method-specific-id>
// For example: did:key:z6MkpTHR8VNsBxYAAA...
type NeuronDID interface {
	// String returns the full DID string representation.
	// Example: "did:key:z6MkpTHR8VNsBxYAAA..."
	String() string

	// Method returns the DID method identifier.
	// Example: "key", "web", "hedera", "neuron"
	Method() string

	// Identifier returns the method-specific identifier portion.
	// For "did:key:z6MkpTHR...", this returns "z6MkpTHR..."
	Identifier() string

	// Validate checks if the DID is well-formed according to its method's rules.
	// Returns nil if valid, error otherwise.
	Validate() error

	// Equal compares this DID with another for equality.
	// Two DIDs are equal if they have the same method and identifier.
	Equal(other NeuronDID) bool
}

// NeuronDIDWithKey extends NeuronDID with public key operations.
// This interface is implemented by DID methods that encode or can derive
// the public key (e.g., did:key, did:pkh).
type NeuronDIDWithKey interface {
	NeuronDID

	// PublicKey returns the NeuronPublicKey associated with this DID.
	// For did:key, this is extracted directly from the DID.
	// For other methods, this may require resolution.
	// Returns error if the public key cannot be determined.
	PublicKey() (keylib.NeuronPublicKey, error)

	// MatchesKey checks if this DID corresponds to the given public key.
	// This is useful for validating that a DID belongs to an account.
	MatchesKey(pubKey keylib.NeuronPublicKey) bool
}

// DIDDocument represents a DID Document as per W3C DID Core specification.
// This is a simplified representation for the essential fields needed by NeuronAccount.
// Full DID Document support may be added in future phases.
type DIDDocument struct {
	// Context is the JSON-LD context, typically ["https://www.w3.org/ns/did/v1"]
	Context []string `json:"@context"`

	// ID is the DID this document describes
	ID string `json:"id"`

	// Controller is the DID(s) that control this DID document
	Controller []string `json:"controller,omitempty"`

	// VerificationMethod lists the cryptographic public keys
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`

	// Authentication lists verification methods for authentication
	Authentication []string `json:"authentication,omitempty"`

	// AssertionMethod lists verification methods for assertions
	AssertionMethod []string `json:"assertionMethod,omitempty"`

	// Service lists service endpoints
	Service []ServiceEndpoint `json:"service,omitempty"`
}

// VerificationMethod represents a cryptographic verification method in a DID Document.
type VerificationMethod struct {
	// ID is the verification method identifier (e.g., "did:key:z6Mk...#z6Mk...")
	ID string `json:"id"`

	// Type is the verification method type (e.g., "EcdsaSecp256k1VerificationKey2019")
	Type string `json:"type"`

	// Controller is the DID that controls this verification method
	Controller string `json:"controller"`

	// PublicKeyMultibase is the multibase-encoded public key (preferred)
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`

	// PublicKeyJwk is the public key in JWK format (alternative)
	PublicKeyJwk map[string]any `json:"publicKeyJwk,omitempty"`
}

// ServiceEndpoint represents a service endpoint in a DID Document.
type ServiceEndpoint struct {
	// ID is the service endpoint identifier
	ID string `json:"id"`

	// Type is the service type (e.g., "NeuronCommChannel")
	Type string `json:"type"`

	// ServiceEndpoint is the endpoint URL or structured data
	ServiceEndpoint any `json:"serviceEndpoint"`
}

// DIDMethodKey is the did:key method - self-describing, derived from public key.
const DIDMethodKey = "key"
