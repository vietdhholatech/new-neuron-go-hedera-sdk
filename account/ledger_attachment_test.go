package account

import (
	"testing"
)

// =============================================================================
// AttachmentState Tests
// =============================================================================

func TestAttachmentState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    AttachmentState
		expected string
	}{
		{"Detached", AttachmentStateDetached, "Detached"},
		{"Attached", AttachmentStateAttached, "Attached"},
		{"Verified", AttachmentStateVerified, "Verified"},
		{"Unknown", AttachmentState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("AttachmentState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAttachmentState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state AttachmentState
		valid bool
	}{
		{"Detached is valid", AttachmentStateDetached, true},
		{"Attached is valid", AttachmentStateAttached, true},
		{"Verified is valid", AttachmentStateVerified, true},
		{"Negative is invalid", AttachmentState(-1), false},
		{"Out of range is invalid", AttachmentState(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.valid {
				t.Errorf("AttachmentState.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

// =============================================================================
// VerificationStatus Tests
// =============================================================================

func TestVerificationStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   VerificationStatus
		expected string
	}{
		{"None", VerificationStatusNone, "None"},
		{"Pending", VerificationStatusPending, "Pending"},
		{"Verified", VerificationStatusVerified, "Verified"},
		{"Failed", VerificationStatusFailed, "Failed"},
		{"Unknown", VerificationStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("VerificationStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// NewLedgerAttachment Tests
// =============================================================================

func TestNewLedgerAttachment(t *testing.T) {
	t.Run("creates valid attachment", func(t *testing.T) {
		attachment, err := NewLedgerAttachment("ethereum-mainnet", "0x1234567890abcdef")
		if err != nil {
			t.Fatalf("NewLedgerAttachment() error = %v", err)
		}

		if attachment.LedgerIdentifier() != "ethereum-mainnet" {
			t.Errorf("LedgerIdentifier() = %v, want %v", attachment.LedgerIdentifier(), "ethereum-mainnet")
		}
		if attachment.AttachedAddress() != "0x1234567890abcdef" {
			t.Errorf("AttachedAddress() = %v, want %v", attachment.AttachedAddress(), "0x1234567890abcdef")
		}
		if attachment.State() != AttachmentStateDetached {
			t.Errorf("State() = %v, want Detached", attachment.State())
		}
		if attachment.VerificationStatus() != VerificationStatusNone {
			t.Errorf("VerificationStatus() = %v, want None", attachment.VerificationStatus())
		}
	})

	t.Run("empty ledgerID returns error", func(t *testing.T) {
		_, err := NewLedgerAttachment("", "0x1234")
		if err == nil {
			t.Error("expected error for empty ledgerID")
		}
	})

	t.Run("empty address returns error", func(t *testing.T) {
		_, err := NewLedgerAttachment("ethereum-mainnet", "")
		if err == nil {
			t.Error("expected error for empty address")
		}
	})
}

// =============================================================================
// LedgerAttachment Accessors Tests
// =============================================================================

func TestLedgerAttachment_Accessors(t *testing.T) {
	t.Run("nil attachment returns defaults", func(t *testing.T) {
		var attachment *LedgerAttachment

		if attachment.LedgerIdentifier() != "" {
			t.Error("nil LedgerIdentifier should return empty string")
		}
		if attachment.AttachedAddress() != "" {
			t.Error("nil AttachedAddress should return empty string")
		}
		if attachment.State() != AttachmentStateDetached {
			t.Error("nil State should return Detached")
		}
		if attachment.VerificationStatus() != VerificationStatusNone {
			t.Error("nil VerificationStatus should return None")
		}
	})
}

// =============================================================================
// LedgerAttachment State Checks Tests
// =============================================================================

func TestLedgerAttachment_IsAttached(t *testing.T) {
	t.Run("nil returns false", func(t *testing.T) {
		var attachment *LedgerAttachment
		if attachment.IsAttached() {
			t.Error("nil IsAttached() should return false")
		}
	})

	t.Run("Detached returns false", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if attachment.IsAttached() {
			t.Error("Detached state should return IsAttached=false")
		}
	})

	t.Run("Attached returns true", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if !attachment.IsAttached() {
			t.Error("Attached state should return IsAttached=true")
		}
	})

	t.Run("Verified returns true", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if err := attachment.SetVerified(); err != nil {
			t.Fatalf("SetVerified() error = %v", err)
		}
		if !attachment.IsAttached() {
			t.Error("Verified state should return IsAttached=true")
		}
	})
}

func TestLedgerAttachment_IsVerified(t *testing.T) {
	t.Run("nil returns false", func(t *testing.T) {
		var attachment *LedgerAttachment
		if attachment.IsVerified() {
			t.Error("nil IsVerified() should return false")
		}
	})

	t.Run("Detached returns false", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if attachment.IsVerified() {
			t.Error("Detached state should return IsVerified=false")
		}
	})

	t.Run("Attached returns false", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if attachment.IsVerified() {
			t.Error("Attached state should return IsVerified=false")
		}
	})

	t.Run("Verified returns true", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if err := attachment.SetVerified(); err != nil {
			t.Fatalf("SetVerified() error = %v", err)
		}
		if !attachment.IsVerified() {
			t.Error("Verified state should return IsVerified=true")
		}
	})
}

// =============================================================================
// LedgerAttachment State Transitions Tests
// =============================================================================

func TestLedgerAttachment_SetAttached(t *testing.T) {
	t.Run("transitions from Detached to Attached", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetAttached(); err != nil {
			t.Fatalf("SetAttached() error = %v", err)
		}
		if attachment.State() != AttachmentStateAttached {
			t.Errorf("State() = %v, want Attached", attachment.State())
		}
	})

	t.Run("rejects transition from Attached", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		if err := attachment.SetAttached(); err == nil {
			t.Error("expected error when transitioning from Attached to Attached")
		}
	})

	t.Run("rejects transition from Verified", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		_ = attachment.SetVerified()
		if err := attachment.SetAttached(); err == nil {
			t.Error("expected error when transitioning from Verified to Attached")
		}
	})

	t.Run("nil is safe", func(t *testing.T) {
		var attachment *LedgerAttachment
		if err := attachment.SetAttached(); err != nil {
			t.Errorf("nil SetAttached() should return nil, got %v", err)
		}
	})
}

