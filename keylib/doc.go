// Package keylib provides type-safe cryptographic key management for the Neuron SDK.
//
// # Overview
//
// The keylib package elevates cryptographic keys to first-class citizens, replacing
// brittle string-based key handling with a robust, type-safe system. It provides
// seamless interoperability between Hedera, Ethereum, and libp2p ecosystems.
//
// # Core Types
//
// The package defines several immutable key types:
//
//   - [NeuronPrivateKey]: ECDSA secp256k1 private key with signing capabilities
//   - [NeuronPublicKey]: ECDSA secp256k1 public key with verification
//   - [EVMAddress]: 20-byte Ethereum address with EIP-55 checksum support
//   - [PeerID]: libp2p peer identifier for P2P networking
//   - [Signature]: ECDSA signature with recovery support
//   - [EncryptedPrivateKey]: Password-protected private key storage
//
// # Key Strategy
//
// The library standardizes on ECDSA secp256k1 keys for maximum interoperability:
//
//   - A single private key controls both Hedera accounts and Ethereum addresses
//   - The same key derives libp2p PeerIDs for P2P networking
//   - Ed25519 keys are explicitly rejected with clear error messages
//
// # Design Philosophy
//
// The package follows these principles:
//
//   - Type safety: Functions accept and return types, not strings
//   - Validation at boundaries: All Parse* functions validate immediately
//   - No panics: All fallible operations return (T, error)
//   - Zero-value safety: Zero values are invalid; methods return errors
//   - Constant-time operations: All key comparisons prevent timing attacks
//   - Memory safety: Zeroize() clears sensitive data from memory
//
// # Basic Usage
//
// Generate a new key:
//
//	privKey, err := keylib.GeneratePrivateKey()
//	if err != nil {
//	    return err
//	}
//	pubKey := privKey.PublicKey()
//
// Parse an existing key:
//
//	privKey, err := keylib.ParsePrivateKeyHex("0x...")
//	if err != nil {
//	    return err
//	}
//
// Derive addresses:
//
//	evmAddr := pubKey.EVMAddress()
//	peerID, err := pubKey.PeerID()
//
// Sign and verify:
//
//	sig, err := privKey.SignMessage([]byte("hello"))
//	if err != nil {
//	    return err
//	}
//	valid := pubKey.Verify([]byte("hello"), sig)
//
// # Mnemonic Support
//
// Generate or restore keys from BIP39 mnemonics:
//
//	// Generate new mnemonic
//	mnemonic, err := keylib.GenerateMnemonic(12)
//
//	// Restore key from mnemonic
//	privKey, err := keylib.PrivateKeyFromMnemonic(mnemonic)
//
// # Key Protection
//
// Encrypt private keys for secure storage:
//
//	encrypted, err := privKey.Scramble("password")
//	if err != nil {
//	    return err
//	}
//
//	// Store encrypted as JSON
//	jsonData, _ := json.Marshal(encrypted)
//
//	// Later, decrypt
//	restored, err := keylib.UnscramblePrivateKey(encrypted, "password")
//
// # Hedera Integration
//
// Convert to/from Hedera SDK types:
//
//	hederaPrivKey := privKey.ToHederaPrivateKey()
//	hederaPubKey := pubKey.ToHederaPublicKey()
//
//	// From Hedera key
//	neuronKey, err := keylib.PrivateKeyFromHedera(hederaPrivKey)
//
// # Error Handling
//
// All errors are of type *KeyError with detailed context:
//
//	privKey, err := keylib.ParsePrivateKeyHex("invalid")
//	if err != nil {
//	    var keyErr *keylib.KeyError
//	    if errors.As(err, &keyErr) {
//	        fmt.Printf("Operation: %s, Kind: %s, Details: %s\n",
//	            keyErr.Op, keyErr.Kind, keyErr.Details)
//	    }
//	}
//
// # Security Considerations
//
// When handling private keys:
//
//   - Call Zeroize() when done with a private key
//   - Use Scramble() for persistent storage
//   - Never log or print private key hex values
//   - Use constant-time Matches* methods for comparisons
package keylib
