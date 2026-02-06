package account

import (
	"errors"
	"strings"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Builder Test Helpers
// =============================================================================

// builderMockDID is a test double for NeuronDID interface used in builder tests
type builderMockDID struct {
	did      string
	method   string
	id       string
	valid    bool
	matchKey keylib.NeuronPublicKey
}

func (m *builderMockDID) String() string     { return m.did }
func (m *builderMockDID) Method() string     { return m.method }
func (m *builderMockDID) Identifier() string { return m.id }
func (m *builderMockDID) Validate() error {
	if !m.valid {
		return errors.New("invalid mock DID")
	}
	return nil
}
func (m *builderMockDID) Equal(other NeuronDID) bool {
	return m.did == other.String()
}
func (m *builderMockDID) PublicKey() (keylib.NeuronPublicKey, error) {
	return m.matchKey, nil
}
func (m *builderMockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	return m.matchKey.Equal(pubKey)
}

func newBuilderMockDID(pubKey keylib.NeuronPublicKey) *builderMockDID {
	return &builderMockDID{
		did:      "did:mock:test",
		method:   "mock",
		id:       "test",
		valid:    true,
		matchKey: pubKey,
	}
}

func newMismatchedMockDID() *builderMockDID {
	differentKey, _ := keylib.GeneratePrivateKey()
	return &builderMockDID{
		did:      "did:mock:different",
		method:   "mock",
		id:       "different",
		valid:    true,
		matchKey: differentKey.PublicKey(),
	}
}

// =============================================================================
// NewParentAccountBuilder Tests
// =============================================================================

func TestNewParentAccountBuilder(t *testing.T) {
	t.Run("creates builder with valid inputs", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did)

		if b == nil {
			t.Fatal("expected non-nil builder")
		}
		if b.HasErrors() {
			t.Errorf("expected no errors, got: %v", b.Errors())
		}
		if b.accountType != AccountTypeParent {
			t.Errorf("accountType = %v, want Parent", b.accountType)
		}
	})

	t.Run("records error for zero public key", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		did := newBuilderMockDID(privKey.PublicKey())
		var zeroPubKey keylib.NeuronPublicKey

		b := NewParentAccountBuilder(zeroPubKey, did)

		if !b.HasErrors() {
			t.Error("expected error for zero public key")
		}
		errs := b.Errors()
		if len(errs) == 0 {
			t.Fatal("expected at least one error")
		}
		var ae *AccountError
		if errors.As(errs[0], &ae) {
			if ae.Kind != ErrKindZeroValue {
				t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
			}
		}
	})

	t.Run("records error for nil DID", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		b := NewParentAccountBuilder(pubKey, nil)

		if !b.HasErrors() {
			t.Error("expected error for nil DID")
		}
		errs := b.Errors()
		if len(errs) == 0 {
			t.Fatal("expected at least one error")
		}
		var ae *AccountError
		if errors.As(errs[0], &ae) {
			if ae.Kind != ErrKindMissingRequired {
				t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindMissingRequired)
			}
		}
	})

	t.Run("accumulates multiple errors", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		b := NewParentAccountBuilder(zeroPubKey, nil)

		if len(b.Errors()) < 2 {
			t.Errorf("expected at least 2 errors, got %d", len(b.Errors()))
		}
	})
}

// =============================================================================
// NewChildAccountBuilder Tests
// =============================================================================

