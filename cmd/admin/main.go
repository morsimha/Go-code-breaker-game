package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run admin_client.go <server_address>")
		fmt.Println("Example: go run admin_client.go localhost:8080")
		return
	}

	serverAddress := os.Args[1]

	fmt.Println("Code Breaker Admin Client")
	fmt.Println("========================")
	fmt.Println("Available commands:")
	fmt.Println("  stats - Display game statistics")
	fmt.Println("  exit - Exit the admin client")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nEnter command: ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		command = strings.TrimSpace(command)

		if command == "exit" {
			fmt.Println("Exiting admin client.")
			break
		}

		if command == "stats" {
			// Connect to the server
			conn, err := net.Dial("tcp", serverAddress)
			if err != nil {
				fmt.Printf("Error connecting to server: %v\n", err)
				continue
			}

			// Send the command
			_, err = conn.Write([]byte(command))
			if err != nil {
				fmt.Printf("Error sending command: %v\n", err)
				conn.Close()
				continue
			}

			// Read the response
			buffer := make([]byte, 8192) // Large buffer for stats
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				conn.Close()
				continue
			}

			// Display the response
			fmt.Println("\n" + string(buffer[:n]))

			conn.Close()
		} else {
			fmt.Println("Unknown command. Available commands: stats, exit")
		}
	}
}
