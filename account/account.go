package account

import (
	"encoding/json"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// NeuronAccount is the portable, blockchain-agnostic representation of an agent's
// identity and endpoints.
//
// A NeuronAccount:
//   - Is identified by a Neuron Key (NeuronPublicKey from keylib)
//   - Does NOT perform communication or message passing
//   - Describes WHERE communication happens and HOW the agent can be reached
//   - Can be anchored to one or more blockchains
//
// The Neuron key is the root of identity. All other identifiers (PeerID, EVMAddress)
// are derived from it, never authoritative on their own.
type NeuronAccount struct {
	// === Identity (from keylib) ===

	// publicKey is the root of identity for this account.
	// All other identifiers are derived from this key.
	publicKey keylib.NeuronPublicKey

	// peerID is the libp2p peer ID, derived from publicKey.
	// Cached for efficiency; always derivable from publicKey.
	peerID keylib.PeerID

	// evmAddress is the Ethereum-compatible address, derived from publicKey.
	// Cached for efficiency; always derivable from publicKey.
	evmAddress keylib.EVMAddress

	// === Hierarchy ===

	// accountType indicates whether this is a Parent or Child account.
	accountType AccountType

	// did is the Decentralized Identifier for this account.
	// Required for Parent accounts; empty for Child accounts.
	did NeuronDID

	// parentPubKey is the public key of the parent account.
	// Only set for Child accounts; zero-value for Parent accounts.
	parentPubKey keylib.NeuronPublicKey

	// === Communication Endpoints ===

	// stdIn is where others send messages to this agent.
	stdIn CommAddress

	// stdOut is where this agent publishes outputs/heartbeats.
	stdOut CommAddress

	// stdErr is where this agent publishes errors/diagnostics.
	stdErr CommAddress

	// === Reachability ===

	// reachableAddrs are the canonical "find/dial me" endpoints.
	// All addresses must have PeerIDs matching this account's derived PeerID.
	reachableAddrs ReachableAddrs
}

// === Identity Accessors ===

// PublicKey returns the NeuronPublicKey that is the root of this account's identity.
func (a NeuronAccount) PublicKey() keylib.NeuronPublicKey {
	return a.publicKey
}

// PeerID returns the libp2p PeerID derived from the public key.
// This is cached but always derivable from the public key.
func (a NeuronAccount) PeerID() keylib.PeerID {
	return a.peerID
}

// EVMAddress returns the Ethereum-compatible address derived from the public key.
// This is cached but always derivable from the public key.
func (a NeuronAccount) EVMAddress() keylib.EVMAddress {
	return a.evmAddress
}

// === Hierarchy Accessors ===

// AccountType returns whether this is a Parent or Child account.
func (a NeuronAccount) AccountType() AccountType {
	return a.accountType
}

// IsParent returns true if this is a Parent account.
func (a NeuronAccount) IsParent() bool {
	return a.accountType == AccountTypeParent
}

// IsChild returns true if this is a Child account.
func (a NeuronAccount) IsChild() bool {
	return a.accountType == AccountTypeChild
}

// DID returns the Decentralized Identifier for this account.
// Returns nil for Child accounts.
func (a NeuronAccount) DID() NeuronDID {
	return a.did
}

// ParentPublicKey returns the parent's public key for Child accounts.
// Returns zero-value for Parent accounts.
func (a NeuronAccount) ParentPublicKey() keylib.NeuronPublicKey {
	return a.parentPubKey
}

// === Communication Endpoint Accessors ===

// StdIn returns the address where others send messages to this agent.
func (a NeuronAccount) StdIn() CommAddress {
	return a.stdIn
}

// StdOut returns the address where this agent publishes outputs.
func (a NeuronAccount) StdOut() CommAddress {
	return a.stdOut
}

// StdErr returns the address where this agent publishes errors.
func (a NeuronAccount) StdErr() CommAddress {
	return a.stdErr
}

// === Reachability Accessors ===

// ReachableAddrs returns the collection of reachable connection addresses.
func (a NeuronAccount) ReachableAddrs() ReachableAddrs {
	return a.reachableAddrs
}

// FirstReachableAddr returns the first reachable address, or zero value if none.
func (a NeuronAccount) FirstReachableAddr() ReachableAddr {
	return a.reachableAddrs.First()
}

// === Validation ===

// IsZero returns true if this is a zero-value NeuronAccount.
func (a NeuronAccount) IsZero() bool {
	return a.publicKey.IsZero()
}

// Validate performs comprehensive validation of this account.
// Returns the first validation error encountered, or nil if valid.
func (a NeuronAccount) Validate() error {
	const op = "NeuronAccount.Validate"

	// Validate public key (required for all accounts)
	if a.publicKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	// Validate account type
	if !a.accountType.IsValid() {
		return errInvalidAccount(op, "invalid account type")
	}

	// Validate hierarchy-specific requirements
	if a.accountType == AccountTypeParent {
		if err := ValidateParentAccount(a.publicKey, a.did); err != nil {
			return wrapAccountError(op, ErrKindValidation, "parent account validation failed", err)
		}
	} else if a.accountType == AccountTypeChild {
		if err := ValidateChildAccount(a.publicKey, a.parentPubKey); err != nil {
			return wrapAccountError(op, ErrKindValidation, "child account validation failed", err)
		}
	}

	// Validate DID matches public key (for Parent accounts with DID)
	if a.did != nil {
		if err := ValidateDIDMatchesKey(a.did, a.publicKey); err != nil {
			return wrapAccountError(op, ErrKindInvalidDID, "DID-key mismatch", err)
		}
	}

	// Validate communication addresses (if set)
	if !a.stdIn.IsZero() {
		if err := a.stdIn.Validate(); err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdIn address", err)
		}
	}
	if !a.stdOut.IsZero() {
		if err := a.stdOut.Validate(); err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdOut address", err)
		}
	}
	if !a.stdErr.IsZero() {
		if err := a.stdErr.Validate(); err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdErr address", err)
		}
	}

	// Validate reachable addresses (PeerID consistency)
	if !a.reachableAddrs.IsEmpty() {
		if err := a.reachableAddrs.ValidateAll(a.peerID); err != nil {
			return wrapAccountError(op, ErrKindPeerIDMismatch, "reachable address validation failed", err)
		}
	}

	return nil
}

