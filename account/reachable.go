package account

import (
	"fmt"
	"strings"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ReachableAddr represents a reachable connection address using libp2p multiaddr.
// It is the canonical "find / dial me" endpoint for a NeuronAccount.
//
// Key invariants:
//   - Must be a valid libp2p multiaddr
//   - Must contain exactly one /p2p/<PeerID>
//   - The PeerID must match the account's derived PeerID
//
// For circuit relay addresses, the final /p2p/<TargetPeerID> is validated.
type ReachableAddr struct {
	// ma is the parsed multiaddr
	ma multiaddr.Multiaddr

	// peerID is the extracted PeerID from the multiaddr
	peerID keylib.PeerID

	// raw is the original string representation
	raw string
}

// ParseReachableAddr parses a multiaddr string and extracts the PeerID.
// Returns an error if:
//   - The string is not a valid multiaddr
//   - The multiaddr doesn't contain /p2p/<PeerID>
func ParseReachableAddr(s string) (ReachableAddr, error) {
	const op = "ParseReachableAddr"

	if s == "" {
		return ReachableAddr{}, errInvalidAddress(op, "empty multiaddr string", nil)
	}

	ma, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		return ReachableAddr{}, errInvalidAddress(op,
			fmt.Sprintf("invalid multiaddr format: %s", s), err)
	}

	peerID, err := extractPeerIDFromMultiaddr(ma)
	if err != nil {
		return ReachableAddr{}, errInvalidAddress(op, err.Error(), nil)
	}

	return ReachableAddr{
		ma:     ma,
		peerID: peerID,
		raw:    s,
	}, nil
}

// MustParseReachableAddr parses a multiaddr string and panics on error.
// This is intended for use in tests and initialization of known-valid addresses.
func MustParseReachableAddr(s string) ReachableAddr {
	addr, err := ParseReachableAddr(s)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewReachableAddrWithValidation creates a ReachableAddr from a multiaddr,
// validating that the PeerID matches the expected value.
//
// This is the preferred constructor when creating addresses for a NeuronAccount,
// as it ensures the PeerID consistency rule is enforced.
func NewReachableAddrWithValidation(ma multiaddr.Multiaddr, expectedPeerID keylib.PeerID) (ReachableAddr, error) {
	const op = "NewReachableAddrWithValidation"

	if ma == nil {
		return ReachableAddr{}, errInvalidAddress(op, "multiaddr is nil", nil)
	}

	if expectedPeerID.IsZero() {
		return ReachableAddr{}, errInvalidAddress(op, "expected PeerID is zero-value", nil)
	}

	peerID, err := extractPeerIDFromMultiaddr(ma)
	if err != nil {
		return ReachableAddr{}, errInvalidAddress(op, err.Error(), nil)
	}

	if !peerID.Equal(expectedPeerID) {
		return ReachableAddr{}, errPeerIDMismatch(op, expectedPeerID.String(), peerID.String())
	}

	return ReachableAddr{
		ma:     ma,
		peerID: peerID,
		raw:    ma.String(),
	}, nil
}

// NewReachableAddrFromStringWithValidation parses a multiaddr string and validates the PeerID.
func NewReachableAddrFromStringWithValidation(s string, expectedPeerID keylib.PeerID) (ReachableAddr, error) {
	const op = "NewReachableAddrFromStringWithValidation"

	ma, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		return ReachableAddr{}, errInvalidAddress(op,
			fmt.Sprintf("invalid multiaddr format: %s", s), err)
	}

	return NewReachableAddrWithValidation(ma, expectedPeerID)
}

// Multiaddr returns the underlying libp2p multiaddr.
func (a ReachableAddr) Multiaddr() multiaddr.Multiaddr {
	return a.ma
}

// PeerID returns the PeerID extracted from this address.
func (a ReachableAddr) PeerID() keylib.PeerID {
	return a.peerID
}

// String returns the string representation of this address.
func (a ReachableAddr) String() string {
	if a.IsZero() {
		return ""
	}
	return a.raw
}

// IsZero returns true if this is a zero-value ReachableAddr.
func (a ReachableAddr) IsZero() bool {
	return a.ma == nil
}

// IsCircuitRelay returns true if this address is a circuit relay address.
// Circuit relay addresses contain /p2p-circuit/ in their path.
func (a ReachableAddr) IsCircuitRelay() bool {
	if a.ma == nil {
		return false
	}
	return strings.Contains(a.raw, "/p2p-circuit/")
}

