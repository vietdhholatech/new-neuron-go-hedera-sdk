package keylib

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerID wraps a libp2p peer.ID with type safety.
// It represents a unique identifier for a node in the libp2p network.
type PeerID struct {
	id peer.ID
}

// ParsePeerID parses a peer ID from its string representation.
// The string should be a base58-encoded multihash.
// Example: "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
func ParsePeerID(s string) (PeerID, error) {
	const op = "ParsePeerID"

	if s == "" {
		return PeerID{}, errInvalidLength(op, 1, 0, "characters (minimum)")
	}

	id, err := peer.Decode(s)
	if err != nil {
		return PeerID{}, wrapError(op, ErrKindInvalidFormat, "invalid peer ID format", err)
	}

	return PeerID{id: id}, nil
}

// peerIDFromPublicKeyBytes derives a PeerID from secp256k1 public key bytes.
// The input should be 33 bytes (compressed SEC1 format).
func peerIDFromPublicKeyBytes(pubKeyBytes []byte) (PeerID, error) {
	const op = "peerIDFromPublicKeyBytes"

	// Unmarshal as secp256k1 public key for libp2p
	libp2pPubKey, err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		return PeerID{}, wrapError(op, ErrKindInvalidKey, "failed to unmarshal secp256k1 public key for libp2p", err)
	}

	// Derive peer ID from the public key
	id, err := peer.IDFromPublicKey(libp2pPubKey)
	if err != nil {
		return PeerID{}, wrapError(op, ErrKindDerivation, "failed to derive peer ID from public key", err)
	}

	return PeerID{id: id}, nil
}

// IsZero returns true if this is a zero-value PeerID.
func (p PeerID) IsZero() bool {
	return p.id == ""
}

// String returns the base58-encoded string representation of the peer ID.
func (p PeerID) String() string {
	return p.id.String()
}

// Bytes returns the raw bytes of the peer ID.
func (p PeerID) Bytes() []byte {
	return []byte(p.id)
}

// ToLibp2p returns the underlying libp2p peer.ID.
// Useful for interoperability with libp2p libraries.
func (p PeerID) ToLibp2p() peer.ID {
	return p.id
}

// Equal returns true if this peer ID equals another.
// Uses constant-time comparison to prevent timing attacks.
func (p PeerID) Equal(other PeerID) bool {
	return constantTimeEqualStrings(string(p.id), string(other.id))
}

// PeerIDFromLibp2p creates a PeerID from a libp2p peer.ID.
func PeerIDFromLibp2p(id peer.ID) PeerID {
	return PeerID{id: id}
}

// Validate checks if the peer ID is valid and non-empty.
func (p PeerID) Validate() error {
	if p.IsZero() {
		return errZeroValue("PeerID.Validate", "PeerID")
	}
	return p.id.Validate()
}
