package account

import (
	"errors"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Mock LedgerVerifier for testing
// =============================================================================

type mockLedgerVerifier struct {
	ledgerID               string
	accountExists          bool
	keyOwnershipValid      bool
	multisigOwnershipValid bool
	semanticConsistency    bool
	parentChildValid       bool
	err                    error
}

func (m *mockLedgerVerifier) LedgerIdentifier() string {
	return m.ledgerID
}

func (m *mockLedgerVerifier) VerifyAccountExists(address string) (bool, error) {
	return m.accountExists, m.err
}

func (m *mockLedgerVerifier) VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error) {
	return m.keyOwnershipValid, m.err
}

func (m *mockLedgerVerifier) VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error) {
	return m.multisigOwnershipValid, m.err
}

func (m *mockLedgerVerifier) VerifySemanticConsistency(address string, pubKey keylib.NeuronPublicKey) (bool, error) {
	return m.semanticConsistency, m.err
}

func (m *mockLedgerVerifier) VerifyParentChildRelationship(parentAddr, childAddr string) (bool, error) {
	return m.parentChildValid, m.err
}

// =============================================================================
// VerificationResult Tests
// =============================================================================

func TestVerificationResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   *VerificationResult
		expected bool
	}{
		{
			name:     "verified with no error is success",
			result:   &VerificationResult{Verified: true, Error: nil},
			expected: true,
		},
		{
			name:     "not verified is not success",
			result:   &VerificationResult{Verified: false, Error: nil},
			expected: false,
		},
		{
			name:     "verified with error is not success",
			result:   &VerificationResult{Verified: true, Error: errors.New("err")},
			expected: false,
		},
		{
			name:     "not verified with error is not success",
			result:   &VerificationResult{Verified: false, Error: errors.New("err")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsSuccess(); got != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// VerifyAccountAttachment Tests
// =============================================================================

func TestVerifyAccountAttachment(t *testing.T) {
	t.Run("zero account returns error", func(t *testing.T) {
		var account NeuronAccount
		verifier := &mockLedgerVerifier{}

		_, err := VerifyAccountAttachment(account, verifier)
		if err == nil {
			t.Error("expected error for zero account")
		}
	})

	t.Run("account without attachment returns error", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).Build()
		verifier := &mockLedgerVerifier{}

		_, err := VerifyAccountAttachment(account, verifier)
		if err == nil {
			t.Error("expected error for account without attachment")
		}
	})

	t.Run("ledger mismatch returns verification failure", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{ledgerID: "hedera-mainnet"}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected verification to fail for ledger mismatch")
		}
		if result.Message != "verifier ledger does not match attachment ledger" {
			t.Errorf("unexpected message: %s", result.Message)
		}
	})

	t.Run("account not exists returns verification failure", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:      "ethereum-mainnet",
			accountExists: false,
		}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected verification to fail when account doesn't exist")
		}
	})

	t.Run("account exists check error returns failure", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:      "ethereum-mainnet",
			accountExists: true,
			err:           errors.New("network error"),
		}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected verification to fail on network error")
		}
		if result.Error == nil {
			t.Error("expected error to be set in result")
		}
	})

	t.Run("key ownership mismatch returns failure", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:          "ethereum-mainnet",
			accountExists:     true,
			keyOwnershipValid: false,
		}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected verification to fail on key ownership mismatch")
		}
	})

	t.Run("semantic consistency mismatch returns failure", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:            "ethereum-mainnet",
			accountExists:       true,
			keyOwnershipValid:   true,
			semanticConsistency: false,
		}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected verification to fail on semantic consistency mismatch")
		}
	})

	t.Run("successful verification", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		attachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithLedgerAttachment(attachment.LedgerIdentifier(), attachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:            "ethereum-mainnet",
			accountExists:       true,
			keyOwnershipValid:   true,
			semanticConsistency: true,
		}

		result, err := VerifyAccountAttachment(account, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Verified {
			t.Errorf("expected verification to succeed, got: %s", result.Message)
		}
	})
}

// =============================================================================
// VerifyParentChildLink Tests
// =============================================================================

