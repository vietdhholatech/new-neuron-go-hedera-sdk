package account

import (
	"fmt"
)

// AccountType categorizes a NeuronAccount in the hierarchy.
//
// The NeuronAccount hierarchy is strictly two-level:
//   - Parent accounts have no parent, may have children, and must have a DID
//   - Child accounts have exactly one parent, no children, and no DID
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
	AccountTypeParent

	// AccountTypeChild indicates a subordinate account.
	// Child accounts:
	//   - Have exactly one parent account (referenced by parent's public key)
	//   - Cannot have children (no grandchildren in hierarchy)
	//   - Do not have their own DID
	AccountTypeChild
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
	default:
		return fmt.Sprintf("Unknown(%d)", int(t))
	}
}

// IsValid returns true if this is a valid, specified account type.
func (t AccountType) IsValid() bool {
	return t == AccountTypeParent || t == AccountTypeChild
}

// IsParent returns true if this is a parent account type.
func (t AccountType) IsParent() bool {
	return t == AccountTypeParent
}

// IsChild returns true if this is a child account type.
func (t AccountType) IsChild() bool {
	return t == AccountTypeChild
}

// RequiresDID returns true if accounts of this type must have a DID.
func (t AccountType) RequiresDID() bool {
	return t == AccountTypeParent
}

// RequiresParent returns true if accounts of this type must have a parent.
func (t AccountType) RequiresParent() bool {
	return t == AccountTypeChild
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
