package handlers

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// PrivateKeyFromBytes godoc
// @Summary      Construct private key from raw bytes
// @Description  Construct a private key from base64-encoded raw 32-byte value
// @Tags         bytes
// @Accept       json
// @Produce      json
// @Param        request  body      dto.PrivateKeyFromBytesRequest  true  "Base64-encoded 32 bytes"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /bytes/private-key-from-bytes [post]
func PrivateKeyFromBytes(c *gin.Context) {
	var req dto.PrivateKeyFromBytesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Decode base64 bytes
	rawBytes, err := base64.StdEncoding.DecodeString(req.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid base64: " + err.Error()})
		return
	}

	if len(rawBytes) != 32 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "private key must be exactly 32 bytes"})
		return
	}

	var keyBytes [32]byte
	copy(keyBytes[:], rawBytes)

	privKey, err := keylib.PrivateKeyFromBytes(keyBytes)
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

// PublicKeyFromBytes godoc
// @Summary      Construct public key from raw bytes
// @Description  Construct a public key from base64-encoded raw SEC1 bytes (33 compressed or 65 uncompressed)
// @Tags         bytes
// @Accept       json
// @Produce      json
// @Param        request  body      dto.PublicKeyFromBytesRequest  true  "Base64-encoded 33 or 65 bytes"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /bytes/public-key-from-bytes [post]
func PublicKeyFromBytes(c *gin.Context) {
	var req dto.PublicKeyFromBytesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Decode base64 bytes
	rawBytes, err := base64.StdEncoding.DecodeString(req.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid base64: " + err.Error()})
		return
	}

	if len(rawBytes) != 33 && len(rawBytes) != 65 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "public key must be 33 bytes (compressed) or 65 bytes (uncompressed)"})
		return
	}

	pubKey, err := keylib.PublicKeyFromBytes(rawBytes)
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

// ExportPublicKeyBytes godoc
// @Summary      Export public key as raw bytes
// @Description  Export a public key in multiple formats (compressed/uncompressed, base64/hex)
// @Tags         bytes
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ExportBytesRequest  true  "Public key hex"
// @Success      200      {object}  dto.KeyBytesResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /bytes/export-public-key [post]
func ExportPublicKeyBytes(c *gin.Context) {
	var req dto.ExportBytesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	compressed := pubKey.CompressedBytes()
	uncompressed := pubKey.UncompressedBytes()

	c.JSON(http.StatusOK, dto.KeyBytesResponse{
		CompressedBase64:   base64.StdEncoding.EncodeToString(compressed[:]),
		UncompressedBase64: base64.StdEncoding.EncodeToString(uncompressed[:]),
		CompressedHex:      "0x" + hex.EncodeToString(compressed[:]),
		UncompressedHex:    "0x" + hex.EncodeToString(uncompressed[:]),
	})
}
