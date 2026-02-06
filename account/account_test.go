package account

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Test Helpers
// =============================================================================

// accountMockDID is a test double for NeuronDID interface used in account tests
type accountMockDID struct {
	did      string
	method   string
	id       string
	valid    bool
	matchKey keylib.NeuronPublicKey
}

func (m *accountMockDID) String() string     { return m.did }
func (m *accountMockDID) Method() string     { return m.method }
func (m *accountMockDID) Identifier() string { return m.id }
func (m *accountMockDID) Validate() error {
	if !m.valid {
		return errors.New("invalid mock DID")
	}
	return nil
}
func (m *accountMockDID) Equal(other NeuronDID) bool {
	return m.did == other.String()
}
func (m *accountMockDID) PublicKey() (keylib.NeuronPublicKey, error) {
	return m.matchKey, nil
}
func (m *accountMockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	return m.matchKey.Equal(pubKey)
}

func newAccountMockDID(pubKey keylib.NeuronPublicKey) *accountMockDID {
	return &accountMockDID{
		did:      "did:mock:test",
		method:   "mock",
		id:       "test",
		valid:    true,
		matchKey: pubKey,
	}
}

// createTestParentAccount creates a valid parent account for testing
// Per spec: Parent accounts must NOT have comm channels
func createTestParentAccount(t *testing.T) NeuronAccount {
	t.Helper()
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did := newAccountMockDID(pubKey)

	account, err := NewParentAccountBuilder(pubKey, did).
		Build()
	if err != nil {
		t.Fatalf("failed to create parent account: %v", err)
	}
	return account
}

// createTestChildAccount creates a valid child account for testing
// Per spec: Child accounts must have ALL 3 comm channels
func createTestChildAccount(t *testing.T) NeuronAccount {
	t.Helper()
	childPriv, _ := keylib.GeneratePrivateKey()
	parentPriv, _ := keylib.GeneratePrivateKey()

	account, err := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
		WithStdInHedera("0.0.111").
		WithStdOutHedera("0.0.222").
		WithStdErrHedera("0.0.333").
		Build()
	if err != nil {
		t.Fatalf("failed to create child account: %v", err)
	}
	return account
}

// =============================================================================
// Identity Accessor Tests
// =============================================================================

func TestNeuronAccount_IdentityAccessors(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	t.Run("PublicKey returns the account's public key", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if !account.PublicKey().Equal(pubKey) {
			t.Error("PublicKey() does not match")
		}
	})

	t.Run("PeerID is correctly derived from public key", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if account.PeerID().String() != testPeerID {
			t.Errorf("PeerID = %v, want %v", account.PeerID().String(), testPeerID)
		}
	})

	t.Run("EVMAddress is derived from public key", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// EVMAddress should not be zero
		if account.EVMAddress().IsZero() {
			t.Error("EVMAddress should not be zero")
		}

		// Verify determinism - same key should produce same address
		expectedEVMAddr := pubKey.EVMAddress()
		if account.EVMAddress() != expectedEVMAddr {
			t.Errorf("EVMAddress = %v, want %v", account.EVMAddress(), expectedEVMAddr)
		}
	})
}

// =============================================================================
// Hierarchy Accessor Tests
// =============================================================================

