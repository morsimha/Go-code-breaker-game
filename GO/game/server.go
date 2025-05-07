package game

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Player struct {
	conn      net.Conn
	id        int
	name      string
	readyNext bool
}

type GameSession struct {
	players       []*Player
	currentPlayer int
	secretCode    int
	gameOver      bool
	guessCount    int
	mutex         sync.Mutex
}

func StartServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	fmt.Println("Server started, waiting for players...")

	// Create a new game session
	session := &GameSession{
		players:       make([]*Player, 0, 2),
		currentPlayer: 0,
		secretCode:    GenerateSecretCode(),
		gameOver:      false,
		guessCount:    0,
	}

	// Accept connections from two players
	for i := 0; i < 2; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		player := &Player{
			conn: conn,
			id:   i + 1,
			name: fmt.Sprintf("Player %d", i+1),
		}

		session.players = append(session.players, player)
		fmt.Printf("%s has connected.\n", player.name)

		// Send welcome message
		writeToClient(conn, fmt.Sprintf("Welcome %s! Waiting for opponent...", player.name))
	}

	// Start the game
	startGame(session)
}

func startGame(session *GameSession) {
	// Notify players that the game is starting
	for _, player := range session.players {
		writeToClient(player.conn, "\nBoth players have connected. Game is starting!")
		writeToClient(player.conn, "Try to guess the 4-digit code. Enter your guess when it's your turn.")
	}

	// Notify the first player that it's their turn
	writeToClient(session.players[0].conn, "\nIt's your turn. Enter your guess:")
	writeToClient(session.players[1].conn, "\nWaiting for Player 1 to make a guess...")

	// Main game loop - handle player guesses
	for !session.gameOver {
		currentPlayer := session.players[session.currentPlayer]
		handlePlayerGuess(session, currentPlayer)
	}
}

func handlePlayerGuess(session *GameSession, player *Player) {
	// Read the player's guess
	buffer := make([]byte, 1024)
	n, err := player.conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from %s: %v", player.name, err)
		handlePlayerDisconnect(session, player)
		return
	}

	// Process the guess
	guess := string(buffer[:n])
	fmt.Printf("Received guess from %s: %s\n", player.name, guess)

	// Validate the guess
	numGuess, err := ValidateGuess(guess)
	if err != nil {
		writeToClient(player.conn, err.Error())
		writeToClient(player.conn, "\nTry again:")
		return
	}

	// Increment guess count
	session.mutex.Lock()
	session.guessCount++
	totalGuesses := session.guessCount
	session.mutex.Unlock()

	// Check if the guess is correct
	var response string
	if numGuess == session.secretCode {
		// Game over - player wins
		prefix := GenerateTimestampPrefix()
		response = prefix + "Congratulations! You guessed the correct number!"
		session.gameOver = true
		
		// Notify both players
		broadcastMessage(session, fmt.Sprintf("\n%s guessed the correct code (%d) and won the game!", player.name, numGuess))
		
		// Ask if they want to play again
		for _, p := range session.players {
			writeToClient(p.conn, "GAME_OVER")
			writeToClient(p.conn, "\nWould you like to play again? (yes/no)")
		}
		
		// Handle restart logic in a separate goroutine
		go handleGameRestart(session)
	} else {
		response = "Try again!"
		
		// Switch turns
		session.mutex.Lock()
		session.currentPlayer = (session.currentPlayer + 1) % 2
		nextPlayer := session.players[session.currentPlayer]
		session.mutex.Unlock()
		
		// Notify both players about the guess and whose turn it is
		otherPlayer := session.players[(session.currentPlayer+1)%2]
		
		// Send response to current player
		writeToClient(player.conn, response)
		broadcastMessage(session, fmt.Sprintf("\n%s guessed %d (incorrect). Total guesses: %d", player.name, numGuess, totalGuesses))
		
		// Update players about whose turn it is
		writeToClient(nextPlayer.conn, "\nIt's your turn. Enter your guess:")
		writeToClient(otherPlayer.conn, fmt.Sprintf("\nWaiting for %s to make a guess...", nextPlayer.name))
	}
}

func handlePlayerDisconnect(session *GameSession, player *Player) {
	log.Printf("%s has disconnected.", player.name)
	
	// Notify the other player
	otherPlayerIdx := 0
	if player.id == 1 {
		otherPlayerIdx = 1
	}
	
	if len(session.players) > otherPlayerIdx {
		otherPlayer := session.players[otherPlayerIdx]
		writeToClient(otherPlayer.conn, fmt.Sprintf("\n%s has disconnected. Game over.", player.name))
		writeToClient(otherPlayer.conn, "GAME_OVER")
	}
	
	// End the game
	session.gameOver = true
}

func handleGameRestart(session *GameSession) {
	responses := make([]string, 2)
	playersReady := 0
	
	// Wait for both players to respond
	for i, player := range session.players {
		buffer := make([]byte, 1024)
		n, err := player.conn.Read(buffer)
		if err != nil {
			handlePlayerDisconnect(session, player)
			return
		}
		
		responses[i] = string(buffer[:n])
		
		if responses[i] == "yes" {
			playersReady++
			player.readyNext = true
		}
	}
	
	// Check if both players want to play again
	if playersReady == 2 {
		// Reset the game
		session.mutex.Lock()
		session.gameOver = false
		session.guessCount = 0
		session.secretCode = GenerateSecretCode()
		session.currentPlayer = 0
		session.mutex.Unlock()
		
		// Start a new game
		broadcastMessage(session, "\nStarting a new game!")
		startGame(session)
	} else {
		// End the game
		broadcastMessage(session, "\nGame ended. Thank you for playing!")
		
		// Close connections
		for _, player := range session.players {
			player.conn.Close()
		}
	}
}

func broadcastMessage(session *GameSession, message string) {
	for _, player := range session.players {
		writeToClient(player.conn, message)
	}
}

func writeToClient(conn net.Conn, s string) {
	_, err := conn.Write([]byte(s))
	if err != nil {
		log.Printf("Error writing to client: %v", err)
		return
	}
}
