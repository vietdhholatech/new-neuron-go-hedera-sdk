// Package main provides the entry point for the keylib API server.
//
//	@title			Keylib API
//	@version		1.0
//	@description	Interactive API for testing Neuron keylib functionality
//	@host			localhost:8080
//	@BasePath		/api/v1
package main

import (
	"log"

	"github.com/aspect-build/neuron-go-hedera-sdk/api"
)

func main() {
	server := api.NewServer()
	log.Println("Starting keylib API server on :8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
