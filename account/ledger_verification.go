package account

import (
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// LedgerVerifier defines the interface for verifying account-ledger relationships.
// Implementations of this interface connect to specific ledgers (Ethereum, Hedera, etc.)
// to verify that accounts exist and that key ownership can be proven.
//
// This interface is used for:
//   - Verifying that an account exists on a specific ledger
//   - Proving key ownership (that a Neuron private key corresponds to a ledger account)
//   - Verifying semantic consistency between Neuron and ledger public key derivations
//   - Verifying parent-child relationships on-ledger or in account registries
//
// Implementations are expected to be provided by ledger-specific packages.
type LedgerVerifier interface {
	// LedgerIdentifier returns the identifier for this ledger verifier.
	// Examples: "ethereum-mainnet", "ethereum-goerli", "hedera-mainnet", "hedera-testnet"
	LedgerIdentifier() string

	// VerifyAccountExists checks if an account exists on the ledger.
	// The address is the ledger-specific account identifier (e.g., Ethereum address).
	// Returns true if the account exists, false otherwise.
	// Returns error if the verification cannot be completed (network error, etc.).
	VerifyAccountExists(address string) (bool, error)

	// VerifyKeyOwnership verifies that a Neuron public key corresponds to a ledger account.
	// This proves that the owner of the Neuron private key also controls the ledger account.
	// The address is the ledger-specific account identifier.
	// Returns true if the public key derives to the given address, false otherwise.
	VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error)

	// VerifyMultisigOwnership verifies that a MultisigKey configuration matches a ledger account.
	// This is used for Shared accounts attached to ledger multisig accounts.
	// The address is the ledger-specific multisig account identifier.
	// Returns true if the MultisigKey matches the ledger's multisig structure, false otherwise.
	VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error)

	// VerifySemanticConsistency verifies that the public key semantics match between
	// Neuron and the ledger. Specifically, it checks that the public key that derives
	// the address in Neuron matches the ledger account's public key semantics.
	// This is required by FR-019 to ensure consistent key derivation.
	// Returns true if semantics match, false if they don't match.
	// Returns error if verification cannot be completed.
	VerifySemanticConsistency(address string, pubKey keylib.NeuronPublicKey) (bool, error)

	// VerifyParentChildRelationship verifies that a parent-child relationship exists
	// on the ledger or in an account registry.
	// This is required by FR-017 for verifying hierarchical account relationships.
	// Returns true if the relationship is valid, false otherwise.
	// Returns error if verification cannot be completed.
	VerifyParentChildRelationship(parentAddr, childAddr string) (bool, error)
}

// VerificationResult represents the result of a ledger verification operation.
type VerificationResult struct {
	// Verified indicates whether the verification succeeded.
	Verified bool

	// Message provides additional information about the verification result.
	Message string

	// Error contains any error that occurred during verification.
	Error error
}

// IsSuccess returns true if the verification succeeded without errors.
func (r *VerificationResult) IsSuccess() bool {
	return r.Verified && r.Error == nil
}