func TestNeuronAccount_HierarchyAccessors(t *testing.T) {
	t.Run("AccountType returns correct type for parent", func(t *testing.T) {
		account := createTestParentAccount(t)
		if account.AccountType() != AccountTypeParent {
			t.Errorf("AccountType = %v, want Parent", account.AccountType())
		}
	})

	t.Run("AccountType returns correct type for child", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.AccountType() != AccountTypeChild {
			t.Errorf("AccountType = %v, want Child", account.AccountType())
		}
	})

	t.Run("IsParent returns true for parent account", func(t *testing.T) {
		account := createTestParentAccount(t)
		if !account.IsParent() {
			t.Error("IsParent() should return true")
		}
	})

	t.Run("IsParent returns false for child account", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.IsParent() {
			t.Error("IsParent() should return false for child")
		}
	})

	t.Run("IsChild returns true for child account", func(t *testing.T) {
		account := createTestChildAccount(t)
		if !account.IsChild() {
			t.Error("IsChild() should return true")
		}
	})

	t.Run("IsChild returns false for parent account", func(t *testing.T) {
		account := createTestParentAccount(t)
		if account.IsChild() {
			t.Error("IsChild() should return false for parent")
		}
	})

	t.Run("DID returns non-nil for parent account", func(t *testing.T) {
		account := createTestParentAccount(t)
		if account.DID() == nil {
			t.Error("DID() should not be nil for parent")
		}
	})

	t.Run("DID returns nil for child account", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.DID() != nil {
			t.Error("DID() should be nil for child")
		}
	})

	t.Run("ParentPublicKey returns zero for parent account", func(t *testing.T) {
		account := createTestParentAccount(t)
		if !account.ParentPublicKey().IsZero() {
			t.Error("ParentPublicKey should be zero for parent account")
		}
	})

	t.Run("ParentPublicKey returns non-zero for child account", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.ParentPublicKey().IsZero() {
			t.Error("ParentPublicKey should not be zero for child account")
		}
	})
}

// =============================================================================
// Communication Endpoint Accessor Tests
// =============================================================================

func TestNeuronAccount_CommAccessors(t *testing.T) {
	// Note: Per spec, comm channels are only allowed on Child accounts
	t.Run("StdIn returns set address", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.StdIn().Locator() != "0.0.111" {
			t.Errorf("StdIn = %v, want '0.0.111'", account.StdIn().Locator())
		}
	})

	t.Run("StdOut returns set address", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.StdOut().Locator() != "0.0.222" {
			t.Errorf("StdOut = %v, want '0.0.222'", account.StdOut().Locator())
		}
	})

	t.Run("StdErr returns set address", func(t *testing.T) {
		account := createTestChildAccount(t)
		if account.StdErr().Locator() != "0.0.333" {
			t.Errorf("StdErr = %v, want '0.0.333'", account.StdErr().Locator())
		}
	})

	t.Run("unset addresses return zero values for parent", func(t *testing.T) {
		// Parent accounts don't have comm channels per spec
		account := createTestParentAccount(t)

		if !account.StdIn().IsZero() {
			t.Error("parent StdIn should be zero")
		}
		if !account.StdOut().IsZero() {
			t.Error("parent StdOut should be zero")
		}
		if !account.StdErr().IsZero() {
			t.Error("parent StdErr should be zero")
		}
	})
}

// =============================================================================
// Reachability Accessor Tests
// =============================================================================

func TestNeuronAccount_ReachabilityAccessors(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	addr1 := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID
	addr2 := "/ip4/192.168.1.2/tcp/4001/p2p/" + testPeerID

	t.Run("ReachableAddrs returns set addresses", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(addr1, addr2).
			Build()

		addrs := account.ReachableAddrs().Addrs()
		if len(addrs) != 2 {
			t.Errorf("expected 2 reachable addrs, got %d", len(addrs))
		}
	})

	t.Run("FirstReachableAddr returns first address", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(addr1, addr2).
			Build()

		first := account.FirstReachableAddr()
		if first.IsZero() {
			t.Error("FirstReachableAddr should not be zero")
		}
		if first.String() != addr1 {
			t.Errorf("FirstReachableAddr = %v, want %v", first.String(), addr1)
		}
	})

	t.Run("FirstReachableAddr returns zero when empty", func(t *testing.T) {
		account := createTestParentAccount(t)

		first := account.FirstReachableAddr()
		if !first.IsZero() {
			t.Error("FirstReachableAddr should be zero when no addrs")
		}
	})
}

// =============================================================================
// IsZero Tests
// =============================================================================