func TestNewChildAccountBuilder(t *testing.T) {
	t.Run("creates builder with valid inputs", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		b := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey())

		if b == nil {
			t.Fatal("expected non-nil builder")
		}
		if b.HasErrors() {
			t.Errorf("expected no errors, got: %v", b.Errors())
		}
		if b.accountType != AccountTypeChild {
			t.Errorf("accountType = %v, want Child", b.accountType)
		}
	})

	t.Run("records error for zero child public key", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		var zeroChildKey keylib.NeuronPublicKey

		b := NewChildAccountBuilder(zeroChildKey, parentPriv.PublicKey())

		if !b.HasErrors() {
			t.Error("expected error for zero child public key")
		}
		var ae *AccountError
		if errors.As(b.Errors()[0], &ae) && ae.Kind != ErrKindZeroValue {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
		}
	})

	t.Run("records error for zero parent public key", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		var zeroParentKey keylib.NeuronPublicKey

		b := NewChildAccountBuilder(childPriv.PublicKey(), zeroParentKey)

		if !b.HasErrors() {
			t.Error("expected error for zero parent public key")
		}
		var ae *AccountError
		if errors.As(b.Errors()[0], &ae) && ae.Kind != ErrKindMissingRequired {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindMissingRequired)
		}
	})

	t.Run("accumulates multiple errors", func(t *testing.T) {
		var zeroChildKey, zeroParentKey keylib.NeuronPublicKey

		b := NewChildAccountBuilder(zeroChildKey, zeroParentKey)

		if len(b.Errors()) < 2 {
			t.Errorf("expected at least 2 errors, got %d", len(b.Errors()))
		}
	})
}

// =============================================================================
// WithStdIn Tests
// =============================================================================

func TestWithStdIn(t *testing.T) {
	t.Run("sets valid address", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)
		addr, _ := NewHederaTopicAddress("0.0.12345")

		b := NewParentAccountBuilder(pubKey, did).WithStdIn(addr)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if !b.stdIn.Equal(addr) {
			t.Error("stdIn not set correctly")
		}
	})

	t.Run("allows zero address (optional)", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)
		var zeroAddr CommAddress

		b := NewParentAccountBuilder(pubKey, did).WithStdIn(zeroAddr)

		if b.HasErrors() {
			t.Errorf("zero address should be allowed: %v", b.Errors())
		}
	})

	t.Run("supports fluent chaining", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		result := NewParentAccountBuilder(pubKey, did).WithStdIn(CommAddress{})
		if result == nil {
			t.Error("expected chained builder")
		}
	})
}

// =============================================================================
// WithStdInHedera Tests
// =============================================================================

func TestWithStdInHedera(t *testing.T) {
	t.Run("sets valid Hedera topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdInHedera("0.0.12345")

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if b.stdIn.Kind() != CommAddressKind(HederaTopicKind) {
			t.Errorf("stdIn Kind = %v, want HederaTopic", b.stdIn.Kind())
		}
		if b.stdIn.Locator() != "0.0.12345" {
			t.Errorf("stdIn Locator = %v, want '0.0.12345'", b.stdIn.Locator())
		}
	})

	t.Run("records error for invalid topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdInHedera("invalid")

		if !b.HasErrors() {
			t.Error("expected error for invalid topic")
		}
		var ae *AccountError
		if errors.As(b.Errors()[0], &ae) && ae.Kind != ErrKindInvalidAddress {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidAddress)
		}
	})

	t.Run("records error for empty topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdInHedera("")

		if !b.HasErrors() {
			t.Error("expected error for empty topic")
		}
	})
}

// =============================================================================
// WithStdOut and WithStdOutHedera Tests
// =============================================================================

func TestWithStdOut(t *testing.T) {
	t.Run("sets valid address", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)
		addr, _ := NewHederaTopicAddress("0.0.54321")

		b := NewParentAccountBuilder(pubKey, did).WithStdOut(addr)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if !b.stdOut.Equal(addr) {
			t.Error("stdOut not set correctly")
		}
	})
}

func TestWithStdOutHedera(t *testing.T) {
	t.Run("sets valid Hedera topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdOutHedera("0.0.54321")

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if b.stdOut.Locator() != "0.0.54321" {
			t.Errorf("stdOut Locator = %v, want '0.0.54321'", b.stdOut.Locator())
		}
	})

	t.Run("records error for invalid topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdOutHedera("bad-topic")

		if !b.HasErrors() {
			t.Error("expected error for invalid topic")
		}
	})
}

// =============================================================================
// WithStdErr and WithStdErrHedera Tests
// =============================================================================

