package account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// init registers the mock DID parser for E2E tests.
// This allows JSON round-trip to work by parsing "did:key:z..." strings.
func init() {
	RegisterDIDParser(DIDMethodKey, parseE2EMockDID)
}

// =============================================================================
// E2E Lifecycle Test - Complete Neuron SDK Lifecycle
// =============================================================================

// TestE2E_FullLifecycle exercises the complete Neuron SDK lifecycle:
// Key Creation -> Account Creation -> Ledger Attachment -> Verification
//
// This test serves as executable documentation of the system's behavior
// and validates all critical invariants across the lifecycle.
func TestE2E_FullLifecycle(t *testing.T) {
	report := newTestReport("E2E Full Lifecycle Test")
	defer report.print(t)

	// Phase 1: Key Creation
	t.Run("Phase1_KeyCreation", func(t *testing.T) {
		testKeyCreationPhase(t, report)
	})

	// Phase 2: Account Creation
	t.Run("Phase2_AccountCreation", func(t *testing.T) {
		testAccountCreationPhase(t, report)
	})

	// Phase 3: Ledger Attachment
	t.Run("Phase3_LedgerAttachment", func(t *testing.T) {
		testLedgerAttachmentPhase(t, report)
	})

	// Phase 4: Verification
	t.Run("Phase4_Verification", func(t *testing.T) {
		testVerificationPhase(t, report)
	})

	// Phase 5: Invariant Assertions
	t.Run("Phase5_Invariants", func(t *testing.T) {
		testInvariantAssertions(t, report)
	})
}

// =============================================================================
// E2E Mock DID (to avoid import cycle with didkey package)
// =============================================================================

// e2eMockDID implements NeuronDID and NeuronDIDWithKey for E2E testing.
// This allows testing without importing the didkey package, which would
// cause an import cycle (didkey imports account).
type e2eMockDID struct {
	did      string
	method   string
	id       string
	valid    bool
	matchKey keylib.NeuronPublicKey
}

func (m *e2eMockDID) String() string     { return m.did }
func (m *e2eMockDID) Method() string     { return m.method }
func (m *e2eMockDID) Identifier() string { return m.id }
func (m *e2eMockDID) Validate() error {
	if !m.valid {
		return errors.New("invalid mock DID")
	}
	return nil
}
func (m *e2eMockDID) Equal(other NeuronDID) bool {
	if other == nil {
		return false
	}
	return m.did == other.String()
}
func (m *e2eMockDID) PublicKey() (keylib.NeuronPublicKey, error) {
	if m.matchKey.IsZero() {
		return keylib.NeuronPublicKey{}, errors.New("no public key")
	}
	return m.matchKey, nil
}
func (m *e2eMockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	// If we have the original key stored, compare directly
	if !m.matchKey.IsZero() {
		return m.matchKey.Equal(pubKey)
	}
	// Otherwise, regenerate what the DID would be from the key and compare strings
	// This allows MatchesKey to work after JSON round-trip
	expectedDID := newE2EMockDID(pubKey)
	return m.did == expectedDID.did
}

// newE2EMockDID creates a mock DID that simulates did:key behavior.
// It generates a deterministic DID string based on the public key.
func newE2EMockDID(pubKey keylib.NeuronPublicKey) *e2eMockDID {
	// Create a deterministic identifier from the public key
	// This simulates the did:key format without the actual encoding
	hexSuffix := pubKey.Hex()[2:18] // Take first 16 chars after 0x
	identifier := "z" + hexSuffix
	return &e2eMockDID{
		did:      "did:key:" + identifier,
		method:   DIDMethodKey,
		id:       identifier,
		valid:    true,
		matchKey: pubKey,
	}
}

// parseE2EMockDID parses a DID string back into e2eMockDID.
// This is used during JSON unmarshaling to reconstruct the DID.
func parseE2EMockDID(didString string) (NeuronDID, error) {
	// Format: did:key:<identifier>
	const prefix = "did:key:"
	if !strings.HasPrefix(didString, prefix) {
		return nil, errors.New("invalid did:key format")
	}
	identifier := didString[len(prefix):]
	if len(identifier) == 0 {
		return nil, errors.New("empty identifier")
	}
	return &e2eMockDID{
		did:    didString,
		method: DIDMethodKey,
		id:     identifier,
		valid:  true,
		// matchKey is not recoverable from string, but Equal() uses String() comparison
	}, nil
}