// Equal compares this address with another for equality.
func (a ReachableAddr) Equal(other ReachableAddr) bool {
	if a.IsZero() && other.IsZero() {
		return true
	}
	if a.IsZero() || other.IsZero() {
		return false
	}
	return a.ma.Equal(other.ma)
}

// Validate checks if this address is well-formed.
// Note: This doesn't validate that the PeerID matches a specific account;
// use ValidateForAccount for that.
func (a ReachableAddr) Validate() error {
	const op = "ReachableAddr.Validate"

	if a.IsZero() {
		return errZeroValue(op, "ReachableAddr")
	}

	if a.peerID.IsZero() {
		return errInvalidAddress(op, "ReachableAddr has zero-value PeerID", nil)
	}

	return nil
}

// ValidateForAccount validates that this address is appropriate for the given PeerID.
// This enforces the PeerID consistency rule from the spec.
func (a ReachableAddr) ValidateForAccount(expectedPeerID keylib.PeerID) error {
	const op = "ReachableAddr.ValidateForAccount"

	if err := a.Validate(); err != nil {
		return err
	}

	if expectedPeerID.IsZero() {
		return errInvalidAddress(op, "expected PeerID is zero-value", nil)
	}

	if !a.peerID.Equal(expectedPeerID) {
		return errPeerIDMismatch(op, expectedPeerID.String(), a.peerID.String())
	}

	return nil
}

// extractPeerIDFromMultiaddr extracts the target PeerID from a multiaddr.
// For regular addresses: extracts from /p2p/<PeerID>
// For circuit relay: extracts the final /p2p/<TargetPeerID> after /p2p-circuit/
func extractPeerIDFromMultiaddr(ma multiaddr.Multiaddr) (keylib.PeerID, error) {
	// Check if this is a circuit relay address
	maStr := ma.String()
	if strings.Contains(maStr, "/p2p-circuit/") {
		return extractTargetPeerIDFromCircuitRelay(ma)
	}

	// Regular address: extract /p2p/<PeerID>
	return extractP2PComponent(ma)
}

// extractP2PComponent extracts the PeerID from a /p2p/<PeerID> component.
func extractP2PComponent(ma multiaddr.Multiaddr) (keylib.PeerID, error) {
	// Try to extract peer ID using the standard method
	addrInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		// Fallback: manually search for /p2p/ component
		return extractP2PManually(ma)
	}

	return keylib.PeerIDFromLibp2p(addrInfo.ID), nil
}

// extractP2PManually searches for /p2p/<peerID> in the multiaddr components.
func extractP2PManually(ma multiaddr.Multiaddr) (keylib.PeerID, error) {
	var foundPeerID keylib.PeerID
	found := false

	multiaddr.ForEach(ma, func(c multiaddr.Component) bool {
		if c.Protocol().Code == multiaddr.P_P2P {
			peerIDStr := c.Value()
			pid, err := peer.Decode(peerIDStr)
			if err == nil {
				foundPeerID = keylib.PeerIDFromLibp2p(pid)
				found = true
				return false // stop iteration
			}
		}
		return true // continue
	})

	if !found {
		return keylib.PeerID{}, fmt.Errorf("multiaddr missing /p2p/<PeerID> component")
	}

	return foundPeerID, nil
}

// extractTargetPeerIDFromCircuitRelay extracts the target PeerID from a circuit relay address.
// The target is the final /p2p/<PeerID> after /p2p-circuit/.
//
// Example: /ip4/1.2.3.4/tcp/4001/p2p/RelayPeerID/p2p-circuit/p2p/TargetPeerID
// Returns: TargetPeerID
func extractTargetPeerIDFromCircuitRelay(ma multiaddr.Multiaddr) (keylib.PeerID, error) {
	maStr := ma.String()

	// Find the /p2p-circuit/ position
	circuitIdx := strings.Index(maStr, "/p2p-circuit/")
	if circuitIdx == -1 {
		return keylib.PeerID{}, fmt.Errorf("circuit relay address missing /p2p-circuit/")
	}

	// Get everything after /p2p-circuit/
	afterCircuit := maStr[circuitIdx+len("/p2p-circuit/"):]

	// Parse the remaining part as a multiaddr
	targetMA, err := multiaddr.NewMultiaddr("/" + afterCircuit)
	if err != nil {
		return keylib.PeerID{}, fmt.Errorf("invalid target portion in circuit relay: %w", err)
	}

	// Extract /p2p/<PeerID> from the target portion
	return extractP2PManually(targetMA)
}

// ReachableAddrs is a collection of ReachableAddr for a NeuronAccount.
// Multiple addresses allow for redundancy and different connectivity options.
type ReachableAddrs struct {
	addrs []ReachableAddr
}

