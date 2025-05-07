package main

import (
	"log"
	"os"
	"ccs_interview/game" // Replace with your actual module name
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
		// Connect client to server
		address := "localhost:8080"
		
		// If an address is provided, use it
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
