package account

import (
	"errors"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
	"github.com/multiformats/go-multiaddr"
)

// =============================================================================
// Test Vectors - Known Deterministic Values
// =============================================================================
//
// These vectors are derived from a known secp256k1 public key.
// They ensure that our implementations are deterministic and cross-compatible.

var (
	// Test public key (compressed secp256k1)
	testReachablePubKeyHex = "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"

	// Expected PeerID derived from the test public key
	testReachableExpectedPeerID = "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	// Relay PeerID (different from target)
	testRelayPeerID = "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"
)

// getTestPeerID returns a PeerID from the test public key for testing.
func getTestPeerID(t *testing.T) keylib.PeerID {
	t.Helper()
	pubKey, err := keylib.ParsePublicKeyHex(testReachablePubKeyHex)
	if err != nil {
		t.Fatalf("failed to parse test public key: %v", err)
	}
	peerID, err := pubKey.PeerID()
	if err != nil {
		t.Fatalf("failed to derive PeerID: %v", err)
	}
	return peerID
}

// =============================================================================
// ParseReachableAddr Tests - Valid Formats
// =============================================================================

func TestParseReachableAddr_ValidFormats(t *testing.T) {
	expectedPeerID := getTestPeerID(t)
	peerIDStr := expectedPeerID.String()

	validCases := []struct {
		name string
		addr string
	}{
		{
			name: "ip4/tcp/p2p",
			addr: "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "ip6/tcp/p2p",
			addr: "/ip6/::1/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "dns/tcp/p2p",
			addr: "/dns/example.com/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "dns4/tcp/p2p",
			addr: "/dns4/example.com/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "dns6/tcp/p2p",
			addr: "/dns6/example.com/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "ip4/udp/quic/p2p",
			addr: "/ip4/192.168.1.1/udp/4001/quic/p2p/" + peerIDStr,
		},
		{
			name: "ip4/udp/quic-v1/p2p",
			addr: "/ip4/192.168.1.1/udp/4001/quic-v1/p2p/" + peerIDStr,
		},
		{
			name: "dnsaddr/p2p",
			addr: "/dnsaddr/bootstrap.libp2p.io/p2p/" + peerIDStr,
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := ParseReachableAddr(tc.addr)
			if err != nil {
				t.Fatalf("ParseReachableAddr() error = %v", err)
			}
			if addr.IsZero() {
				t.Error("parsed address should not be zero")
			}
			if !addr.PeerID().Equal(expectedPeerID) {
				t.Errorf("PeerID = %v, want %v", addr.PeerID(), expectedPeerID)
			}
		})
	}
}

// =============================================================================
// ParseReachableAddr Tests - Invalid Formats
// =============================================================================

func TestParseReachableAddr_InvalidFormats(t *testing.T) {
	invalidCases := []struct {
		name        string
		addr        string
		errContains string
	}{
		{
			name:        "empty string",
			addr:        "",
			errContains: "empty",
		},
		{
			name:        "invalid multiaddr - missing slash prefix",
			addr:        "ip4/192.168.1.1/tcp/4001",
			errContains: "invalid multiaddr",
		},
		{
			name:        "missing p2p component",
			addr:        "/ip4/192.168.1.1/tcp/4001",
			errContains: "missing /p2p/",
		},
		{
			name:        "invalid peer ID",
			addr:        "/ip4/192.168.1.1/tcp/4001/p2p/invalid",
			errContains: "invalid multiaddr",
		},
		{
			name:        "empty peer ID",
			addr:        "/ip4/192.168.1.1/tcp/4001/p2p/",
			errContains: "invalid multiaddr",
		},
		{
			name:        "garbage input",
			addr:        "not-a-multiaddr",
			errContains: "invalid multiaddr",
		},
		{
			name:        "http URL",
			addr:        "http://example.com",
			errContains: "invalid multiaddr",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseReachableAddr(tc.addr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var ae *AccountError
			if !errors.As(err, &ae) {
				t.Error("error should be AccountError")
			}
		})
	}
}