func TestLedgerAttachment_SetVerified(t *testing.T) {
	t.Run("transitions from Attached to Verified", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		if err := attachment.SetVerified(); err != nil {
			t.Fatalf("SetVerified() error = %v", err)
		}
		if attachment.State() != AttachmentStateVerified {
			t.Errorf("State() = %v, want Verified", attachment.State())
		}
		if attachment.VerificationStatus() != VerificationStatusVerified {
			t.Errorf("VerificationStatus() = %v, want Verified", attachment.VerificationStatus())
		}
	})

	t.Run("rejects transition from Detached", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetVerified(); err == nil {
			t.Error("expected error when transitioning from Detached to Verified")
		}
	})

	t.Run("rejects transition from Verified", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		_ = attachment.SetVerified()
		if err := attachment.SetVerified(); err == nil {
			t.Error("expected error when transitioning from Verified to Verified")
		}
	})

	t.Run("nil is safe", func(t *testing.T) {
		var attachment *LedgerAttachment
		if err := attachment.SetVerified(); err != nil {
			t.Errorf("nil SetVerified() should return nil, got %v", err)
		}
	})
}

func TestLedgerAttachment_SetVerificationFailed(t *testing.T) {
	t.Run("sets verification failed from Attached", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		if err := attachment.SetVerificationFailed(); err != nil {
			t.Fatalf("SetVerificationFailed() error = %v", err)
		}
		if attachment.VerificationStatus() != VerificationStatusFailed {
			t.Errorf("VerificationStatus() = %v, want Failed", attachment.VerificationStatus())
		}
		// State should remain Attached
		if attachment.State() != AttachmentStateAttached {
			t.Errorf("State() = %v, want Attached (should not change)", attachment.State())
		}
	})

	t.Run("rejects from Detached", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.SetVerificationFailed(); err == nil {
			t.Error("expected error when marking failed from Detached")
		}
	})

	t.Run("rejects from Verified", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		_ = attachment.SetAttached()
		_ = attachment.SetVerified()
		if err := attachment.SetVerificationFailed(); err == nil {
			t.Error("expected error when marking failed from Verified")
		}
	})

	t.Run("nil is safe", func(t *testing.T) {
		var attachment *LedgerAttachment
		if err := attachment.SetVerificationFailed(); err != nil {
			t.Errorf("nil SetVerificationFailed() should return nil, got %v", err)
		}
	})
}

// =============================================================================
// LedgerAttachment Validate Tests
// =============================================================================

func TestLedgerAttachment_Validate(t *testing.T) {
	t.Run("nil is valid", func(t *testing.T) {
		var attachment *LedgerAttachment
		if err := attachment.Validate(); err != nil {
			t.Errorf("nil Validate() should return nil, got %v", err)
		}
	})

	t.Run("valid attachment passes", func(t *testing.T) {
		attachment, _ := NewLedgerAttachment("ledger", "addr")
		if err := attachment.Validate(); err != nil {
			t.Errorf("valid attachment Validate() should pass, got %v", err)
		}
	})

	t.Run("empty ledgerIdentifier fails", func(t *testing.T) {
		attachment := &LedgerAttachment{
			attachedAddress: "addr",
		}
		if err := attachment.Validate(); err == nil {
			t.Error("empty ledgerIdentifier should fail validation")
		}
	})

	t.Run("empty attachedAddress fails", func(t *testing.T) {
		attachment := &LedgerAttachment{
			ledgerIdentifier: "ledger",
		}
		if err := attachment.Validate(); err == nil {
			t.Error("empty attachedAddress should fail validation")
		}
	})

	t.Run("invalid attachment state fails", func(t *testing.T) {
		attachment := &LedgerAttachment{
			ledgerIdentifier: "ledger",
			attachedAddress:  "addr",
			attachmentState:  AttachmentState(99),
		}
		if err := attachment.Validate(); err == nil {
			t.Error("invalid attachment state should fail validation")
		}
	})
}