// NewReachableAddrs creates a new collection from the given addresses.
func NewReachableAddrs(addrs ...ReachableAddr) ReachableAddrs {
	// Filter out zero values
	filtered := make([]ReachableAddr, 0, len(addrs))
	for _, addr := range addrs {
		if !addr.IsZero() {
			filtered = append(filtered, addr)
		}
	}
	return ReachableAddrs{addrs: filtered}
}

// ParseReachableAddrs parses multiple multiaddr strings.
// All addresses must be valid; returns error on first invalid address.
func ParseReachableAddrs(strs ...string) (ReachableAddrs, error) {
	addrs := make([]ReachableAddr, 0, len(strs))
	for _, s := range strs {
		if s == "" {
			continue
		}
		addr, err := ParseReachableAddr(s)
		if err != nil {
			return ReachableAddrs{}, err
		}
		addrs = append(addrs, addr)
	}
	return NewReachableAddrs(addrs...), nil
}

// NewReachableAddrsWithValidation creates addresses with PeerID validation.
// All addresses must have the expected PeerID.
func NewReachableAddrsWithValidation(strs []string, expectedPeerID keylib.PeerID) (ReachableAddrs, error) {
	addrs := make([]ReachableAddr, 0, len(strs))
	for _, s := range strs {
		if s == "" {
			continue
		}
		addr, err := NewReachableAddrFromStringWithValidation(s, expectedPeerID)
		if err != nil {
			return ReachableAddrs{}, err
		}
		addrs = append(addrs, addr)
	}
	return NewReachableAddrs(addrs...), nil
}

// Addrs returns a copy of the addresses in this collection.
func (r ReachableAddrs) Addrs() []ReachableAddr {
	if len(r.addrs) == 0 {
		return nil
	}
	result := make([]ReachableAddr, len(r.addrs))
	copy(result, r.addrs)
	return result
}

// Strings returns the string representations of all addresses.
func (r ReachableAddrs) Strings() []string {
	if len(r.addrs) == 0 {
		return nil
	}
	result := make([]string, len(r.addrs))
	for i, addr := range r.addrs {
		result[i] = addr.String()
	}
	return result
}

// Multiaddrs returns the underlying multiaddrs.
func (r ReachableAddrs) Multiaddrs() []multiaddr.Multiaddr {
	if len(r.addrs) == 0 {
		return nil
	}
	result := make([]multiaddr.Multiaddr, len(r.addrs))
	for i, addr := range r.addrs {
		result[i] = addr.Multiaddr()
	}
	return result
}

// IsEmpty returns true if there are no addresses.
func (r ReachableAddrs) IsEmpty() bool {
	return len(r.addrs) == 0
}

// Len returns the number of addresses.
func (r ReachableAddrs) Len() int {
	return len(r.addrs)
}

// First returns the first address, or zero value if empty.
func (r ReachableAddrs) First() ReachableAddr {
	if len(r.addrs) == 0 {
		return ReachableAddr{}
	}
	return r.addrs[0]
}

// ValidateAll validates all addresses against the expected PeerID.
func (r ReachableAddrs) ValidateAll(expectedPeerID keylib.PeerID) error {
	const op = "ReachableAddrs.ValidateAll"

	if expectedPeerID.IsZero() {
		return errInvalidAddress(op, "expected PeerID is zero-value", nil)
	}

	for i, addr := range r.addrs {
		if err := addr.ValidateForAccount(expectedPeerID); err != nil {
			return wrapAccountError(op, ErrKindPeerIDMismatch,
				fmt.Sprintf("address at index %d failed validation", i), err)
		}
	}

	return nil
}

// Contains checks if the given address is in this collection.
func (r ReachableAddrs) Contains(addr ReachableAddr) bool {
	for _, a := range r.addrs {
		if a.Equal(addr) {
			return true
		}
	}
	return false
}

// DirectAddrs returns only the direct (non-relay) addresses.
func (r ReachableAddrs) DirectAddrs() []ReachableAddr {
	result := make([]ReachableAddr, 0, len(r.addrs))
	for _, addr := range r.addrs {
		if !addr.IsCircuitRelay() {
			result = append(result, addr)
		}
	}
	return result
}

// RelayAddrs returns only the circuit relay addresses.
func (r ReachableAddrs) RelayAddrs() []ReachableAddr {
	result := make([]ReachableAddr, 0, len(r.addrs))
	for _, addr := range r.addrs {
		if addr.IsCircuitRelay() {
			result = append(result, addr)
		}
	}
	return result
}