// VerifyAccountAttachment performs comprehensive verification of an account's ledger attachment.
// It verifies:
//  1. The account exists on the ledger
//  2. The Neuron key corresponds to the ledger account
//  3. The public key semantics match
//
// For Shared accounts, it also verifies the MultisigKey configuration.
func VerifyAccountAttachment(account NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error) {
	const op = "VerifyAccountAttachment"

	if account.IsZero() {
		return nil, errZeroValue(op, "NeuronAccount")
	}

	attachment := account.LedgerAttachment()
	if attachment == nil {
		return nil, errMissingRequired(op, "LedgerAttachment")
	}

	// Verify the verifier matches the attachment's ledger
	if verifier.LedgerIdentifier() != attachment.LedgerIdentifier() {
		return &VerificationResult{
			Verified: false,
			Message:  "verifier ledger does not match attachment ledger",
		}, nil
	}

	address := attachment.AttachedAddress()

	// Step 1: Verify account exists
	exists, err := verifier.VerifyAccountExists(address)
	if err != nil {
		return &VerificationResult{
			Verified: false,
			Message:  "failed to verify account existence",
			Error:    err,
		}, nil
	}
	if !exists {
		return &VerificationResult{
			Verified: false,
			Message:  "account does not exist on ledger",
		}, nil
	}

	// Step 2: Verify key ownership based on account type
	if account.IsShared() {
		// Shared accounts use MultisigKey
		multisigKey := account.MultisigKey()
		if multisigKey == nil || multisigKey.IsZero() {
			return &VerificationResult{
				Verified: false,
				Message:  "shared account missing MultisigKey",
			}, nil
		}

		matches, err := verifier.VerifyMultisigOwnership(address, *multisigKey)
		if err != nil {
			return &VerificationResult{
				Verified: false,
				Message:  "failed to verify multisig ownership",
				Error:    err,
			}, nil
		}
		if !matches {
			return &VerificationResult{
				Verified: false,
				Message:  "MultisigKey does not match ledger multisig structure",
			}, nil
		}
	} else {
		// Parent and Child accounts use single public key
		pubKey := account.PublicKey()
		if pubKey.IsZero() {
			return &VerificationResult{
				Verified: false,
				Message:  "account missing public key",
			}, nil
		}

		matches, err := verifier.VerifyKeyOwnership(address, pubKey)
		if err != nil {
			return &VerificationResult{
				Verified: false,
				Message:  "failed to verify key ownership",
				Error:    err,
			}, nil
		}
		if !matches {
			return &VerificationResult{
				Verified: false,
				Message:  "public key does not derive to attached address",
			}, nil
		}

		// Step 3: Verify semantic consistency
		semanticsMatch, err := verifier.VerifySemanticConsistency(address, pubKey)
		if err != nil {
			return &VerificationResult{
				Verified: false,
				Message:  "failed to verify semantic consistency",
				Error:    err,
			}, nil
		}
		if !semanticsMatch {
			return &VerificationResult{
				Verified: false,
				Message:  "public key semantics do not match between Neuron and ledger",
			}, nil
		}
	}

	return &VerificationResult{
		Verified: true,
		Message:  "account attachment verified successfully",
	}, nil
}

// VerifyParentChildLink verifies the parent-child relationship between two accounts.
// Both accounts must be attached to the same ledger.
func VerifyParentChildLink(parent, child NeuronAccount, verifier LedgerVerifier) (*VerificationResult, error) {
	const op = "VerifyParentChildLink"

	if parent.IsZero() {
		return nil, errZeroValue(op, "parent NeuronAccount")
	}
	if child.IsZero() {
		return nil, errZeroValue(op, "child NeuronAccount")
	}

	// Verify parent is actually a parent account
	if !parent.IsParent() {
		return &VerificationResult{
			Verified: false,
			Message:  "first account is not a Parent account",
		}, nil
	}

	// Verify child is actually a child account
	if !child.IsChild() {
		return &VerificationResult{
			Verified: false,
			Message:  "second account is not a Child account",
		}, nil
	}

	// Verify object-level relationship
	if !child.ParentPublicKey().Equal(parent.PublicKey()) {
		return &VerificationResult{
			Verified: false,
			Message:  "child's parent reference does not match parent's public key",
		}, nil
	}

	// Get ledger attachments
	parentAttachment := parent.LedgerAttachment()
	childAttachment := child.LedgerAttachment()

	// If both have attachments, verify on-ledger relationship
	if parentAttachment != nil && childAttachment != nil {
		if parentAttachment.LedgerIdentifier() != childAttachment.LedgerIdentifier() {
			return &VerificationResult{
				Verified: false,
				Message:  "parent and child are attached to different ledgers",
			}, nil
		}

		matches, err := verifier.VerifyParentChildRelationship(
			parentAttachment.AttachedAddress(),
			childAttachment.AttachedAddress(),
		)
		if err != nil {
			return &VerificationResult{
				Verified: false,
				Message:  "failed to verify on-ledger parent-child relationship",
				Error:    err,
			}, nil
		}
		if !matches {
			return &VerificationResult{
				Verified: false,
				Message:  "parent-child relationship not verified on ledger",
			}, nil
		}
	}

	return &VerificationResult{
		Verified: true,
		Message:  "parent-child relationship verified",
	}, nil
}
