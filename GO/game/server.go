package game

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Player struct {
	conn      net.Conn
	id        int
	name      string
	readyNext bool
}

type GameSession struct {
	players          []*Player
	currentPlayer    int
	secretCode       int
	gameOver         bool
	guessCount       int
	mutex            sync.Mutex
	gameStarted      bool
	maxPlayers       int
	acceptingPlayers bool
	singlePlayerMode bool
	turnTimeLimit    time.Duration // Time limit for each player's turn
	analytics        *GameStats    // Analytics for this game session
}

// Global analytics tracker
var globalAnalytics *GameAnalytics

// InitAnalytics initializes the global analytics tracker
func InitAnalytics() {
	globalAnalytics = NewGameAnalytics()
}

// StartMultiplayerServer starts the server in multiplayer mode
func StartMultiplayerServer() {
	// Initialize analytics if not already done
	if globalAnalytics == nil {
		InitAnalytics()
	}

	// Get the maximum number of players from environment variable or argument
	// Default to 2 if not specified
	maxPlayers := 2
	if len(os.Args) > 2 {
		if val, err := strconv.Atoi(os.Args[2]); err == nil && val > 1 {
			maxPlayers = val
		}
	}

	startServer(maxPlayers, false)
}

// StartSinglePlayerServer starts the server in single-player mode
func StartSinglePlayerServer() {
	// Initialize analytics if not already done
	if globalAnalytics == nil {
		InitAnalytics()
	}

	startServer(1, true)
}

// Common server starting function with mode parameter
func startServer(maxPlayers int, singlePlayerMode bool) {
	modeStr := "multiplayer"
	if singlePlayerMode {
		modeStr = "single-player"
	}

	log.Printf("Starting server in %s mode with support for %d players...", modeStr, maxPlayers)
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	// Create command listener for admin commands
	go startCommandListener()

	for {
		// Generate secret code for this session
		secretCode := GenerateSecretCode()
		
		// Create a new game session
		session := &GameSession{
			players:          make([]*Player, 0, maxPlayers),
			currentPlayer:    0,
			secretCode:       secretCode,
			gameOver:         false,
			guessCount:       0,
			gameStarted:      false,
			maxPlayers:       maxPlayers,
			acceptingPlayers: true,
			singlePlayerMode: singlePlayerMode,
			turnTimeLimit:    30 * time.Second, // 30-second time limit for each turn
		}

		// Initialize analytics for this game
		session.analytics = globalAnalytics.StartGame(secretCode, maxPlayers)

		if singlePlayerMode {
			log.Printf("New game session created. Waiting for a player to connect...")
		} else {
			log.Printf("New game session created. Waiting for up to %d players to connect...", maxPlayers)
		}

		// Start accepting players in a separate goroutine
		playersConnected := make(chan struct{})
		go acceptPlayers(listener, session, playersConnected)

		// Wait until we have enough players to start or all player slots are filled
		<-playersConnected

		// Run the game session
		runGameSession(session)
	}
}

// startCommandListener starts a goroutine to listen for admin commands
func startCommandListener() {
	commandPort := 8081 // Different port for admin commands
	commandListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", commandPort))
	if err != nil {
		log.Printf("Error starting command listener: %v", err)
		return
	}
	defer commandListener.Close()

	log.Printf("Command listener started on port %d", commandPort)

	for {
		conn, err := commandListener.Accept()
		if err != nil {
			log.Printf("Error accepting command connection: %v", err)
			continue
		}

		go handleCommandConnection(conn)
	}
}

// handleCommandConnection processes admin commands
func handleCommandConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading command: %v", err)
		return
	}

	command := strings.TrimSpace(string(buffer[:n]))
	
	switch command {
	case "stats":
		// Generate and send analytics report
		report := globalAnalytics.GetAnalyticsReport()
		conn.Write([]byte(report))
	default:
		conn.Write([]byte("Unknown command. Available commands: stats"))
	}
}

