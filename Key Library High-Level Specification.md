# Key Library High-Level Specification

## 1. Intent & Philosophy

The goal of this library is to replace brittle, string-based key handling with a robust, type-safe system. Currently, key conversions often rely on loose string passing, which is error-prone and hard to debug.

We want to elevate "Keys" to be first-class citizens in the codebase. A developer should never have to guess if a string is a raw hex, a 0x-prefixed hex, or a base58 encoded value. The library must handle these distinctions internally and expose strict, high-level **Neuron Types**.

## 2. Core Requirements

### A. The "Neuron Key" Elevation

The application should primarily interact with **Neuron Keys**, not raw Hedera or Ethereum keys.

- **Elevation**: Any primitive key (Hedera Public Key, raw bytes) must be "elevated" into a `NeuronPublicKey` or `NeuronPrivateKey`.
- **Unified Interface**: Whether the underlying key is a Hedera-native Ed25519 key or an Ethereum-compatible Secp256k1 key, it is wrapped in a `Neuron` type.
- **Benefit**: This abstraction layer allows the rest of the application to be agnostic to the specific underlying crypto, while ensuring all keys are valid and potent.

### B. Primary Key Strategy: ECDSA & Ethereum Compatibility

To ensure seamless interoperability between the Hedera and Ethereum ecosystems, **we will standardize on ECDSA (Secp256k1) keys**.

- **Why**: This ensures that a single private key controls both a Hedera Account and an Ethereum Address.
- **Requirement**: While the library may support Ed25519 for legacy reasons or specific Libp2p needs, all primary application logic should default to or favor ECDSA.

### C. Robust Conversion Matrix

The library must provide a "Rosetta Stone" for keys, allowing conversions between:

1.  **Neuron Public Keys** (wrapping Hedera/Eth keys)
2.  **Ethereum Addresses** (EVM)
3.  **Libp2p PeerIDs**

**Critical Functionality:**

- **PeerID from Neuron Key**: deterministically derive the Libp2p PeerID from a Neuron Public Key.
- **Ethereum Address from Neuron Key**: Extract the EVM address from a Neuron Public Key.
- **EMM address to PeerID and back**: Will require knoweledge of the Nueorn Public key

### D. Input Handling & Validation

"Don't accept stupid strings."

- **Strict Parsing**: Functions accepting strings must validate all preconditions (prefixes like `0x`, hex validity, length checks).
- **Descriptive Errors**: If a key is invalid, the error must explain _why_ (e.g., "invalid hex character", "wrong length for Secp256k1", "checksum mismatch").

### D. Type Safety

- Functions should accept and return **Types**, not strings.
- Example: `ConvertToPeerID(pubKey PublicKey) PeerID` is preferred over `ConvertToPeerID(pubKey string) string`.
- This ensures that once a key is instantiated, it is guaranteed to be valid.

## 3. Scope of Functionality

### Key Management

- **Elevation**: Factory methods to take various inputs (strings, SDK types) and return `NeuronPublicKey` / `NeuronPrivateKey`.
- **Extraction**: Ability to retrieve the underlying raw keys or SDK-specific types (e.g., `ToHederaKey()`) when interaction with legacy code or specific SDK functions is required.
- **Generation**: Ability to generate new `NeuronPrivateKey`s (defaulting to ECDSA).
- **Restoration**: Ability to restore `NeuronKeys` from Mnemonic phrases.
- **Signer Interface**: `NeuronPrivateKey` must satisfy standard signing interfaces.
- **Key Safety**: Provide methods to `Scramble` (encrypt) and `Unscramble` (decrypt) private keys using a password.
  - _Goal_: A basic security layer for storing keys at rest, preferred over plaintext storage.
- **Key Matching Verification**: Detailed boolean functions to verify relationships.
  - `PrivateKey.Matches(PublicKey)`
  - `PrivateKey.Matches(EVMAddress)`
  - `PublicKey.Matches(PeerID)`
- **Signing & Verification**:
  - `Sign(message)`: Comprehensive signing using the elevated key.
  - `Verify(message, signature)`: Verify signatures from various formats (Hedera, raw ECDSA).

### Inter-Ecosystem Conversions

- **Neuron <-> Ethereum**: Seamless mapping between Neuron Keys and Ethereum Addresses.
- **Neuron <-> Libp2p**: Deriving network identifiers (PeerIDs) from Neuron Keys.

## 4. Developer Experience

- **Predictability**: If a function returns a `NeuronKey`, it is valid.
- **Safety**: Impossible to mix up an Ethereum address with a Neuron Key due to distinct types.
- **Clarity**: API names should clearly indicate if they are performing a computation (deriving an address) or a transformation (elevating a type).