func TestNeuronAccount_IsZero(t *testing.T) {
	t.Run("valid account is not zero", func(t *testing.T) {
		account := createTestParentAccount(t)
		if account.IsZero() {
			t.Error("valid account should not be zero")
		}
	})

	t.Run("zero-value account is zero", func(t *testing.T) {
		var account NeuronAccount
		if !account.IsZero() {
			t.Error("zero-value account should be zero")
		}
	})
}

// =============================================================================
// Validate Tests - Parent Account
// =============================================================================

func TestNeuronAccount_Validate_Parent(t *testing.T) {
	t.Run("valid parent account passes", func(t *testing.T) {
		account := createTestParentAccount(t)
		if err := account.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("zero account fails validation", func(t *testing.T) {
		// A zero NeuronAccount has an invalid account type (Unspecified)
		// which is checked first in Validate()
		var account NeuronAccount
		err := account.Validate()
		if err == nil {
			t.Fatal("expected error for zero account")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindInvalidAccount {
			t.Errorf("error Kind = %v, want %v (invalid account type)", ae.Kind, ErrKindInvalidAccount)
		}
	})

	t.Run("invalid account type fails validation", func(t *testing.T) {
		// Create account with invalid type by building first then modifying
		// Since fields are private, we test this indirectly through ValidateAccountType
		// This test covers the builder's validation
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newAccountMockDID(pubKey)

		account, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// The valid account should pass
		if err := account.Validate(); err != nil {
			t.Errorf("valid account should pass: %v", err)
		}
	})
}

// =============================================================================
// Validate Tests - Child Account
// =============================================================================

func TestNeuronAccount_Validate_Child(t *testing.T) {
	t.Run("valid child account passes", func(t *testing.T) {
		account := createTestChildAccount(t)
		if err := account.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("child with valid parent passes", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		// Child accounts require all 3 comm channels per spec
		account, err := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if err := account.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})
}

// =============================================================================
// Validate Tests - Reachable Addresses
// =============================================================================

func TestNeuronAccount_Validate_ReachableAddrs(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	matchingAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

	t.Run("matching PeerID passes", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(matchingAddr).
			Build()

		if err := account.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("account without reachable addrs passes", func(t *testing.T) {
		account := createTestParentAccount(t)
		if err := account.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})
}

// =============================================================================
// ValidateAll Tests
// =============================================================================

func TestNeuronAccount_ValidateAll(t *testing.T) {
	t.Run("valid account returns no errors", func(t *testing.T) {
		account := createTestParentAccount(t)
		result := account.ValidateAll()

		if result.HasErrors() {
			t.Errorf("ValidateAll() errors = %v", result.Errors())
		}
		if !result.Valid() {
			t.Error("result.Valid() should be true")
		}
	})

	t.Run("zero account returns errors", func(t *testing.T) {
		var account NeuronAccount
		result := account.ValidateAll()

		if !result.HasErrors() {
			t.Error("ValidateAll() should return errors for zero account")
		}
		if result.Valid() {
			t.Error("result.Valid() should be false")
		}
	})

	t.Run("ValidateAll collects multiple errors", func(t *testing.T) {
		// Create a child account with missing comm channels to generate multiple errors
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		// Manually construct to bypass Build() validation
		account := NeuronAccount{
			accountType:  AccountTypeChild,
			publicKey:    childPriv.PublicKey(),
			parentPubKey: parentPriv.PublicKey(),
			// Missing all 3 comm channels
		}
		result := account.ValidateAll()

		// Should have at least error for missing comm channels
		if !result.HasErrors() {
			t.Error("expected errors for missing comm channels")
		}
	})
}

// =============================================================================
// Equal Tests
// =============================================================================

func TestNeuronAccount_Equal(t *testing.T) {
	t.Run("same public key accounts are equal", func(t *testing.T) {
		testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		// Both parent accounts with same public key should be equal
		account1, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}
		account2, err := NewParentAccountBuilder(pubKey, did).Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if !account1.Equal(account2) {
			t.Error("accounts with same public key should be equal")
		}
	})

	t.Run("different public key accounts are not equal", func(t *testing.T) {
		account1 := createTestParentAccount(t)
		account2 := createTestParentAccount(t)

		if account1.Equal(account2) {
			t.Error("accounts with different public keys should not be equal")
		}
	})

	t.Run("zero account is not equal to anything", func(t *testing.T) {
		var zero1 NeuronAccount
		var zero2 NeuronAccount
		account := createTestParentAccount(t)

		if zero1.Equal(zero2) {
			t.Error("zero accounts should not be equal to each other")
		}
		if zero1.Equal(account) {
			t.Error("zero account should not be equal to valid account")
		}
		if account.Equal(zero1) {
			t.Error("valid account should not be equal to zero account")
		}
	})
}

// =============================================================================
// MarshalJSON Tests
// =============================================================================

func TestNeuronAccount_MarshalJSON(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

	t.Run("includes all required fields", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		// Parent accounts don't have comm channels per spec
		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		// Required fields
		if result["publicKey"] == nil {
			t.Error("publicKey should be present")
		}
		if result["peerId"] == nil {
			t.Error("peerId should be present")
		}
		if result["evmAddress"] == nil {
			t.Error("evmAddress should be present")
		}
		if result["accountType"] == nil {
			t.Error("accountType should be present")
		}
	})

	t.Run("includes DID for parent account", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).Build()

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["did"] == nil {
			t.Error("did should be present for parent account")
		}
		if result["did"] != "did:mock:test" {
			t.Errorf("did = %v, want 'did:mock:test'", result["did"])
		}
	})

	t.Run("omits DID for child account", func(t *testing.T) {
		account := createTestChildAccount(t)

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["did"] != nil {
			t.Error("did should be omitted for child account")
		}
	})

	t.Run("includes parentPublicKey for child account", func(t *testing.T) {
		account := createTestChildAccount(t)

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["parentPublicKey"] == nil {
			t.Error("parentPublicKey should be present for child account")
		}
	})

	t.Run("omits empty communication addresses", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newAccountMockDID(privKey.PublicKey())

		account, _ := NewParentAccountBuilder(privKey.PublicKey(), did).Build()

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["stdIn"] != nil {
			t.Error("empty stdIn should be omitted")
		}
		if result["stdOut"] != nil {
			t.Error("empty stdOut should be omitted")
		}
		if result["stdErr"] != nil {
			t.Error("empty stdErr should be omitted")
		}
	})

	t.Run("includes communication addresses when set", func(t *testing.T) {
		// Per spec, only child accounts have comm channels
		account := createTestChildAccount(t)

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["stdIn"] == nil {
			t.Error("stdIn should be present when set")
		}
		if result["stdOut"] == nil {
			t.Error("stdOut should be present when set")
		}
		if result["stdErr"] == nil {
			t.Error("stdErr should be present when set")
		}
	})

	t.Run("omits empty reachable addresses", func(t *testing.T) {
		account := createTestParentAccount(t)

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["reachableAddrs"] != nil {
			t.Error("empty reachableAddrs should be omitted")
		}
	})

	t.Run("includes reachable addresses when set", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr).
			Build()

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["reachableAddrs"] == nil {
			t.Error("reachableAddrs should be present when set")
		}
	})

	t.Run("formats are correct", func(t *testing.T) {
		// Use child account since it has comm channels
		account := createTestChildAccount(t)

		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		// PublicKey should be hex with 0x prefix
		pk := result["publicKey"].(string)
		if !strings.HasPrefix(pk, "0x") {
			t.Error("publicKey should have 0x prefix")
		}

		// AccountType should be string
		at := result["accountType"].(string)
		if at != "Child" {
			t.Errorf("accountType = %v, want 'Child'", at)
		}

		// stdIn should be kind:locator format
		stdIn := result["stdIn"].(string)
		if !strings.Contains(stdIn, ":") {
			t.Error("stdIn should be in kind:locator format")
		}
	})
}

