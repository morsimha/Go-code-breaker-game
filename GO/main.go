package main

import (
	"ccs_interview/game"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <mode>\nMode can be 'server' or 'client'")
	}

	mode := os.Args[1]

	switch mode {
	case "server":
		// Start the server in multiplayer mode
		log.Println("Starting server in multiplayer mode...")
		game.StartServer()
	case "client":
		address := "server:8080"
		// If an address is provided, use it (e.g., "localhost:8080")
		// Otherwise, default to "server:8080"
		if len(os.Args) > 2 {
			address = os.Args[2]
		}
		log.Printf("Connecting to server at %s...\n", address)
		err := game.StartClient(address)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Invalid mode. Use 'server' or 'client'.")
	}
}