// =============================================================================
// MustParseReachableAddr Tests
// =============================================================================

func TestMustParseReachableAddr(t *testing.T) {
	peerIDStr := getTestPeerID(t).String()

	t.Run("valid address does not panic", func(t *testing.T) {
		validAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr
		addr := MustParseReachableAddr(validAddr)
		if addr.IsZero() {
			t.Error("expected non-zero address")
		}
	})

	t.Run("invalid address panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic, got none")
			}
		}()
		MustParseReachableAddr("invalid")
	})
}

// =============================================================================
// NewReachableAddrWithValidation Tests - PeerID Consistency (SPEC 4.2)
// =============================================================================

func TestNewReachableAddrWithValidation_PeerIDConsistency(t *testing.T) {
	expectedPeerID := getTestPeerID(t)
	peerIDStr := expectedPeerID.String()
	validAddrStr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr

	t.Run("matching PeerID passes", func(t *testing.T) {
		ma, _ := multiaddr.NewMultiaddr(validAddrStr)
		addr, err := NewReachableAddrWithValidation(ma, expectedPeerID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !addr.PeerID().Equal(expectedPeerID) {
			t.Errorf("PeerID = %v, want %v", addr.PeerID(), expectedPeerID)
		}
	})

	t.Run("mismatched PeerID fails with ErrKindPeerIDMismatch", func(t *testing.T) {
		// Use a different peer ID in the address
		otherAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID
		ma, _ := multiaddr.NewMultiaddr(otherAddr)

		_, err := NewReachableAddrWithValidation(ma, expectedPeerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Fatal("error should be AccountError")
		}
		if ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})

	t.Run("nil multiaddr fails", func(t *testing.T) {
		_, err := NewReachableAddrWithValidation(nil, expectedPeerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Fatal("error should be AccountError")
		}
	})

	t.Run("zero expected PeerID fails", func(t *testing.T) {
		ma, _ := multiaddr.NewMultiaddr(validAddrStr)
		_, err := NewReachableAddrWithValidation(ma, keylib.PeerID{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Fatal("error should be AccountError")
		}
	})
}

func TestNewReachableAddrFromStringWithValidation(t *testing.T) {
	expectedPeerID := getTestPeerID(t)
	peerIDStr := expectedPeerID.String()

	t.Run("valid string with matching PeerID passes", func(t *testing.T) {
		validAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr
		addr, err := NewReachableAddrFromStringWithValidation(validAddr, expectedPeerID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !addr.PeerID().Equal(expectedPeerID) {
			t.Errorf("PeerID mismatch")
		}
	})

	t.Run("invalid multiaddr string fails", func(t *testing.T) {
		_, err := NewReachableAddrFromStringWithValidation("invalid", expectedPeerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("mismatched PeerID fails", func(t *testing.T) {
		otherAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID
		_, err := NewReachableAddrFromStringWithValidation(otherAddr, expectedPeerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// =============================================================================
// Circuit Relay Tests - PeerID Extraction (SPEC 4.5)
// =============================================================================

func TestCircuitRelay_PeerIDExtraction(t *testing.T) {
	targetPeerID := getTestPeerID(t)
	targetPeerIDStr := targetPeerID.String()

	t.Run("extracts TARGET PeerID from circuit relay address", func(t *testing.T) {
		// Format: /ip4/.../p2p/RELAY/p2p-circuit/p2p/TARGET
		// The TARGET is what should be extracted, NOT the RELAY
		circuitAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr

		addr, err := ParseReachableAddr(circuitAddr)
		if err != nil {
			t.Fatalf("ParseReachableAddr() error = %v", err)
		}

		// The extracted PeerID should be the TARGET, not the RELAY
		if !addr.PeerID().Equal(targetPeerID) {
			t.Errorf("extracted PeerID = %v, want TARGET %v", addr.PeerID(), targetPeerID)
		}

		// Verify it's identified as a circuit relay
		if !addr.IsCircuitRelay() {
			t.Error("expected IsCircuitRelay() = true")
		}
	})

	t.Run("circuit relay with validation passes for TARGET PeerID", func(t *testing.T) {
		circuitAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr

		addr, err := NewReachableAddrFromStringWithValidation(circuitAddr, targetPeerID)
		if err != nil {
			t.Fatalf("expected no error for TARGET validation, got: %v", err)
		}
		if !addr.PeerID().Equal(targetPeerID) {
			t.Error("PeerID should match TARGET")
		}
	})

	t.Run("circuit relay with validation fails for RELAY PeerID", func(t *testing.T) {
		circuitAddr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr

		// Parse the relay peer ID directly
		relayPeerID, err := keylib.ParsePeerID(testRelayPeerID)
		if err != nil {
			t.Fatalf("failed to parse relay peer ID: %v", err)
		}
		// Validate against relay PeerID should fail (we want TARGET, not RELAY)
		_, err = NewReachableAddrFromStringWithValidation(circuitAddr, relayPeerID)
		if err == nil {
			t.Fatal("expected error when validating against RELAY PeerID, got nil")
		}
	})

	t.Run("circuit relay IPv6", func(t *testing.T) {
		circuitAddr := "/ip6/::1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr

		addr, err := ParseReachableAddr(circuitAddr)
		if err != nil {
			t.Fatalf("ParseReachableAddr() error = %v", err)
		}
		if !addr.PeerID().Equal(targetPeerID) {
			t.Error("should extract TARGET PeerID from IPv6 circuit relay")
		}
	})

	t.Run("circuit relay with dns", func(t *testing.T) {
		circuitAddr := "/dns/relay.example.com/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr

		addr, err := ParseReachableAddr(circuitAddr)
		if err != nil {
			t.Fatalf("ParseReachableAddr() error = %v", err)
		}
		if !addr.PeerID().Equal(targetPeerID) {
			t.Error("should extract TARGET PeerID from DNS circuit relay")
		}
	})
}

// =============================================================================
// ReachableAddr Methods Tests
// =============================================================================

func TestReachableAddr_Methods(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()
	addrStr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr

	addr, err := ParseReachableAddr(addrStr)
	if err != nil {
		t.Fatalf("ParseReachableAddr() error = %v", err)
	}

	t.Run("Multiaddr returns non-nil", func(t *testing.T) {
		if addr.Multiaddr() == nil {
			t.Error("Multiaddr() should not return nil")
		}
	})

	t.Run("PeerID returns correct value", func(t *testing.T) {
		if !addr.PeerID().Equal(peerID) {
			t.Errorf("PeerID() = %v, want %v", addr.PeerID(), peerID)
		}
	})

	t.Run("String returns original address", func(t *testing.T) {
		// String may be canonicalized, so just check it's non-empty and parseable
		s := addr.String()
		if s == "" {
			t.Error("String() should not be empty")
		}
		// Verify round-trip
		reparsed, err := ParseReachableAddr(s)
		if err != nil {
			t.Errorf("String() produced unparseable result: %v", err)
		}
		if !reparsed.Equal(addr) {
			t.Error("String() round-trip failed")
		}
	})

	t.Run("IsZero returns false for valid address", func(t *testing.T) {
		if addr.IsZero() {
			t.Error("IsZero() should return false for valid address")
		}
	})

	t.Run("IsCircuitRelay returns false for direct address", func(t *testing.T) {
		if addr.IsCircuitRelay() {
			t.Error("IsCircuitRelay() should return false for direct address")
		}
	})
}

func TestReachableAddr_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var addr ReachableAddr
		if !addr.IsZero() {
			t.Error("zero value should be zero")
		}
	})

	t.Run("zero value String returns empty", func(t *testing.T) {
		var addr ReachableAddr
		if addr.String() != "" {
			t.Errorf("zero value String() = %v, want empty", addr.String())
		}
	})

	t.Run("zero value IsCircuitRelay returns false", func(t *testing.T) {
		var addr ReachableAddr
		if addr.IsCircuitRelay() {
			t.Error("zero value IsCircuitRelay() should return false")
		}
	})
}

func TestReachableAddr_Equal(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
	addr2, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
	addr3, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + peerIDStr)

	t.Run("same address equals itself", func(t *testing.T) {
		if !addr1.Equal(addr1) {
			t.Error("address should equal itself")
		}
	})

	t.Run("equivalent addresses are equal", func(t *testing.T) {
		if !addr1.Equal(addr2) {
			t.Error("equivalent addresses should be equal")
		}
	})

	t.Run("different addresses are not equal", func(t *testing.T) {
		if addr1.Equal(addr3) {
			t.Error("different addresses should not be equal")
		}
	})

	t.Run("zero values are equal", func(t *testing.T) {
		var zero1, zero2 ReachableAddr
		if !zero1.Equal(zero2) {
			t.Error("zero values should be equal")
		}
	})

	t.Run("zero and non-zero are not equal", func(t *testing.T) {
		var zero ReachableAddr
		if addr1.Equal(zero) {
			t.Error("zero and non-zero should not be equal")
		}
	})
}

func TestReachableAddr_Validate(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	t.Run("valid address passes validation", func(t *testing.T) {
		addr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		if err := addr.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("zero address fails validation", func(t *testing.T) {
		var addr ReachableAddr
		err := addr.Validate()
		if err == nil {
			t.Error("zero address should fail validation")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
		if ae.Kind != ErrKindZeroValue {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindZeroValue)
		}
	})
}

func TestReachableAddr_ValidateForAccount(t *testing.T) {
	expectedPeerID := getTestPeerID(t)
	peerIDStr := expectedPeerID.String()

	t.Run("matching PeerID passes", func(t *testing.T) {
		addr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		if err := addr.ValidateForAccount(expectedPeerID); err != nil {
			t.Errorf("ValidateForAccount() error = %v", err)
		}
	})

	t.Run("mismatched PeerID fails", func(t *testing.T) {
		addr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testRelayPeerID)
		err := addr.ValidateForAccount(expectedPeerID)
		if err == nil {
			t.Fatal("expected error for mismatched PeerID")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Fatal("error should be AccountError")
		}
		if ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})

	t.Run("zero expected PeerID fails", func(t *testing.T) {
		addr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		err := addr.ValidateForAccount(keylib.PeerID{})
		if err == nil {
			t.Fatal("expected error for zero expected PeerID")
		}
	})

	t.Run("zero address fails", func(t *testing.T) {
		var addr ReachableAddr
		err := addr.ValidateForAccount(expectedPeerID)
		if err == nil {
			t.Fatal("expected error for zero address")
		}
	})
}

// =============================================================================
// ReachableAddrs Collection Tests
// =============================================================================

func TestNewReachableAddrs(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
	addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + peerIDStr)

	t.Run("creates collection from addresses", func(t *testing.T) {
		addrs := NewReachableAddrs(addr1, addr2)
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2", addrs.Len())
		}
	})

	t.Run("filters out zero values", func(t *testing.T) {
		var zero ReachableAddr
		addrs := NewReachableAddrs(addr1, zero, addr2)
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2 (zero value should be filtered)", addrs.Len())
		}
	})

	t.Run("empty input creates empty collection", func(t *testing.T) {
		addrs := NewReachableAddrs()
		if !addrs.IsEmpty() {
			t.Error("expected empty collection")
		}
	})
}

func TestParseReachableAddrs(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	t.Run("parses multiple valid addresses", func(t *testing.T) {
		addrs, err := ParseReachableAddrs(
			"/ip4/192.168.1.1/tcp/4001/p2p/"+peerIDStr,
			"/ip4/192.168.1.2/tcp/4001/p2p/"+peerIDStr,
		)
		if err != nil {
			t.Fatalf("ParseReachableAddrs() error = %v", err)
		}
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2", addrs.Len())
		}
	})

	t.Run("skips empty strings", func(t *testing.T) {
		addrs, err := ParseReachableAddrs(
			"/ip4/192.168.1.1/tcp/4001/p2p/"+peerIDStr,
			"",
			"/ip4/192.168.1.2/tcp/4001/p2p/"+peerIDStr,
		)
		if err != nil {
			t.Fatalf("ParseReachableAddrs() error = %v", err)
		}
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2", addrs.Len())
		}
	})

	t.Run("returns error on first invalid address", func(t *testing.T) {
		_, err := ParseReachableAddrs(
			"/ip4/192.168.1.1/tcp/4001/p2p/"+peerIDStr,
			"invalid",
		)
		if err == nil {
			t.Fatal("expected error for invalid address")
		}
	})
}

func TestNewReachableAddrsWithValidation(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	t.Run("passes when all addresses match expected PeerID", func(t *testing.T) {
		strs := []string{
			"/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr,
			"/ip4/192.168.1.2/tcp/4001/p2p/" + peerIDStr,
		}
		addrs, err := NewReachableAddrsWithValidation(strs, peerID)
		if err != nil {
			t.Fatalf("NewReachableAddrsWithValidation() error = %v", err)
		}
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2", addrs.Len())
		}
	})

	t.Run("fails when any address has mismatched PeerID", func(t *testing.T) {
		strs := []string{
			"/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr,
			"/ip4/192.168.1.2/tcp/4001/p2p/" + testRelayPeerID, // Different PeerID
		}
		_, err := NewReachableAddrsWithValidation(strs, peerID)
		if err == nil {
			t.Fatal("expected error for mismatched PeerID")
		}
	})

	t.Run("skips empty strings", func(t *testing.T) {
		strs := []string{
			"/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr,
			"",
		}
		addrs, err := NewReachableAddrsWithValidation(strs, peerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addrs.Len() != 1 {
			t.Errorf("Len() = %d, want 1", addrs.Len())
		}
	})
}

func TestReachableAddrs_Methods(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
	addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + peerIDStr)
	addrs := NewReachableAddrs(addr1, addr2)

	t.Run("Addrs returns copy of addresses", func(t *testing.T) {
		result := addrs.Addrs()
		if len(result) != 2 {
			t.Errorf("Addrs() len = %d, want 2", len(result))
		}
		// Verify it's a copy
		result[0] = ReachableAddr{}
		if addrs.Addrs()[0].IsZero() {
			t.Error("Addrs() should return a copy, not original slice")
		}
	})

	t.Run("Strings returns string representations", func(t *testing.T) {
		strs := addrs.Strings()
		if len(strs) != 2 {
			t.Errorf("Strings() len = %d, want 2", len(strs))
		}
		for _, s := range strs {
			if s == "" {
				t.Error("Strings() should not contain empty strings")
			}
		}
	})

	t.Run("Multiaddrs returns multiaddr objects", func(t *testing.T) {
		mas := addrs.Multiaddrs()
		if len(mas) != 2 {
			t.Errorf("Multiaddrs() len = %d, want 2", len(mas))
		}
		for _, ma := range mas {
			if ma == nil {
				t.Error("Multiaddrs() should not contain nil")
			}
		}
	})

	t.Run("IsEmpty returns false for non-empty collection", func(t *testing.T) {
		if addrs.IsEmpty() {
			t.Error("IsEmpty() should return false")
		}
	})

	t.Run("Len returns correct count", func(t *testing.T) {
		if addrs.Len() != 2 {
			t.Errorf("Len() = %d, want 2", addrs.Len())
		}
	})

	t.Run("First returns first address", func(t *testing.T) {
		first := addrs.First()
		if !first.Equal(addr1) {
			t.Error("First() should return first address")
		}
	})

	t.Run("Contains finds existing address", func(t *testing.T) {
		if !addrs.Contains(addr1) {
			t.Error("Contains() should find existing address")
		}
	})

	t.Run("Contains returns false for non-existing address", func(t *testing.T) {
		other, _ := ParseReachableAddr("/ip4/10.0.0.1/tcp/4001/p2p/" + peerIDStr)
		if addrs.Contains(other) {
			t.Error("Contains() should return false for non-existing address")
		}
	})
}

func TestReachableAddrs_EmptyCollection(t *testing.T) {
	var addrs ReachableAddrs

	t.Run("IsEmpty returns true", func(t *testing.T) {
		if !addrs.IsEmpty() {
			t.Error("empty collection IsEmpty() should return true")
		}
	})

	t.Run("Len returns 0", func(t *testing.T) {
		if addrs.Len() != 0 {
			t.Errorf("Len() = %d, want 0", addrs.Len())
		}
	})

	t.Run("First returns zero value", func(t *testing.T) {
		if !addrs.First().IsZero() {
			t.Error("First() on empty should return zero value")
		}
	})

	t.Run("Addrs returns nil", func(t *testing.T) {
		if addrs.Addrs() != nil {
			t.Error("Addrs() on empty should return nil")
		}
	})

	t.Run("Strings returns nil", func(t *testing.T) {
		if addrs.Strings() != nil {
			t.Error("Strings() on empty should return nil")
		}
	})

	t.Run("Multiaddrs returns nil", func(t *testing.T) {
		if addrs.Multiaddrs() != nil {
			t.Error("Multiaddrs() on empty should return nil")
		}
	})
}

func TestReachableAddrs_ValidateAll(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	t.Run("passes when all addresses match expected PeerID", func(t *testing.T) {
		addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + peerIDStr)
		addrs := NewReachableAddrs(addr1, addr2)

		if err := addrs.ValidateAll(peerID); err != nil {
			t.Errorf("ValidateAll() error = %v", err)
		}
	})

	t.Run("fails when any address has mismatched PeerID", func(t *testing.T) {
		addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + testRelayPeerID)
		addrs := NewReachableAddrs(addr1, addr2)

		err := addrs.ValidateAll(peerID)
		if err == nil {
			t.Fatal("expected error for mismatched PeerID")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Fatal("error should be AccountError")
		}
		if ae.Kind != ErrKindPeerIDMismatch {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindPeerIDMismatch)
		}
	})

	t.Run("fails when expected PeerID is zero", func(t *testing.T) {
		addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr)
		addrs := NewReachableAddrs(addr1)

		err := addrs.ValidateAll(keylib.PeerID{})
		if err == nil {
			t.Fatal("expected error for zero PeerID")
		}
	})

	t.Run("empty collection passes", func(t *testing.T) {
		var addrs ReachableAddrs
		if err := addrs.ValidateAll(peerID); err != nil {
			t.Errorf("empty collection ValidateAll() should pass, got: %v", err)
		}
	})
}