func acceptPlayers(listener net.Listener, session *GameSession, playersConnected chan struct{}) {
	// Set a timer for max wait time (3 minutes)
	timer := time.NewTimer(3 * time.Minute)
	defer timer.Stop()

	// Channel to receive new connections
	connChan := make(chan net.Conn)

	// Start accepting connections in a separate goroutine
	go func() {
		for session.acceptingPlayers {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			connChan <- conn
		}
	}()

	// Handle connections until we reach players number
	for session.acceptingPlayers {
		conn := <-connChan
		// New player connected
		session.mutex.Lock()
		if len(session.players) < session.maxPlayers && !session.gameStarted {
			playerID := len(session.players) + 1
			player := &Player{
				conn:      conn,
				id:        playerID,
				name:      fmt.Sprintf("Player %d", playerID),
				readyNext: false,
			}

			session.players = append(session.players, player)
			log.Printf("%s has connected. Total players: %d/%d", player.name, len(session.players), session.maxPlayers)

			// Send welcome message to the new player
			if session.singlePlayerMode {
				writeToClient(conn, fmt.Sprintf("Welcome %s! You are playing in single-player mode against the computer.",
					player.name))
				writeToClient(conn, fmt.Sprintf("You have %d seconds to make each guess!", int(session.turnTimeLimit.Seconds())))
			} else {
				writeToClient(conn, fmt.Sprintf("Welcome %s! Waiting for other players... (%d/%d connected)",
					player.name, len(session.players), session.maxPlayers))
				writeToClient(conn, fmt.Sprintf("You will have %d seconds to make each guess!", int(session.turnTimeLimit.Seconds())))

				// Broadcast to other players that someone new joined
				for _, p := range session.players {
					if p.id != playerID {
						writeToClient(p.conn, fmt.Sprintf("\n%s has joined the game. (%d/%d players connected)",
							player.name, len(session.players), session.maxPlayers))
					}
				}
			}

			// Check if we have reached max players
			if len(session.players) == session.maxPlayers {
				session.acceptingPlayers = false
				session.mutex.Unlock()
				close(playersConnected)
				return
			}

		} else {
			// Game already started or max players reached, reject connection
			writeToClient(conn, "Sorry, this game has already started or is full. Please try again later.")
			conn.Close()
		}
		session.mutex.Unlock()
	}
}

func runGameSession(session *GameSession) {
	session.mutex.Lock()
	
	// Single-player mode only needs 1 player
	minPlayers := 2
	if session.singlePlayerMode {
		minPlayers = 1
	}
	
	if len(session.players) < minPlayers {
		log.Println("Not enough players to start the game.")
		for _, player := range session.players {
			writeToClient(player.conn, "Not enough players to start the game. Please try again later.")
			player.conn.Close()
		}
		session.mutex.Unlock()
		
		// End game analytics with no winner
		globalAnalytics.EndGame(session.analytics, 0)
		return
	}
	
	session.gameStarted = true
	session.acceptingPlayers = false
	session.mutex.Unlock()

	if session.singlePlayerMode {
		// Single-player mode
		player := session.players[0]
		writeToClient(player.conn, "\nGame is starting in single-player mode!")
		writeToClient(player.conn, "Try to guess the 4-digit code.")
		writeToClient(player.conn, fmt.Sprintf("\nYou have %d seconds for each guess!", int(session.turnTimeLimit.Seconds())))
		writeToClient(player.conn, "\nIt's your turn. Enter your guess:")
		
		// Run the single-player game loop
		for !session.gameOver {
			handlePlayerGuess(session, player)
		}
		
		// When game is over, ask if player wants to restart
		handleSinglePlayerRestart(session, player)
	} else {
		// Multiplayer mode
		// Notify players that the game is starting
		broadcastMessage(session, "\nGame is starting with "+strconv.Itoa(len(session.players))+" players!")
		broadcastMessage(session, "Try to guess the 4-digit code. Players will take turns in order.")
		broadcastMessage(session, fmt.Sprintf("\nEach player has %d seconds to make their guess!", int(session.turnTimeLimit.Seconds())))

		// Show player list
		playerList := "\nPlayers in this game:"
		for _, player := range session.players {
			playerList += "\n- " + player.name
		}
		broadcastMessage(session, playerList)

		// Notify the first player that it's their turn
		writeToClient(session.players[0].conn, "\nIt's your turn. Enter your guess:")

		// Notify other players they're waiting
		for i, player := range session.players {
			if i != 0 {
				writeToClient(player.conn, "\nWaiting for Player 1 to make a guess...")
			}
		}

		// Main game loop
		for !session.gameOver {
			session.mutex.Lock()
			currentPlayerIndex := session.currentPlayer
			currentPlayer := session.players[currentPlayerIndex]
			session.mutex.Unlock()

			handlePlayerGuess(session, currentPlayer)
		}

		// When game is over, wait for player responses about restarting
		handleGameRestart(session)
	}
}

