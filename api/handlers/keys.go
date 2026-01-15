package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// GenerateKey godoc
// @Summary      Generate new keypair
// @Description  Generate a new random ECDSA secp256k1 private key and derive all identifiers
// @Tags         keys
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.KeyPairResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /keys/generate [post]
func GenerateKey(c *gin.Context) {
	privKey, err := keylib.GeneratePrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
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

// ParsePrivateKey godoc
// @Summary      Parse private key from hex
// @Description  Parse and validate a private key from hex string
// @Tags         keys
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParsePrivateKeyRequest  true  "Private key hex"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /keys/parse/private [post]
func ParsePrivateKey(c *gin.Context) {
	var req dto.ParsePrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.Hex)
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

// ParsePublicKey godoc
// @Summary      Parse public key from hex
// @Description  Parse and validate a public key from hex string (compressed or uncompressed)
// @Tags         keys
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParsePublicKeyRequest  true  "Public key hex"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /keys/parse/public [post]
func ParsePublicKey(c *gin.Context) {
	var req dto.ParsePublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.Hex)
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