// =============================================================================
// Phase 1: Key Creation Tests
// =============================================================================

func testKeyCreationPhase(t *testing.T, report *testReport) {
	phase := report.startPhase("Key Creation")

	// Generate private key
	t.Run("GeneratePrivateKey", func(t *testing.T) {
		privKey, err := keylib.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GeneratePrivateKey() error = %v", err)
		}
		if privKey.IsZero() {
			t.Fatal("generated key is zero-value")
		}
		phase.addResult("PrivateKey generated", truncateHex(privKey.Hex(), 20)+"...")

		// Store in report for later phases
		report.storeValue("privKey", privKey)
	})

	// Derive public key
	t.Run("DerivePublicKey", func(t *testing.T) {
		privKey := report.getPrivateKey("privKey")
		pubKey := privKey.PublicKey()
		if pubKey.IsZero() {
			t.Fatal("derived public key is zero-value")
		}
		phase.addResult("PublicKey derived", pubKey.Hex())
		report.storeValue("pubKey", pubKey)
	})

	// Derive PeerID
	t.Run("DerivePeerID", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		peerID, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("PeerID() error = %v", err)
		}
		if peerID.IsZero() {
			t.Fatal("derived PeerID is zero-value")
		}
		phase.addResult("PeerID derived", peerID.String())
		report.storeValue("peerID", peerID)
	})

	// Derive EVMAddress
	t.Run("DeriveEVMAddress", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		evmAddr := pubKey.EVMAddress()
		if evmAddr.IsZero() {
			t.Fatal("derived EVMAddress is zero")
		}
		phase.addResult("EVMAddress derived", evmAddr.Hex())
		report.storeValue("evmAddr", evmAddr)
	})

	// Verify determinism
	t.Run("VerifyDeterminism", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		storedPeerID := report.getPeerID("peerID")
		storedEVMAddr := report.getEVMAddress("evmAddr")

		// Derive PeerID again (10 times)
		for i := 0; i < 10; i++ {
			peerID, _ := pubKey.PeerID()
			if !peerID.Equal(storedPeerID) {
				t.Errorf("PeerID derivation not deterministic on iteration %d", i)
			}
		}

		// Derive EVMAddress again (10 times)
		for i := 0; i < 10; i++ {
			evmAddr := pubKey.EVMAddress()
			if !evmAddr.Equal(storedEVMAddr) {
				t.Errorf("EVMAddress derivation not deterministic on iteration %d", i)
			}
		}

		phase.addValidation("Determinism verified (10 iterations)", true)
	})

	// Sign and verify message
	t.Run("SignAndVerify", func(t *testing.T) {
		privKey := report.getPrivateKey("privKey")
		pubKey := report.getPublicKey("pubKey")

		message := []byte("Neuron E2E Test Message - Lifecycle Verification")

		sig, err := privKey.SignMessage(message)
		if err != nil {
			t.Fatalf("SignMessage() error = %v", err)
		}

		if !pubKey.Verify(message, sig) {
			t.Error("signature verification failed for correct message")
		}

		// Verify wrong message fails
		if pubKey.Verify([]byte("wrong message"), sig) {
			t.Error("verification should fail for wrong message")
		}

		phase.addValidation("Sign/Verify works correctly", true)
		phase.addResult("Signature size", fmt.Sprintf("%d bytes", len(sig.Bytes())))
		report.storeValue("testMessage", message)
		report.storeValue("testSignature", sig)
	})

	phase.complete()
}

// =============================================================================
// Phase 2: Account Creation Tests
// =============================================================================

