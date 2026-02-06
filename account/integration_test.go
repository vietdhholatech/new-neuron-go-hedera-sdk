package account

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Integration Test Helpers
// =============================================================================

// integrationMockDID is a test double for NeuronDID interface
type integrationMockDID struct {
	did      string
	method   string
	id       string
	valid    bool
	matchKey keylib.NeuronPublicKey
}

func (m *integrationMockDID) String() string     { return m.did }
func (m *integrationMockDID) Method() string     { return m.method }
func (m *integrationMockDID) Identifier() string { return m.id }
func (m *integrationMockDID) Validate() error {
	if !m.valid {
		return errors.New("invalid mock DID")
	}
	return nil
}
func (m *integrationMockDID) Equal(other NeuronDID) bool {
	return m.did == other.String()
}
func (m *integrationMockDID) PublicKey() (keylib.NeuronPublicKey, error) {
	return m.matchKey, nil
}
func (m *integrationMockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	return m.matchKey.Equal(pubKey)
}

func newIntegrationMockDID(pubKey keylib.NeuronPublicKey) *integrationMockDID {
	return &integrationMockDID{
		did:      "did:mock:integration",
		method:   "mock",
		id:       "integration",
		valid:    true,
		matchKey: pubKey,
	}
}

// =============================================================================
// Known Test Vectors
// =============================================================================

const (
	// Known public key (secp256k1 compressed)
	knownPubKeyHex = "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"

	// Expected PeerID derived from the public key
	knownPeerID = "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	// A different PeerID for mismatch testing
	differentPeerID = "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"
)

// =============================================================================
// Parent Account Lifecycle Tests
// =============================================================================

func TestIntegration_ParentAccountLifecycle(t *testing.T) {
	t.Run("complete lifecycle: generate -> build -> validate -> serialize", func(t *testing.T) {
		// 1. Generate a key
		privKey, err := keylib.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GeneratePrivateKey error: %v", err)
		}
		pubKey := privKey.PublicKey()

		// 2. Create DID (mock)
		did := newIntegrationMockDID(pubKey)

		// 3. Derive PeerID for reachable address
		peerID, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("PeerID derivation error: %v", err)
		}
		reachableAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		// 4. Build the account (Parent accounts must NOT have comm channels)
		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(reachableAddr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// 5. Validate
		if err := account.Validate(); err != nil {
			t.Errorf("Validate error: %v", err)
		}

		// 6. Verify all accessors
		if account.IsZero() {
			t.Error("account should not be zero")
		}
		if !account.IsParent() {
			t.Error("should be parent")
		}
		if account.IsChild() {
			t.Error("should not be child")
		}
		if !account.PublicKey().Equal(pubKey) {
			t.Error("PublicKey mismatch")
		}
		if !account.PeerID().Equal(peerID) {
			t.Error("PeerID mismatch")
		}
		if account.EVMAddress().IsZero() {
			t.Error("EVMAddress should not be zero")
		}
		if account.DID() == nil {
			t.Error("DID should not be nil")
		}
		if !account.ParentPublicKey().IsZero() {
			t.Error("ParentPublicKey should be zero for parent")
		}

		// 7. Serialize to JSON
		jsonData, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if len(jsonData) == 0 {
			t.Error("JSON should not be empty")
		}

		// 8. Verify JSON structure
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if jsonMap["publicKey"] == nil {
			t.Error("JSON missing publicKey")
		}
		if jsonMap["peerId"] == nil {
			t.Error("JSON missing peerId")
		}
		if jsonMap["did"] == nil {
			t.Error("JSON missing did for parent")
		}
		if jsonMap["accountType"] != "Parent" {
			t.Errorf("JSON accountType = %v, want 'Parent'", jsonMap["accountType"])
		}
	})
}

// =============================================================================
// Child Account Lifecycle Tests
// =============================================================================

