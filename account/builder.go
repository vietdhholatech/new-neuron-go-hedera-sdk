package account

import (
	"fmt"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// AccountBuilder provides a fluent API for constructing NeuronAccount instances.
// Use NewParentAccountBuilder or NewChildAccountBuilder to create a builder.
type AccountBuilder struct {
	// Core identity
	publicKey    keylib.NeuronPublicKey
	accountType  AccountType
	did          NeuronDID
	parentPubKey keylib.NeuronPublicKey

	// Communication endpoints
	stdIn  CommAddress
	stdOut CommAddress
	stdErr CommAddress

	// Reachable addresses
	reachableAddrs []ReachableAddr

	// Build errors
	errors []error
}

// NewParentAccountBuilder creates a builder for a Parent NeuronAccount.
// Parent accounts require a public key and DID.
func NewParentAccountBuilder(publicKey keylib.NeuronPublicKey, did NeuronDID) *AccountBuilder {
	b := &AccountBuilder{
		publicKey:   publicKey,
		accountType: AccountTypeParent,
		did:         did,
	}

	// Validate required fields upfront
	if publicKey.IsZero() {
		b.addError(errZeroValue("NewParentAccountBuilder", "NeuronPublicKey"))
	}
	if did == nil {
		b.addError(errMissingRequired("NewParentAccountBuilder", "DID"))
	}

	return b
}

// NewChildAccountBuilder creates a builder for a Child NeuronAccount.
// Child accounts require a public key and parent's public key.
func NewChildAccountBuilder(publicKey, parentPubKey keylib.NeuronPublicKey) *AccountBuilder {
	b := &AccountBuilder{
		publicKey:    publicKey,
		accountType:  AccountTypeChild,
		parentPubKey: parentPubKey,
	}

	// Validate required fields upfront
	if publicKey.IsZero() {
		b.addError(errZeroValue("NewChildAccountBuilder", "NeuronPublicKey"))
	}
	if parentPubKey.IsZero() {
		b.addError(errMissingRequired("NewChildAccountBuilder", "parent public key"))
	}

	return b
}

// addError records a build error.
func (b *AccountBuilder) addError(err error) {
	if err != nil {
		b.errors = append(b.errors, err)
	}
}

// WithStdIn sets the stdIn communication address.
func (b *AccountBuilder) WithStdIn(addr CommAddress) *AccountBuilder {
	if err := addr.Validate(); err != nil && !addr.IsZero() {
		b.addError(wrapAccountError("WithStdIn", ErrKindInvalidAddress, "invalid stdIn", err))
	}
	b.stdIn = addr
	return b
}

// WithStdInBackend sets the stdIn address using a specific backend kind and locator.
// This is the generic method for setting stdIn with any registered backend.
//
// Example:
//
//	builder.WithStdInBackend("pulsar-topic", "persistent://tenant/namespace/topic")
func (b *AccountBuilder) WithStdInBackend(kind, locator string) *AccountBuilder {
	addr, err := NewCommAddress(kind, locator)
	if err != nil {
		b.addError(wrapAccountError("WithStdInBackend", ErrKindInvalidAddress,
			fmt.Sprintf("invalid %s address", kind), err))
		return b
	}
	b.stdIn = addr
	return b
}

// WithStdInHedera sets the stdIn address as a Hedera topic.
// This is a convenience wrapper around WithStdInBackend("hedera-topic", topicID).
func (b *AccountBuilder) WithStdInHedera(topicID string) *AccountBuilder {
	return b.WithStdInBackend(HederaTopicKind, topicID)
}

// WithStdInKafka sets the stdIn address as a Kafka topic.
// This is a convenience wrapper around WithStdInBackend("kafka-topic", topicName).
func (b *AccountBuilder) WithStdInKafka(topicName string) *AccountBuilder {
	return b.WithStdInBackend(KafkaTopicKind, topicName)
}

// WithStdOut sets the stdOut communication address.
func (b *AccountBuilder) WithStdOut(addr CommAddress) *AccountBuilder {
	if err := addr.Validate(); err != nil && !addr.IsZero() {
		b.addError(wrapAccountError("WithStdOut", ErrKindInvalidAddress, "invalid stdOut", err))
	}
	b.stdOut = addr
	return b
}

// WithStdOutBackend sets the stdOut address using a specific backend kind and locator.
// This is the generic method for setting stdOut with any registered backend.
func (b *AccountBuilder) WithStdOutBackend(kind, locator string) *AccountBuilder {
	addr, err := NewCommAddress(kind, locator)
	if err != nil {
		b.addError(wrapAccountError("WithStdOutBackend", ErrKindInvalidAddress,
			fmt.Sprintf("invalid %s address", kind), err))
		return b
	}
	b.stdOut = addr
	return b
}

// WithStdOutHedera sets the stdOut address as a Hedera topic.
// This is a convenience wrapper around WithStdOutBackend("hedera-topic", topicID).
func (b *AccountBuilder) WithStdOutHedera(topicID string) *AccountBuilder {
	return b.WithStdOutBackend(HederaTopicKind, topicID)
}

// WithStdOutKafka sets the stdOut address as a Kafka topic.
// This is a convenience wrapper around WithStdOutBackend("kafka-topic", topicName).
func (b *AccountBuilder) WithStdOutKafka(topicName string) *AccountBuilder {
	return b.WithStdOutBackend(KafkaTopicKind, topicName)
}

// WithStdErr sets the stdErr communication address.
func (b *AccountBuilder) WithStdErr(addr CommAddress) *AccountBuilder {
	if err := addr.Validate(); err != nil && !addr.IsZero() {
		b.addError(wrapAccountError("WithStdErr", ErrKindInvalidAddress, "invalid stdErr", err))
	}
	b.stdErr = addr
	return b
}

// WithStdErrBackend sets the stdErr address using a specific backend kind and locator.
// This is the generic method for setting stdErr with any registered backend.
func (b *AccountBuilder) WithStdErrBackend(kind, locator string) *AccountBuilder {
	addr, err := NewCommAddress(kind, locator)
	if err != nil {
		b.addError(wrapAccountError("WithStdErrBackend", ErrKindInvalidAddress,
			fmt.Sprintf("invalid %s address", kind), err))
		return b
	}
	b.stdErr = addr
	return b
}

// WithStdErrHedera sets the stdErr address as a Hedera topic.
// This is a convenience wrapper around WithStdErrBackend("hedera-topic", topicID).
func (b *AccountBuilder) WithStdErrHedera(topicID string) *AccountBuilder {
	return b.WithStdErrBackend(HederaTopicKind, topicID)
}

// WithStdErrKafka sets the stdErr address as a Kafka topic.
// This is a convenience wrapper around WithStdErrBackend("kafka-topic", topicName).
func (b *AccountBuilder) WithStdErrKafka(topicName string) *AccountBuilder {
	return b.WithStdErrBackend(KafkaTopicKind, topicName)
}

// WithBackendTopics sets all three communication addresses using a specific backend.
// This is a convenience method for using the same backend for all addresses.
//
// Example:
//
//	builder.WithBackendTopics("kafka-topic", "stdin-topic", "stdout-topic", "stderr-topic")
func (b *AccountBuilder) WithBackendTopics(kind, stdInLocator, stdOutLocator, stdErrLocator string) *AccountBuilder {
	return b.WithStdInBackend(kind, stdInLocator).
		WithStdOutBackend(kind, stdOutLocator).
		WithStdErrBackend(kind, stdErrLocator)
}

// WithHederaTopics sets all three communication addresses to Hedera topics.
// This is a convenience method for the common case.
func (b *AccountBuilder) WithHederaTopics(stdInTopic, stdOutTopic, stdErrTopic string) *AccountBuilder {
	return b.WithStdInHedera(stdInTopic).
		WithStdOutHedera(stdOutTopic).
		WithStdErrHedera(stdErrTopic)
}

// WithKafkaTopics sets all three communication addresses to Kafka topics.
// This is a convenience method for the common case.
func (b *AccountBuilder) WithKafkaTopics(stdInTopic, stdOutTopic, stdErrTopic string) *AccountBuilder {
	return b.WithStdInKafka(stdInTopic).
		WithStdOutKafka(stdOutTopic).
		WithStdErrKafka(stdErrTopic)
}

// WithReachableAddr adds a reachable address from a multiaddr string.
// The PeerID will be validated against the account's public key during Build().
func (b *AccountBuilder) WithReachableAddr(multiaddr string) *AccountBuilder {
	addr, err := ParseReachableAddr(multiaddr)
	if err != nil {
		b.addError(wrapAccountError("WithReachableAddr", ErrKindInvalidAddress, "invalid multiaddr", err))
		return b
	}
	b.reachableAddrs = append(b.reachableAddrs, addr)
	return b
}

// WithReachableAddrs adds multiple reachable addresses from multiaddr strings.
func (b *AccountBuilder) WithReachableAddrs(multiaddrs ...string) *AccountBuilder {
	for _, ma := range multiaddrs {
		b.WithReachableAddr(ma)
	}
	return b
}

// WithReachableAddrValidated adds a reachable address with immediate PeerID validation.
// Use this when you want to validate during building rather than at Build() time.
func (b *AccountBuilder) WithReachableAddrValidated(multiaddr string) *AccountBuilder {
	if b.publicKey.IsZero() {
		b.addError(errZeroValue("WithReachableAddrValidated", "cannot validate without public key"))
		return b
	}

	expectedPeerID, err := b.publicKey.PeerID()
	if err != nil {
		b.addError(wrapAccountError("WithReachableAddrValidated", ErrKindValidation, "failed to derive PeerID", err))
		return b
	}
	addr, err := NewReachableAddrFromStringWithValidation(multiaddr, expectedPeerID)
	if err != nil {
		b.addError(err)
		return b
	}
	b.reachableAddrs = append(b.reachableAddrs, addr)
	return b
}

// Build constructs the NeuronAccount.
// Returns an error if any validation fails.
func (b *AccountBuilder) Build() (NeuronAccount, error) {
	const op = "AccountBuilder.Build"

	// Check for accumulated errors
	if len(b.errors) > 0 {
		// Return the first error
		return NeuronAccount{}, b.errors[0]
	}

	// Final validation based on account type
	switch b.accountType {
	case AccountTypeParent:
		if err := ValidateParentAccount(b.publicKey, b.did); err != nil {
			return NeuronAccount{}, wrapAccountError(op, ErrKindValidation, "parent validation failed", err)
		}
		// Validate DID matches public key
		if err := ValidateDIDMatchesKey(b.did, b.publicKey); err != nil {
			return NeuronAccount{}, wrapAccountError(op, ErrKindInvalidDID, "DID-key mismatch", err)
		}
	case AccountTypeChild:
		if err := ValidateChildAccount(b.publicKey, b.parentPubKey); err != nil {
			return NeuronAccount{}, wrapAccountError(op, ErrKindValidation, "child validation failed", err)
		}
	default:
		return NeuronAccount{}, errInvalidAccount(op, "invalid account type")
	}

	// Derive identifiers from public key
	peerID, err := b.publicKey.PeerID()
	if err != nil {
		return NeuronAccount{}, wrapAccountError(op, ErrKindValidation, "failed to derive PeerID", err)
	}
	evmAddress := b.publicKey.EVMAddress()

	// Validate all reachable addresses have matching PeerID
	for i, addr := range b.reachableAddrs {
		if err := addr.ValidateForAccount(peerID); err != nil {
			return NeuronAccount{}, wrapAccountError(op, ErrKindPeerIDMismatch,
				fmt.Sprintf("reachable address at index %d has wrong PeerID", i), err)
		}
	}

	// Construct the account
	account := NeuronAccount{
		publicKey:      b.publicKey,
		peerID:         peerID,
		evmAddress:     evmAddress,
		accountType:    b.accountType,
		did:            b.did,
		parentPubKey:   b.parentPubKey,
		stdIn:          b.stdIn,
		stdOut:         b.stdOut,
		stdErr:         b.stdErr,
		reachableAddrs: NewReachableAddrs(b.reachableAddrs...),
	}

	return account, nil
}

// MustBuild constructs the NeuronAccount and panics on error.
// This is intended for use in tests and initialization of known-valid accounts.
func (b *AccountBuilder) MustBuild() NeuronAccount {
	account, err := b.Build()
	if err != nil {
		panic(err)
	}
	return account
}

// Errors returns all accumulated build errors.
// Useful for debugging when Build() fails.
func (b *AccountBuilder) Errors() []error {
	return b.errors
}

// HasErrors returns true if there are any build errors.
func (b *AccountBuilder) HasErrors() bool {
	return len(b.errors) > 0
}
