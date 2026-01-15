package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aspect-build/neuron-go-hedera-sdk/api/dto"
	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// GenerateMnemonic godoc
// @Summary      Generate a new mnemonic
// @Description  Generate a new BIP39 mnemonic phrase with specified word count
// @Tags         mnemonic
// @Accept       json
// @Produce      json
// @Param        request  body      dto.GenerateMnemonicRequest  true  "Word count (12, 15, 18, 21, or 24)"
// @Success      200      {object}  dto.MnemonicResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /mnemonic/generate [post]
func GenerateMnemonic(c *gin.Context) {
	var req dto.GenerateMnemonicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	mnemonic, err := keylib.GenerateMnemonic(req.WordCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MnemonicResponse{
		Mnemonic:  mnemonic,
		WordCount: req.WordCount,
	})
}

// ValidateMnemonic godoc
// @Summary      Validate a mnemonic
// @Description  Validate a BIP39 mnemonic phrase
// @Tags         mnemonic
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ValidateMnemonicRequest  true  "Mnemonic phrase"
// @Success      200      {object}  dto.ValidateResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /mnemonic/validate [post]
func ValidateMnemonic(c *gin.Context) {
	var req dto.ValidateMnemonicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	err := keylib.ValidateMnemonic(req.Mnemonic)
	if err != nil {
		c.JSON(http.StatusOK, dto.ValidateResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ValidateResponse{
		Valid: true,
	})
}

// DeriveKeyFromMnemonic godoc
// @Summary      Derive key from mnemonic
// @Description  Derive a private key from a BIP39 mnemonic using default path (m/44'/60'/0'/0/0)
// @Tags         mnemonic
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromMnemonicRequest  true  "Mnemonic phrase"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /mnemonic/derive-key [post]
func DeriveKeyFromMnemonic(c *gin.Context) {
	var req dto.DeriveFromMnemonicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	privKey, err := keylib.PrivateKeyFromMnemonic(req.Mnemonic)
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

// DeriveKeyFull godoc
// @Summary      Derive key with options
// @Description  Derive a private key from a BIP39 mnemonic with custom passphrase and derivation path
// @Tags         mnemonic
// @Accept       json
// @Produce      json
// @Param        request  body      dto.DeriveFromMnemonicFullRequest  true  "Mnemonic with options"
// @Success      200      {object}  dto.KeyPairResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /mnemonic/derive-key-full [post]
func DeriveKeyFull(c *gin.Context) {
	var req dto.DeriveFromMnemonicFullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	var privKey keylib.NeuronPrivateKey
	var err error

	// Use options if provided
	if req.Path != "" || req.Passphrase != "" {
		path := req.Path
		if path == "" {
			path = "m/44'/60'/0'/0/0" // Default Ethereum path
		}
		privKey, err = keylib.PrivateKeyFromMnemonicWithOptions(req.Mnemonic, req.Passphrase, path)
	} else {
		privKey, err = keylib.PrivateKeyFromMnemonic(req.Mnemonic)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	pubKey := privKey.PublicKey()
	peerID, _ := pubKey.PeerID()

	// Return response with path info
	response := dto.KeyPairResponse{
		PrivateKeyHex:         privKey.Hex(),
		PublicKeyHex:          pubKey.Hex(),
		PublicKeyUncompressed: pubKey.HexUncompressed(),
		EVMAddress:            pubKey.EVMAddress().Hex(),
		EVMAddressChecksum:    pubKey.EVMAddress().ChecksumHex(),
		PeerID:                peerID.String(),
	}

	// Add path info to response if we want to extend it
	c.JSON(http.StatusOK, gin.H{
		"privateKeyHex":            response.PrivateKeyHex,
		"publicKeyHex":             response.PublicKeyHex,
		"publicKeyHexUncompressed": response.PublicKeyUncompressed,
		"evmAddress":               response.EVMAddress,
		"evmAddressChecksum":       response.EVMAddressChecksum,
		"peerID":                   response.PeerID,
		"derivationPath":           getPath(req.Path),
		"wordCount":                len(strings.Fields(req.Mnemonic)),
	})
}

func getPath(path string) string {
	if path == "" {
		return "m/44'/60'/0'/0/0"
	}
	return path
}
