package dto

// ===============================
// Common Responses
// ===============================

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid hex format"`
}

// ===============================
// Key Responses
// ===============================

// KeyPairResponse is the response containing a full key pair with derived identifiers.
type KeyPairResponse struct {
	PrivateKeyHex         string `json:"privateKeyHex" example:"0x1234567890abcdef..."`
	PublicKeyHex          string `json:"publicKeyHex" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	PublicKeyUncompressed string `json:"publicKeyHexUncompressed" example:"0x04759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae..."`
	EVMAddress            string `json:"evmAddress" example:"0xe364f2f1e5f4f03d1df682322500b9c68c997ec3"`
	EVMAddressChecksum    string `json:"evmAddressChecksum" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
	PeerID                string `json:"peerID" example:"16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"`
}

// PublicKeyResponse is the response containing public key and derived identifiers.
type PublicKeyResponse struct {
	PublicKeyHex          string `json:"publicKeyHex" example:"0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"`
	PublicKeyUncompressed string `json:"publicKeyHexUncompressed" example:"0x04759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae..."`
	EVMAddress            string `json:"evmAddress" example:"0xe364f2f1e5f4f03d1df682322500b9c68c997ec3"`
	EVMAddressChecksum    string `json:"evmAddressChecksum" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
	PeerID                string `json:"peerID" example:"16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"`
}

// ===============================
// Derivation Responses
// ===============================

// EVMAddressResponse is the response containing an EVM address.
type EVMAddressResponse struct {
	Address         string `json:"address" example:"0xe364f2f1e5f4f03d1df682322500b9c68c997ec3"`
	AddressChecksum string `json:"addressChecksum" example:"0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"`
	IsZero          bool   `json:"isZero" example:"false"`
}

// PeerIDResponse is the response containing a peer ID.
type PeerIDResponse struct {
	PeerID string `json:"peerID" example:"16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"`
}

// ===============================
// Signature Responses
// ===============================

// SignatureResponse is the response containing a signature with all formats.
type SignatureResponse struct {
	SignatureHex    string `json:"signatureHex" example:"0x...65bytes..."`
	SignatureEthHex string `json:"signatureEthHex" example:"0x...65bytes..."`
	R               string `json:"r" example:"0x..."`
	S               string `json:"s" example:"0x..."`
	V               int    `json:"v" example:"0"`
	VEthereum       int    `json:"vEthereum" example:"27"`
}

// VerifyResponse is the response for signature verification.
type VerifyResponse struct {
	Valid bool `json:"valid" example:"true"`
}

// SignatureComponentsResponse is the response containing signature components.
type SignatureComponentsResponse struct {
	R         string `json:"r" example:"0x..."`
	S         string `json:"s" example:"0x..."`
	V         int    `json:"v" example:"0"`
	VEthereum int    `json:"vEthereum" example:"27"`
	IsZero    bool   `json:"isZero" example:"false"`
}

// ===============================
// Mnemonic Responses
// ===============================

// MnemonicResponse is the response containing a generated mnemonic.
type MnemonicResponse struct {
	Mnemonic  string `json:"mnemonic" example:"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"`
	WordCount int    `json:"wordCount" example:"12"`
}

// ValidateResponse is the response for mnemonic validation.
type ValidateResponse struct {
	Valid bool   `json:"valid" example:"true"`
	Error string `json:"error,omitempty" example:""`
}

// ===============================
// Encryption Responses
// ===============================

// ScrambleResponse is the response containing an encrypted private key.
type ScrambleResponse struct {
	Encrypted EncryptedKeyDTO `json:"encrypted"`
}

// ===============================
// Matching Responses
// ===============================

// MatchResponse is the response for matching operations.
type MatchResponse struct {
	Matches bool `json:"matches" example:"true"`
}

// ===============================
// Hedera Responses
// ===============================

// HederaKeyResponse is the response containing Hedera SDK key representations.
type HederaKeyResponse struct {
	HederaKeyString string `json:"hederaKeyString" example:"302e020100300506032b6570042204..."`
	KeyType         string `json:"keyType" example:"ECDSA_SECP256K1"`
}

// KeyTypeResponse is the response for key type detection.
type KeyTypeResponse struct {
	KeyType     string `json:"keyType" example:"ECDSA_SECP256K1"`
	IsEd25519   bool   `json:"isEd25519" example:"false"`
	IsECDSA     bool   `json:"isECDSA" example:"true"`
	Description string `json:"description" example:"ECDSA secp256k1 key compatible with Ethereum"`
}

// ===============================
// Bytes Responses
// ===============================

// KeyBytesResponse is the response containing raw key bytes in multiple formats.
type KeyBytesResponse struct {
	CompressedBase64   string `json:"compressedBase64" example:"base64..."`
	UncompressedBase64 string `json:"uncompressedBase64" example:"base64..."`
	CompressedHex      string `json:"compressedHex" example:"0x02..."`
	UncompressedHex    string `json:"uncompressedHex" example:"0x04..."`
}