func handlePlayerGuess(session *GameSession, player *Player) {
    // Create a context with cancel function to abort the read operation if needed
    ctx, cancel := context.WithTimeout(context.Background(), session.turnTimeLimit)
    defer cancel() // Ensure we always cancel the context
    
    // Create channels for the guess and timeout
    guessChan := make(chan string)
    errChan := make(chan error)
    
    // Start a goroutine to read the player's guess
    go func() {
        // Set read deadline based on our timeout
        player.conn.SetReadDeadline(time.Now().Add(session.turnTimeLimit))
        buffer := make([]byte, 1024)
        n, err := player.conn.Read(buffer)

        // Check if the context was canceled before we check the error
        select {
        case <-ctx.Done():
            // The context was canceled, so we just return without sending
            // to the channels, effectively abandoning this read attempt
            return
        default:
            // Context wasn't canceled, so proceed normally
            if err != nil {
                errChan <- err
                return
            }
            guessChan <- string(buffer[:n])
        }
    }()
    
    // Wait for either a guess, an error, or a timeout
    select {
    case guess := <-guessChan:
        // Reset read deadline
        player.conn.SetReadDeadline(time.Time{})
        
        // Process the guess
        log.Printf("Received guess from %s: %s", player.name, guess)
        
        // Validate the guess
        numGuess, err := ValidateGuess(guess)
        if err != nil {
            writeToClient(player.conn, err.Error())
            writeToClient(player.conn, "\nTry again:")
            return
        }
        
        // Record this guess in analytics
        globalAnalytics.RecordGuess(session.analytics, player.id, numGuess)
        
        // Increment guess count
        session.mutex.Lock()
        session.guessCount++
        totalGuesses := session.guessCount
        session.mutex.Unlock()
        
        // Check if the guess is correct
        if numGuess == session.secretCode {
            // Game over - player wins
            prefix := GenerateTimestampPrefix()
            response := prefix + "Congratulations! You guessed the correct number!"
            writeToClient(player.conn, response)
            
            session.mutex.Lock()
            session.gameOver = true
            session.mutex.Unlock()
            
            // Update analytics for game end with winner
            globalAnalytics.EndGame(session.analytics, player.id)
            
            if session.singlePlayerMode {
                // Single-player mode - notify only current player
                writeToClient(player.conn, fmt.Sprintf("\nYou guessed the correct code (%d)!", numGuess))
                writeToClient(player.conn, fmt.Sprintf("\nSecret code was: %d", session.secretCode))
                writeToClient(player.conn, fmt.Sprintf("\nTotal guesses: %d", totalGuesses))
                
                // Ask if they want to play again
                writeToClient(player.conn, "\nWould you like to play again? (yes/no)")
            } else {
                // Multiplayer mode - notify all players
                broadcastMessage(session, fmt.Sprintf("\n%s guessed the correct code (%d) and won the game!", player.name, numGuess))
                broadcastMessage(session, fmt.Sprintf("\nSecret code was: %d", session.secretCode))
                broadcastMessage(session, fmt.Sprintf("\nTotal guesses: %d", totalGuesses))
                
                // Ask if they want to play again
                broadcastMessage(session, "\nWould you like to play again? (yes/no)")
            }
        } else {
            response := "Try again!"
            
            if session.singlePlayerMode {
                // Single-player mode - just notify the player
                writeToClient(player.conn, response)
                writeToClient(player.conn, fmt.Sprintf("\nYou guessed %d (incorrect). Total guesses: %d", numGuess, totalGuesses))
                writeToClient(player.conn, "\nIt's your turn. Enter your guess:")
            } else {
                // Multiplayer mode - switch turns to next player
                session.mutex.Lock()
                playerCount := len(session.players)
                session.currentPlayer = (session.currentPlayer + 1) % playerCount
                nextPlayer := session.players[session.currentPlayer]
                session.mutex.Unlock()
                
                // Send response to current player
                writeToClient(player.conn, response)
                broadcastMessage(session, fmt.Sprintf("\n%s guessed %d (incorrect). Total guesses: %d", player.name, numGuess, totalGuesses))
                
                // Update players about whose turn it is
                writeToClient(nextPlayer.conn, "\nIt's your turn. Enter your guess:")
                
                // Tell other players to wait
                for _, p := range session.players {
                    if p.id != nextPlayer.id {
                        writeToClient(p.conn, fmt.Sprintf("\nWaiting for %s to make a guess...", nextPlayer.name))
                    }
                }
            }
        }
    
    case err := <-errChan:
        // Handle error (could be a disconnection or timeout)
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            // This is a timeout error
            log.Printf("%s timed out on their turn", player.name)
            
            // Cancel the context to abort the read operation
            cancel()
            
            if session.singlePlayerMode {
                // In single-player, just tell them they timed out and give another chance
                writeToClient(player.conn, fmt.Sprintf("\nTime's up! You took longer than %d seconds. Try again:", int(session.turnTimeLimit.Seconds())))
                // Reset read deadline
                player.conn.SetReadDeadline(time.Time{})
            } else {
                // In multiplayer, forfeit their turn
                writeToClient(player.conn, fmt.Sprintf("\nTime's up! You took longer than %d seconds. Your turn is forfeited.", int(session.turnTimeLimit.Seconds())))
                
                // Move to next player
                session.mutex.Lock()
                playerCount := len(session.players)
                session.currentPlayer = (session.currentPlayer + 1) % playerCount
                nextPlayer := session.players[session.currentPlayer]
                session.mutex.Unlock()
                
                // Broadcast timeout message
                broadcastMessage(session, fmt.Sprintf("\n%s ran out of time and forfeited their turn!", player.name))
                
                // Update players about whose turn it is
                writeToClient(nextPlayer.conn, "\nIt's your turn. Enter your guess:")
                
                // Tell other players to wait
                for _, p := range session.players {
                    if p.id != nextPlayer.id {
                        writeToClient(p.conn, fmt.Sprintf("\nWaiting for %s to make a guess...", nextPlayer.name))
                    }
                }
                
                // Reset read deadline for the player who timed out
                player.conn.SetReadDeadline(time.Time{})
            }
        } else {
            // Some other error (like disconnection)
            log.Printf("Error reading from %s: %v", player.name, err)
            handlePlayerDisconnect(session, player)
        }
    
    case <-ctx.Done():
        // Context timeout - the turn has expired
        log.Printf("%s timed out on their turn (context timeout)", player.name)
        
        if session.singlePlayerMode {
            writeToClient(player.conn, fmt.Sprintf("\nTime's up! You took longer than %d seconds. Try again:", int(session.turnTimeLimit.Seconds())))
            // Reset read deadline
            player.conn.SetReadDeadline(time.Time{})
        } else {
            writeToClient(player.conn, fmt.Sprintf("\nTime's up! You took longer than %d seconds. Your turn is forfeited.", int(session.turnTimeLimit.Seconds())))
            
            // Move to next player
            session.mutex.Lock()
            playerCount := len(session.players)
            session.currentPlayer = (session.currentPlayer + 1) % playerCount
            nextPlayer := session.players[session.currentPlayer]
            session.mutex.Unlock()
            
            broadcastMessage(session, fmt.Sprintf("\n%s ran out of time and forfeited their turn!", player.name))
            writeToClient(nextPlayer.conn, "\nIt's your turn. Enter your guess:")
            
            for _, p := range session.players {
                if p.id != nextPlayer.id {
                    writeToClient(p.conn, fmt.Sprintf("\nWaiting for %s to make a guess...", nextPlayer.name))
                }
            }
            
            // Reset read deadline
            player.conn.SetReadDeadline(time.Time{})
        }
    }
}