func TestVerifyParentChildLink(t *testing.T) {
	t.Run("zero parent returns error", func(t *testing.T) {
		var parent NeuronAccount
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		child, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		verifier := &mockLedgerVerifier{}

		_, err := VerifyParentChildLink(parent, child, verifier)
		if err == nil {
			t.Error("expected error for zero parent")
		}
	})

	t.Run("zero child returns error", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		did := newConcurrentMockDID(parentPriv.PublicKey())

		parent, _ := NewParentAccountBuilder(parentPriv.PublicKey(), did).Build()
		var child NeuronAccount

		verifier := &mockLedgerVerifier{}

		_, err := VerifyParentChildLink(parent, child, verifier)
		if err == nil {
			t.Error("expected error for zero child")
		}
	})

	t.Run("first account not parent returns failure", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		// Create two child accounts
		child1, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()
		child2, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.444").
			WithStdOutHedera("0.0.555").
			WithStdErrHedera("0.0.666").
			Build()

		verifier := &mockLedgerVerifier{}

		result, err := VerifyParentChildLink(child1, child2, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected failure when first account is not parent")
		}
	})

	t.Run("second account not child returns failure", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		parent2Priv, _ := keylib.GeneratePrivateKey()

		did1 := newConcurrentMockDID(parentPriv.PublicKey())
		did2 := newConcurrentMockDID(parent2Priv.PublicKey())

		parent1, _ := NewParentAccountBuilder(parentPriv.PublicKey(), did1).Build()
		parent2, _ := NewParentAccountBuilder(parent2Priv.PublicKey(), did2).Build()

		verifier := &mockLedgerVerifier{}

		result, err := VerifyParentChildLink(parent1, parent2, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected failure when second account is not child")
		}
	})

	t.Run("parent reference mismatch returns failure", func(t *testing.T) {
		parent1Priv, _ := keylib.GeneratePrivateKey()
		parent2Priv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		did1 := newConcurrentMockDID(parent1Priv.PublicKey())

		// Parent 1 is the account we're checking against
		parent1, _ := NewParentAccountBuilder(parent1Priv.PublicKey(), did1).Build()

		// Child references parent 2 (not parent 1)
		child, _ := NewChildAccountBuilder(childPriv.PublicKey(), parent2Priv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		verifier := &mockLedgerVerifier{}

		result, err := VerifyParentChildLink(parent1, child, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected failure when child's parent reference doesn't match")
		}
	})

	t.Run("successful verification without attachments", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		did := newConcurrentMockDID(parentPriv.PublicKey())

		parent, _ := NewParentAccountBuilder(parentPriv.PublicKey(), did).Build()
		child, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		verifier := &mockLedgerVerifier{}

		result, err := VerifyParentChildLink(parent, child, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Verified {
			t.Errorf("expected verification to succeed: %s", result.Message)
		}
	})

	t.Run("different ledgers returns failure", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		did := newConcurrentMockDID(parentPriv.PublicKey())

		parentAttachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		childAttachment, _ := NewLedgerAttachment("hedera-mainnet", "0.0.123")

		parent, _ := NewParentAccountBuilder(parentPriv.PublicKey(), did).
			WithLedgerAttachment(parentAttachment.LedgerIdentifier(), parentAttachment.AttachedAddress()).
			Build()
		child, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			WithLedgerAttachment(childAttachment.LedgerIdentifier(), childAttachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{}

		result, err := VerifyParentChildLink(parent, child, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Verified {
			t.Error("expected failure when attached to different ledgers")
		}
	})

	t.Run("successful verification with matching attachments", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		did := newConcurrentMockDID(parentPriv.PublicKey())

		parentAttachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x1234")
		childAttachment, _ := NewLedgerAttachment("ethereum-mainnet", "0x5678")

		parent, _ := NewParentAccountBuilder(parentPriv.PublicKey(), did).
			WithLedgerAttachment(parentAttachment.LedgerIdentifier(), parentAttachment.AttachedAddress()).
			Build()
		child, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			WithLedgerAttachment(childAttachment.LedgerIdentifier(), childAttachment.AttachedAddress()).
			Build()

		verifier := &mockLedgerVerifier{
			ledgerID:         "ethereum-mainnet",
			parentChildValid: true,
		}

		result, err := VerifyParentChildLink(parent, child, verifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Verified {
			t.Errorf("expected verification to succeed: %s", result.Message)
		}
	})
}