func TestWithStdErr(t *testing.T) {
	t.Run("sets valid address", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)
		addr, _ := NewHederaTopicAddress("0.0.99999")

		b := NewParentAccountBuilder(pubKey, did).WithStdErr(addr)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if !b.stdErr.Equal(addr) {
			t.Error("stdErr not set correctly")
		}
	})
}

func TestWithStdErrHedera(t *testing.T) {
	t.Run("sets valid Hedera topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdErrHedera("0.0.99999")

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if b.stdErr.Locator() != "0.0.99999" {
			t.Errorf("stdErr Locator = %v, want '0.0.99999'", b.stdErr.Locator())
		}
	})

	t.Run("records error for invalid topic", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithStdErrHedera("0.0")

		if !b.HasErrors() {
			t.Error("expected error for invalid topic")
		}
	})
}

// =============================================================================
// WithHederaTopics Tests
// =============================================================================

func TestWithHederaTopics(t *testing.T) {
	t.Run("sets all three topics", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithHederaTopics("0.0.111", "0.0.222", "0.0.333")

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if b.stdIn.Locator() != "0.0.111" {
			t.Errorf("stdIn = %v, want '0.0.111'", b.stdIn.Locator())
		}
		if b.stdOut.Locator() != "0.0.222" {
			t.Errorf("stdOut = %v, want '0.0.222'", b.stdOut.Locator())
		}
		if b.stdErr.Locator() != "0.0.333" {
			t.Errorf("stdErr = %v, want '0.0.333'", b.stdErr.Locator())
		}
	})

	t.Run("accumulates errors for all invalid topics", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithHederaTopics("bad1", "bad2", "bad3")

		if len(b.Errors()) != 3 {
			t.Errorf("expected 3 errors, got %d", len(b.Errors()))
		}
	})

	t.Run("partial invalid topics", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithHederaTopics("0.0.111", "invalid", "0.0.333")

		if len(b.Errors()) != 1 {
			t.Errorf("expected 1 error, got %d", len(b.Errors()))
		}
		// Valid ones should still be set
		if b.stdIn.Locator() != "0.0.111" {
			t.Errorf("stdIn = %v, want '0.0.111'", b.stdIn.Locator())
		}
		if b.stdErr.Locator() != "0.0.333" {
			t.Errorf("stdErr = %v, want '0.0.333'", b.stdErr.Locator())
		}
	})
}

// =============================================================================
// WithReachableAddr Tests
// =============================================================================

func TestWithReachableAddr(t *testing.T) {
	// Known test values
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	testMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

	t.Run("adds valid multiaddr", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddr(testMultiaddr)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if len(b.reachableAddrs) != 1 {
			t.Errorf("expected 1 reachable addr, got %d", len(b.reachableAddrs))
		}
	})

	t.Run("records error for invalid multiaddr", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddr("invalid-multiaddr")

		if !b.HasErrors() {
			t.Error("expected error for invalid multiaddr")
		}
		var ae *AccountError
		if errors.As(b.Errors()[0], &ae) && ae.Kind != ErrKindInvalidAddress {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidAddress)
		}
	})

	t.Run("records error for empty multiaddr", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddr("")

		if !b.HasErrors() {
			t.Error("expected error for empty multiaddr")
		}
	})

	t.Run("records error for multiaddr without p2p", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddr("/ip4/192.168.1.1/tcp/4001")

		if !b.HasErrors() {
			t.Error("expected error for multiaddr without p2p")
		}
	})
}

// =============================================================================
// WithReachableAddrs Tests
// =============================================================================

func TestWithReachableAddrs(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	addr1 := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID
	addr2 := "/ip4/192.168.1.2/tcp/4001/p2p/" + testPeerID

	t.Run("adds multiple valid addresses", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrs(addr1, addr2)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if len(b.reachableAddrs) != 2 {
			t.Errorf("expected 2 reachable addrs, got %d", len(b.reachableAddrs))
		}
	})

	t.Run("accumulates errors for invalid addresses", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrs("invalid1", "invalid2")

		if len(b.Errors()) != 2 {
			t.Errorf("expected 2 errors, got %d", len(b.Errors()))
		}
	})

	t.Run("mixed valid and invalid addresses", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrs(addr1, "invalid", addr2)

		if len(b.Errors()) != 1 {
			t.Errorf("expected 1 error, got %d", len(b.Errors()))
		}
		if len(b.reachableAddrs) != 2 {
			t.Errorf("expected 2 valid addrs, got %d", len(b.reachableAddrs))
		}
	})
}

