# NeuronAccount Implementation Gaps Analysis

## Date: 2026-01-27

## Summary
The current implementation matches v1.3.0 spec but has gaps vs the new NeuronAccount.md feature specification.

## Critical Missing Features

### 1. Shared Account Type (AccountType = 3)
- Location: account/account_type.go
- Missing: `AccountTypeShared AccountType = 3`
- Missing: NewSharedAccountBuilder
- Missing: Shared-specific validation

### 2. MultisigKey in keylib
- Location: keylib/ (needs new file)
- Purpose: Threshold signing (e.g., 2-of-3)
- Required fields: public keys list, threshold (M-of-N)

### 3. LedgerAttachment Entity
- Location: account/ (needs new file)
- Fields needed:
  - ledgerIdentifier (string)
  - attachedAddress (EVMAddress or derived)
  - attachmentState (attached/detached)
  - verificationStatus
- Methods needed:
  - VerifyAttachment()
  - ProveOwnership()
  - VerifySemantics()

### 4. Financial Fields in NeuronAccount
- currencySymbol (string) - all accounts
- creditBalance (*big.Int) - Parent only
- balanceAllocation (*big.Int) - Child only
- balance (*big.Int) - Shared only

## Specification Conflicts

### Parent Account Comm Channels
- Spec: Parent MUST NOT have communication channels
- Code: Parent CAN have communication channels
- Action: Add validation to reject comm channels on Parent

### Child Account Comm Channels
- Spec: Child MUST have all 3 communication channels
- Code: All channels are optional
- Action: Add validation to require all 3 for Child

## Files to Modify/Create

1. account/account_type.go - Add AccountTypeShared
2. account/account.go - Add financial/ledger fields
3. account/ledger_attachment.go - NEW FILE
4. account/builder.go - Add NewSharedAccountBuilder
5. account/validation.go - Add new validation rules
6. keylib/multisig_key.go - NEW FILE

## Related Documents
- docs/neuron-sdk-spec/NeuronAccount.md (new spec)
- docs/neuron-account-specification.md (current spec v1.3.0)
