package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// PrivateMatchesPublic godoc
// @Summary      Check if private key matches public key
// @Description  Verify that a private key corresponds to a given public key
// @Tags         matching
// @Accept       json
// @Produce      json
// @Param        request  body      dto.MatchPrivatePublicRequest  true  "Private and public keys"
// @Success      200      {object}  dto.MatchResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /matching/private-matches-public [post]
func PrivateMatchesPublic(c *gin.Context) {
	var req dto.MatchPrivatePublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid private key: " + err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid public key: " + err.Error()})
		return
	}

	matches := privKey.MatchesPublicKey(pubKey)

	c.JSON(http.StatusOK, dto.MatchResponse{
		Matches: matches,
	})
}

// PrivateMatchesEVM godoc
// @Summary      Check if private key matches EVM address
// @Description  Verify that a private key derives to a given EVM address
// @Tags         matching
// @Accept       json
// @Produce      json
// @Param        request  body      dto.MatchPrivateEVMRequest  true  "Private key and EVM address"
// @Success      200      {object}  dto.MatchResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /matching/private-matches-evm [post]
func PrivateMatchesEVM(c *gin.Context) {
	var req dto.MatchPrivateEVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.ParsePrivateKeyHex(req.PrivateKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid private key: " + err.Error()})
		return
	}

	evmAddr, err := keylib.ParseEVMAddress(req.EVMAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid EVM address: " + err.Error()})
		return
	}

	matches := privKey.MatchesEVMAddress(evmAddr)

	c.JSON(http.StatusOK, dto.MatchResponse{
		Matches: matches,
	})
}

// PublicMatchesEVM godoc
// @Summary      Check if public key matches EVM address
// @Description  Verify that a public key derives to a given EVM address
// @Tags         matching
// @Accept       json
// @Produce      json
// @Param        request  body      dto.MatchPublicEVMRequest  true  "Public key and EVM address"
// @Success      200      {object}  dto.MatchResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /matching/public-matches-evm [post]
func PublicMatchesEVM(c *gin.Context) {
	var req dto.MatchPublicEVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid public key: " + err.Error()})
		return
	}

	evmAddr, err := keylib.ParseEVMAddress(req.EVMAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid EVM address: " + err.Error()})
		return
	}

	matches := pubKey.MatchesEVMAddress(evmAddr)

	c.JSON(http.StatusOK, dto.MatchResponse{
		Matches: matches,
	})
}

// PublicMatchesPeer godoc
// @Summary      Check if public key matches peer ID
// @Description  Verify that a public key derives to a given libp2p peer ID
// @Tags         matching
// @Accept       json
// @Produce      json
// @Param        request  body      dto.MatchPublicPeerIDRequest  true  "Public key and peer ID"
// @Success      200      {object}  dto.MatchResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /matching/public-matches-peer [post]
func PublicMatchesPeer(c *gin.Context) {
	var req dto.MatchPublicPeerIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey, err := keylib.ParsePublicKeyHex(req.PublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid public key: " + err.Error()})
		return
	}

	peerID, err := keylib.ParsePeerID(req.PeerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid peer ID: " + err.Error()})
		return
	}

	matches := pubKey.MatchesPeerID(peerID)

	c.JSON(http.StatusOK, dto.MatchResponse{
		Matches: matches,
	})
}

// ParseEVMAddress godoc
// @Summary      Parse an EVM address
// @Description  Parse and validate an Ethereum address from hex string
// @Tags         identifiers
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParseEVMAddressRequest  true  "EVM address"
// @Success      200      {object}  dto.EVMAddressResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /identifiers/parse-evm-address [post]
func ParseEVMAddress(c *gin.Context) {
	var req dto.ParseEVMAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	evmAddr, err := keylib.ParseEVMAddress(req.Address)
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

// ParsePeerID godoc
// @Summary      Parse a peer ID
// @Description  Parse and validate a libp2p peer ID string
// @Tags         identifiers
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParsePeerIDRequest  true  "Peer ID"
// @Success      200      {object}  dto.PeerIDResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /identifiers/parse-peer-id [post]
func ParsePeerID(c *gin.Context) {
	var req dto.ParsePeerIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	peerID, err := keylib.ParsePeerID(req.PeerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PeerIDResponse{
		PeerID: peerID.String(),
	})
}

// EVMChecksum godoc
// @Summary      Get EIP-55 checksum address
// @Description  Convert an EVM address to its EIP-55 checksummed form
// @Tags         identifiers
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ParseEVMAddressRequest  true  "EVM address"
// @Success      200      {object}  dto.EVMAddressResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /identifiers/evm-checksum [post]
func EVMChecksum(c *gin.Context) {
	var req dto.ParseEVMAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	evmAddr, err := keylib.ParseEVMAddress(req.Address)
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
