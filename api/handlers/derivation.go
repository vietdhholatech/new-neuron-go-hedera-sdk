package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// DerivePublicKey godoc
// @Summary      Derive public key from private key
// @Description  Derive the public key from a private key hex string
// @Tags         derive
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPrivateKeyRequest  true  "Private key hex"
// @Success      200      {object}  dto.PublicKeyResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /derive/public-key [post]
func DerivePublicKey(c *gin.Context) {
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

	pubKey := privKey.PublicKey()
	peerID, _ := pubKey.PeerID()

	c.JSON(http.StatusOK, dto.PublicKeyResponse{
		PublicKeyHex:          pubKey.Hex(),
		PublicKeyUncompressed: pubKey.HexUncompressed(),
		EVMAddress:            pubKey.EVMAddress().Hex(),
		EVMAddressChecksum:    pubKey.EVMAddress().ChecksumHex(),
		PeerID:                peerID.String(),
	})
}

// DeriveEVMAddress godoc
// @Summary      Derive EVM address from public key
// @Description  Derive the Ethereum address from a public key hex string
// @Tags         derive
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPublicKeyRequest  true  "Public key hex"
// @Success      200      {object}  dto.EVMAddressResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /derive/evm-address [post]
func DeriveEVMAddress(c *gin.Context) {
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

	evmAddr := pubKey.EVMAddress()

	c.JSON(http.StatusOK, dto.EVMAddressResponse{
		Address:         evmAddr.Hex(),
		AddressChecksum: evmAddr.ChecksumHex(),
		IsZero:          evmAddr.IsZero(),
	})
}

// DeriveEVMAddressSafe godoc
// @Summary      Derive EVM address with error checking
// @Description  Derive the Ethereum address from a public key with explicit error handling for zero keys
// @Tags         derive
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPublicKeyRequest  true  "Public key hex"
// @Success      200      {object}  dto.EVMAddressResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /derive/evm-address-safe [post]
func DeriveEVMAddressSafe(c *gin.Context) {
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

	evmAddr, err := pubKey.EVMAddressSafe()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.EVMAddressResponse{
		Address:         evmAddr.Hex(),
		AddressChecksum: evmAddr.ChecksumHex(),
		IsZero:          evmAddr.IsZero(),
	})
}

// DerivePeerID godoc
// @Summary      Derive libp2p peer ID from public key
// @Description  Derive the libp2p peer ID from a public key hex string
// @Tags         derive
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromPublicKeyRequest  true  "Public key hex"
// @Success      200      {object}  dto.PeerIDResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /derive/peer-id [post]
func DerivePeerID(c *gin.Context) {
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

	peerID, err := pubKey.PeerID()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PeerIDResponse{
		PeerID: peerID.String(),
	})
}