func handleSinglePlayerRestart(session *GameSession, player *Player) {
	log.Println("Single-player game over, waiting for player to decide if they want to restart...")
	
	// Read the player's response with a timeout
	player.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	buffer := make([]byte, 1024)
	n, err := player.conn.Read(buffer)
	player.conn.SetReadDeadline(time.Time{}) // Reset deadline
	
	if err != nil {
		log.Printf("Error reading restart response from %s: %v", player.name, err)
		writeToClient(player.conn, "\nNo response received. Ending game. Thank you for playing!")
		player.conn.Close()
		return
	}
	
	response := strings.TrimSpace(string(buffer[:n]))
	log.Printf("%s responded with: %s", player.name, response)
	
	if response == "yes" {
		// Player wants to continue
		session.mutex.Lock()
		
		// Reset for new game
		newSecretCode := GenerateSecretCode()
		session.gameOver = false
		session.guessCount = 0
		session.secretCode = newSecretCode
		
		// Create new analytics for this game
		session.analytics = globalAnalytics.StartGame(newSecretCode, 1)
		
		session.mutex.Unlock()
		
		// Start a new game
		writeToClient(player.conn, "\nStarting a new game!")
		writeToClient(player.conn, "\nTry to guess the 4-digit code.")
		writeToClient(player.conn, fmt.Sprintf("\nYou have %d seconds for each guess!", int(session.turnTimeLimit.Seconds())))
		writeToClient(player.conn, "\nIt's your turn. Enter your guess:")
		
		// Run the single-player game session again
		for !session.gameOver {
			handlePlayerGuess(session, player)
		}
		
		// When game is over, ask if player wants to restart again
		handleSinglePlayerRestart(session, player)
	} else {
		// Player doesn't want to continue
		writeToClient(player.conn, "\nThank you for playing! Goodbye.")
		player.conn.Close()
	}
}