func TestIntegration_ChildAccountLifecycle(t *testing.T) {
	t.Run("complete lifecycle: parent exists -> generate child -> build -> validate", func(t *testing.T) {
		// 1. Generate parent key
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()
		parentDID := newIntegrationMockDID(parentPubKey)

		// Build parent account
		parentAccount, err := NewParentAccountBuilder(parentPubKey, parentDID).Build()
		if err != nil {
			t.Fatalf("Parent Build error: %v", err)
		}

		// 2. Generate child key
		childPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()
		childPeerID, _ := childPubKey.PeerID()
		childReachableAddr := "/ip4/192.168.1.2/tcp/4001/p2p/" + childPeerID.String()

		// 3. Build child account (Child accounts must have all 3 comm channels)
		childAccount, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdInHedera("0.0.444").
			WithStdOutHedera("0.0.555").
			WithStdErrHedera("0.0.666").
			WithReachableAddr(childReachableAddr).
			Build()
		if err != nil {
			t.Fatalf("Child Build error: %v", err)
		}

		// 4. Validate child
		if err := childAccount.Validate(); err != nil {
			t.Errorf("Validate error: %v", err)
		}

		// 5. Verify child-specific properties
		if !childAccount.IsChild() {
			t.Error("should be child")
		}
		if childAccount.IsParent() {
			t.Error("should not be parent")
		}
		if childAccount.DID() != nil {
			t.Error("child should have no DID")
		}
		if !childAccount.ParentPublicKey().Equal(parentPubKey) {
			t.Error("ParentPublicKey mismatch")
		}

		// 6. Verify parent-child relationship
		if !childAccount.ParentPublicKey().Equal(parentAccount.PublicKey()) {
			t.Error("child's parent key should match parent's public key")
		}

		// 7. Child and parent should have different public keys
		if childAccount.PublicKey().Equal(parentAccount.PublicKey()) {
			t.Error("child and parent should have different public keys")
		}

		// 8. Serialize and verify JSON
		jsonData, err := json.Marshal(childAccount)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var jsonMap map[string]interface{}
		json.Unmarshal(jsonData, &jsonMap)

		if jsonMap["parentPublicKey"] == nil {
			t.Error("JSON missing parentPublicKey for child")
		}
		if jsonMap["did"] != nil {
			t.Error("JSON should not have did for child")
		}
		if jsonMap["accountType"] != "Child" {
			t.Errorf("JSON accountType = %v, want 'Child'", jsonMap["accountType"])
		}
	})
}

// =============================================================================
// Shared Account Lifecycle Tests
// =============================================================================