func testAccountCreationPhase(t *testing.T, report *testReport) {
	phase := report.startPhase("Account Creation")

	// Create DID (using mock that simulates did:key)
	t.Run("CreateDID", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")

		// Create mock DID (simulates did:key without import cycle)
		did := newE2EMockDID(pubKey)
		if did == nil {
			t.Fatal("DID is nil")
		}

		// Verify DID format
		if !strings.HasPrefix(did.String(), "did:key:z") {
			t.Errorf("DID format invalid, got: %s", did.String())
		}
		if did.Method() != DIDMethodKey {
			t.Errorf("DID method = %s, want 'key'", did.Method())
		}
		if err := did.Validate(); err != nil {
			t.Errorf("DID Validate() error = %v", err)
		}

		phase.addResult("DID created", truncateString(did.String(), 40)+"...")
		report.storeValue("did", did)
	})

	// Verify DID matches key
	t.Run("VerifyDIDMatchesKey", func(t *testing.T) {
		did := report.getDID("did")
		pubKey := report.getPublicKey("pubKey")

		if !did.MatchesKey(pubKey) {
			t.Error("DID does not match public key")
		}

		// Verify extracted key matches
		extractedKey, err := did.PublicKey()
		if err != nil {
			t.Fatalf("did.PublicKey() error = %v", err)
		}
		if !extractedKey.Equal(pubKey) {
			t.Error("extracted key does not match original")
		}

		// Verify DID equality works
		sameDID := newE2EMockDID(pubKey)
		if !did.Equal(sameDID) {
			t.Error("DID should equal another DID with same string")
		}

		// Verify different key produces different DID
		otherPriv, _ := keylib.GeneratePrivateKey()
		otherDID := newE2EMockDID(otherPriv.PublicKey())
		if did.Equal(otherDID) {
			t.Error("DID should not equal DID from different key")
		}

		phase.addValidation("DID-Key match verified", true)
	})

	// Build Parent account
	t.Run("BuildParentAccount", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")

		// Create reachable address with correct PeerID
		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		acct, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		if acct.IsZero() {
			t.Fatal("built account is zero-value")
		}

		phase.addResult("Account type", acct.AccountType().String())
		phase.addResult("Has reachable addrs", fmt.Sprintf("%d", len(acct.ReachableAddrs().Addrs())))
		report.storeValue("parentAccount", acct)
	})

	// Validate account structure
	t.Run("ValidateAccount", func(t *testing.T) {
		acct := report.getAccount("parentAccount")

		// Single validation
		if err := acct.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}

		// Comprehensive validation
		result := acct.ValidateAll()
		if result.HasErrors() {
			for _, e := range result.Errors() {
				t.Errorf("ValidateAll() error: %v", e)
			}
		}

		phase.addValidation("Validate() passed", true)
		phase.addValidation("ValidateAll() passed", result.Valid())
	})

	// Verify accessors
	t.Run("VerifyAccessors", func(t *testing.T) {
		acct := report.getAccount("parentAccount")
		pubKey := report.getPublicKey("pubKey")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")
		did := report.getDID("did")

		// Verify all accessors return expected values
		if !acct.PublicKey().Equal(pubKey) {
			t.Error("PublicKey() mismatch")
		}
		if !acct.PeerID().Equal(peerID) {
			t.Error("PeerID() mismatch")
		}
		if !acct.EVMAddress().Equal(evmAddr) {
			t.Error("EVMAddress() mismatch")
		}
		if !acct.DID().Equal(did) {
			t.Error("DID() mismatch")
		}
		if !acct.IsParent() {
			t.Error("IsParent() should be true")
		}
		if acct.IsChild() {
			t.Error("IsChild() should be false")
		}
		if acct.IsShared() {
			t.Error("IsShared() should be false")
		}
		if !acct.ParentPublicKey().IsZero() {
			t.Error("ParentPublicKey() should be zero for parent")
		}
		// Parent accounts must NOT have comm channels
		if !acct.StdIn().IsZero() {
			t.Error("Parent StdIn should be zero")
		}
		if !acct.StdOut().IsZero() {
			t.Error("Parent StdOut should be zero")
		}
		if !acct.StdErr().IsZero() {
			t.Error("Parent StdErr should be zero")
		}

		phase.addValidation("All accessors verified", true)
	})

	phase.complete()
}

// =============================================================================
// Phase 3: Ledger Attachment Tests
// =============================================================================