func handlePlayerDisconnect(session *GameSession, player *Player) {
	log.Printf("%s has disconnected.", player.name)

	// In single-player mode, just end the game
	if session.singlePlayerMode {
		session.mutex.Lock()
		session.gameOver = true
		session.mutex.Unlock()
		
		// Update analytics for game end with no winner
		globalAnalytics.EndGame(session.analytics, 0)
		return
	}

	// Remove the player from the session
	session.mutex.Lock()
	for i, p := range session.players {
		if p.id == player.id {
			// Remove this player
			session.players = append(session.players[:i], session.players[i+1:]...)
			break
		}
	}

	// Check if we still have enough players to continue
	if len(session.players) < 2 {
		session.gameOver = true
		session.mutex.Unlock()
		
		// Update analytics for game end with no winner
		globalAnalytics.EndGame(session.analytics, 0)

		// Notify remaining players
		for _, p := range session.players {
			writeToClient(p.conn, fmt.Sprintf("\n%s has disconnected. Not enough players to continue.", player.name))
			writeToClient(p.conn, "\nGame over. Thank you for playing!")
			p.conn.Close()
		}
	} else {
		// Adjust current player index if needed
		if session.currentPlayer >= len(session.players) {
			session.currentPlayer = 0
		}

		nextPlayer := session.players[session.currentPlayer]
		session.mutex.Unlock()

		// Notify remaining players
		broadcastMessage(session, fmt.Sprintf("\n%s has disconnected. Continuing with %d players.",
			player.name, len(session.players)))

		// Update turn if it was the disconnected player's turn
		writeToClient(nextPlayer.conn, "\nIt's your turn. Enter your guess:")

		for _, p := range session.players {
			if p.id != nextPlayer.id {
				writeToClient(p.conn, fmt.Sprintf("\nWaiting for %s to make a guess...", nextPlayer.name))
			}
		}
	}
}

