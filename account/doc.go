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
// NeuronAccounts follow a strict two-level hierarchy:
//
//   - Parent Account: Has no parent, may have children, MUST have a DID
//   - Child Account: Has exactly one parent, no children, no DID
//
// This structure ensures clear ownership and administrative boundaries while maintaining
// system simplicity. Nesting is limited to one level (no grandchildren).
//
// # Communication Addresses
//
// A NeuronAccount exposes three public communication addresses:
//   - stdIn: where others send messages to the agent
//   - stdOut: where the agent publishes outputs/heartbeats
//   - stdErr: where the agent publishes errors/diagnostics
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
// Creating a Parent account with did:key:
//
//	import (
//	    "github.com/aspect-build/neuron-go-hedera-sdk/account"
//	    "github.com/aspect-build/neuron-go-hedera-sdk/account/didkey"
//	    "github.com/aspect-build/neuron-go-hedera-sdk/keylib"
//	)
//
//	// Generate or load a key pair
//	privKey, err := keylib.GenerateKey()
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
//	// Build the account
//	acct, err := account.NewParentAccountBuilder(pubKey, did).
//	    WithStdInHedera("0.0.1001").
//	    WithStdOutHedera("0.0.1002").
//	    WithStdErrHedera("0.0.1003").
//	    WithReachableAddr("/ip4/203.0.113.10/tcp/4001/p2p/" + pubKey.PeerID().String()).
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
// # Blockchain Interfaces
//
// The package defines interfaces for blockchain operations but does NOT implement them:
//   - AccountCreator: Creates on-chain account identities
//   - AccountVerifier: Verifies account existence on-chain
//   - EndpointVerifier: Verifies communication endpoints exist
//   - RegistryResolver: Resolves accounts from yellow pages registries
//
// These interfaces allow blockchain-specific implementations to be provided
// without coupling this package to any particular chain's SDK.
//
// # Sub-packages
//
//   - didkey: did:key implementation for secp256k1 keys
package account