func testLedgerAttachmentPhase(t *testing.T, report *testReport) {
	phase := report.startPhase("Ledger Attachment")

	// Create LedgerAttachment
	t.Run("CreateLedgerAttachment", func(t *testing.T) {
		evmAddr := report.getEVMAddress("evmAddr")

		attachment, err := NewLedgerAttachment("hedera-mainnet", evmAddr.Hex())
		if err != nil {
			t.Fatalf("NewLedgerAttachment() error = %v", err)
		}

		// Verify initial state
		if attachment.State() != AttachmentStateDetached {
			t.Errorf("initial state = %v, want Detached", attachment.State())
		}
		if attachment.VerificationStatus() != VerificationStatusNone {
			t.Errorf("initial verification status = %v, want None", attachment.VerificationStatus())
		}
		if attachment.LedgerIdentifier() != "hedera-mainnet" {
			t.Error("LedgerIdentifier mismatch")
		}
		if attachment.AttachedAddress() != evmAddr.Hex() {
			t.Error("AttachedAddress mismatch")
		}
		if attachment.IsAttached() {
			t.Error("IsAttached() should be false initially")
		}
		if attachment.IsVerified() {
			t.Error("IsVerified() should be false initially")
		}

		phase.addResult("Attachment created", attachment.LedgerIdentifier())
		phase.addResult("Initial state", attachment.State().String())
		report.storeValue("attachment", attachment)
	})

	// Test invalid state transitions
	t.Run("InvalidStateTransitions", func(t *testing.T) {
		// Create fresh attachment for this test
		attachment, _ := NewLedgerAttachment("hedera-testnet", "0x1234567890abcdef")

		// Cannot transition from Detached directly to Verified
		if err := attachment.SetVerified(); err == nil {
			t.Error("SetVerified() from Detached should fail")
		} else {
			phase.addValidation("Detached->Verified blocked", true)
		}

		// Cannot set verification failed from Detached
		if err := attachment.SetVerificationFailed(); err == nil {
			t.Error("SetVerificationFailed() from Detached should fail")
		} else {
			phase.addValidation("SetVerificationFailed from Detached blocked", true)
		}

		// After attaching, cannot attach again
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if err := attachment.SetAttached(); err == nil {
			t.Error("SetAttached() from Attached should fail")
		} else {
			phase.addValidation("Attached->Attached blocked", true)
		}

		// After verifying, cannot verify again
		attachment2, _ := NewLedgerAttachment("hedera-testnet", "0xabcdef")
		_ = attachment2.SetAttached()
		_ = attachment2.SetVerified()
		if err := attachment2.SetVerified(); err == nil {
			t.Error("SetVerified() from Verified should fail")
		} else {
			phase.addValidation("Verified->Verified blocked", true)
		}
		if err := attachment2.SetAttached(); err == nil {
			t.Error("SetAttached() from Verified should fail")
		} else {
			phase.addValidation("Verified->Attached blocked", true)
		}
	})

	// Transition to Attached state
	t.Run("TransitionToAttached", func(t *testing.T) {
		attachment := report.getLedgerAttachment("attachment")

		// Valid transition: Detached -> Attached
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}

		if attachment.State() != AttachmentStateAttached {
			t.Errorf("state = %v, want Attached", attachment.State())
		}
		if !attachment.IsAttached() {
			t.Error("IsAttached() should be true")
		}
		if attachment.IsVerified() {
			t.Error("IsVerified() should be false")
		}

		phase.addResult("State after SetAttached", attachment.State().String())
	})

	// Build account with ledger attachment
	t.Run("BuildAccountWithAttachment", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		acct, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("hedera-mainnet", evmAddr.Hex()).
			WithCurrencySymbol("HBAR").
			Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		attachment := acct.LedgerAttachment()
		if attachment == nil {
			t.Fatal("LedgerAttachment() returned nil")
		}
		if attachment.LedgerIdentifier() != "hedera-mainnet" {
			t.Error("LedgerIdentifier mismatch")
		}
		if acct.CurrencySymbol() != "HBAR" {
			t.Errorf("CurrencySymbol = %s, want HBAR", acct.CurrencySymbol())
		}

		// Validate account with attachment
		if err := acct.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}

		phase.addResult("Account with attachment built", "success")
		phase.addResult("Currency symbol", acct.CurrencySymbol())
		report.storeValue("accountWithAttachment", acct)
	})

	// Test currency requirement when ledger attached
	t.Run("CurrencyRequiredWhenAttached", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		// Build account with attachment but NO currency - should fail validation
		acct, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("hedera-mainnet", evmAddr.Hex()).
			// Intentionally omit WithCurrencySymbol
			Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		// Validation should fail due to missing currency
		if err := acct.Validate(); err == nil {
			t.Error("Validate() should fail when ledger attached but no currency")
		} else {
			phase.addValidation("Currency required when attached", true)
		}
	})

	phase.complete()
}

