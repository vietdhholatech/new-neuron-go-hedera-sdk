package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// ToHederaPrivateKey godoc
// @Summary      Convert Neuron private key to Hedera format
// @Description  Convert a Neuron private key (hex) to Hedera SDK DER-encoded string format
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPrivateKeyRequest  true  "Private key hex"
// @Success      200      {object}  dto.HederaKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/to-private-key [post]
func ToHederaPrivateKey(c *gin.Context) {
	var req dto.DeriveFromPrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	hederaKey, err := privKey.ToHederaPrivateKeySafe()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.HederaKeyResponse{
		HederaKeyString: hederaKey.String(),
		KeyType:         "ECDSA_SECP256K1",
	})
}

// ToHederaPublicKey godoc
// @Summary      Convert Neuron public key to Hedera format
// @Description  Convert a Neuron public key (hex) to Hedera SDK DER-encoded string format
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPublicKeyRequest  true  "Public key hex"
// @Success      200      {object}  dto.HederaKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/to-public-key [post]
func ToHederaPublicKey(c *gin.Context) {
	var req dto.DeriveFromPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	hederaKey, err := pubKey.ToHederaPublicKeySafe()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.HederaKeyResponse{
		HederaKeyString: hederaKey.String(),
		KeyType:         "ECDSA_SECP256K1",
	})
}

// FromHederaPrivateKey godoc
// @Summary      Elevate Hedera private key to Neuron format
// @Description  Convert a Hedera SDK DER-encoded private key string to Neuron format. Rejects Ed25519 keys.
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.HederaPrivateKeyRequest  true  "Hedera DER-encoded key string"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/from-private-key [post]
func FromHederaPrivateKey(c *gin.Context) {
	var req dto.HederaPrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Parse Hedera key from DER string
	hederaKey, err := hiero.PrivateKeyFromString(req.HederaKeyString)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid Hedera key string: " + err.Error()})
		return
	}

	// Convert to Neuron key (will reject Ed25519)
	privKey, err := keylib.PrivateKeyFromHedera(hederaKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey := privKey.PublicKey()
	peerID, _ := pubKey.PeerID()

	c.JSON(http.StatusOK, dto.KeyPairResponse{
		PrivateKeyHex:         privKey.Hex(),
		PublicKeyHex:          pubKey.Hex(),
		PublicKeyUncompressed: pubKey.HexUncompressed(),
		EVMAddress:            pubKey.EVMAddress().Hex(),
		EVMAddressChecksum:    pubKey.EVMAddress().ChecksumHex(),
		PeerID:                peerID.String(),
	})
}

// FromHederaPublicKey godoc
// @Summary      Elevate Hedera public key to Neuron format
// @Description  Convert a Hedera SDK DER-encoded public key string to Neuron format. Rejects Ed25519 keys.
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.HederaPublicKeyRequest  true  "Hedera DER-encoded key string"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/from-public-key [post]
func FromHederaPublicKey(c *gin.Context) {
	var req dto.HederaPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Parse Hedera key from DER string
	hederaKey, err := hiero.PublicKeyFromString(req.HederaKeyString)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid Hedera key string: " + err.Error()})
		return
	}

	// Convert to Neuron key (will reject Ed25519)
	pubKey, err := keylib.PublicKeyFromHedera(hederaKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	peerID, _ := pubKey.PeerID()

	c.JSON(http.StatusOK, dto.PublicKeyResponse{
		PublicKeyHex:          pubKey.Hex(),
		PublicKeyUncompressed: pubKey.HexUncompressed(),
		EVMAddress:            pubKey.EVMAddress().Hex(),
		EVMAddressChecksum:    pubKey.EVMAddress().ChecksumHex(),
		PeerID:                peerID.String(),
	})
}

// DetectKeyType godoc
// @Summary      Detect Hedera private key type
// @Description  Detect if a Hedera private key is Ed25519 or ECDSA secp256k1
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.HederaPrivateKeyRequest  true  "Hedera DER-encoded key string"
// @Success      200      {object}  dto.KeyTypeResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/detect-key-type [post]
func DetectKeyType(c *gin.Context) {
	var req dto.HederaPrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Parse Hedera key from DER string
	hederaKey, err := hiero.PrivateKeyFromString(req.HederaKeyString)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid Hedera key string: " + err.Error()})
		return
	}

	isEd25519 := keylib.IsEd25519Key(hederaKey)

	var keyType, description string
	if isEd25519 {
		keyType = "ED25519"
		description = "Ed25519 key - NOT compatible with Ethereum, cannot derive EVM address"
	} else {
		keyType = "ECDSA_SECP256K1"
		description = "ECDSA secp256k1 key - compatible with Ethereum, can derive EVM address"
	}

	c.JSON(http.StatusOK, dto.KeyTypeResponse{
		KeyType:     keyType,
		IsEd25519:   isEd25519,
		IsECDSA:     !isEd25519,
		Description: description,
	})
}

// DetectPublicKeyType godoc
// @Summary      Detect Hedera public key type
// @Description  Detect if a Hedera public key is Ed25519 or ECDSA secp256k1
// @Tags         hedera
// @Accept       json
// @Produce      json
// @Param        request  body      dto.HederaPublicKeyRequest  true  "Hedera DER-encoded key string"
// @Success      200      {object}  dto.KeyTypeResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /hedera/detect-public-key-type [post]
func DetectPublicKeyType(c *gin.Context) {
	var req dto.HederaPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Parse Hedera key from DER string
	hederaKey, err := hiero.PublicKeyFromString(req.HederaKeyString)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid Hedera key string: " + err.Error()})
		return
	}

	isEd25519 := keylib.IsEd25519PublicKey(hederaKey)

	var keyType, description string
	if isEd25519 {
		keyType = "ED25519"
		description = "Ed25519 public key - NOT compatible with Ethereum, cannot derive EVM address"
	} else {
		keyType = "ECDSA_SECP256K1"
		description = "ECDSA secp256k1 public key - compatible with Ethereum, can derive EVM address"
	}

	c.JSON(http.StatusOK, dto.KeyTypeResponse{
		KeyType:     keyType,
		IsEd25519:   isEd25519,
		IsECDSA:     !isEd25519,
		Description: description,
	})
}
