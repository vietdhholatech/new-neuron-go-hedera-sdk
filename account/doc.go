// Package account provides the NeuronAccount abstraction for agent identity and endpoints.
//
// # Overview
//
// A NeuronAccount is the portable, blockchain-agnostic representation of an agent's
// identity and endpoints. It is:
//   - Identified by a Neuron Key (NeuronPublicKey from keylib)
//   - NOT responsible for communication or message passing
//   - Responsible for describing WHERE communication happens and HOW the agent can be reached
//   - Capable of being anchored to one or more blockchains
//
// # Identity
//
// A NeuronAccount is identified by a NeuronPublicKey, which is the root of identity.
// All other identifiers are derived from this key:
//   - PeerID (libp2p) - for P2P networking
//   - EVMAddress (Ethereum-compatible) - for blockchain interactions
//   - DID (Decentralized Identifier) - for self-sovereign identity
//
// The Neuron key is always authoritative; derived identifiers are never authoritative on their own.
//
// # Account Hierarchy
//
// NeuronAccounts follow a strict hierarchy with three account types:
//
//   - Parent Account: Has no parent, may have children, MUST have a DID,
//     must NOT have communication channels
//   - Child Account: Has exactly one parent, no children, no DID,
//     MUST have all three communication channels (stdIn, stdOut, stdErr)
//   - Shared Account: M-of-N multisig threshold account using MultisigKey,
//     no DID, no communication channels, no parent reference
//
// This structure ensures clear ownership and administrative boundaries while maintaining
// system simplicity. Nesting is limited to one level (no grandchildren).
//
// # Communication Addresses
//
// Child accounts expose three public communication addresses (all required):
//   - stdIn: where others send messages to the agent
//   - stdOut: where the agent publishes outputs/heartbeats
//   - stdErr: where the agent publishes errors/diagnostics
//
// Parent and Shared accounts must NOT have communication channels.
//
// These are represented as CommAddress values, which are technology-agnostic.
// Currently supported backends include:
//   - Hedera Consensus Service topics
//   - Kafka topics
//
// The account does NOT send, receive, or subscribe to messages. It only declares
// where these channels exist. Actual communication is handled by separate packages.
//
// # Reachable Addresses
//
// A NeuronAccount exposes reachable connection addresses using libp2p multiaddr.
// These are the agent's canonical "find / dial me" endpoints, used for:
//   - Direct P2P dialing
//   - NAT traversal
//   - Relay-based connectivity
//
// Every reachable address MUST contain /p2p/<PeerID> where the PeerID matches
// the account's derived PeerID. This is enforced during construction and validation.
//
// # DID Support
//
// The package defines a NeuronDID interface for Decentralized Identifiers.
// The default implementation is did:key, provided in the didkey sub-package.
// The did:key format for secp256k1 keys is:
//
//	did:key:z<base58btc(0xe7 + compressed-pubkey-33-bytes)>
//
// Other DID methods (did:web, did:hedera, etc.) can be implemented by satisfying
// the NeuronDID interface.
//
// # Usage Example
//
// Creating a Parent account with did:key (Parent accounts have no comm channels):
//
//	import (
//	    "github.com/aspect-build/neuron-go-hedera-sdk/account"
//	    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
//	    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
//	)
//
//	// Generate or load a key pair
//	privKey, err := keylib.GeneratePrivateKey()
//	if err != nil {
//	    return err
//	}
//	pubKey := privKey.PublicKey()
//
//	// Create did:key from public key
//	did, err := didkey.FromPublicKey(pubKey)
//	if err != nil {
//	    return err
//	}
//
//	// Build the Parent account (no comm channels)
//	parent, err := account.NewParentAccountBuilder(pubKey, did).Build()
//	if err != nil {
//	    return err
//	}
//
// Creating a Child account (all 3 comm channels required):
//
//	childPrivKey, _ := keylib.GeneratePrivateKey()
//	childPubKey := childPrivKey.PublicKey()
//
//	child, err := account.NewChildAccountBuilder(childPubKey, pubKey).
//	    WithStdInHedera("0.0.1001").
//	    WithStdOutHedera("0.0.1002").
//	    WithStdErrHedera("0.0.1003").
//	    Build()
//	if err != nil {
//	    return err
//	}
//
// # Thread Safety
//
// NeuronAccount is immutable after construction and is safe for concurrent use.
// The AccountBuilder is NOT thread-safe and should only be used from a single goroutine.
//
// # Ledger Attachment
//
// A NeuronAccount can be attached to ledger infrastructure (blockchain) via LedgerAttachment:
//   - LedgerAttachment: Links account to a specific ledger address
//   - AttachmentState: Tracks attachment lifecycle (Detached, Attached, Verified)
//   - VerificationStatus: Tracks verification result (None, Pending, Verified, Failed)
//
// # Blockchain Interfaces
//
// The package defines interfaces for blockchain operations but does NOT implement them:
//   - AccountCreator: Creates on-chain account identities
//   - AccountVerifier: Verifies account existence on-chain
//   - EndpointVerifier: Verifies communication endpoints exist
//   - RegistryResolver: Resolves accounts from yellow pages registries
//   - LedgerVerifier: Verifies account-ledger relationships including key ownership,
//     semantic consistency, and parent-child relationships
//
// These interfaces allow blockchain-specific implementations to be provided
// without coupling this package to any particular chain's SDK.
//
// # Sub-packages
//
//   - didkey: did:key implementation for secp256k1 keys
package account
