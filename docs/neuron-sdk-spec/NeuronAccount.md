# Feature Specification: NeuronAccount Module

**Feature Branch**: `001-neuron-account-module`  
**Created**: 2026-01-23  
**Status**: Draft  
**Input**: User description: "# NeuronAccount Module Technical Specification"

## Purpose

The NeuronAccount module implements a hierarchical account system for agent identity and endpoint management in distributed systems. The module provides three account types:

**Parent Accounts** serve as the root identity and primary financial account in the system. Each Parent account is:
- Identified by a [neuron public key](specs/002-key-library/spec.md) (secp256k1)
- Associated with a Decentralized Identifier (DID) document
- Capable of creating and managing Child accounts
- The primary recipient of revenue from data transactions performed by Child accounts
- The main account holder for financial assets (credit balance)

**Child Accounts** represent agent identities, devices, or service endpoints that operate under a Parent account. Each Child account:
- References its Parent account's [neuron public key](specs/002-key-library/spec.md) for hierarchical identification
- Does not require its own DID document (parent relationship provides identity resolution)
- Maintains a separate balance allocation for operational expenses such as transaction signing fees
- Provides communication endpoints for public and third-party communications (stdIn, stdOut, stdErr)
- Provides peer-to-peer connection endpoints (multiaddresses) for private, direct communications

**Ledger Attachment**: Accounts must be attached to a settlement infrastructure (e.g., a public DLT like Ethereum, Base, Hedera, a private ledger or just a centralised system, or any ledger system that works with the cryptographic key standards used by Neuron). The account object in this module represents an account that may or may not be attached to a ledger. When attached, the account links to a ledger account using the derived address (e.g., Ethereum address) as the account identifier. Note that some ledgers do not identify accounts by public key directly - multiple accounts can share the same public key, and the discriminator is the Ethereum address (or equivalent derived address). Therefore, the Ethereum address (or ledger-specific derived address) is adopted as the account identity for ledger attachment, not the public key itself. The attachment relationship must be provable - the module must provide methods to verify that the account exists on the specified ledger and that the Neuron private key corresponds to the ledger account (by proving ownership of the Ethereum address or derived address). Account creation on the ledger itself is deferred (out of scope); this module manages the object-level representation and attachment verification.

**Shared Accounts** represent multisig accounts that require threshold signatures (e.g., 2-of-3 multisig) for operations. Each Shared account:
- Uses [MultisigKey](specs/002-key-library/spec.md) with threshold configuration (e.g., 2-of-3)
- Requires multiple [neuron public keys](specs/002-key-library/spec.md) for threshold signing
- Can be attached to settlement infrastructure (ledgers) like Parent and Child accounts
- Maintains currency symbol and balance information linked to the attached ledger account
- Does NOT have communication channels (like Parent accounts)
- Does NOT have a DID document (like Child accounts)

**Currency and Balance**: Each account maintains a currency symbol identifier and balance information. The balance is linked to the attached ledger account. Parent accounts serve as the primary financial account holder, while Child accounts maintain separate balance allocations for operational expenses. Shared accounts maintain their own balance for multisig-controlled operations.

**Ledger Attachment Semantics Verification**: When an account is attached to a ledger, the system MUST verify that the public key semantics match between Neuron and the ledger. Specifically, the [neuron public key](specs/002-key-library/spec.md) that derives the Ethereum address in Neuron MUST have matching semantics with the ledger account. If the ledger account has a public key and a different Ethereum address (or derived address), they do NOT match and attachment verification MUST fail. Parent and Child accounts MUST have matching semantics - the public key → Ethereum address derivation must be consistent between the Neuron account object and the ledger account. For Shared accounts, the multisig threshold key configuration must match the ledger's multisig account structure.

This hierarchical model enables organizations to manage multiple agents or devices under a single Parent account while maintaining financial control and identity relationships. The Parent account receives revenue from data transactions, while Child accounts operate with allocated funds for their operational needs. Shared accounts provide additional security through threshold multisig requirements.

## Clarifications

### Session 2026-01-23