func handleGameRestart(session *GameSession) {
	log.Println("Game over, waiting for players to decide if they want to restart...")

	// Reset player ready flags
	for _, player := range session.players {
		player.readyNext = false
	}

	// Track player responses
	responses := make(map[int]bool)
	var responseMutex sync.Mutex
	var wg sync.WaitGroup

	// Process each player's restart decision in separate goroutines
	session.mutex.Lock()
	playerCount := len(session.players)
	playersArray := make([]*Player, playerCount)
	copy(playersArray, session.players)
	session.mutex.Unlock()

	wg.Add(playerCount)

	// Set restart decision timeout
	decisionTimeout := 30 * time.Second

	for i := range playersArray {
		go func(playerIndex int) {
			defer wg.Done()
			player := playersArray[playerIndex]

			// Read the player's response with a timeout
			player.conn.SetReadDeadline(time.Now().Add(decisionTimeout))
			buffer := make([]byte, 1024)
			n, err := player.conn.Read(buffer)
			player.conn.SetReadDeadline(time.Time{}) // Reset deadline
			
			if err != nil {
				log.Printf("Error or timeout reading restart response from %s: %v", player.name, err)
				responseMutex.Lock()
				responses[player.id] = false
				responseMutex.Unlock()
				writeToClient(player.conn, "\nNo response received in time. You'll be disconnected when the game restarts.")
				return
			}

			response := strings.TrimSpace(string(buffer[:n]))
			log.Printf("%s responded with: %s", player.name, response)

			responseMutex.Lock()
			if response == "yes" {
				responses[player.id] = true
				player.readyNext = true
				writeToClient(player.conn, "\nYou chose to continue. Waiting for other players' responses...")
			} else {
				responses[player.id] = false
				writeToClient(player.conn, "\nYou chose not to continue. Waiting for other players...")
			}
			responseMutex.Unlock()
		}(i)
	}

	// Wait for all players to respond (or timeout)
	wg.Wait()

	// Count yes responses
	session.mutex.Lock()
	yesCount := 0
	for _, player := range session.players {
		if player.readyNext {
			yesCount++
		}
	}

	// Check if we have enough players to restart (at least 2)
	if yesCount >= 2 {
		// Create a new array with only players who want to continue
		continuingPlayers := make([]*Player, 0, yesCount)
		for _, player := range session.players {
			if player.readyNext {
				continuingPlayers = append(continuingPlayers, player)
			} else {
				// Close connection for players who don't want to continue
				writeToClient(player.conn, "\nThank you for playing! Goodbye.")
				player.conn.Close()
			}
		}

		// Generate a new secret code for the next game
		newSecretCode := GenerateSecretCode()
		
		// Update the session with only continuing players
		session.players = continuingPlayers
		session.gameOver = false
		session.guessCount = 0
		session.secretCode = newSecretCode
		session.currentPlayer = 0
		
		// Initialize analytics for this new game
		session.analytics = globalAnalytics.StartGame(newSecretCode, len(continuingPlayers))
		
		session.mutex.Unlock()

		// Start a new game
		broadcastMessage(session, fmt.Sprintf("\n%d players want to continue. Starting a new game!", yesCount))
		broadcastMessage(session, fmt.Sprintf("\nEach player has %d seconds to make their guess!", int(session.turnTimeLimit.Seconds())))

		// Run the game session again
		runGameSession(session)
	} else {
		// Not enough players to restart
		session.mutex.Unlock()
		broadcastMessage(session, "\nNot enough players want to continue. Game ended.")

		// Close all connections
		for _, player := range session.players {
			writeToClient(player.conn, "Thank you for playing! Goodbye.")
			player.conn.Close()
		}
	}
}

func broadcastMessage(session *GameSession, message string) {
	session.mutex.Lock()
	for _, player := range session.players {
		writeToClient(player.conn, message)
	}
	session.mutex.Unlock()
}

func writeToClient(conn net.Conn, s string) {
	_, err := conn.Write([]byte(s))
	if err != nil {
		log.Printf("Error writing to client: %v", err)
		return
	}
}