func TestIntegration_SharedAccountLifecycle(t *testing.T) {
	t.Run("complete lifecycle: create multisig -> build -> validate -> serialize", func(t *testing.T) {
		// 1. Generate keys for 2-of-3 multisig
		priv1, _ := keylib.GeneratePrivateKey()
		priv2, _ := keylib.GeneratePrivateKey()
		priv3, _ := keylib.GeneratePrivateKey()

		pubKeys := []keylib.NeuronPublicKey{
			priv1.PublicKey(),
			priv2.PublicKey(),
			priv3.PublicKey(),
		}

		// 2. Create MultisigKey (2-of-3)
		multisigKey, err := keylib.NewMultisigKey(pubKeys, 2)
		if err != nil {
			t.Fatalf("NewMultisigKey error: %v", err)
		}

		// 3. Build Shared account via NewSharedAccountBuilder
		account, err := NewSharedAccountBuilder(multisigKey).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// 4. Validate
		if err := account.Validate(); err != nil {
			t.Errorf("Validate error: %v", err)
		}

		// 5. Verify IsZero() returns false for valid account
		if account.IsZero() {
			t.Error("valid Shared account should not be zero")
		}

		// 6. Verify IsShared() returns true
		if !account.IsShared() {
			t.Error("should be shared account")
		}
		if account.IsParent() {
			t.Error("should not be parent")
		}
		if account.IsChild() {
			t.Error("should not be child")
		}

		// 7. Verify prohibited fields are empty (DID, stdIn, stdOut, stdErr, parentPubKey, publicKey)
		if account.DID() != nil {
			t.Error("Shared account should have no DID")
		}
		if !account.StdIn().IsZero() {
			t.Error("Shared account should have no StdIn")
		}
		if !account.StdOut().IsZero() {
			t.Error("Shared account should have no StdOut")
		}
		if !account.StdErr().IsZero() {
			t.Error("Shared account should have no StdErr")
		}
		if !account.ParentPublicKey().IsZero() {
			t.Error("Shared account should have no ParentPublicKey")
		}
		if !account.PublicKey().IsZero() {
			t.Error("Shared account should have zero PublicKey (uses MultisigKey)")
		}

		// 8. Verify MultisigKey is accessible
		retrievedMultisig := account.MultisigKey()
		if retrievedMultisig == nil {
			t.Fatal("MultisigKey should not be nil")
		}
		if retrievedMultisig.Threshold() != 2 {
			t.Errorf("Threshold = %d, want 2", retrievedMultisig.Threshold())
		}
		if retrievedMultisig.Total() != 3 {
			t.Errorf("Total = %d, want 3", retrievedMultisig.Total())
		}

		// 9. JSON marshal and verify fields
		jsonData, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if jsonMap["accountType"] != "Shared" {
			t.Errorf("JSON accountType = %v, want 'Shared'", jsonMap["accountType"])
		}
		// Multisig fields are serialized as separate JSON fields
		if jsonMap["multisigThreshold"] == nil {
			t.Error("JSON missing multisigThreshold for Shared account")
		}
		if jsonMap["multisigTotal"] == nil {
			t.Error("JSON missing multisigTotal for Shared account")
		}
		if jsonMap["multisigKeys"] == nil {
			t.Error("JSON missing multisigKeys for Shared account")
		}
		if jsonMap["did"] != nil {
			t.Error("JSON should not have did for Shared account")
		}
		if jsonMap["parentPublicKey"] != nil {
			t.Error("JSON should not have parentPublicKey for Shared account")
		}

		// 10. JSON unmarshal and verify Equal() with original
		var reconstructed NeuronAccount
		if err := json.Unmarshal(jsonData, &reconstructed); err != nil {
			t.Fatalf("Unmarshal into NeuronAccount error: %v", err)
		}

		if !account.Equal(reconstructed) {
			t.Error("reconstructed account should equal original")
		}
	})

	t.Run("IsZero behavior", func(t *testing.T) {
		// Valid Shared account: IsZero() == false
		priv1, _ := keylib.GeneratePrivateKey()
		priv2, _ := keylib.GeneratePrivateKey()

		multisigKey, _ := keylib.NewMultisigKey(
			[]keylib.NeuronPublicKey{priv1.PublicKey(), priv2.PublicKey()},
			2,
		)

		account, err := NewSharedAccountBuilder(multisigKey).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if account.IsZero() {
			t.Error("valid Shared account should return IsZero() == false")
		}

		// Zero-value NeuronAccount: IsZero() == true
		var zeroAccount NeuronAccount
		if !zeroAccount.IsZero() {
			t.Error("zero-value NeuronAccount should return IsZero() == true")
		}
	})

	t.Run("Equal behavior between Shared accounts", func(t *testing.T) {
		// Generate keys
		priv1, _ := keylib.GeneratePrivateKey()
		priv2, _ := keylib.GeneratePrivateKey()
		priv3, _ := keylib.GeneratePrivateKey()

		pubKeys := []keylib.NeuronPublicKey{
			priv1.PublicKey(),
			priv2.PublicKey(),
		}

		// Same multisig config: Equal() == true
		multisigKey1, _ := keylib.NewMultisigKey(pubKeys, 2)
		multisigKey2, _ := keylib.NewMultisigKey(pubKeys, 2) // Same keys and threshold

		account1, _ := NewSharedAccountBuilder(multisigKey1).Build()
		account2, _ := NewSharedAccountBuilder(multisigKey2).Build()

		if !account1.Equal(account2) {
			t.Error("Shared accounts with same multisig config should be equal")
		}

		// Different multisig config (different threshold): Equal() == false
		multisigKey3, _ := keylib.NewMultisigKey(pubKeys, 1) // Different threshold
		account3, _ := NewSharedAccountBuilder(multisigKey3).Build()

		if account1.Equal(account3) {
			t.Error("Shared accounts with different thresholds should not be equal")
		}

		// Different multisig config (different keys): Equal() == false
		differentKeys := []keylib.NeuronPublicKey{
			priv2.PublicKey(),
			priv3.PublicKey(),
		}
		multisigKey4, _ := keylib.NewMultisigKey(differentKeys, 2)
		account4, _ := NewSharedAccountBuilder(multisigKey4).Build()

		if account1.Equal(account4) {
			t.Error("Shared accounts with different keys should not be equal")
		}

		// Shared vs Parent: Equal() == false
		parentDID := newIntegrationMockDID(priv1.PublicKey())
		parentAccount, _ := NewParentAccountBuilder(priv1.PublicKey(), parentDID).Build()

		if account1.Equal(parentAccount) {
			t.Error("Shared account should not equal Parent account")
		}
		if parentAccount.Equal(account1) {
			t.Error("Parent account should not equal Shared account")
		}
	})
}