// =============================================================================
// Phase 4: Verification Tests
// =============================================================================

func testVerificationPhase(t *testing.T, report *testReport) {
	phase := report.startPhase("Verification")

	// Successful verification path
	t.Run("SuccessfulVerification", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		// Build account with attachment
		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()
		acct, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("hedera-mainnet", evmAddr.Hex()).
			WithCurrencySymbol("HBAR").
			Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		// Create mock verifier with success responses
		verifier := newE2EMockVerifier("hedera-mainnet")

		// Verify attachment
		result, err := VerifyAccountAttachment(acct, verifier)
		if err != nil {
			t.Fatalf("VerifyAccountAttachment() error = %v", err)
		}
		if !result.Verified {
			t.Errorf("verification failed: %s", result.Message)
		}
		if !result.IsSuccess() {
			t.Error("IsSuccess() should be true")
		}

		// Verify all expected methods were called
		if !verifier.wasMethodCalled("VerifyAccountExists") {
			t.Error("VerifyAccountExists was not called")
		}
		if !verifier.wasMethodCalled("VerifyKeyOwnership") {
			t.Error("VerifyKeyOwnership was not called")
		}
		if !verifier.wasMethodCalled("VerifySemanticConsistency") {
			t.Error("VerifySemanticConsistency was not called")
		}

		phase.addResult("Verification result", "Verified")
		phase.addValidation("All verifier methods called", true)

		// Complete attachment state transition
		attachment := acct.LedgerAttachment()
		_ = attachment.SetAttached()
		if err := attachment.SetVerified(); err != nil {
			t.Fatalf("SetVerified() error = %v", err)
		}
		phase.addResult("Final attachment state", attachment.State().String())
	})

	// Verification failure path: Account not found
	t.Run("VerificationFailure_AccountNotFound", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()
		acct, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("hedera-mainnet", evmAddr.Hex()).
			WithCurrencySymbol("HBAR").
			Build()

		// Create mock verifier that simulates account not found
		verifier := newE2EMockVerifier("hedera-mainnet")
		verifier.accountExistsResult = false

		result, err := VerifyAccountAttachment(acct, verifier)
		if err != nil {
			t.Fatalf("VerifyAccountAttachment() error = %v", err)
		}
		if result.Verified {
			t.Error("verification should have failed")
		}
		if result.Message != "account does not exist on ledger" {
			t.Errorf("unexpected message: %s", result.Message)
		}

		phase.addValidation("Failure: account not found", !result.Verified)
	})

	// Verification failure path: Key ownership mismatch
	t.Run("VerificationFailure_KeyMismatch", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()
		acct, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("hedera-mainnet", evmAddr.Hex()).
			WithCurrencySymbol("HBAR").
			Build()

		// Create mock verifier that simulates key mismatch
		verifier := newE2EMockVerifier("hedera-mainnet")
		verifier.keyOwnershipResult = false

		result, err := VerifyAccountAttachment(acct, verifier)
		if err != nil {
			t.Fatalf("VerifyAccountAttachment() error = %v", err)
		}
		if result.Verified {
			t.Error("verification should have failed")
		}
		if result.Message != "public key does not derive to attached address" {
			t.Errorf("unexpected message: %s", result.Message)
		}

		phase.addValidation("Failure: key mismatch", !result.Verified)
	})

	// Verification failure with SetVerificationFailed
	t.Run("SetVerificationFailed", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("hedera-mainnet", "0x1234")
		_ = attachment.SetAttached()

		if err := attachment.SetVerificationFailed(); err != nil {
			t.Fatalf("SetVerificationFailed() error = %v", err)
		}

		if attachment.VerificationStatus() != VerificationStatusFailed {
			t.Errorf("VerificationStatus = %v, want Failed", attachment.VerificationStatus())
		}
		// State should remain Attached
		if attachment.State() != AttachmentStateAttached {
			t.Errorf("State = %v, want Attached", attachment.State())
		}

		phase.addValidation("SetVerificationFailed works", true)
	})

	// Ledger mismatch
	t.Run("LedgerMismatch", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")
		evmAddr := report.getEVMAddress("evmAddr")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()
		acct, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			WithLedgerAttachment("ethereum-mainnet", evmAddr.Hex()). // Different ledger
			WithCurrencySymbol("ETH").
			Build()

		// Verifier is for hedera, but account is attached to ethereum
		verifier := newE2EMockVerifier("hedera-mainnet")

		result, err := VerifyAccountAttachment(acct, verifier)
		if err != nil {
			t.Fatalf("VerifyAccountAttachment() error = %v", err)
		}
		if result.Verified {
			t.Error("verification should fail for ledger mismatch")
		}
		if result.Message != "verifier ledger does not match attachment ledger" {
			t.Errorf("unexpected message: %s", result.Message)
		}

		phase.addValidation("Ledger mismatch detected", !result.Verified)
	})

	phase.complete()
}

