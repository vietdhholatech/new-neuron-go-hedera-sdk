package keylib

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
)

// =============================================================================
// ParsePeerID Tests
// =============================================================================

func TestParsePeerID_Comprehensive(t *testing.T) {
	// Known valid peer ID from test vectors
	validPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	t.Run("parses valid peer ID", func(t *testing.T) {
		pid, err := ParsePeerID(validPeerID)
		if err != nil {
			t.Fatalf("ParsePeerID failed: %v", err)
		}
		if pid.IsZero() {
			t.Error("parsed peer ID should not be zero")
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		_, err := ParsePeerID("")
		if err == nil {
			t.Error("empty string should fail")
		}
	})

	t.Run("rejects invalid format", func(t *testing.T) {
		_, err := ParsePeerID("not-a-valid-peer-id")
		if err == nil {
			t.Error("invalid format should fail")
		}
	})

	t.Run("rejects random base58", func(t *testing.T) {
		// Random base58 that is not a valid peer ID
		_, err := ParsePeerID("ABC123xyz")
		if err == nil {
			t.Error("random base58 should fail")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		pid, _ := ParsePeerID(validPeerID)
		str := pid.String()
		restored, err := ParsePeerID(str)
		if err != nil {
			t.Fatalf("round trip failed: %v", err)
		}
		if !pid.Equal(restored) {
			t.Error("peer ID should equal after round trip")
		}
	})
}

// =============================================================================
// PeerID.IsZero Tests
// =============================================================================

func TestPeerID_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var pid PeerID
		if !pid.IsZero() {
			t.Error("zero value should be zero")
		}
	})

	t.Run("parsed peer ID is not zero", func(t *testing.T) {
		pid, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")
		if pid.IsZero() {
			t.Error("parsed peer ID should not be zero")
		}
	})

	t.Run("derived peer ID is not zero", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pid, _ := key.PublicKey().PeerID()
		if pid.IsZero() {
			t.Error("derived peer ID should not be zero")
		}
	})
}

// =============================================================================
// PeerID.String Tests
// =============================================================================

func TestPeerID_String(t *testing.T) {
	validPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	pid, _ := ParsePeerID(validPeerID)

	t.Run("returns correct string", func(t *testing.T) {
		str := pid.String()
		if str != validPeerID {
			t.Errorf("expected %s, got %s", validPeerID, str)
		}
	})

	t.Run("zero peer ID returns empty", func(t *testing.T) {
		var zeroPID PeerID
		str := zeroPID.String()
		if str != "" {
			t.Errorf("zero peer ID should return empty string, got %s", str)
		}
	})
}

// =============================================================================
// PeerID.Bytes Tests
// =============================================================================

func TestPeerID_Bytes(t *testing.T) {
	pid, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")

	t.Run("returns non-empty bytes", func(t *testing.T) {
		bytes := pid.Bytes()
		if len(bytes) == 0 {
			t.Error("bytes should not be empty")
		}
	})

	t.Run("zero peer ID returns empty bytes", func(t *testing.T) {
		var zeroPID PeerID
		bytes := zeroPID.Bytes()
		if len(bytes) != 0 {
			t.Errorf("zero peer ID should return empty bytes, got %d bytes", len(bytes))
		}
	})
}

// =============================================================================
// PeerID.Equal Tests
// =============================================================================

func TestPeerID_Equal(t *testing.T) {
	pid1, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")
	pid2, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")

	t.Run("same peer ID equals", func(t *testing.T) {
		if !pid1.Equal(pid2) {
			t.Error("same peer ID should equal")
		}
	})

	t.Run("different peer IDs not equal", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pid3, _ := key.PublicKey().PeerID()
		if pid1.Equal(pid3) {
			t.Error("different peer IDs should not equal")
		}
	})

	t.Run("zero peer IDs equal", func(t *testing.T) {
		var z1, z2 PeerID
		if !z1.Equal(z2) {
			t.Error("zero peer IDs should equal")
		}
	})

	t.Run("zero not equal to valid", func(t *testing.T) {
		var zeroPID PeerID
		if zeroPID.Equal(pid1) || pid1.Equal(zeroPID) {
			t.Error("zero peer ID should not equal valid peer ID")
		}
	})
}