// =============================================================================
// PeerID Consistency Rule Tests (SPEC 4.2)
// =============================================================================

func TestIntegration_PeerIDConsistencyRule(t *testing.T) {
	t.Run("account PeerID matches all reachable addresses", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)

		// Multiple addresses with same (correct) PeerID
		addr1 := "/ip4/192.168.1.1/tcp/4001/p2p/" + knownPeerID
		addr2 := "/ip4/192.168.1.2/tcp/4001/p2p/" + knownPeerID
		addr3 := "/ip6/::1/tcp/4001/p2p/" + knownPeerID

		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(addr1, addr2, addr3).
			Build()

		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// All addresses should have matching PeerID
		for i, addr := range account.ReachableAddrs().Addrs() {
			if addr.PeerID().String() != knownPeerID {
				t.Errorf("addr[%d] PeerID = %v, want %v", i, addr.PeerID(), knownPeerID)
			}
		}

		// Account PeerID should match
		if account.PeerID().String() != knownPeerID {
			t.Errorf("account PeerID = %v, want %v", account.PeerID(), knownPeerID)
		}
	})

	t.Run("rejects address with mismatched PeerID at Build time", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)

		// One address with correct PeerID, one with wrong
		correctAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + knownPeerID
		wrongAddr := "/ip4/192.168.1.2/tcp/4001/p2p/" + differentPeerID

		_, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(correctAddr, wrongAddr).
			Build()

		if err == nil {
			t.Fatal("Build should fail when PeerID mismatches")
		}

		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})

	t.Run("WithReachableAddrValidated catches mismatch immediately", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)
		wrongAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + differentPeerID

		b := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrValidated(wrongAddr)

		if !b.HasErrors() {
			t.Error("WithReachableAddrValidated should catch PeerID mismatch")
		}

		errs := b.Errors()
		var ae *AccountError
		for _, e := range errs {
			if errors.As(e, &ae) && ae.Kind == ErrKindPeerIDMismatch {
				return // Found expected error
			}
		}
		t.Error("expected ErrKindPeerIDMismatch error")
	})
}

// =============================================================================
// Known Vectors Tests
// =============================================================================

func TestIntegration_KnownVectors(t *testing.T) {
	t.Run("known public key produces known PeerID", func(t *testing.T) {
		pubKey, err := keylib.ParsePublicKeyHex(knownPubKeyHex)
		if err != nil {
			t.Fatalf("ParsePublicKeyHex error: %v", err)
		}

		peerID, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("PeerID error: %v", err)
		}

		if peerID.String() != knownPeerID {
			t.Errorf("PeerID = %v, want %v", peerID.String(), knownPeerID)
		}
	})

	t.Run("known public key produces consistent EVMAddress", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)

		evmAddr := pubKey.EVMAddress()
		if evmAddr.IsZero() {
			t.Error("EVMAddress should not be zero")
		}

		// Verify determinism
		evmAddr2 := pubKey.EVMAddress()
		if evmAddr != evmAddr2 {
			t.Error("EVMAddress should be deterministic")
		}
	})

	t.Run("account built with known key has correct identifiers", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + knownPeerID

		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// Verify all derived identifiers
		if account.PeerID().String() != knownPeerID {
			t.Errorf("account PeerID = %v, want %v", account.PeerID(), knownPeerID)
		}

		// Verify the public key matches
		if account.PublicKey().Hex() != knownPubKeyHex {
			t.Errorf("account PublicKey = %v, want %v", account.PublicKey().Hex(), knownPubKeyHex)
		}
	})
}

