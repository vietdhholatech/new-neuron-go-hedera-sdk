package account

// AttachmentState represents the current state of a ledger attachment.
type AttachmentState int

const (
	// AttachmentStateDetached indicates the account is not linked to a ledger.
	AttachmentStateDetached AttachmentState = iota
	// AttachmentStateAttached indicates the account is linked but not verified.
	AttachmentStateAttached
	// AttachmentStateVerified indicates the account is linked and verified.
	AttachmentStateVerified
)

// String returns a human-readable name for the attachment state.
func (s AttachmentState) String() string {
	switch s {
	case AttachmentStateDetached:
		return "Detached"
	case AttachmentStateAttached:
		return "Attached"
	case AttachmentStateVerified:
		return "Verified"
	default:
		return "Unknown"
	}
}

// IsValid returns true if this is a valid attachment state.
func (s AttachmentState) IsValid() bool {
	return s >= AttachmentStateDetached && s <= AttachmentStateVerified
}

// VerificationStatus represents the result of ledger verification.
type VerificationStatus int

const (
	// VerificationStatusNone indicates verification has not been attempted.
	VerificationStatusNone VerificationStatus = iota
	// VerificationStatusPending indicates verification is in progress.
	VerificationStatusPending
	// VerificationStatusVerified indicates verification succeeded.
	VerificationStatusVerified
	// VerificationStatusFailed indicates verification failed.
	VerificationStatusFailed
)

// String returns a human-readable name for the verification status.
func (s VerificationStatus) String() string {
	switch s {
	case VerificationStatusNone:
		return "None"
	case VerificationStatusPending:
		return "Pending"
	case VerificationStatusVerified:
		return "Verified"
	case VerificationStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// LedgerAttachment represents the link between a NeuronAccount and
// settlement infrastructure (blockchain or ledger).
type LedgerAttachment struct {
	// ledgerIdentifier identifies the ledger (e.g., "ethereum-mainnet", "hedera-mainnet").
	ledgerIdentifier string

	// attachedAddress is the ledger-specific account address (e.g., Ethereum address).
	attachedAddress string

	// attachmentState is the current state of the attachment.
	attachmentState AttachmentState

	// verificationStatus is the result of verification operations.
	verificationStatus VerificationStatus
}

// NewLedgerAttachment creates a new LedgerAttachment in Detached state.
func NewLedgerAttachment(ledgerID, address string) (*LedgerAttachment, error) {
	const op = "NewLedgerAttachment"

	if ledgerID == "" {
		return nil, errMissingRequired(op, "ledgerIdentifier")
	}
	if address == "" {
		return nil, errMissingRequired(op, "attachedAddress")
	}

	return &LedgerAttachment{
		ledgerIdentifier:   ledgerID,
		attachedAddress:    address,
		attachmentState:    AttachmentStateDetached,
		verificationStatus: VerificationStatusNone,
	}, nil
}

// LedgerIdentifier returns the ledger identifier.
func (a *LedgerAttachment) LedgerIdentifier() string {
	if a == nil {
		return ""
	}
	return a.ledgerIdentifier
}

// AttachedAddress returns the ledger-specific account address.
func (a *LedgerAttachment) AttachedAddress() string {
	if a == nil {
		return ""
	}
	return a.attachedAddress
}

// State returns the current attachment state.
func (a *LedgerAttachment) State() AttachmentState {
	if a == nil {
		return AttachmentStateDetached
	}
	return a.attachmentState
}

// VerificationStatus returns the verification status.
func (a *LedgerAttachment) VerificationStatus() VerificationStatus {
	if a == nil {
		return VerificationStatusNone
	}
	return a.verificationStatus
}

// IsAttached returns true if the attachment state is Attached or Verified.
func (a *LedgerAttachment) IsAttached() bool {
	if a == nil {
		return false
	}
	return a.attachmentState >= AttachmentStateAttached
}

// IsVerified returns true if the attachment state is Verified.
func (a *LedgerAttachment) IsVerified() bool {
	if a == nil {
		return false
	}
	return a.attachmentState == AttachmentStateVerified
}

// Validate checks if the LedgerAttachment is valid.
func (a *LedgerAttachment) Validate() error {
	const op = "LedgerAttachment.Validate"

	if a == nil {
		return nil // nil attachment is valid (optional field)
	}

	if a.ledgerIdentifier == "" {
		return errMissingRequired(op, "ledgerIdentifier")
	}
	if a.attachedAddress == "" {
		return errMissingRequired(op, "attachedAddress")
	}
	if !a.attachmentState.IsValid() {
		return errInvalidAccount(op, "invalid attachment state")
	}

	return nil
}

// SetAttached transitions the attachment from Detached to Attached state.
// Returns an error if the current state is not Detached.
func (a *LedgerAttachment) SetAttached() error {
	if a == nil {
		return nil
	}
	if a.attachmentState != AttachmentStateDetached {
		return errInvalidAccount("LedgerAttachment.SetAttached",
			"invalid state transition: can only transition to Attached from Detached, current state is "+a.attachmentState.String())
	}
	a.attachmentState = AttachmentStateAttached
	return nil
}

// SetVerified transitions the attachment from Attached to Verified state.
// Returns an error if the current state is not Attached.
func (a *LedgerAttachment) SetVerified() error {
	if a == nil {
		return nil
	}
	if a.attachmentState != AttachmentStateAttached {
		return errInvalidAccount("LedgerAttachment.SetVerified",
			"invalid state transition: can only transition to Verified from Attached, current state is "+a.attachmentState.String())
	}
	a.attachmentState = AttachmentStateVerified
	a.verificationStatus = VerificationStatusVerified
	return nil
}

// SetVerificationFailed marks verification as failed.
// Returns an error if the current state is not Attached.
func (a *LedgerAttachment) SetVerificationFailed() error {
	if a == nil {
		return nil
	}
	if a.attachmentState != AttachmentStateAttached {
		return errInvalidAccount("LedgerAttachment.SetVerificationFailed",
			"invalid state transition: can only mark verification failed from Attached, current state is "+a.attachmentState.String())
	}
	a.verificationStatus = VerificationStatusFailed
	return nil
}