// =============================================================================
// Phase 5: Invariant Assertions
// =============================================================================

func testInvariantAssertions(t *testing.T, report *testReport) {
	phase := report.startPhase("Invariant Assertions")

	// Key immutability: same key produces same derived values
	t.Run("KeyImmutability", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		storedPeerID := report.getPeerID("peerID")
		storedEVMAddr := report.getEVMAddress("evmAddr")

		// Repeat derivations many times
		for i := 0; i < 100; i++ {
			derivedPeerID, _ := pubKey.PeerID()
			derivedEVMAddr := pubKey.EVMAddress()

			if !derivedPeerID.Equal(storedPeerID) {
				t.Fatalf("PeerID changed on iteration %d", i)
			}
			if !derivedEVMAddr.Equal(storedEVMAddr) {
				t.Fatalf("EVMAddress changed on iteration %d", i)
			}
		}

		phase.addValidation("Key derivation immutable (100 iterations)", true)
	})

	// DID determinism (using mock)
	t.Run("DIDDeterminism", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		originalDID := report.getDID("did")

		// Create DID again from same key (10 times)
		for i := 0; i < 10; i++ {
			did := newE2EMockDID(pubKey)

			if !did.Equal(originalDID) {
				t.Errorf("DID not deterministic on iteration %d", i)
			}
			if did.String() != originalDID.String() {
				t.Errorf("DID string not identical on iteration %d", i)
			}
		}

		phase.addValidation("DID generation deterministic", true)
	})

	// Account validity after rebuild
	t.Run("AccountRebuild", func(t *testing.T) {
		pubKey := report.getPublicKey("pubKey")
		did := report.getDID("did")
		peerID := report.getPeerID("peerID")

		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		var firstAccount NeuronAccount

		// Build account multiple times
		for i := 0; i < 10; i++ {
			acct, err := NewParentAccountBuilder(pubKey, did).
				WithReachableAddr(reachableAddr).
				Build()
			if err != nil {
				t.Fatalf("Build() iteration %d error = %v", i, err)
			}
			if err := acct.Validate(); err != nil {
				t.Fatalf("Validate() iteration %d error = %v", i, err)
			}
			if i == 0 {
				firstAccount = acct
			} else {
				// Verify equality
				if !acct.Equal(firstAccount) {
					t.Errorf("Rebuilt account not equal on iteration %d", i)
				}
			}
		}

		phase.addValidation("Account rebuild consistent (10 iterations)", true)
	})

	// State machine correctness
	t.Run("StateMachineCorrectness", func(t *testing.T) {
		// Cannot skip states
		attachment, _ := NewLedgerAttachment("hedera-mainnet", "0x1234")

		// Try to skip to Verified (should fail)
		if err := attachment.SetVerified(); err == nil {
			t.Error("should not be able to skip to Verified from Detached")
		}

		// Proper sequence
		attachment2, _ := NewLedgerAttachment("hedera-mainnet", "0x1234")
		if err := attachment2.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if err := attachment2.SetVerified(); err != nil {
			t.Fatalf("SetVerified() error = %v", err)
		}

		// Cannot go back
		if err := attachment2.SetAttached(); err == nil {
			t.Error("should not be able to go back to Attached from Verified")
		}

		phase.addValidation("State machine enforced", true)
	})

	// JSON round-trip preserves data
	t.Run("JSONRoundTrip", func(t *testing.T) {
		acct := report.getAccount("parentAccount")

		// Marshal to JSON
		jsonData, err := json.Marshal(acct)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		// Unmarshal back
		var restored NeuronAccount
		if err := json.Unmarshal(jsonData, &restored); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		// Verify equality
		if !acct.Equal(restored) {
			t.Error("JSON round-trip should preserve account equality")
		}

		// Verify validation still passes
		if err := restored.Validate(); err != nil {
			t.Errorf("restored account Validate() error = %v", err)
		}

		// Verify key fields preserved
		if !acct.PublicKey().Equal(restored.PublicKey()) {
			t.Error("PublicKey not preserved")
		}
		if acct.AccountType() != restored.AccountType() {
			t.Error("AccountType not preserved")
		}

		phase.addValidation("JSON round-trip preserves data", true)
	})

	// Signature with recovered key
	t.Run("SignatureKeyRecovery", func(t *testing.T) {
		privKey := report.getPrivateKey("privKey")
		pubKey := report.getPublicKey("pubKey")

		message := []byte("key recovery test message - E2E lifecycle")
		sig, err := privKey.SignMessage(message)
		if err != nil {
			t.Fatalf("SignMessage() error = %v", err)
		}

		recovered, err := keylib.RecoverPublicKey(message, sig)
		if err != nil {
			t.Fatalf("RecoverPublicKey() error = %v", err)
		}

		if !recovered.Equal(pubKey) {
			t.Error("recovered key does not match original")
		}

		phase.addValidation("Key recovery from signature works", true)
	})

	// Account Equal() reflexivity and symmetry
	t.Run("EqualityProperties", func(t *testing.T) {
		acct := report.getAccount("parentAccount")

		// Reflexivity: a == a
		if !acct.Equal(acct) {
			t.Error("Equal() should be reflexive")
		}

		// Zero account handling
		var zero NeuronAccount
		if zero.Equal(acct) {
			t.Error("zero should not equal valid account")
		}
		if acct.Equal(zero) {
			t.Error("valid account should not equal zero")
		}
		if zero.Equal(zero) {
			t.Error("zero should not equal zero (per implementation)")
		}

		phase.addValidation("Equality properties verified", true)
	})

	phase.complete()
}