// =============================================================================
// Error Handling Integration Tests
// =============================================================================

func TestIntegration_ErrorHandling(t *testing.T) {
	t.Run("builder accumulates multiple errors", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		// Build with multiple problems
		b := NewParentAccountBuilder(zeroPubKey, nil).
			WithStdInHedera("invalid1").
			WithStdOutHedera("invalid2").
			WithReachableAddr("invalid-multiaddr")

		// Should have accumulated many errors
		errs := b.Errors()
		if len(errs) < 4 {
			t.Errorf("expected at least 4 errors, got %d", len(errs))
		}

		// Build should fail with first error
		_, err := b.Build()
		if err == nil {
			t.Fatal("Build should fail")
		}
	})

	t.Run("validation catches all issues", func(t *testing.T) {
		var account NeuronAccount
		result := account.ValidateAll()

		if !result.HasErrors() {
			t.Error("zero account should have validation errors")
		}

		// Zero account has invalid account type (Unspecified)
		// Public key check is only done for valid account types
		if len(result.Errors()) < 1 {
			t.Errorf("expected at least 1 error, got %d", len(result.Errors()))
		}
	})

	t.Run("error types are correct", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		_, err := NewParentAccountBuilder(zeroPubKey, nil).Build()
		if err == nil {
			t.Fatal("expected error")
		}

		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
		if ae.Kind == 0 {
			t.Error("error should have a Kind")
		}
	})
}

// =============================================================================
// Communication Endpoints Integration Tests
// =============================================================================

func TestIntegration_CommunicationEndpoints(t *testing.T) {
	t.Run("Hedera topics are correctly set on child account", func(t *testing.T) {
		// Child accounts must have all 3 comm channels (Parent accounts cannot have any)
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()

		childPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()

		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithHederaTopics("0.0.111", "0.0.222", "0.0.333").
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// Verify all topics
		if account.StdIn().Kind() != CommAddressKind(HederaTopicKind) {
			t.Error("StdIn should be HederaTopic")
		}
		if account.StdIn().Locator() != "0.0.111" {
			t.Errorf("StdIn Locator = %v, want '0.0.111'", account.StdIn().Locator())
		}

		if account.StdOut().Locator() != "0.0.222" {
			t.Errorf("StdOut Locator = %v, want '0.0.222'", account.StdOut().Locator())
		}

		if account.StdErr().Locator() != "0.0.333" {
			t.Errorf("StdErr Locator = %v, want '0.0.333'", account.StdErr().Locator())
		}
	})

	t.Run("Kafka topics are correctly set on child account", func(t *testing.T) {
		// Child accounts must have all 3 comm channels
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()

		childPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()

		kafkaStdIn, _ := NewKafkaTopicAddress("agent.stdin")
		kafkaStdOut, _ := NewKafkaTopicAddress("agent.stdout")
		kafkaStdErr, _ := NewKafkaTopicAddress("agent.stderr")

		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdIn(kafkaStdIn).
			WithStdOut(kafkaStdOut).
			WithStdErr(kafkaStdErr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if account.StdIn().Kind() != CommAddressKind(KafkaTopicKind) {
			t.Error("StdIn should be KafkaTopic")
		}
		if account.StdIn().Locator() != "agent.stdin" {
			t.Errorf("StdIn Locator = %v, want 'agent.stdin'", account.StdIn().Locator())
		}
	})
}

// =============================================================================
// Reachable Address Integration Tests
// =============================================================================

func TestIntegration_ReachableAddresses(t *testing.T) {
	t.Run("multiple address types work together", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)

		// Different transport protocols, same PeerID
		tcpAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + knownPeerID
		quicAddr := "/ip4/192.168.1.1/udp/4001/quic-v1/p2p/" + knownPeerID
		wsAddr := "/ip4/192.168.1.1/tcp/443/ws/p2p/" + knownPeerID

		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(tcpAddr, quicAddr, wsAddr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		addrs := account.ReachableAddrs()
		if len(addrs.Addrs()) != 3 {
			t.Errorf("expected 3 addrs, got %d", len(addrs.Addrs()))
		}

		// All should have same PeerID
		for _, addr := range addrs.Addrs() {
			if addr.PeerID().String() != knownPeerID {
				t.Errorf("addr PeerID = %v, want %v", addr.PeerID(), knownPeerID)
			}
		}
	})

	t.Run("circuit relay addresses work", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)

		// Circuit relay through a different peer
		relayAddr := "/ip4/192.168.1.100/tcp/4001/p2p/" + differentPeerID + "/p2p-circuit/p2p/" + knownPeerID

		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(relayAddr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		addrs := account.ReachableAddrs()
		if len(addrs.Addrs()) != 1 {
			t.Fatalf("expected 1 addr, got %d", len(addrs.Addrs()))
		}

		// The extracted PeerID should be the TARGET (knownPeerID), not the relay
		addr := addrs.Addrs()[0]
		if addr.PeerID().String() != knownPeerID {
			t.Errorf("circuit relay target PeerID = %v, want %v", addr.PeerID(), knownPeerID)
		}

		// Should be identified as circuit relay
		if !addr.IsCircuitRelay() {
			t.Error("should be identified as circuit relay")
		}
	})
}

