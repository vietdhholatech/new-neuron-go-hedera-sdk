package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// ScrambleKey godoc
// @Summary      Encrypt a private key
// @Description  Encrypt a private key using AES-256-GCM with Argon2id key derivation
// @Tags         encryption
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ScrambleRequest  true  "Private key and password"
// @Success      200      {object}  dto.ScrambleResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /encryption/scramble [post]
func ScrambleKey(c *gin.Context) {
	var req dto.ScrambleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	encrypted, err := privKey.Scramble(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ScrambleResponse{
		Encrypted: dto.EncryptedKeyDTO{
			Version:    encrypted.Version,
			Salt:       base64.StdEncoding.EncodeToString(encrypted.Salt),
			Nonce:      base64.StdEncoding.EncodeToString(encrypted.Nonce),
			Ciphertext: base64.StdEncoding.EncodeToString(encrypted.Ciphertext),
		},
	})
}

// UnscrambleKey godoc
// @Summary      Decrypt a private key
// @Description  Decrypt an encrypted private key using the password
// @Tags         encryption
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UnscrambleRequest  true  "Encrypted key and password"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /encryption/unscramble [post]
func UnscrambleKey(c *gin.Context) {
	var req dto.UnscrambleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Decode base64 fields
	salt, err := base64.StdEncoding.DecodeString(req.Encrypted.Salt)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid salt: " + err.Error()})
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(req.Encrypted.Nonce)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid nonce: " + err.Error()})
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(req.Encrypted.Ciphertext)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid ciphertext: " + err.Error()})
		return
	}

	encrypted := keylib.EncryptedPrivateKey{
		Version:    req.Encrypted.Version,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}

	privKey, err := keylib.UnscramblePrivateKey(encrypted, req.Password)
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