// =============================================================================
// DirectAddrs and RelayAddrs Tests
// =============================================================================

func TestReachableAddrs_DirectAndRelayAddrs(t *testing.T) {
	targetPeerID := getTestPeerID(t)
	targetPeerIDStr := targetPeerID.String()

	directAddr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + targetPeerIDStr)
	directAddr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + targetPeerIDStr)
	relayAddr1, _ := ParseReachableAddr("/ip4/10.0.0.1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr)
	relayAddr2, _ := ParseReachableAddr("/ip4/10.0.0.2/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + targetPeerIDStr)

	addrs := NewReachableAddrs(directAddr1, relayAddr1, directAddr2, relayAddr2)

	t.Run("DirectAddrs returns only non-relay addresses", func(t *testing.T) {
		direct := addrs.DirectAddrs()
		if len(direct) != 2 {
			t.Errorf("DirectAddrs() len = %d, want 2", len(direct))
		}
		for _, addr := range direct {
			if addr.IsCircuitRelay() {
				t.Error("DirectAddrs() should not contain relay addresses")
			}
		}
	})

	t.Run("RelayAddrs returns only relay addresses", func(t *testing.T) {
		relay := addrs.RelayAddrs()
		if len(relay) != 2 {
			t.Errorf("RelayAddrs() len = %d, want 2", len(relay))
		}
		for _, addr := range relay {
			if !addr.IsCircuitRelay() {
				t.Error("RelayAddrs() should only contain relay addresses")
			}
		}
	})

	t.Run("DirectAddrs on only-relay collection returns empty", func(t *testing.T) {
		relayOnly := NewReachableAddrs(relayAddr1, relayAddr2)
		direct := relayOnly.DirectAddrs()
		if len(direct) != 0 {
			t.Errorf("DirectAddrs() on relay-only should be empty, got %d", len(direct))
		}
	})

	t.Run("RelayAddrs on only-direct collection returns empty", func(t *testing.T) {
		directOnly := NewReachableAddrs(directAddr1, directAddr2)
		relay := directOnly.RelayAddrs()
		if len(relay) != 0 {
			t.Errorf("RelayAddrs() on direct-only should be empty, got %d", len(relay))
		}
	})
}

// =============================================================================
// Round-Trip Tests
// =============================================================================

func TestReachableAddr_RoundTrip(t *testing.T) {
	peerID := getTestPeerID(t)
	peerIDStr := peerID.String()

	testCases := []struct {
		name string
		addr string
	}{
		{
			name: "direct address",
			addr: "/ip4/192.168.1.1/tcp/4001/p2p/" + peerIDStr,
		},
		{
			name: "circuit relay address",
			addr: "/ip4/10.0.0.1/tcp/4001/p2p/" + testRelayPeerID + "/p2p-circuit/p2p/" + peerIDStr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse
			original, err := ParseReachableAddr(tc.addr)
			if err != nil {
				t.Fatalf("ParseReachableAddr() error = %v", err)
			}

			// Convert to string
			str := original.String()

			// Parse again
			restored, err := ParseReachableAddr(str)
			if err != nil {
				t.Fatalf("second ParseReachableAddr() error = %v", err)
			}

			// Verify equality
			if !original.Equal(restored) {
				t.Errorf("round-trip failed: original = %v, restored = %v", original, restored)
			}

			// Verify PeerID preserved
			if !original.PeerID().Equal(restored.PeerID()) {
				t.Error("PeerID not preserved in round-trip")
			}
		})
	}
}

// =============================================================================
// Known Vector Tests
// =============================================================================

func TestReachableAddr_KnownVectors(t *testing.T) {
	t.Run("test public key derives to expected PeerID", func(t *testing.T) {
		pubKey, err := keylib.ParsePublicKeyHex(testReachablePubKeyHex)
		if err != nil {
			t.Fatalf("failed to parse public key: %v", err)
		}

		peerID, err := pubKey.PeerID()
		if err != nil {
			t.Fatalf("failed to derive PeerID: %v", err)
		}
		if peerID.String() != testReachableExpectedPeerID {
			t.Errorf("PeerID = %v, want %v", peerID.String(), testReachableExpectedPeerID)
		}
	})

	t.Run("address with expected PeerID parses correctly", func(t *testing.T) {
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + testReachableExpectedPeerID
		parsed, err := ParseReachableAddr(addr)
		if err != nil {
			t.Fatalf("ParseReachableAddr() error = %v", err)
		}

		if parsed.PeerID().String() != testReachableExpectedPeerID {
			t.Errorf("extracted PeerID = %v, want %v", parsed.PeerID().String(), testReachableExpectedPeerID)
		}
	})
}