// =============================================================================
// JSON Serialization Integration Tests
// =============================================================================

func TestIntegration_JSONSerialization(t *testing.T) {
	t.Run("full parent account serializes correctly", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + knownPeerID

		// Parent accounts must NOT have comm channels
		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr).
			Build()

		jsonData, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(jsonData, &result)

		// Check all expected fields (parent accounts don't have comm channels)
		expectedFields := []string{"publicKey", "peerId", "evmAddress", "accountType", "did", "reachableAddrs"}
		for _, field := range expectedFields {
			if result[field] == nil {
				t.Errorf("JSON missing field: %s", field)
			}
		}

		// Verify specific values
		if result["publicKey"] != knownPubKeyHex {
			t.Errorf("publicKey = %v, want %v", result["publicKey"], knownPubKeyHex)
		}
		if result["peerId"] != knownPeerID {
			t.Errorf("peerId = %v, want %v", result["peerId"], knownPeerID)
		}
	})

	t.Run("child account JSON differs from parent", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		// Child accounts must have all 3 comm channels
		childAccount, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		jsonData, _ := json.Marshal(childAccount)

		var result map[string]interface{}
		json.Unmarshal(jsonData, &result)

		// Child should have parentPublicKey, not did
		if result["parentPublicKey"] == nil {
			t.Error("child JSON should have parentPublicKey")
		}
		if result["did"] != nil {
			t.Error("child JSON should not have did")
		}
	})
}

// =============================================================================
// Equality Integration Tests
// =============================================================================

func TestIntegration_Equality(t *testing.T) {
	t.Run("accounts with same public key are equal regardless of other fields", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(knownPubKeyHex)
		did := newIntegrationMockDID(pubKey)

		// Build two accounts with same key (Parent accounts must NOT have comm channels)
		account1, _ := NewParentAccountBuilder(pubKey, did).
			Build()

		account2, _ := NewParentAccountBuilder(pubKey, did).
			Build()

		if !account1.Equal(account2) {
			t.Error("accounts with same public key should be equal")
		}
	})

	t.Run("parent and child with different keys are not equal", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()
		parentDID := newIntegrationMockDID(parentPubKey)

		childPriv, _ := keylib.GeneratePrivateKey()

		parentAccount, _ := NewParentAccountBuilder(parentPubKey, parentDID).Build()
		// Child accounts must have all 3 comm channels
		childAccount, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPubKey).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		if parentAccount.Equal(childAccount) {
			t.Error("parent and child with different keys should not be equal")
		}
	})
}