// =============================================================================
// Mock Verifier for E2E Tests
// =============================================================================

// e2eMockVerifier implements LedgerVerifier for E2E testing.
type e2eMockVerifier struct {
	mu sync.Mutex

	// Configuration
	ledgerID string

	// Simulated verification results (defaults to true)
	accountExistsResult           bool
	keyOwnershipResult            bool
	multisigOwnershipResult       bool
	semanticConsistencyResult     bool
	parentChildRelationshipResult bool

	// Call tracking
	callLog []string
}

func newE2EMockVerifier(ledgerID string) *e2eMockVerifier {
	return &e2eMockVerifier{
		ledgerID:                      ledgerID,
		accountExistsResult:           true,
		keyOwnershipResult:            true,
		semanticConsistencyResult:     true,
		multisigOwnershipResult:       true,
		parentChildRelationshipResult: true,
		callLog:                       make([]string, 0),
	}
}

func (m *e2eMockVerifier) LedgerIdentifier() string {
	return m.ledgerID
}

func (m *e2eMockVerifier) VerifyAccountExists(address string) (bool, error) {
	m.mu.Lock()
	m.callLog = append(m.callLog, "VerifyAccountExists:"+address)
	m.mu.Unlock()
	return m.accountExistsResult, nil
}

func (m *e2eMockVerifier) VerifyKeyOwnership(address string, pubKey keylib.NeuronPublicKey) (bool, error) {
	m.mu.Lock()
	m.callLog = append(m.callLog, "VerifyKeyOwnership:"+address)
	m.mu.Unlock()
	return m.keyOwnershipResult, nil
}

func (m *e2eMockVerifier) VerifyMultisigOwnership(address string, multisigKey keylib.MultisigKey) (bool, error) {
	m.mu.Lock()
	m.callLog = append(m.callLog, "VerifyMultisigOwnership:"+address)
	m.mu.Unlock()
	return m.multisigOwnershipResult, nil
}

func (m *e2eMockVerifier) VerifySemanticConsistency(address string, pubKey keylib.NeuronPublicKey) (bool, error) {
	m.mu.Lock()
	m.callLog = append(m.callLog, "VerifySemanticConsistency:"+address)
	m.mu.Unlock()
	return m.semanticConsistencyResult, nil
}

