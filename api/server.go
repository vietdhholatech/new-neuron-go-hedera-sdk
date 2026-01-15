package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/aspect-build/neuron-go-hedera-sdk/api/docs"
	"github.com/aspect-build/neuron-go-hedera-sdk/api/handlers"
)

// Server represents the HTTP API server.
type Server struct {
	router *gin.Engine
}

// NewServer creates a new API server with all routes configured.
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware
	router.Use(corsMiddleware())
	router.Use(errorHandler())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Keys - Generation and parsing
		keys := v1.Group("/keys")
		{
			keys.POST("/generate", handlers.GenerateKey)
			keys.POST("/parse/private", handlers.ParsePrivateKey)
			keys.POST("/parse/public", handlers.ParsePublicKey)
		}

		// Derivation - Derive identifiers from keys
		derive := v1.Group("/derive")
		{
			derive.POST("/public-key", handlers.DerivePublicKey)
			derive.POST("/evm-address", handlers.DeriveEVMAddress)
			derive.POST("/evm-address-safe", handlers.DeriveEVMAddressSafe)
			derive.POST("/peer-id", handlers.DerivePeerID)
		}

		// Signing - Sign and verify messages
		signing := v1.Group("/signing")
		{
			signing.POST("/sign-message", handlers.SignMessage)
			signing.POST("/sign-digest", handlers.SignDigest)
			signing.POST("/verify-message", handlers.VerifyMessage)
			signing.POST("/verify-digest", handlers.VerifyDigest)
		}

		// Signatures - Parse and recover from signatures
		signatures := v1.Group("/signatures")
		{
			signatures.POST("/parse", handlers.ParseSignature)
			signatures.POST("/recover-pubkey", handlers.RecoverPublicKey)
			signatures.POST("/recover-from-digest", handlers.RecoverFromDigest)
			signatures.POST("/components", handlers.GetSignatureComponents)
		}

		// Mnemonic - Generate and derive from mnemonics
		mnemonic := v1.Group("/mnemonic")
		{
			mnemonic.POST("/generate", handlers.GenerateMnemonic)
			mnemonic.POST("/validate", handlers.ValidateMnemonic)
			mnemonic.POST("/derive-key", handlers.DeriveKeyFromMnemonic)
			mnemonic.POST("/derive-key-full", handlers.DeriveKeyFull)
		}

		// Encryption - Scramble and unscramble keys
		encryption := v1.Group("/encryption")
		{
			encryption.POST("/scramble", handlers.ScrambleKey)
			encryption.POST("/unscramble", handlers.UnscrambleKey)
		}

		// Matching - Check key/identifier relationships
		matching := v1.Group("/matching")
		{
			matching.POST("/private-matches-public", handlers.PrivateMatchesPublic)
			matching.POST("/private-matches-evm", handlers.PrivateMatchesEVM)
			matching.POST("/public-matches-evm", handlers.PublicMatchesEVM)
			matching.POST("/public-matches-peer", handlers.PublicMatchesPeer)
		}

		// Identifiers - Parse EVM addresses and peer IDs
		identifiers := v1.Group("/identifiers")
		{
			identifiers.POST("/parse-evm-address", handlers.ParseEVMAddress)
			identifiers.POST("/parse-peer-id", handlers.ParsePeerID)
			identifiers.POST("/evm-checksum", handlers.EVMChecksum)
		}
	}

	return &Server{router: router}
}

// Run starts the server on the given address.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
