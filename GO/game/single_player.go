package game

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// StartSinglePlayerGame starts a single-player version of the Code Breaker Game
func StartSinglePlayerGame() {
	fmt.Println("Welcome to the Code Breaker Game (Single Player Mode)!")
	fmt.Println("Try to guess the 4-digit code.")
	
	reader := bufio.NewReader(os.Stdin)
	playAgain := true
	
	for playAgain {
		// Generate a secret code for this game
		secretCode := GenerateSecretCode()
		guessCount := 0
		gameWon := false
		
		// Game loop for one round
		for !gameWon {
			fmt.Print("\nEnter your guess (4 digits) or 'exit' to quit: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				return
			}
			
			input = strings.TrimSpace(input)
			
			// Check for exit command
			if input == "exit" {
				fmt.Println("Exiting the game.")
				return
			}
			
			// Validate the guess
			numGuess, err := ValidateGuess(input)
			if err != nil {
				fmt.Printf("Invalid input: %s\n", err.Error())
				continue
			}
			
			// Increment guess count
			guessCount++
			
			// Check if the guess is correct
			if numGuess == secretCode {
				prefix := GenerateTimestampPrefix()
				fmt.Printf("%sCorrect! You guessed it in %d attempts.\n", prefix, guessCount)
				gameWon = true
			} else {
				// Provide feedback on the guess
				fmt.Printf("Incorrect. Try again! (Attempts: %d)\n", guessCount)
				
				// Optional: Add hint functionality for single player mode
				hint := generateHint(numGuess, secretCode)
				fmt.Printf("Hint: %s\n", hint)
			}
		}
		
		// Ask if player wants to play again
		fmt.Print("\nWould you like to play again? (yes/no): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		playAgain = (response == "yes" || response == "y")
	}
	
	fmt.Println("Thanks for playing! Goodbye.")
}

// Helper function to generate hints for single player mode
func generateHint(guess, secretCode int) string {
	guessStr := fmt.Sprintf("%04d", guess)
	secretStr := fmt.Sprintf("%04d", secretCode)
	
	// Count correct digits in correct positions and correct digits in wrong positions
	correctPosition := 0
	correctDigit := 0
	
	// Track which positions we've already matched
	usedSecret := [4]bool{}
	usedGuess := [4]bool{}
	
	// First pass: find correct positions
	for i := 0; i < 4; i++ {
		if guessStr[i] == secretStr[i] {
			correctPosition++
			usedSecret[i] = true
			usedGuess[i] = true
		}
	}
	
	// Second pass: find correct digits in wrong positions
	for i := 0; i < 4; i++ {
		if usedGuess[i] {
			continue
		}
		
		for j := 0; j < 4; j++ {
			if !usedSecret[j] && guessStr[i] == secretStr[j] {
				correctDigit++
				usedSecret[j] = true
				break
			}
		}
	}
	
	return fmt.Sprintf("%d correct position, %d correct digit but wrong position", 
		correctPosition, correctDigit)
}