func (m *e2eMockVerifier) VerifyParentChildRelationship(parentAddr, childAddr string) (bool, error) {
	m.mu.Lock()
	m.callLog = append(m.callLog, "VerifyParentChildRelationship:"+parentAddr+":"+childAddr)
	m.mu.Unlock()
	return m.parentChildRelationshipResult, nil
}

func (m *e2eMockVerifier) wasMethodCalled(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.callLog {
		if strings.HasPrefix(call, method) {
			return true
		}
	}
	return false
}

// =============================================================================
// Test Report Structure
// =============================================================================

type testReport struct {
	name      string
	startTime time.Time
	phases    []*testPhase
	keyStore  map[string]interface{}
	mu        sync.Mutex
}

type testPhase struct {
	name        string
	results     []phaseResult
	validations []phaseValidation
	completed   bool
}

type phaseResult struct {
	name  string
	value string
}

type phaseValidation struct {
	name   string
	passed bool
}

func newTestReport(name string) *testReport {
	return &testReport{
		name:      name,
		startTime: time.Now(),
		phases:    make([]*testPhase, 0),
		keyStore:  make(map[string]interface{}),
	}
}

func (r *testReport) startPhase(name string) *testPhase {
	phase := &testPhase{
		name:        name,
		results:     make([]phaseResult, 0),
		validations: make([]phaseValidation, 0),
	}
	r.phases = append(r.phases, phase)
	return phase
}

func (p *testPhase) addResult(name, value string) {
	p.results = append(p.results, phaseResult{name: name, value: value})
}

func (p *testPhase) addValidation(name string, passed bool) {
	p.validations = append(p.validations, phaseValidation{name: name, passed: passed})
}

func (p *testPhase) complete() {
	p.completed = true
}

func (r *testReport) print(t *testing.T) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "\n")
	fmt.Fprintf(&buf, "============================================================\n")
	fmt.Fprintf(&buf, "  %s\n", r.name)
	fmt.Fprintf(&buf, "  Duration: %v\n", time.Since(r.startTime))
	fmt.Fprintf(&buf, "============================================================\n")

	for _, phase := range r.phases {
		status := "INCOMPLETE"
		if phase.completed {
			status = "COMPLETE"
		}
		fmt.Fprintf(&buf, "\n[%s] %s\n", status, phase.name)
		fmt.Fprintf(&buf, "------------------------------------------------------------\n")

		if len(phase.results) > 0 {
			fmt.Fprintf(&buf, "Results:\n")
			for _, result := range phase.results {
				fmt.Fprintf(&buf, "  - %-30s : %s\n", result.name, result.value)
			}
		}

		if len(phase.validations) > 0 {
			fmt.Fprintf(&buf, "Validations:\n")
			for _, v := range phase.validations {
				mark := "PASS"
				if !v.passed {
					mark = "FAIL"
				}
				fmt.Fprintf(&buf, "  - [%s] %s\n", mark, v.name)
			}
		}
	}

	fmt.Fprintf(&buf, "\n============================================================\n")
	t.Logf("%s", buf.String())
}

// Key storage methods
func (r *testReport) storeValue(name string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keyStore[name] = value
}

func (r *testReport) getPrivateKey(name string) keylib.NeuronPrivateKey {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(keylib.NeuronPrivateKey)
}

func (r *testReport) getPublicKey(name string) keylib.NeuronPublicKey {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(keylib.NeuronPublicKey)
}

func (r *testReport) getPeerID(name string) keylib.PeerID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(keylib.PeerID)
}

func (r *testReport) getEVMAddress(name string) keylib.EVMAddress {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(keylib.EVMAddress)
}

func (r *testReport) getDID(name string) *e2eMockDID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(*e2eMockDID)
}

func (r *testReport) getAccount(name string) NeuronAccount {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(NeuronAccount)
}

func (r *testReport) getLedgerAttachment(name string) *LedgerAttachment {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.keyStore[name].(*LedgerAttachment)
}

// =============================================================================
// Utility Functions
// =============================================================================

func truncateHex(hex string, maxLen int) string {
	if len(hex) <= maxLen {
		return hex
	}
	return hex[:maxLen]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