// ValidateAll performs comprehensive validation and returns all errors.
// Use this when you want to see all validation issues at once.
func (a NeuronAccount) ValidateAll() *ValidationResult {
	result := NewValidationResult()

	// Validate public key
	if a.publicKey.IsZero() {
		result.AddError(errZeroValue("ValidateAll", "NeuronPublicKey"))
	}

	// Validate account type
	result.AddError(a.accountType.Validate())

	// Validate hierarchy-specific requirements
	if a.accountType == AccountTypeParent {
		result.AddError(ValidateParentAccount(a.publicKey, a.did))
	} else if a.accountType == AccountTypeChild {
		result.AddError(ValidateChildAccount(a.publicKey, a.parentPubKey))
	}

	// Validate DID matches key
	if a.did != nil {
		result.AddError(ValidateDIDMatchesKey(a.did, a.publicKey))
	}

	// Validate communication addresses
	if !a.stdIn.IsZero() {
		result.AddError(a.stdIn.Validate())
	}
	if !a.stdOut.IsZero() {
		result.AddError(a.stdOut.Validate())
	}
	if !a.stdErr.IsZero() {
		result.AddError(a.stdErr.Validate())
	}

	// Validate reachable addresses
	if !a.reachableAddrs.IsEmpty() && !a.peerID.IsZero() {
		result.AddError(a.reachableAddrs.ValidateAll(a.peerID))
	}

	return result
}

// === Comparison ===

// Equal compares this account with another for identity equality.
// Two accounts are equal if they have the same public key.
func (a NeuronAccount) Equal(other NeuronAccount) bool {
	if a.IsZero() || other.IsZero() {
		return false
	}
	return a.publicKey.Equal(other.publicKey)
}

// === Serialization ===

// neuronAccountJSON is the JSON representation of NeuronAccount.
type neuronAccountJSON struct {
	PublicKey      string   `json:"publicKey"`
	PeerID         string   `json:"peerId"`
	EVMAddress     string   `json:"evmAddress"`
	AccountType    string   `json:"accountType"`
	DID            string   `json:"did,omitempty"`
	ParentPubKey   string   `json:"parentPublicKey,omitempty"`
	StdIn          string   `json:"stdIn,omitempty"`
	StdOut         string   `json:"stdOut,omitempty"`
	StdErr         string   `json:"stdErr,omitempty"`
	ReachableAddrs []string `json:"reachableAddrs,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (a NeuronAccount) MarshalJSON() ([]byte, error) {
	j := neuronAccountJSON{
		PublicKey:   a.publicKey.Hex(),
		PeerID:      a.peerID.String(),
		EVMAddress:  a.evmAddress.Hex(),
		AccountType: a.accountType.String(),
	}

	if a.did != nil {
		j.DID = a.did.String()
	}

	if !a.parentPubKey.IsZero() {
		j.ParentPubKey = a.parentPubKey.Hex()
	}

	if !a.stdIn.IsZero() {
		j.StdIn = a.stdIn.String()
	}
	if !a.stdOut.IsZero() {
		j.StdOut = a.stdOut.String()
	}
	if !a.stdErr.IsZero() {
		j.StdErr = a.stdErr.String()
	}

	if !a.reachableAddrs.IsEmpty() {
		j.ReachableAddrs = a.reachableAddrs.Strings()
	}

	return json.Marshal(j)
}

// String returns a summary string representation of this account.
func (a NeuronAccount) String() string {
	if a.IsZero() {
		return "NeuronAccount{zero-value}"
	}

	typeStr := a.accountType.String()
	if a.did != nil {
		return "NeuronAccount{" + typeStr + ", did=" + a.did.String() + "}"
	}
	return "NeuronAccount{" + typeStr + ", pubKey=" + a.publicKey.Hex()[:16] + "...}"
}
