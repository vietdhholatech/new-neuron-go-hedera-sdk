package dto

// ===============================
// Key Operations
// ===============================

// ParsePrivateKeyRequest is the request for parsing a private key from hex.
type ParsePrivateKeyRequest struct {
	Hex string `json:"hex" binding:"required" example:"0x1234567890abcdef..."`
}

// ParsePublicKeyRequest is the request for parsing a public key from hex.
type ParsePublicKeyRequest struct {
	Hex string `json:"hex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
}

// ===============================
// Derivation Operations
// ===============================

// DeriveFromPrivateKeyRequest is the request for deriving values from a private key.
type DeriveFromPrivateKeyRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
}

// DeriveFromPublicKeyRequest is the request for deriving values from a public key.
type DeriveFromPublicKeyRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
}

// ===============================
// Signing Operations
// ===============================

// SignMessageRequest is the request for signing a message.
type SignMessageRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
	Message       string `json:"message" binding:"required" example:"Hello, World!"`
}

// SignDigestRequest is the request for signing a pre-hashed digest.
type SignDigestRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
	DigestHex     string `json:"digestHex" binding:"required,len=66" example:"0x1234567890abcdef...32bytes..."`
}

// VerifyMessageRequest is the request for verifying a message signature.
type VerifyMessageRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	Message      string `json:"message" binding:"required" example:"Hello, World!"`
	SignatureHex string `json:"signatureHex" binding:"required" example:"0x...65bytes..."`
}

// VerifyDigestRequest is the request for verifying a digest signature.
type VerifyDigestRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	DigestHex    string `json:"digestHex" binding:"required,len=66" example:"0x1234567890abcdef...32bytes..."`
	SignatureHex string `json:"signatureHex" binding:"required" example:"0x...65bytes..."`
}

// ===============================
// Signature Operations
// ===============================

// ParseSignatureRequest is the request for parsing a signature from hex.
type ParseSignatureRequest struct {
	SignatureHex string `json:"signatureHex" binding:"required" example:"0x...65bytes..."`
}

// RecoverPublicKeyRequest is the request for recovering a public key from a signed message.
type RecoverPublicKeyRequest struct {
	Message      string `json:"message" binding:"required" example:"Hello, World!"`
	SignatureHex string `json:"signatureHex" binding:"required" example:"0x...65bytes..."`
}

// RecoverFromDigestRequest is the request for recovering a public key from a signed digest.
type RecoverFromDigestRequest struct {
	DigestHex    string `json:"digestHex" binding:"required,len=66" example:"0x1234567890abcdef...32bytes..."`
	SignatureHex string `json:"signatureHex" binding:"required" example:"0x...65bytes..."`
}

// ===============================
// Mnemonic Operations
// ===============================

// GenerateMnemonicRequest is the request for generating a mnemonic.
type GenerateMnemonicRequest struct {
	WordCount int `json:"wordCount" binding:"required,oneof=12 15 18 21 24" example:"12"`
}

// ValidateMnemonicRequest is the request for validating a mnemonic.
type ValidateMnemonicRequest struct {
	Mnemonic string `json:"mnemonic" binding:"required" example:"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"`
}

// DeriveFromMnemonicRequest is the request for deriving a key from a mnemonic.
type DeriveFromMnemonicRequest struct {
	Mnemonic string `json:"mnemonic" binding:"required" example:"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"`
}

// DeriveFromMnemonicFullRequest is the request for deriving a key with options.
type DeriveFromMnemonicFullRequest struct {
	Mnemonic   string `json:"mnemonic" binding:"required" example:"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"`
	Passphrase string `json:"passphrase,omitempty" example:""`
	Path       string `json:"path,omitempty" example:"m/44'/60'/0'/0/0"`
}

// ===============================
// Encryption Operations
// ===============================

// ScrambleRequest is the request for encrypting a private key.
type ScrambleRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
	Password      string `json:"password" binding:"required,min=8" example:"mypassword123"`
}

// UnscrambleRequest is the request for decrypting an encrypted private key.
type UnscrambleRequest struct {
	Encrypted EncryptedKeyDTO `json:"encrypted" binding:"required"`
	Password  string          `json:"password" binding:"required" example:"mypassword123"`
}

// EncryptedKeyDTO represents an encrypted private key.
type EncryptedKeyDTO struct {
	Version    int    `json:"version" example:"1"`
	Salt       string `json:"salt" example:"base64..."`
	Nonce      string `json:"nonce" example:"base64..."`
	Ciphertext string `json:"ciphertext" example:"base64..."`
}

// ===============================
// Matching Operations
// ===============================

// MatchPrivatePublicRequest is the request for checking if private matches public key.
type MatchPrivatePublicRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
	PublicKeyHex  string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
}

// MatchPrivateEVMRequest is the request for checking if private key matches EVM address.
type MatchPrivateEVMRequest struct {
	PrivateKeyHex string `json:"privateKeyHex" binding:"required" example:"0x1234567890abcdef..."`
	EVMAddress    string `json:"evmAddress" binding:"required" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
}

// MatchPublicEVMRequest is the request for checking if public key matches EVM address.
type MatchPublicEVMRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	EVMAddress   string `json:"evmAddress" binding:"required" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
}

// MatchPublicPeerIDRequest is the request for checking if public key matches peer ID.
type MatchPublicPeerIDRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	PeerID       string `json:"peerID" binding:"required" example:"16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"`
}

// ===============================
// Identifier Operations
// ===============================

// ParseEVMAddressRequest is the request for parsing an EVM address.
type ParseEVMAddressRequest struct {
	Address string `json:"address" binding:"required" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
}

// ParsePeerIDRequest is the request for parsing a peer ID.
type ParsePeerIDRequest struct {
	PeerID string `json:"peerID" binding:"required" example:"16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"`
}

// ===============================
// Hedera Operations
// ===============================

// HederaPrivateKeyRequest is the request containing a Hedera DER-encoded private key string.
type HederaPrivateKeyRequest struct {
	HederaKeyString string `json:"hederaKeyString" binding:"required" example:"302e020100300506032b6570042204..."`
}

// HederaPublicKeyRequest is the request containing a Hedera DER-encoded public key string.
type HederaPublicKeyRequest struct {
	HederaKeyString string `json:"hederaKeyString" binding:"required" example:"302a300506032b6570032100..."`
}

// ===============================
// Bytes Operations
// ===============================

// PrivateKeyFromBytesRequest is the request for constructing a private key from raw bytes.
type PrivateKeyFromBytesRequest struct {
	Bytes string `json:"bytes" binding:"required" example:"base64-encoded-32-bytes"`
}

// PublicKeyFromBytesRequest is the request for constructing a public key from raw bytes.
type PublicKeyFromBytesRequest struct {
	Bytes string `json:"bytes" binding:"required" example:"base64-encoded-33-or-65-bytes"`
}

// ExportBytesRequest is the request for exporting key bytes.
type ExportBytesRequest struct {
	PublicKeyHex string `json:"publicKeyHex" binding:"required" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
}