// =============================================================================
// PeerID.Validate Tests
// =============================================================================

func TestPeerID_Validate(t *testing.T) {
	t.Run("valid peer ID validates", func(t *testing.T) {
		pid, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")
		err := pid.Validate()
		if err != nil {
			t.Errorf("valid peer ID should validate: %v", err)
		}
	})

	t.Run("zero peer ID fails validation", func(t *testing.T) {
		var zeroPID PeerID
		err := zeroPID.Validate()
		if err == nil {
			t.Error("zero peer ID should fail validation")
		}
	})

	t.Run("derived peer ID validates", func(t *testing.T) {
		key, _ := GeneratePrivateKey()
		pid, _ := key.PublicKey().PeerID()
		err := pid.Validate()
		if err != nil {
			t.Errorf("derived peer ID should validate: %v", err)
		}
	})
}

// =============================================================================
// PeerID Interop Tests
// =============================================================================

func TestPeerID_Interop(t *testing.T) {
	pid, _ := ParsePeerID("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")

	t.Run("ToLibp2p returns correct type", func(t *testing.T) {
		libp2pID := pid.ToLibp2p()
		var _ peer.ID = libp2pID // Type assertion
	})

	t.Run("ToLibp2p round trip", func(t *testing.T) {
		libp2pID := pid.ToLibp2p()
		restored := PeerIDFromLibp2p(libp2pID)
		if !pid.Equal(restored) {
			t.Error("peer ID should equal after ToLibp2p round trip")
		}
	})

	t.Run("PeerIDFromLibp2p", func(t *testing.T) {
		libp2pID, _ := peer.Decode("16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq")
		restored := PeerIDFromLibp2p(libp2pID)
		if !pid.Equal(restored) {
			t.Error("PeerIDFromLibp2p should produce equal peer ID")
		}
	})
}

// =============================================================================
// peerIDFromPublicKeyBytes Tests
// =============================================================================

func TestPeerIDFromPublicKeyBytes(t *testing.T) {
	key, _ := GeneratePrivateKey()
	pubKey := key.PublicKey()

	t.Run("derives peer ID from compressed bytes", func(t *testing.T) {
		bytes := pubKey.CompressedBytes()
		pid, err := peerIDFromPublicKeyBytes(bytes[:])
		if err != nil {
			t.Fatalf("peerIDFromPublicKeyBytes failed: %v", err)
		}
		if pid.IsZero() {
			t.Error("derived peer ID should not be zero")
		}
	})

	t.Run("matches PeerID method", func(t *testing.T) {
		bytes := pubKey.CompressedBytes()
		pid1, _ := peerIDFromPublicKeyBytes(bytes[:])
		pid2, _ := pubKey.PeerID()
		if !pid1.Equal(pid2) {
			t.Error("peerIDFromPublicKeyBytes should match PeerID method")
		}
	})

	t.Run("rejects invalid bytes", func(t *testing.T) {
		invalidBytes := []byte{1, 2, 3}
		_, err := peerIDFromPublicKeyBytes(invalidBytes)
		if err == nil {
			t.Error("invalid bytes should fail")
		}
	})
}

// =============================================================================
// Known Test Vector Tests
// =============================================================================

func TestPeerID_KnownVector(t *testing.T) {
	pubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	expectedPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	pubKey, _ := ParsePublicKeyHex(pubKeyHex)
	derivedPID, _ := pubKey.PeerID()

	t.Run("derives correct peer ID", func(t *testing.T) {
		if derivedPID.String() != expectedPeerID {
			t.Errorf("expected %s, got %s", expectedPeerID, derivedPID.String())
		}
	})

	t.Run("parsed peer ID matches derived", func(t *testing.T) {
		parsedPID, _ := ParsePeerID(expectedPeerID)
		if !derivedPID.Equal(parsedPID) {
			t.Error("derived peer ID should equal parsed")
		}
	})
}