- Q: How should keys be referenced in this specification? → A: Use "neuron public key" terminology and link to the [Key Library specification](specs/002-key-library/spec.md) when discussing keys. This ensures consistency and provides developers with the authoritative source for key-related details.
- Q: Do Parent accounts have communication channels? → A: No. Parent accounts do NOT have communication channels (stdIn, stdOut, stdErr). However, Parent accounts have a credit balance and can receive money. Sending money (actual transfer execution) is NOT part of this spec. Signing transactions (including money transactions) is a capability but is out of scope and not discussed in this specification. Communication channels are only for Child accounts.
- Q: How are parent-child account associations handled? → A: Parent-child associations are maintained at the object level (in-memory representation) in this module. While DLTs may create accounts through sending money to non-existent accounts or explicit account creation primitives, and the association may exist in the ledger, this module keeps the association at object level. The module MUST provide a verification function that can verify the parent-child link against on-ledger data or other account registry systems if the underlying technology is not exactly a ledger.
- Q: How are accounts linked to ledgers? → A: Accounts must be attached to settlement infrastructure (public DLTs, private ledgers, or any ledger system compatible with Neuron's cryptographic key standards). The account object in this module may be attached or detached from a ledger. When attached, the account links to a ledger account using the derived address (e.g., Ethereum address) as the account identifier. Note that some ledgers do not identify accounts by public key directly - multiple accounts can share the same public key, and the discriminator is the Ethereum address. Therefore, the Ethereum address (or ledger-specific derived address) is adopted as the account identity for ledger attachment, not the public key itself. The module MUST provide methods to prove attachment by verifying: (1) the account exists on the specified ledger, (2) the Neuron private key corresponds to the ledger account by proving ownership of the Ethereum address or derived address, (3) the public key semantics match - the public key that derives the Ethereum address in Neuron must have matching semantics with the ledger account (if the ledger has a public key and a different Ethereum address, they do not match). Parent and Child accounts MUST have matching semantics. Account creation on the ledger is out of scope; this module manages object-level representation and attachment verification. Each account must have a currency symbol associated with the ledger and balance.
- Q: What are Shared accounts? → A: Shared accounts are multisig accounts that require threshold signatures (e.g., 2-of-3 multisig) for operations. They use [MultisigKey](specs/002-key-library/spec.md) with threshold configuration. Shared accounts can be attached to settlement infrastructure like Parent and Child accounts, maintain currency symbol and balance, but do NOT have communication channels (like Parent accounts) and do NOT have DID documents (like Child accounts). For Shared accounts attached to ledgers, the multisig threshold key configuration must match the ledger's multisig account structure.

## Out of Scope

The following are explicitly out of scope for the NeuronAccount Module:

- **Sending Money (Transfer Execution)**: While Parent accounts have credit balance and can receive money (balance increases), sending money (actual transfer execution) is NOT part of this specification. The module describes account identity and balance tracking, not transaction execution.

- **Transaction Signing**: Signing transactions (including money transactions) is a capability but is out of scope for this specification. Transaction signing functionality would be provided by other modules (e.g., the Key Library for cryptographic signing operations).

- **Account Creation on Ledger**: While DLTs may create accounts through sending money to non-existent accounts or explicit account creation primitives, actual account creation on the ledger is out of scope. This module works with account representations at the object level and manages attachment to existing ledger accounts. The module provides attachment verification but does not create accounts on the ledger.

- **Account Management Operations**: Account lifecycle operations beyond creation, validation, and relationship verification (e.g., updates, deletion) are not specified in this document.

- **Message Transport or Routing**: The module describes WHERE and HOW agents can be reached (communication addresses), but does NOT perform actual message transport or routing.

- **Key Generation or Storage**: Key management is handled by the [Key Library](specs/002-key-library/spec.md). This module uses keys but does not generate or store them.

- **Network Connectivity**: The module does not establish or manage network connections. It describes endpoints but does not perform connection operations.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Parent Account with Identity and Financial Management (Priority: P1)

A developer needs to create a root identity (Parent account) that serves as the primary account holder and can create and manage Child accounts. The Parent account includes cryptographic identity ([neuron public key](specs/002-key-library/spec.md)), DID document, and credit balance management. Parent accounts receive revenue from data transactions performed by their Child accounts. Parent accounts can receive money (balance increases), but sending money (actual transfer execution) is NOT part of this spec. Parent accounts do NOT have communication channels. Signing transactions (including money transactions) is a capability but is out of scope for this specification.

**Why this priority**: This is the foundational capability - without the ability to create parent accounts, no agent identities can exist. Parent accounts serve as the root of the hierarchical identity system and the primary financial account holder, receiving revenue from Child account operations.

**Independent Test**: Can be fully tested by creating a parent account with a [neuron public key](specs/002-key-library/spec.md), generating its DID, and verifying it has a credit balance. The test validates that the account structure is correct, the DID is properly formatted, and the account has credit balance tracking capabilities.

**Acceptance Scenarios**:

1. **Given** a developer has a [neuron public key](specs/002-key-library/spec.md) (secp256k1), **When** they create a Parent account, **Then** the account has a valid DID document and a credit balance (initialized to zero or specified amount)
2. **Given** a developer creates a Parent account, **When** they query the account, **Then** the account has no communication channels (stdIn, stdOut, stdErr are not present or are nil)
3. **Given** a developer creates a Parent account without a DID, **When** they attempt to validate the account, **Then** validation fails with clear error messages indicating that Parent accounts require a DID

---

### User Story 2 - Create Child Account for Agents and Devices with Relationship Verification (Priority: P2)

A developer needs to create a Child account that represents an agent, device, or service endpoint operating under a Parent account. Child accounts reference their Parent account's [neuron public key](specs/002-key-library/spec.md) for hierarchical identification, enabling identity resolution without requiring a full DID document. Child accounts maintain their own balance allocation for operational expenses (e.g., transaction signing fees) while the Parent account serves as the primary financial account holder. The parent-child association is maintained at the object level, and the system must provide verification capabilities to verify the relationship against on-ledger data or account registry systems.

**Why this priority**: Enables the hierarchical account model where agents, devices, and service endpoints operate under a Parent account. This structure supports organizational management, delegated operations, and financial control while allowing Child accounts to operate independently with allocated funds. Relationship verification ensures the object-level association can be validated against authoritative sources (ledgers or account registries).

**Independent Test**: Can be fully tested by creating a child account with a parent's [neuron public key](specs/002-key-library/spec.md) reference, then verifying the relationship. The test validates that the child account correctly references the parent at object level, does not require a DID, passes validation rules, and can verify the relationship against on-ledger or account registry data.

**Acceptance Scenarios**:

1. **Given** a developer has a Parent account's [neuron public key](specs/002-key-library/spec.md), **When** they create a Child account referencing that parent, **Then** the child account is created without a DID, correctly references the parent [neuron public key](specs/002-key-library/spec.md) at object level, and passes validation
2. **Given** a developer attempts to create a Child account, **When** they do not provide a parent [neuron public key](specs/002-key-library/spec.md), **Then** validation fails with an error indicating that child accounts require a parent reference
3. **Given** a developer creates both Parent and Child accounts, **When** they query the relationship, **Then** the system can identify that the child belongs to the specified parent at object level
4. **Given** a developer has Parent and Child accounts with an object-level association, **When** they verify the parent-child relationship, **Then** the system can verify the link against on-ledger data or account registry systems (if the underlying technology is not exactly a ledger)

---

### User Story 3 - Configure Communication Endpoints for Child Accounts with Extensible Backends (Priority: P1)

A developer needs to configure communication endpoints for Child accounts to enable public communications, third-party integrations, and private peer-to-peer communications. Child accounts require:
- Public/third-party communication channels (stdIn, stdOut, stdErr) using various messaging technologies (Hedera Consensus Service, Kafka, custom backends)
- Private peer-to-peer connection endpoints (multiaddresses) for direct agent-to-agent communications

The system uses an extensible plugin system that validates addresses according to each backend's rules. Parent accounts do NOT have communication channels.

**Why this priority**: Communication endpoints are essential for Child accounts - they enable agents and devices to participate in public communications, integrate with third-party services, and establish private peer-to-peer connections. The extensibility ensures the system can adapt to new messaging technologies without core changes.

**Independent Test**: Can be fully tested by registering a custom backend, creating communication addresses using that backend, and validating that addresses are checked against the backend's validation rules. The test demonstrates that new backends can be added without modifying core account code.

**Acceptance Scenarios**:

1. **Given** a developer wants to use a Hedera topic for communication, **When** they create a communication address with a Hedera topic locator (e.g., "0.0.12345"), **Then** the address is validated against Hedera backend rules and accepted if valid
2. **Given** a developer wants to use a Kafka topic, **When** they create a communication address with a Kafka topic name, **Then** the address is validated against Kafka backend rules and accepted if valid
3. **Given** a developer implements a custom backend plugin, **When** they register it and create addresses using that backend, **Then** the system uses the custom backend's validation rules without requiring core code changes
4. **Given** a developer provides an invalid locator format for a backend, **When** they attempt to create a communication address, **Then** validation fails with a backend-specific error message explaining the expected format

---

### User Story 4 - Validate Account Structure and Relationships (Priority: P2)

A developer needs to ensure that account structures are valid according to business rules (e.g., Parent accounts must have DIDs and credit balance capabilities, Child accounts must have parent references and communication addresses, Shared accounts must have multisig threshold keys, communication addresses must be valid, public key semantics must match for ledger attachments).

**Why this priority**: Validation ensures data integrity and prevents invalid account states, but is secondary to account creation capabilities. Public key semantics verification ensures consistency between Neuron accounts and ledger accounts.

**Independent Test**: Can be fully tested by creating various invalid account configurations and verifying that validation catches all rule violations with appropriate error messages, including public key semantics mismatches.

**Acceptance Scenarios**:

1. **Given** a developer creates a Parent account without a DID, **When** they attempt to validate the account, **Then** validation fails with an error indicating that Parent accounts require a DID
2. **Given** a developer creates a Child account without a parent [neuron public key](specs/002-key-library/spec.md), **When** they attempt to validate the account, **Then** validation fails with an error indicating that Child accounts require a parent reference
3. **Given** a developer creates a Shared account without a [MultisigKey](specs/002-key-library/spec.md), **When** they attempt to validate the account, **Then** validation fails with an error indicating that Shared accounts require a multisig threshold key
4. **Given** a developer creates an account with an invalid communication address, **When** they attempt to validate the account, **Then** validation fails with an error indicating which address is invalid and why
5. **Given** a developer creates a Parent or Child account attached to a ledger where the public key semantics do not match (e.g., the public key that derives the Ethereum address in Neuron differs from the ledger account's public key semantics), **When** they attempt to verify attachment, **Then** verification fails with a clear error indicating public key semantics mismatch
6. **Given** a developer creates a valid Parent account, **When** they validate it, **Then** validation passes and the account is marked as valid

---

### User Story 5 - Configure Private Peer-to-Peer Connection Endpoints (Priority: P3)

A developer needs to specify private peer-to-peer connection endpoints (reachable addresses) for Child accounts that include libp2p multiaddresses and PeerID information for direct, private agent-to-agent communication. These endpoints complement the public communication channels (stdIn, stdOut, stdErr) by enabling secure, direct connections between agents.

**Why this priority**: Private P2P endpoints enable secure direct communication between agents but are less critical than the core identity and public communication channel specification. This is an advanced feature that supports private communication scenarios.

**Independent Test**: Can be fully tested by adding reachable addresses with multiaddresses to an account and verifying that PeerID consistency is enforced and addresses are properly parsed.

**Acceptance Scenarios**:

1. **Given** a developer has a libp2p multiaddress, **When** they add it as a reachable address to an account, **Then** the address is parsed correctly, PeerID is extracted, and consistency is validated
2. **Given** a developer adds a reachable address with an inconsistent PeerID, **When** they attempt to validate the account, **Then** validation fails with an error indicating PeerID mismatch
3. **Given** a developer creates an account with multiple reachable addresses, **When** they query the account, **Then** all addresses are returned and can be used for connection attempts

---

### User Story 6 - Create Shared Account with Multisig Threshold Keys (Priority: P3)

A developer needs to create a Shared account (multisig threshold account) that requires threshold signatures (e.g., 2-of-3) for operations. Shared accounts use [MultisigKey](specs/002-key-library/spec.md) with threshold configuration and can be attached to settlement infrastructure like Parent and Child accounts.

**Why this priority**: Shared accounts provide additional security through threshold multisig requirements, enabling collaborative control over accounts. This is an advanced feature that supports multisig use cases.

**Independent Test**: Can be fully tested by creating a Shared account with a [MultisigKey](specs/002-key-library/spec.md) (e.g., 2-of-3 threshold), attaching it to a ledger, and verifying that the multisig threshold key configuration matches the ledger's multisig account structure.

**Acceptance Scenarios**:

1. **Given** a developer has multiple [neuron public keys](specs/002-key-library/spec.md) and wants to create a 2-of-3 multisig account, **When** they create a Shared account with a [MultisigKey](specs/002-key-library/spec.md) configured for 2-of-3 threshold, **Then** the Shared account is created with the multisig threshold key, has no DID document, has no communication channels, and passes validation
2. **Given** a developer creates a Shared account, **When** they attach it to a ledger, **Then** the system verifies that the multisig threshold key configuration matches the ledger's multisig account structure, and attachment succeeds if the configuration matches
3. **Given** a developer creates a Shared account with a multisig configuration that does not match the ledger's multisig account structure, **When** they attempt to verify attachment, **Then** verification fails with a clear error indicating the multisig configuration mismatch
4. **Given** a developer creates a Shared account without a [MultisigKey](specs/002-key-library/spec.md), **When** they attempt to validate the account, **Then** validation fails with an error indicating that Shared accounts require a multisig threshold key

---

### Edge Cases

- What happens when a developer creates an account with a communication address using an unregistered backend kind?
- How does the system handle Child accounts with missing required communication channels (stdIn, stdOut, or stdErr)? (Parent accounts do not have communication channels)
- What happens when a Child account references a parent [neuron public key](specs/002-key-library/spec.md) that doesn't correspond to any known Parent account at object level?
- What happens when verifying a parent-child relationship and the verification fails against on-ledger data or account registry? (Return verification failure with clear error indicating the relationship cannot be verified)
- How does the system handle parent-child relationship verification when the underlying technology is not a ledger? (Use account registry systems for verification)
- What happens when attempting to verify ledger attachment for an account that is not attached? (Return clear error indicating the account is in detached state)
- What happens when ledger attachment verification fails (e.g., account does not exist on ledger, private key does not match)? (Return verification failure with specific error indicating the reason: account not found, key mismatch, ledger connection failure, etc.)
- What happens when public key semantics do not match between Neuron and the ledger (e.g., ledger has a public key that derives a different Ethereum address)? (Return verification failure with clear error indicating public key semantics mismatch - the public key that derives the Ethereum address in Neuron must match the ledger account's public key semantics)
- How does the system handle Parent and Child accounts where the public key → Ethereum address derivation is inconsistent between Neuron and the ledger? (Verification MUST fail - Parent and Child accounts MUST have matching semantics)
- How does the system handle Shared accounts where the multisig threshold key configuration does not match the ledger's multisig account structure? (Verification MUST fail - Shared account multisig configuration must match the ledger)
- How does the system handle accounts with different currency symbols? (Currency symbol is associated with the ledger and balance; different ledgers may use different currency symbols)
- What happens when an account is attached to a ledger but the balance query fails? (Return error indicating balance query failure, but attachment state remains valid)
- How does the system handle invalid cryptographic key formats or sizes? (Keys must conform to the [Key Library specification](specs/002-key-library/spec.md))
- What happens when reachable addresses contain malformed multiaddresses?
- How does the system handle accounts where the PeerID derived from the [neuron public key](specs/002-key-library/spec.md) doesn't match the PeerID in reachable addresses?
- What happens when a developer attempts to create an account with an AccountType of Unspecified (0)? (Reject as invalid)
- What happens when a Shared account is created without a [MultisigKey](specs/002-key-library/spec.md)? (Validation fails with error indicating Shared accounts require multisig threshold keys)
- What happens when a Parent or Child account is created with a [MultisigKey](specs/002-key-library/spec.md) instead of a single [neuron public key](specs/002-key-library/spec.md)? (Validation fails - Parent and Child accounts use single keys, not multisig)
- What happens when a Shared account is created with a DID document or communication channels? (Validation fails - Shared accounts do not have DIDs or communication channels)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow developers to create Parent accounts with cryptographic identity ([neuron public key](specs/002-key-library/spec.md) - secp256k1), DID document, currency symbol, and credit balance management. Parent accounts serve as the primary financial account holder, receiving revenue from data transactions performed by Child accounts. Parent accounts can receive money (balance increases), but sending money (actual transfer execution) is NOT part of this spec. Parent accounts MUST be capable of creating and managing Child accounts. Parent accounts MUST NOT have communication channels. Parent accounts MUST support attachment to settlement infrastructure (ledgers) where the balance is linked. Signing transactions (including money transactions) is a capability but is out of scope for this specification
- **FR-002**: System MUST allow developers to create Child accounts that represent agents, devices, or service endpoints. Child accounts MUST reference their Parent account's [neuron public key](specs/002-key-library/spec.md) for hierarchical identification without requiring a DID document. Child accounts MUST maintain their own balance allocation with currency symbol for operational expenses (e.g., transaction signing fees). Child accounts MUST have communication channels (stdIn, stdOut, stdErr for public/third-party communications and multiaddresses for private communications). Child accounts MUST support attachment to settlement infrastructure (ledgers) where the balance is linked. Parent-child associations are maintained at the object level (in-memory representation)
- **FR-003**: System MUST support specifying communication addresses for Child accounts with three public/third-party communication channels: inbound messages (stdIn), outbound messages (stdOut), and error/diagnostic messages (stdErr). Child accounts MUST also support private peer-to-peer connection endpoints (multiaddresses). Parent accounts do NOT have communication channels
- **FR-004**: System MUST validate communication addresses according to backend-specific rules (e.g., Hedera topic format, Kafka topic naming conventions)
- **FR-005**: System MUST support an extensible backend plugin system where new messaging technologies can be registered without modifying core account code
- **FR-006**: System MUST validate that Parent accounts have a DID document, have a credit balance, have a single [neuron public key](specs/002-key-library/spec.md), and do NOT have communication channels, parent reference, or multisig keys. Sending money (actual transfer execution) is NOT part of this spec
- **FR-007**: System MUST validate that Child accounts have a parent [neuron public key](specs/002-key-library/spec.md) reference, have all three communication channels (stdIn, stdOut, stdErr), have a single [neuron public key](specs/002-key-library/spec.md), and do not require a DID document or multisig keys
- **FR-007a**: System MUST validate that Shared accounts have a [MultisigKey](specs/002-key-library/spec.md) with threshold configuration, do NOT have a DID document, do NOT have communication channels, do NOT have a parent reference, and do NOT have a single [neuron public key](specs/002-key-library/spec.md) (they use MultisigKey instead)
- **FR-008**: System MUST derive identity components (PeerID, EVMAddress) from the [neuron public key](specs/002-key-library/spec.md) (secp256k1) consistently
- **FR-009**: System MUST support specifying peer-to-peer connection endpoints (reachable addresses) with libp2p multiaddresses
- **FR-010**: System MUST enforce PeerID consistency between the account's derived PeerID and PeerIDs extracted from reachable addresses
- **FR-011**: System MUST provide a fluent builder API for constructing accounts with validation
- **FR-012**: System MUST generate DID:key format identifiers for Parent accounts
- **FR-013**: System MUST support AccountType enumeration with the following values: Unspecified (0 - invalid), Parent (1), Child (2), Shared (3). System MUST reject accounts with AccountType Unspecified (0) as invalid
- **FR-013a**: System MUST support three valid account types: Parent (1), Child (2), and Shared (3). Each account type has distinct characteristics: Parent accounts have DIDs, single [neuron public key](specs/002-key-library/spec.md), and can create Child accounts; Child accounts have parent references, single [neuron public key](specs/002-key-library/spec.md), and communication channels; Shared accounts use [MultisigKey](specs/002-key-library/spec.md) with threshold configuration and have neither DIDs nor communication channels
- **FR-014**: System MUST provide clear, actionable error messages when validation fails
- **FR-015**: System MUST support serialization and deserialization of account data structures
- **FR-016**: System MUST support credit balance management for Parent accounts, including initial balance setting, balance queries, and balance updates when money is received (e.g., revenue from data transactions performed by Child accounts). System MUST support balance allocation management for Child accounts, allowing Child accounts to maintain separate balance allocations for operational expenses. Sending money (actual transfer execution) is NOT part of this spec. Signing transactions (including money transactions) is a capability but is out of scope for this specification
- **FR-017**: System MUST provide a verification function to verify parent-child account relationships. The verification MUST be able to check the relationship against on-ledger data or other account registry systems if the underlying technology is not exactly a ledger. While DLTs may create accounts through sending money to non-existent accounts or explicit account creation primitives, and the association may exist in the ledger, this module maintains the association at object level and provides verification capabilities
- **FR-018**: System MUST support ledger attachment for accounts. Accounts MUST be linkable to accounts on settlement infrastructure (public DLTs, private ledgers, or any ledger system compatible with Neuron's cryptographic key standards). The link MUST use the derived address (e.g., Ethereum address) as the account identifier, not the public key directly. Note that some ledgers do not identify accounts by public key - multiple accounts can share the same public key, and the discriminator is the Ethereum address. Therefore, the Ethereum address (or ledger-specific derived address) is adopted as the account identity for ledger attachment. Accounts may be in an attached state (linked to a ledger account) or detached state (not linked). Account creation on the ledger is out of scope; this module manages object-level representation and attachment verification
- **FR-019**: System MUST provide methods to prove ledger attachment. The system MUST be able to verify that: (1) the account exists on the specified ledger (e.g., Ethereum mainnet, testnet, private ledger) using the Ethereum address or derived address as the identifier, (2) the Neuron private key corresponds to the ledger account by proving ownership of the Ethereum address or derived address, (3) the public key semantics match - the [neuron public key](specs/002-key-library/spec.md) that derives the Ethereum address in Neuron MUST have matching semantics with the ledger account (if the ledger account has a public key and a different Ethereum address, they do NOT match and verification MUST fail), and (4) the attachment link is valid. Verification MUST use the Ethereum address (or ledger-specific derived address) as the linking identifier, not the public key directly. Parent and Child accounts MUST have matching semantics between Neuron and the ledger. For Shared accounts, the multisig threshold key configuration MUST match the ledger's multisig account structure
- **FR-020**: System MUST support currency symbol identification for accounts. Each account MUST have a currency symbol that identifies the currency/asset type for balance tracking. The currency symbol is associated with the attached ledger and balance information
- **FR-021**: System MUST support Shared accounts (multisig threshold accounts). Shared accounts MUST use [MultisigKey](specs/002-key-library/spec.md) with threshold configuration (e.g., 2-of-3). Shared accounts MUST support attachment to settlement infrastructure (ledgers) where the balance is linked. Shared accounts MUST maintain currency symbol and balance information. Shared accounts MUST NOT have communication channels (like Parent accounts) and MUST NOT have DID documents (like Child accounts). When Shared accounts are attached to ledgers, the multisig threshold key configuration MUST match the ledger's multisig account structure

### Key Entities *(include if feature involves data)*

- **NeuronAccount**: Represents a complete agent identity including cryptographic identity, account type (Parent/Child/Shared), DID document (for Parents only), parent reference (for Children only), multisig threshold key (for Shared accounts), ledger attachment information, currency symbol, and account-specific attributes. For Parent accounts: credit balance (primary financial account holder, receives revenue from Child account data transactions; can receive money, but sending money is NOT part of this spec; transaction signing is out of scope), capability to create and manage Child accounts, ledger attachment, single [neuron public key](specs/002-key-library/spec.md). For Child accounts: balance allocation (operational expenses such as transaction signing fees), communication channels for public/third-party communications (stdIn, stdOut, stdErr), private peer-to-peer connection endpoints (reachable addresses with multiaddresses), ledger attachment, single [neuron public key](specs/002-key-library/spec.md), parent reference. For Shared accounts: balance (for multisig-controlled operations), ledger attachment, [MultisigKey](specs/002-key-library/spec.md) with threshold configuration (e.g., 2-of-3), no communication channels, no DID document. Parent-child associations are maintained at the object level (in-memory representation) and can be verified against on-ledger data or account registry systems. Accounts may be attached to settlement infrastructure (ledgers) or detached. When attached, the public key semantics MUST match between Neuron and the ledger (the public key that derives the Ethereum address in Neuron must have matching semantics with the ledger account). Key attributes: publicKey ([neuron public key](specs/002-key-library/spec.md) for Parent/Child), multisigKey ([MultisigKey](specs/002-key-library/spec.md) for Shared), peerID, evmAddress, accountType (Parent=1, Child=2, Shared=3), did (for Parents only), parentPubKey ([neuron public key](specs/002-key-library/spec.md) for Children only), currencySymbol, creditBalance (for Parents), balanceAllocation (for Children), balance (for Shared), ledgerAttachment (ledger identifier, attached/detached state, linked address), stdIn/stdOut/stdErr communication addresses (for Children only), reachableAddrs (for Children only).

- **CommAddress**: Represents a communication channel endpoint with a backend kind discriminator and technology-specific locator string. Key attributes: kind (backend identifier), locator (backend-specific address format). Relationships: associated with accounts via stdIn, stdOut, stdErr fields.

- **Backend**: Represents a messaging technology plugin that provides validation and parsing rules for communication addresses. Key attributes: kind (unique identifier), validation rules, metadata. Relationships: registered in a global registry, used by CommAddress validation.

- **ReachableAddr**: Represents a private peer-to-peer connection endpoint with a libp2p multiaddress and extracted PeerID for direct, private agent-to-agent communication. Key attributes: multiaddr (parsed address), peerID (extracted identifier). Relationships: part of a Child NeuronAccount's reachableAddrs collection (Parent accounts do not have reachable addresses).

- **NeuronDID**: Represents a decentralized identifier document for Parent accounts, providing a standard identity format. Key attributes: DID string, DID document structure. Relationships: required for Parent accounts, optional (nil) for Child accounts.

- **LedgerAttachment**: Represents the attachment relationship between a NeuronAccount and a settlement infrastructure (ledger). Key attributes: ledgerIdentifier (e.g., "ethereum-mainnet", "ethereum-testnet", "hedera-mainnet", custom ledger identifier), attachedAddress (the Ethereum address or ledger-specific derived address used to link to the ledger account - this is the account identifier, not the public key), attachmentState (attached/detached), verificationStatus. Relationships: each NeuronAccount has zero or one LedgerAttachment. The attachment links the account object to a ledger account using the Ethereum address (or ledger-specific derived address) as the account identifier. Note that some ledgers do not identify accounts by public key directly - multiple accounts can share the same public key, and the discriminator is the Ethereum address. The Neuron private key serves as proof of ownership of the Ethereum address and linkage between the NeuronAccount object and the ledger account.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can create a valid Parent account with all required components (identity, DID, currency symbol, credit balance management, Child account creation capability, ledger attachment support) in a single operation without encountering validation errors. Parent accounts do NOT have communication channels
- **SC-002**: Developers can create a valid Child account (representing an agent, device, or service endpoint) referencing a parent in a single operation without encountering validation errors. Child accounts have currency symbol, balance allocation for operational expenses, communication endpoints, and ledger attachment support. Parent-child associations are maintained at object level and can be verified against on-ledger or account registry data
- **SC-009**: System can verify ledger attachment for accounts, proving that: (1) the account exists on the specified ledger using the Ethereum address (or derived address) as the identifier, (2) the Neuron private key corresponds to the ledger account by proving ownership of the Ethereum address or derived address, and (3) the attachment link is valid. Verification succeeds for attached accounts and fails with clear errors for detached accounts or invalid attachments. Note that verification uses the Ethereum address (or ledger-specific derived address) as the account identifier, not the public key directly, since some ledgers allow multiple accounts to share the same public key
- **SC-003**: System validates communication addresses according to backend-specific rules with 100% accuracy (all valid addresses accepted, all invalid addresses rejected with appropriate errors)
- **SC-004**: New backend plugins can be registered and used for address validation without requiring modifications to core account package code
- **SC-005**: Account validation identifies all rule violations (missing DID for Parents, missing credit balance for Parents, communication channels on Parents, missing parent reference for Children, missing communication channels for Children, missing multisig key for Shared accounts, invalid addresses) with specific, actionable error messages
- **SC-010**: System verifies public key semantics match between Neuron accounts and ledger accounts. For Parent and Child accounts, the public key that derives the Ethereum address in Neuron MUST match the ledger account's public key semantics. For Shared accounts, the multisig threshold key configuration MUST match the ledger's multisig account structure. Verification fails with clear errors when semantics do not match
- **SC-006**: System enforces PeerID consistency between account identity and reachable addresses, rejecting mismatches with clear error messages
- **SC-007**: Developers can successfully serialize and deserialize account structures without data loss or corruption
- **SC-008**: Account creation and validation operations complete in under 100 milliseconds for typical account configurations (excluding network I/O for backend validation)
