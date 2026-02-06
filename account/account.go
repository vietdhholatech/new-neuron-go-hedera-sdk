package account

import (
	"encoding/json"
	"math/big"

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
	// Required for Parent and Child accounts; zero for Shared accounts.
	publicKey keylib.NeuronPublicKey

	// multisigKey is the threshold signing configuration for Shared accounts.
	// Required for Shared accounts; nil for Parent and Child accounts.
	multisigKey *keylib.MultisigKey

	// peerID is the libp2p peer ID, derived from publicKey.
	// Cached for efficiency; always derivable from publicKey.
	peerID keylib.PeerID

	// evmAddress is the Ethereum-compatible address, derived from publicKey.
	// Cached for efficiency; always derivable from publicKey.
	evmAddress keylib.EVMAddress

	// === Hierarchy ===

	// accountType indicates whether this is a Parent, Child, or Shared account.
	accountType AccountType

	// did is the Decentralized Identifier for this account.
	// Required for Parent accounts; must be nil for Child and Shared accounts.
	did NeuronDID

	// parentPubKey is the public key of the parent account.
	// Only set for Child accounts; zero-value for Parent and Shared accounts.
	parentPubKey keylib.NeuronPublicKey

	// === Communication Endpoints ===

	// stdIn is where others send messages to this agent.
	// Required for Child accounts; must be empty for Parent and Shared accounts.
	stdIn CommAddress

	// stdOut is where this agent publishes outputs/heartbeats.
	// Required for Child accounts; must be empty for Parent and Shared accounts.
	stdOut CommAddress

	// stdErr is where this agent publishes errors/diagnostics.
	// Required for Child accounts; must be empty for Parent and Shared accounts.
	stdErr CommAddress

	// === Reachability ===

	// reachableAddrs are the canonical "find/dial me" endpoints.
	// All addresses must have PeerIDs matching this account's derived PeerID.
	reachableAddrs ReachableAddrs

	// === Financial Layer ===

	// currencySymbol identifies the currency for balance tracking (e.g., "HBAR", "ETH").
	currencySymbol string

	// creditBalance is the primary financial balance for Parent accounts.
	// Nil for Child and Shared accounts.
	creditBalance *big.Int

	// balanceAllocation is the allocated funds for Child account operational expenses.
	// Nil for Parent and Shared accounts.
	balanceAllocation *big.Int

	// balance is the multisig-controlled balance for Shared accounts.
	// Nil for Parent and Child accounts.
	balance *big.Int

	// === Ledger Integration Layer ===

	// ledgerAttachment links this account to settlement infrastructure.
	// Optional for all account types.
	ledgerAttachment *LedgerAttachment
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

// IsShared returns true if this is a Shared (multisig) account.
func (a NeuronAccount) IsShared() bool {
	return a.accountType == AccountTypeShared
}

// MultisigKey returns the threshold signing configuration for Shared accounts.
// Returns nil for Parent and Child accounts.
func (a NeuronAccount) MultisigKey() *keylib.MultisigKey {
	return a.multisigKey
}

// DID returns the Decentralized Identifier for this account.
// Returns nil for Child and Shared accounts.
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

// === Financial Accessors ===

// CurrencySymbol returns the currency symbol for balance tracking.
func (a NeuronAccount) CurrencySymbol() string {
	return a.currencySymbol
}

// CreditBalance returns the credit balance for Parent accounts.
// Returns nil for Child and Shared accounts.
// Callers must check for nil before dereferencing.
func (a NeuronAccount) CreditBalance() *big.Int {
	return a.creditBalance
}

// BalanceAllocation returns the allocated funds for Child accounts.
// Returns nil for Parent and Shared accounts.
// Callers must check for nil before dereferencing.
func (a NeuronAccount) BalanceAllocation() *big.Int {
	return a.balanceAllocation
}

// Balance returns the multisig-controlled balance for Shared accounts.
// Returns nil for Parent and Child accounts.
// Callers must check for nil before dereferencing.
func (a NeuronAccount) Balance() *big.Int {
	return a.balance
}

// LedgerAttachment returns the ledger attachment for this account.
// Returns nil if not attached to a ledger.
func (a NeuronAccount) LedgerAttachment() *LedgerAttachment {
	return a.ledgerAttachment
}

// === Validation ===

// IsZero returns true if this is a zero-value NeuronAccount.
// For Shared accounts, this checks the MultisigKey instead of the public key.
func (a NeuronAccount) IsZero() bool {
	// Shared accounts use MultisigKey instead of single public key
	if a.accountType == AccountTypeShared {
		return a.multisigKey == nil || a.multisigKey.IsZero()
	}
	return a.publicKey.IsZero()
}

// Validate performs comprehensive validation of this account.
// Returns the first validation error encountered, or nil if valid.
func (a NeuronAccount) Validate() error {
	const op = "NeuronAccount.Validate"

	// Validate account type
	if !a.accountType.IsValid() {
		return errInvalidAccount(op, "invalid account type")
	}

	// Validate public key (required for Parent and Child accounts, not for Shared)
	if a.accountType.RequiresPublicKey() && a.publicKey.IsZero() {
		return errZeroValue(op, "NeuronPublicKey")
	}

	// Validate hierarchy-specific requirements
	switch a.accountType {
	case AccountTypeParent:
		if err := ValidateParentAccount(a.publicKey, a.did, a.stdIn, a.stdOut, a.stdErr, a.parentPubKey); err != nil {
			return wrapAccountError(op, ErrKindValidation, "parent account validation failed", err)
		}
	case AccountTypeChild:
		if err := ValidateChildAccount(a.publicKey, a.parentPubKey, a.stdIn, a.stdOut, a.stdErr, a.did); err != nil {
			return wrapAccountError(op, ErrKindValidation, "child account validation failed", err)
		}
	case AccountTypeShared:
		if err := ValidateSharedAccount(a.multisigKey, a.did, a.stdIn, a.stdOut, a.stdErr, a.parentPubKey); err != nil {
			return wrapAccountError(op, ErrKindValidation, "shared account validation failed", err)
		}
	}

	// Validate DID matches public key (for Parent accounts with DID)
	if a.did != nil && !a.publicKey.IsZero() {
		if err := ValidateDIDMatchesKey(a.did, a.publicKey); err != nil {
			return wrapAccountError(op, ErrKindInvalidDID, "DID-key mismatch", err)
		}
	}

	// Validate communication addresses (if set and allowed)
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

	// Validate currency symbol (required when ledger is attached, FR-020)
	if err := ValidateCurrencySymbol(a.currencySymbol, a.ledgerAttachment); err != nil {
		return wrapAccountError(op, ErrKindMissingRequired, "currency symbol validation failed", err)
	}

	return nil
}

// ValidateAll performs comprehensive validation and returns all errors.
// Use this when you want to see all validation issues at once.
func (a NeuronAccount) ValidateAll() *ValidationResult {
	result := NewValidationResult()

	// Validate account type
	result.AddError(a.accountType.Validate())

	// Validate public key (required for Parent and Child accounts, not for Shared)
	if a.accountType.RequiresPublicKey() && a.publicKey.IsZero() {
		result.AddError(errZeroValue("ValidateAll", "NeuronPublicKey"))
	}

	// Validate hierarchy-specific requirements
	switch a.accountType {
	case AccountTypeParent:
		result.AddError(ValidateParentAccount(a.publicKey, a.did, a.stdIn, a.stdOut, a.stdErr, a.parentPubKey))
	case AccountTypeChild:
		result.AddError(ValidateChildAccount(a.publicKey, a.parentPubKey, a.stdIn, a.stdOut, a.stdErr, a.did))
	case AccountTypeShared:
		result.AddError(ValidateSharedAccount(a.multisigKey, a.did, a.stdIn, a.stdOut, a.stdErr, a.parentPubKey))
	}

	// Validate DID matches key (for Parent accounts with DID)
	if a.did != nil && !a.publicKey.IsZero() {
		result.AddError(ValidateDIDMatchesKey(a.did, a.publicKey))
	}

	// Validate communication addresses (if set and allowed)
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

	// Validate currency symbol (required when ledger is attached, FR-020)
	result.AddError(ValidateCurrencySymbol(a.currencySymbol, a.ledgerAttachment))

	return result
}

// === Comparison ===

// Equal compares this account with another for identity equality.
// For Parent/Child accounts, equality is based on the public key.
// For Shared accounts, equality is based on the MultisigKey configuration.
func (a NeuronAccount) Equal(other NeuronAccount) bool {
	if a.IsZero() || other.IsZero() {
		return false
	}
	// For Shared accounts, compare multisig configuration
	if a.accountType == AccountTypeShared && other.accountType == AccountTypeShared {
		if a.multisigKey == nil || other.multisigKey == nil {
			return false
		}
		return a.multisigKey.Equal(*other.multisigKey)
	}
	// For Parent/Child accounts, compare public keys
	if a.accountType != AccountTypeShared && other.accountType != AccountTypeShared {
		return a.publicKey.Equal(other.publicKey)
	}
	// Different account types are never equal
	return false
}

// === Serialization ===

// neuronAccountJSON is the JSON representation of NeuronAccount.
type neuronAccountJSON struct {
	// Identity
	PublicKey   string `json:"publicKey,omitempty"`
	PeerID      string `json:"peerId,omitempty"`
	EVMAddress  string `json:"evmAddress,omitempty"`
	AccountType string `json:"accountType"`

	// Multisig (for Shared accounts)
	MultisigThreshold int      `json:"multisigThreshold,omitempty"`
	MultisigTotal     int      `json:"multisigTotal,omitempty"`
	MultisigKeys      []string `json:"multisigKeys,omitempty"`
	MultisigProtocol  string   `json:"multisigProtocol,omitempty"`

	// Hierarchy
	DID          string `json:"did,omitempty"`
	ParentPubKey string `json:"parentPublicKey,omitempty"`

	// Communication
	StdIn          string   `json:"stdIn,omitempty"`
	StdOut         string   `json:"stdOut,omitempty"`
	StdErr         string   `json:"stdErr,omitempty"`
	ReachableAddrs []string `json:"reachableAddrs,omitempty"`

	// Financial
	CurrencySymbol    string `json:"currencySymbol,omitempty"`
	CreditBalance     string `json:"creditBalance,omitempty"`
	BalanceAllocation string `json:"balanceAllocation,omitempty"`
	Balance           string `json:"balance,omitempty"`

	// Ledger
	LedgerIdentifier   string `json:"ledgerIdentifier,omitempty"`
	AttachedAddress    string `json:"attachedAddress,omitempty"`
	AttachmentState    string `json:"attachmentState,omitempty"`
	VerificationStatus string `json:"verificationStatus,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (a NeuronAccount) MarshalJSON() ([]byte, error) {
	j := neuronAccountJSON{
		AccountType: a.accountType.String(),
	}

	// Identity fields
	if !a.publicKey.IsZero() {
		j.PublicKey = a.publicKey.Hex()
		j.PeerID = a.peerID.String()
		j.EVMAddress = a.evmAddress.Hex()
	}

	// Multisig fields (for Shared accounts)
	if a.multisigKey != nil && !a.multisigKey.IsZero() {
		j.MultisigThreshold = a.multisigKey.Threshold()
		j.MultisigTotal = a.multisigKey.Total()
		j.MultisigProtocol = a.multisigKey.Protocol()
		keys := a.multisigKey.PublicKeys()
		j.MultisigKeys = make([]string, len(keys))
		for i, k := range keys {
			j.MultisigKeys[i] = k.Hex()
		}
	}

	// Hierarchy fields
	if a.did != nil {
		j.DID = a.did.String()
	}
	if !a.parentPubKey.IsZero() {
		j.ParentPubKey = a.parentPubKey.Hex()
	}

	// Communication fields
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

	// Financial fields
	if a.currencySymbol != "" {
		j.CurrencySymbol = a.currencySymbol
	}
	if a.creditBalance != nil {
		j.CreditBalance = a.creditBalance.String()
	}
	if a.balanceAllocation != nil {
		j.BalanceAllocation = a.balanceAllocation.String()
	}
	if a.balance != nil {
		j.Balance = a.balance.String()
	}

	// Ledger fields
	if a.ledgerAttachment != nil {
		j.LedgerIdentifier = a.ledgerAttachment.LedgerIdentifier()
		j.AttachedAddress = a.ledgerAttachment.AttachedAddress()
		j.AttachmentState = a.ledgerAttachment.State().String()
		j.VerificationStatus = a.ledgerAttachment.VerificationStatus().String()
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
	if a.multisigKey != nil && !a.multisigKey.IsZero() {
		return "NeuronAccount{" + typeStr + ", " + a.multisigKey.String() + "}"
	}
	return "NeuronAccount{" + typeStr + ", pubKey=" + a.publicKey.Hex()[:16] + "...}"
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *NeuronAccount) UnmarshalJSON(data []byte) error {
	const op = "NeuronAccount.UnmarshalJSON"

	var j neuronAccountJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return wrapAccountError(op, ErrKindValidation, "failed to parse JSON", err)
	}

	// Parse account type
	switch j.AccountType {
	case "Parent":
		a.accountType = AccountTypeParent
	case "Child":
		a.accountType = AccountTypeChild
	case "Shared":
		a.accountType = AccountTypeShared
	default:
		return errInvalidAccount(op, "unknown account type: "+j.AccountType)
	}

	// Parse public key (for Parent and Child accounts)
	if j.PublicKey != "" {
		pubKey, err := keylib.ParsePublicKeyHex(j.PublicKey)
		if err != nil {
			return wrapAccountError(op, ErrKindValidation, "invalid public key", err)
		}
		a.publicKey = pubKey

		// Derive PeerID and EVMAddress
		peerID, err := pubKey.PeerID()
		if err != nil {
			return wrapAccountError(op, ErrKindValidation, "failed to derive PeerID", err)
		}
		a.peerID = peerID
		a.evmAddress = pubKey.EVMAddress()
	}

	// Parse multisig key (for Shared accounts)
	if len(j.MultisigKeys) > 0 && j.MultisigThreshold > 0 {
		keys := make([]keylib.NeuronPublicKey, len(j.MultisigKeys))
		for i, hexKey := range j.MultisigKeys {
			pubKey, err := keylib.ParsePublicKeyHex(hexKey)
			if err != nil {
				return wrapAccountError(op, ErrKindValidation, "invalid multisig key", err)
			}
			keys[i] = pubKey
		}
		var multisigKey keylib.MultisigKey
		var err error
		if j.MultisigProtocol != "" {
			multisigKey, err = keylib.NewMultisigKeyWithProtocol(keys, j.MultisigThreshold, j.MultisigProtocol)
		} else {
			// Backward compat: old JSON without protocol defaults to secp256k1-aggregated
			multisigKey, err = keylib.NewMultisigKey(keys, j.MultisigThreshold)
		}
		if err != nil {
			return wrapAccountError(op, ErrKindValidation, "invalid multisig configuration", err)
		}
		a.multisigKey = &multisigKey
	}

	// Parse parent public key
	if j.ParentPubKey != "" {
		parentPubKey, err := keylib.ParsePublicKeyHex(j.ParentPubKey)
		if err != nil {
			return wrapAccountError(op, ErrKindValidation, "invalid parent public key", err)
		}
		a.parentPubKey = parentPubKey
	}

	// Parse DID (for Parent accounts)
	if j.DID != "" {
		did, err := ParseDID(j.DID)
		if err != nil {
			return wrapAccountError(op, ErrKindInvalidDID, "invalid DID", err)
		}
		a.did = did
	}

	// Parse communication addresses
	if j.StdIn != "" {
		addr, err := ParseCommAddress(j.StdIn)
		if err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdIn", err)
		}
		a.stdIn = addr
	}
	if j.StdOut != "" {
		addr, err := ParseCommAddress(j.StdOut)
		if err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdOut", err)
		}
		a.stdOut = addr
	}
	if j.StdErr != "" {
		addr, err := ParseCommAddress(j.StdErr)
		if err != nil {
			return wrapAccountError(op, ErrKindInvalidAddress, "invalid stdErr", err)
		}
		a.stdErr = addr
	}

	// Parse reachable addresses
	if len(j.ReachableAddrs) > 0 {
		addrs := make([]ReachableAddr, 0, len(j.ReachableAddrs))
		for _, maStr := range j.ReachableAddrs {
			addr, err := ParseReachableAddr(maStr)
			if err != nil {
				return wrapAccountError(op, ErrKindInvalidAddress, "invalid reachable address", err)
			}
			addrs = append(addrs, addr)
		}
		a.reachableAddrs = NewReachableAddrs(addrs...)
	}

	// Parse financial fields
	a.currencySymbol = j.CurrencySymbol

	if j.CreditBalance != "" {
		balance := new(big.Int)
		if _, ok := balance.SetString(j.CreditBalance, 10); !ok {
			return errInvalidAccount(op, "invalid credit balance format")
		}
		a.creditBalance = balance
	}

	if j.BalanceAllocation != "" {
		allocation := new(big.Int)
		if _, ok := allocation.SetString(j.BalanceAllocation, 10); !ok {
			return errInvalidAccount(op, "invalid balance allocation format")
		}
		a.balanceAllocation = allocation
	}

	if j.Balance != "" {
		balance := new(big.Int)
		if _, ok := balance.SetString(j.Balance, 10); !ok {
			return errInvalidAccount(op, "invalid balance format")
		}
		a.balance = balance
	}

	// Parse ledger attachment
	if j.LedgerIdentifier != "" && j.AttachedAddress != "" {
		attachment, err := NewLedgerAttachment(j.LedgerIdentifier, j.AttachedAddress)
		if err != nil {
			return wrapAccountError(op, ErrKindValidation, "invalid ledger attachment", err)
		}
		// Set the attachment state based on the JSON value
		switch j.AttachmentState {
		case "Attached":
			if err := attachment.SetAttached(); err != nil {
				return wrapAccountError(op, ErrKindValidation, "failed to restore attachment state", err)
			}
		case "Verified":
			if err := attachment.SetAttached(); err != nil {
				return wrapAccountError(op, ErrKindValidation, "failed to restore attachment state", err)
			}
			if err := attachment.SetVerified(); err != nil {
				return wrapAccountError(op, ErrKindValidation, "failed to restore attachment state", err)
			}
		}
		a.ledgerAttachment = attachment
	}

	return nil
}
