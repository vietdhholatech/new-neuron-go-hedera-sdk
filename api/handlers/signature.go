package handlers

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// ParseSignature godoc
// @Summary      Parse a signature from hex
// @Description  Parse and validate a signature from hex string (65 bytes: R || S || V)
// @Tags         signatures
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParseSignatureRequest  true  "Signature hex"
// @Success      200      {object}  dto.SignatureComponentsResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signatures/parse [post]
func ParseSignature(c *gin.Context) {
	var req dto.ParseSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	r, s, v := sig.RSV()
	c.JSON(http.StatusOK, dto.SignatureComponentsResponse{
		R:         "0x" + hex.EncodeToString(r.Bytes()),
		S:         "0x" + hex.EncodeToString(s.Bytes()),
		V:         int(v),
		VEthereum: int(sig.VEthereum()),
		IsZero:    sig.IsZero(),
	})
}

// RecoverPublicKey godoc
// @Summary      Recover public key from signed message
// @Description  Recover the public key that signed a message from the signature
// @Tags         signatures
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RecoverPublicKeyRequest  true  "Recover request"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signatures/recover-pubkey [post]
func RecoverPublicKey(c *gin.Context) {
	var req dto.RecoverPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.RecoverPublicKey([]byte(req.Message), sig)
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

// RecoverFromDigest godoc
// @Summary      Recover public key from signed digest
// @Description  Recover the public key that signed a 32-byte digest from the signature
// @Tags         signatures
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RecoverFromDigestRequest  true  "Recover from digest request"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signatures/recover-from-digest [post]
func RecoverFromDigest(c *gin.Context) {
	var req dto.RecoverFromDigestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	digestHex := strings.TrimPrefix(req.DigestHex, "0x")
	digestBytes, err := hex.DecodeString(digestHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid digest hex: " + err.Error()})
		return
	}

	if len(digestBytes) != 32 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "digest must be exactly 32 bytes"})
		return
	}

	var digest [32]byte
	copy(digest[:], digestBytes)

	pubKey, err := keylib.RecoverPublicKeyFromDigest(digest, sig)
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

// GetSignatureComponents godoc
// @Summary      Get signature components
// @Description  Parse a signature and return its R, S, V components
// @Tags         signatures
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParseSignatureRequest  true  "Signature hex"
// @Success      200      {object}  dto.SignatureComponentsResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signatures/components [post]
func GetSignatureComponents(c *gin.Context) {
	var req dto.ParseSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	r, s, v := sig.RSV()
	c.JSON(http.StatusOK, dto.SignatureComponentsResponse{
		R:         "0x" + hex.EncodeToString(r.Bytes()),
		S:         "0x" + hex.EncodeToString(s.Bytes()),
		V:         int(v),
		VEthereum: int(sig.VEthereum()),
		IsZero:    sig.IsZero(),
	})
}
