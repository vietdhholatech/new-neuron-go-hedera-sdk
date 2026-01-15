package handlers

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// SignMessage godoc
// @Summary      Sign a message
// @Description  Sign a message with a private key (message is hashed with Keccak256)
// @Tags         signing
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SignMessageRequest  true  "Sign message request"
// @Success      200      {object}  dto.SignatureResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signing/sign-message [post]
func SignMessage(c *gin.Context) {
	var req dto.SignMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := privKey.SignMessage([]byte(req.Message))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	r, s, v := sig.RSV()
	c.JSON(http.StatusOK, dto.SignatureResponse{
		SignatureHex:    sig.Hex(),
		SignatureEthHex: "0x" + hex.EncodeToString(sig.EthereumBytes()),
		R:               "0x" + hex.EncodeToString(r.Bytes()),
		S:               "0x" + hex.EncodeToString(s.Bytes()),
		V:               int(v),
		VEthereum:       int(sig.VEthereum()),
	})
}

// SignDigest godoc
// @Summary      Sign a pre-hashed digest
// @Description  Sign a 32-byte digest with a private key
// @Tags         signing
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SignDigestRequest  true  "Sign digest request"
// @Success      200      {object}  dto.SignatureResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signing/sign-digest [post]
func SignDigest(c *gin.Context) {
	var req dto.SignDigestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
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

	sig, err := privKey.SignDigest(digest)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	r, s, v := sig.RSV()
	c.JSON(http.StatusOK, dto.SignatureResponse{
		SignatureHex:    sig.Hex(),
		SignatureEthHex: "0x" + hex.EncodeToString(sig.EthereumBytes()),
		R:               "0x" + hex.EncodeToString(r.Bytes()),
		S:               "0x" + hex.EncodeToString(s.Bytes()),
		V:               int(v),
		VEthereum:       int(sig.VEthereum()),
	})
}

// VerifyMessage godoc
// @Summary      Verify a message signature
// @Description  Verify a signature over a message using a public key
// @Tags         signing
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyMessageRequest  true  "Verify message request"
// @Success      200      {object}  dto.VerifyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signing/verify-message [post]
func VerifyMessage(c *gin.Context) {
	var req dto.VerifyMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	valid := pubKey.Verify([]byte(req.Message), sig)

	c.JSON(http.StatusOK, dto.VerifyResponse{
		Valid: valid,
	})
}

// VerifyDigest godoc
// @Summary      Verify a digest signature
// @Description  Verify a signature over a 32-byte digest using a public key
// @Tags         signing
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyDigestRequest  true  "Verify digest request"
// @Success      200      {object}  dto.VerifyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /signing/verify-digest [post]
func VerifyDigest(c *gin.Context) {
	var req dto.VerifyDigestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
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

	sig, err := keylib.ParseSignature(req.SignatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	valid := pubKey.VerifyDigest(digest, sig)

	c.JSON(http.StatusOK, dto.VerifyResponse{
		Valid: valid,
	})
}