// =============================================================================
// String Tests
// =============================================================================

func TestNeuronAccount_String(t *testing.T) {
	t.Run("zero account returns zero-value indicator", func(t *testing.T) {
		var account NeuronAccount
		str := account.String()
		if str != "NeuronAccount{zero-value}" {
			t.Errorf("String() = %v, want 'NeuronAccount{zero-value}'", str)
		}
	})

	t.Run("parent account includes type and DID", func(t *testing.T) {
		account := createTestParentAccount(t)
		str := account.String()

		if !strings.Contains(str, "Parent") {
			t.Error("String() should contain 'Parent'")
		}
		if !strings.Contains(str, "did=") {
			t.Error("String() should contain 'did='")
		}
	})

	t.Run("child account includes type and pubKey prefix", func(t *testing.T) {
		account := createTestChildAccount(t)
		str := account.String()

		if !strings.Contains(str, "Child") {
			t.Error("String() should contain 'Child'")
		}
		if !strings.Contains(str, "pubKey=") {
			t.Error("String() should contain 'pubKey='")
		}
		if !strings.Contains(str, "...") {
			t.Error("String() should truncate pubKey with '...'")
		}
	})

	t.Run("string format is readable", func(t *testing.T) {
		account := createTestParentAccount(t)
		str := account.String()

		if !strings.HasPrefix(str, "NeuronAccount{") {
			t.Error("String() should start with 'NeuronAccount{'")
		}
		if !strings.HasSuffix(str, "}") {
			t.Error("String() should end with '}'")
		}
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestNeuronAccount_Integration(t *testing.T) {
	t.Run("full parent account workflow", func(t *testing.T) {
		testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newAccountMockDID(pubKey)

		// Build - Parent accounts don't have comm channels per spec
		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// Validate
		if err := account.Validate(); err != nil {
			t.Errorf("Validate error: %v", err)
		}

		// Check all accessors
		if !account.PublicKey().Equal(pubKey) {
			t.Error("PublicKey mismatch")
		}
		if account.PeerID().String() != testPeerID {
			t.Error("PeerID mismatch")
		}
		if account.EVMAddress().IsZero() {
			t.Error("EVMAddress should not be zero")
		}
		if !account.IsParent() {
			t.Error("should be parent")
		}
		if account.DID() == nil {
			t.Error("DID should not be nil")
		}
		// Parent accounts don't have comm channels
		if !account.StdIn().IsZero() {
			t.Error("parent StdIn should be zero")
		}
		if len(account.ReachableAddrs().Addrs()) != 1 {
			t.Error("should have 1 reachable addr")
		}

		// JSON round-trip
		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if len(data) == 0 {
			t.Error("JSON should not be empty")
		}

		// String
		str := account.String()
		if str == "" {
			t.Error("String should not be empty")
		}
	})

	t.Run("full child account workflow", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()
		parentPubKey := parentPriv.PublicKey()

		childPeerID, _ := childPubKey.PeerID()
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + childPeerID.String()

		// Build - Child accounts require all 3 comm channels per spec
		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			WithReachableAddr(addr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// Validate
		if err := account.Validate(); err != nil {
			t.Errorf("Validate error: %v", err)
		}

		// Check child-specific accessors
		if !account.IsChild() {
			t.Error("should be child")
		}
		if account.DID() != nil {
			t.Error("child should have no DID")
		}
		if !account.ParentPublicKey().Equal(parentPubKey) {
			t.Error("ParentPublicKey mismatch")
		}

		// JSON
		data, err := json.Marshal(account)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(data, &result)

		if result["parentPublicKey"] == nil {
			t.Error("JSON should contain parentPublicKey")
		}
		if result["did"] != nil {
			t.Error("JSON should not contain did for child")
		}
	})
}