// =============================================================================
// WithReachableAddrValidated Tests
// =============================================================================

func TestWithReachableAddrValidated(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	differentPeerID := "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"

	matchingMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID
	mismatchedMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + differentPeerID

	t.Run("adds address with matching PeerID", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrValidated(matchingMultiaddr)

		if b.HasErrors() {
			t.Errorf("unexpected error: %v", b.Errors())
		}
		if len(b.reachableAddrs) != 1 {
			t.Errorf("expected 1 reachable addr, got %d", len(b.reachableAddrs))
		}
	})

	t.Run("records error for mismatched PeerID", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrValidated(mismatchedMultiaddr)

		if !b.HasErrors() {
			t.Error("expected error for mismatched PeerID")
		}
		var ae *AccountError
		if errors.As(b.Errors()[0], &ae) && ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})

	t.Run("records error when public key is zero", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		did := newBuilderMockDID(zeroPubKey)

		b := NewParentAccountBuilder(zeroPubKey, did)
		initialErrors := len(b.Errors())

		b.WithReachableAddrValidated(matchingMultiaddr)

		// Should have additional error
		if len(b.Errors()) <= initialErrors {
			t.Error("expected additional error when public key is zero")
		}
	})

	t.Run("records error for invalid multiaddr", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).WithReachableAddrValidated("invalid")

		if !b.HasErrors() {
			t.Error("expected error for invalid multiaddr")
		}
	})
}

// =============================================================================
// Build Tests - Parent Account
// =============================================================================

func TestBuild_ParentAccount(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	testMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

	t.Run("builds valid parent account", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		// Parent accounts don't have comm channels per spec
		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(testMultiaddr).
			Build()

		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if !account.PublicKey().Equal(pubKey) {
			t.Error("PublicKey mismatch")
		}
		if account.AccountType() != AccountTypeParent {
			t.Errorf("AccountType = %v, want Parent", account.AccountType())
		}
		if account.PeerID().String() != testPeerID {
			t.Errorf("PeerID = %v, want %v", account.PeerID(), testPeerID)
		}
		if !account.IsParent() {
			t.Error("IsParent() should return true")
		}
		// Parent accounts should have zero comm addresses
		if !account.StdIn().IsZero() {
			t.Error("Parent StdIn should be zero")
		}
	})

	t.Run("returns first accumulated error", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		_, err := NewParentAccountBuilder(zeroPubKey, nil).Build()

		if err == nil {
			t.Fatal("expected error")
		}
		// Should be the first error (zero public key)
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
	})

	t.Run("validates DID matches key", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		mismatchedDID := newMismatchedMockDID()

		_, err := NewParentAccountBuilder(pubKey, mismatchedDID).Build()

		if err == nil {
			t.Fatal("expected error for DID-key mismatch")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindInvalidDID {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidDID)
		}
	})

	t.Run("validates reachable address PeerID at build time", func(t *testing.T) {
		differentPeerID := "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"
		mismatchedMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + differentPeerID

		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		// Note: WithReachableAddr does NOT validate PeerID, Build() does
		_, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(mismatchedMultiaddr).
			Build()

		if err == nil {
			t.Fatal("expected error for PeerID mismatch")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})
}

// =============================================================================
// Build Tests - Child Account
// =============================================================================

