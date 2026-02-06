package account

import (
	"fmt"
)

// AccountType categorizes a NeuronAccount in the hierarchy.
//
// The NeuronAccount system supports three account types:
//   - Parent accounts: root identity with DID, credit balance, no comm channels
//   - Child accounts: subordinate with parent reference, all 3 comm channels required
//   - Shared accounts: multisig threshold accounts using MultisigKey, no DID, no comm channels
//
// This structure ensures clear ownership and administrative boundaries.
type AccountType int

const (
	// AccountTypeUnspecified is the zero value, indicating the type hasn't been set.
	// This is invalid for a fully constructed NeuronAccount.
	AccountTypeUnspecified AccountType = iota

	// AccountTypeParent indicates a top-level account.
	// Parent accounts:
	//   - Have no parent account
	//   - May have one or more Child accounts
	//   - Must include DID (Decentralized Identifier) information
	//   - Must NOT have communication channels (stdIn, stdOut, stdErr)
	//   - Have credit balance for financial operations
	AccountTypeParent

	// AccountTypeChild indicates a subordinate account.
	// Child accounts:
	//   - Have exactly one parent account (referenced by parent's public key)
	//   - Cannot have children (no grandchildren in hierarchy)
	//   - Do not have their own DID
	//   - Must have ALL three communication channels (stdIn, stdOut, stdErr)
	//   - Have balance allocation for operational expenses
	AccountTypeChild

	// AccountTypeShared indicates a multisig threshold account.
	// Shared accounts:
	//   - Use MultisigKey instead of single public key (M-of-N threshold)
	//   - Do NOT have a DID document
	//   - Do NOT have communication channels
	//   - Do NOT have a parent reference
	//   - Have balance for multisig-controlled operations
	AccountTypeShared
)

// String returns a human-readable name for the account type.
func (t AccountType) String() string {
	switch t {
	case AccountTypeUnspecified:
		return "Unspecified"
	case AccountTypeParent:
		return "Parent"
	case AccountTypeChild:
		return "Child"
	case AccountTypeShared:
		return "Shared"
	default:
		return fmt.Sprintf("Unknown(%d)", int(t))
	}
}

// IsValid returns true if this is a valid, specified account type.
func (t AccountType) IsValid() bool {
	return t == AccountTypeParent || t == AccountTypeChild || t == AccountTypeShared
}

// IsParent returns true if this is a parent account type.
func (t AccountType) IsParent() bool {
	return t == AccountTypeParent
}

// IsChild returns true if this is a child account type.
func (t AccountType) IsChild() bool {
	return t == AccountTypeChild
}

// IsShared returns true if this is a shared (multisig) account type.
func (t AccountType) IsShared() bool {
	return t == AccountTypeShared
}

// RequiresDID returns true if accounts of this type must have a DID.
func (t AccountType) RequiresDID() bool {
	return t == AccountTypeParent
}

// RequiresParent returns true if accounts of this type must have a parent.
func (t AccountType) RequiresParent() bool {
	return t == AccountTypeChild
}

// RequiresMultisigKey returns true if accounts of this type must have a MultisigKey.
func (t AccountType) RequiresMultisigKey() bool {
	return t == AccountTypeShared
}

// RequiresPublicKey returns true if accounts of this type must have a single public key.
// Shared accounts use MultisigKey instead.
func (t AccountType) RequiresPublicKey() bool {
	return t == AccountTypeParent || t == AccountTypeChild
}

// RequiresCommChannels returns true if accounts of this type must have communication channels.
// Only Child accounts require all three communication channels.
func (t AccountType) RequiresCommChannels() bool {
	return t == AccountTypeChild
}

// ProhibitsCommChannels returns true if accounts of this type must NOT have communication channels.
// Parent and Shared accounts must not have communication channels.
func (t AccountType) ProhibitsCommChannels() bool {
	return t == AccountTypeParent || t == AccountTypeShared
}

// Validate checks if the account type is valid.
func (t AccountType) Validate() error {
	const op = "AccountType.Validate"

	if !t.IsValid() {
		return errInvalidAccount(op,
			fmt.Sprintf("invalid account type: %s", t.String()))
	}

	return nil
}
