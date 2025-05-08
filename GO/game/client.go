package game

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func StartClient(address string) error {
	// Connect to the server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Welcome to the Code Breaker Game! Connecting to server...")

	// Create a reader to capture input from stdin
	reader := bufio.NewReader(os.Stdin)

	// Use channels to handle incoming server messages in a separate goroutine
	serverMessages := make(chan string)
	clientErrors := make(chan error)
	gameOver := false
	isMyTurn := false // Track if it's this player's turn

	// Start goroutine to listen for server messages
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				clientErrors <- fmt.Errorf("disconnected from server: %v", err)
				return
			}

			message := string(buffer[:n])
			
			// Check for game over condition
			if message == "GAME_OVER" {
				gameOver = true
				continue // Continue to read the next message (play again prompt)
			}
			
			// Check if it's this player's turn
			if strings.Contains(message, "It's your turn") {
				isMyTurn = true
			} else if strings.Contains(message, "Waiting for") || 
                     strings.Contains(message, "ran out of time") ||
                     strings.Contains(message, "turn is forfeited") {
				isMyTurn = false
			}
			
			// Add visual indicator for time-based messages
			if strings.Contains(message, "Time's up!") || 
               strings.Contains(message, "ran out of time") {
				message = "\n⏰ " + message
			}
			
			// Add visual indicator for time limit information
			if strings.Contains(message, "seconds to make") {
				message = "⏱️ " + message
			}
			
			serverMessages <- message
		}
	}()

	// Start the game loop
	for {
		select {
		case err := <-clientErrors:
			return err
		case message := <-serverMessages:
			fmt.Println(message)
			
			// Check if it's this player's turn or a prompt requiring input
			if (isMyTurn && strings.Contains(message, "your turn")) || 
			   strings.Contains(message, "Try again:") ||
			   strings.Contains(message, "play again") {
				var userInput string
				
				// Check if this is a time-based prompt
				if strings.Contains(message, "Time's up!") {
					fmt.Println("⚠️ You ran out of time on your previous turn!")
				}
				
				// For play again prompt
				if strings.Contains(message, "play again") {
					fmt.Print("Enter 'yes' to play again or 'no' to quit: ")
				} else {
					// Regular guess prompt
					fmt.Print("Enter your guess (4 digits) or 'exit' to quit: ")
				}
				
				userInput, err = reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %v", err)
				}
				userInput = strings.TrimSpace(userInput)
				
				// Handle exit command
				if userInput == "exit" && !gameOver {
					fmt.Println("Exiting the game.")
					return nil
				}
				
				// Send input to server
				_, err = conn.Write([]byte(userInput))
				if err != nil {
					return fmt.Errorf("error sending message to server: %v", err)
				}
				
				// After sending input, it's no longer this player's turn
				isMyTurn = false
			}
		case <-time.After(90 * time.Second):
			// Timeout for safety (in case of deadlock)
			// This is longer than the server's turn timeout to account for network latency and processing
			fmt.Println("No response from server in 90 seconds. Please check your connection.")
			return fmt.Errorf("server response timeout")
		}
	}
}