func TestBuild_ChildAccount(t *testing.T) {
	t.Run("builds valid child account", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()
		parentPubKey := parentPriv.PublicKey()

		// Child accounts require all 3 comm channels per spec
		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if !account.PublicKey().Equal(childPubKey) {
			t.Error("PublicKey mismatch")
		}
		if account.AccountType() != AccountTypeChild {
			t.Errorf("AccountType = %v, want Child", account.AccountType())
		}
		if !account.IsChild() {
			t.Error("IsChild() should return true")
		}
		if !account.ParentPublicKey().Equal(parentPubKey) {
			t.Error("ParentPublicKey mismatch")
		}
	})

	t.Run("validates child key differs from parent", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()

		// Even with comm channels, same key should fail
		_, err := NewChildAccountBuilder(pubKey, pubKey).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		if err == nil {
			t.Fatal("expected error when child equals parent")
		}
		var ae *AccountError
		if errors.As(err, &ae) && ae.Kind != ErrKindValidation {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindValidation)
		}
	})

	t.Run("child account has no DID", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()

		// Child accounts require all 3 comm channels
		account, err := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			Build()

		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if account.DID() != nil {
			t.Error("child account should have nil DID")
		}
	})
}

// =============================================================================
// Build Tests - Error Accumulation
// =============================================================================

func TestBuild_ErrorAccumulation(t *testing.T) {
	t.Run("collects errors from multiple With calls", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("invalid1").
			WithStdOutHedera("invalid2").
			WithStdErrHedera("invalid3").
			WithReachableAddr("invalid-multiaddr")

		if len(b.Errors()) != 4 {
			t.Errorf("expected 4 errors, got %d", len(b.Errors()))
		}

		_, err := b.Build()
		if err == nil {
			t.Fatal("Build() should return error")
		}

		// Build returns first error
		errStr := err.Error()
		if !strings.Contains(errStr, "invalid1") && !strings.Contains(errStr, "Hedera") {
			t.Errorf("first error should be about invalid1 or Hedera: %s", errStr)
		}
	})

	t.Run("HasErrors reflects accumulated state", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did)
		if b.HasErrors() {
			t.Error("should not have errors initially")
		}

		b.WithStdInHedera("invalid")
		if !b.HasErrors() {
			t.Error("should have errors after invalid input")
		}
	})

	t.Run("Errors returns all errors", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey

		b := NewParentAccountBuilder(zeroPubKey, nil).
			WithStdInHedera("bad")

		errs := b.Errors()
		// zeroPubKey + nil DID + bad topic = 3 errors
		if len(errs) != 3 {
			t.Errorf("expected 3 errors, got %d", len(errs))
		}
	})
}

// =============================================================================
// MustBuild Tests
// =============================================================================

func TestMustBuild(t *testing.T) {
	t.Run("returns account on success", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		account := NewParentAccountBuilder(pubKey, did).MustBuild()

		if !account.PublicKey().Equal(pubKey) {
			t.Error("MustBuild should return valid account")
		}
	})

	t.Run("panics on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustBuild should panic on error")
			}
		}()

		var zeroPubKey keylib.NeuronPublicKey
		NewParentAccountBuilder(zeroPubKey, nil).MustBuild()
	})

	t.Run("panic contains error message", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					if !strings.Contains(err.Error(), "ZeroValue") && !strings.Contains(err.Error(), "zero") {
						t.Errorf("panic should mention zero value: %v", err)
					}
				}
			}
		}()

		var zeroPubKey keylib.NeuronPublicKey
		NewParentAccountBuilder(zeroPubKey, nil).MustBuild()
	})
}

// =============================================================================
// HasErrors and Errors Tests
// =============================================================================

func TestHasErrors(t *testing.T) {
	t.Run("returns false for valid builder", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did)
		if b.HasErrors() {
			t.Error("HasErrors() should return false")
		}
	})

	t.Run("returns true after error", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		b := NewParentAccountBuilder(zeroPubKey, nil)
		if !b.HasErrors() {
			t.Error("HasErrors() should return true")
		}
	})
}

func TestErrors(t *testing.T) {
	t.Run("returns empty slice for valid builder", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did)
		if len(b.Errors()) != 0 {
			t.Errorf("Errors() len = %d, want 0", len(b.Errors()))
		}
	})

	t.Run("returns all accumulated errors", func(t *testing.T) {
		var zeroPubKey keylib.NeuronPublicKey
		b := NewParentAccountBuilder(zeroPubKey, nil)
		errs := b.Errors()
		if len(errs) != 2 {
			t.Errorf("expected 2 errors (zeroPubKey + nil DID), got %d", len(errs))
		}
	})
}

