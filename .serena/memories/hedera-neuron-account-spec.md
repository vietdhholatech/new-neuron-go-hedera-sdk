# NeuronAccount Module Specification Summary

## Overview
Hierarchical account system for agent identity and endpoint management in distributed systems.

## Three Account Types

### Parent Accounts (AccountType = 1)
- Root identity and primary financial account
- **MUST have**: DID document, credit balance, single `NeuronPublicKey` (secp256k1)
- **MUST NOT have**: Communication channels (stdIn/stdOut/stdErr), parent reference, MultisigKey
- Can create and manage Child accounts
- Receives revenue from Child account data transactions
- Supports ledger attachment

### Child Accounts (AccountType = 2)
- Agent/device/service endpoint identities under a Parent
- **MUST have**: Parent public key reference, all three communication channels (stdIn, stdOut, stdErr), single `NeuronPublicKey`
- **MUST NOT have**: DID document, MultisigKey
- Balance allocation for operational expenses (e.g., transaction signing fees)
- Supports reachable addresses (multiaddresses) for P2P communication
- Supports ledger attachment

### Shared Accounts (AccountType = 3) - NEW
- Multisig threshold accounts (e.g., 2-of-3)
- **MUST have**: `MultisigKey` with threshold configuration
- **MUST NOT have**: DID document, communication channels, parent reference, single public key
- Maintains currency symbol and balance
- Supports ledger attachment
- Multisig config must match ledger's multisig structure

## Ledger Attachment
- Accounts can be attached to settlement infrastructure (Ethereum, Hedera, private ledgers)
- **Uses derived address (e.g., Ethereum address) as account identifier**, NOT public key directly
- Must verify:
  1. Account exists on the ledger
  2. Neuron private key proves ownership of the derived address
  3. Public key semantics match between Neuron and ledger
- Account creation on ledger is OUT OF SCOPE

## Communication Endpoints (Child Accounts Only)
- **stdIn**: Inbound messages
- **stdOut**: Outbound messages  
- **stdErr**: Error/diagnostic messages
- **reachableAddrs**: Private P2P endpoints (libp2p multiaddresses)

### Extensible Backend System
- Hedera Consensus Service topics (e.g., "0.0.12345")
- Kafka topics
- Custom backends via plugin registration

## Key Entities
- `NeuronAccount`: Complete agent identity
- `CommAddress`: Backend kind + locator string
- `Backend`: Messaging technology plugin with validation rules
- `ReachableAddr`: Multiaddress + extracted PeerID
- `NeuronDID`: DID document (Parent accounts only)
- `LedgerAttachment`: Ledger identifier, attached address, state

## Validation Rules
- PeerID must be consistent between account identity and reachable addresses
- Communication addresses validated per backend rules
- AccountType = 0 (Unspecified) is INVALID

## Out of Scope
- Sending money (transfer execution)
- Transaction signing
- Account creation on ledger
- Message transport/routing
- Key generation/storage (handled by keylib)
- Network connectivity

## Key Requirements
- FR-001 to FR-021 define functional requirements
- SC-001 to SC-010 define success criteria
- Operations should complete in <100ms (excluding network I/O)
