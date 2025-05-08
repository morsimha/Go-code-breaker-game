package main

import (
	"ccs_interview/game"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <mode>\nMode can be 'server', 'client', or 'server <num_players>'")
	}

	mode := os.Args[1]

	switch mode {
	case "server":
		// Check if number of players is specified
		if len(os.Args) > 2 {
			// Start the server in multiplayer mode with specified number of players
			log.Println("Starting server in multiplayer mode...")
			game.StartMultiplayerServer()
		} else {
			// Start the server in single-player mode
			log.Println("Starting server in single-player mode...")
			game.StartSinglePlayerServer()
		}
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
		log.Fatal("Invalid mode. Use 'server', 'client', or 'server <num_players>'.")
	}
}