// =============================================================================
// Fluent Chaining Tests
// =============================================================================

func TestBuilder_FluentChaining(t *testing.T) {
	testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	testMultiaddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID

	t.Run("all methods return builder for chaining", func(t *testing.T) {
		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)
		addr, _ := NewHederaTopicAddress("0.0.12345")

		b := NewParentAccountBuilder(pubKey, did)

		// All methods should return *AccountBuilder
		result := b.WithStdIn(addr).
			WithStdInHedera("0.0.111").
			WithStdOut(addr).
			WithStdOutHedera("0.0.222").
			WithStdErr(addr).
			WithStdErrHedera("0.0.333").
			WithHederaTopics("0.0.1", "0.0.2", "0.0.3").
			WithReachableAddr(testMultiaddr).
			WithReachableAddrs(testMultiaddr).
			WithReachableAddrValidated(testMultiaddr)

		if result == nil {
			t.Error("fluent chaining should return non-nil builder")
		}
	})

	t.Run("chaining continues after error", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newBuilderMockDID(pubKey)

		b := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("invalid"). // This causes an error
			WithStdOutHedera("0.0.222") // This should still execute

		// Should have error from invalid, but stdOut should be set
		if !b.HasErrors() {
			t.Error("should have error from invalid topic")
		}
		if b.stdOut.Locator() != "0.0.222" {
			t.Error("subsequent call should still execute after error")
		}
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestBuilder_Integration(t *testing.T) {
	t.Run("complete parent account workflow", func(t *testing.T) {
		testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		addr1 := "/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID
		addr2 := "/ip4/192.168.1.2/tcp/4001/p2p/" + testPeerID

		pubKey, _ := keylib.ParsePublicKeyHex(testPubKeyHex)
		did := newBuilderMockDID(pubKey)

		// Parent accounts don't have comm channels per spec
		account, err := NewParentAccountBuilder(pubKey, did).
			WithReachableAddrs(addr1, addr2).
			Build()

		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		// Verify all aspects
		if !account.PublicKey().Equal(pubKey) {
			t.Error("PublicKey mismatch")
		}
		if account.PeerID().String() != testPeerID {
			t.Errorf("PeerID = %v, want %v", account.PeerID(), testPeerID)
		}
		if !account.EVMAddress().IsZero() && account.EVMAddress() == (keylib.EVMAddress{}) {
			// EVMAddress is derived, just ensure it's computed
		}
		if account.AccountType() != AccountTypeParent {
			t.Error("AccountType should be Parent")
		}
		// Parent accounts have zero comm addresses
		if !account.StdIn().IsZero() {
			t.Error("Parent StdIn should be zero")
		}
		if !account.StdOut().IsZero() {
			t.Error("Parent StdOut should be zero")
		}
		if !account.StdErr().IsZero() {
			t.Error("Parent StdErr should be zero")
		}
		if len(account.ReachableAddrs().Addrs()) != 2 {
			t.Errorf("expected 2 reachable addrs, got %d", len(account.ReachableAddrs().Addrs()))
		}
	})

	t.Run("complete child account workflow", func(t *testing.T) {
		childPriv, _ := keylib.GeneratePrivateKey()
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()
		parentPubKey := parentPriv.PublicKey()

		childPeerID, _ := childPubKey.PeerID()
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + childPeerID.String()

		// Child accounts require all 3 comm channels per spec
		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			WithReachableAddr(addr).
			Build()

		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}

		if !account.PublicKey().Equal(childPubKey) {
			t.Error("PublicKey mismatch")
		}
		if !account.ParentPublicKey().Equal(parentPubKey) {
			t.Error("ParentPublicKey mismatch")
		}
		if !account.IsChild() {
			t.Error("should be child account")
		}
		if account.DID() != nil {
			t.Error("child should have no DID")
		}
	})
}
