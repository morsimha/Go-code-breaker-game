package game

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// GameStats represents statistics for a single game
type GameStats struct {
	SecretCode    int           // The secret code for this game
	GuessCount    int           // Number of guesses made
	Won           bool          // Whether the game was won or not
	StartTime     time.Time     // When the game started
	EndTime       time.Time     // When the game ended
	PlayerCount   int           // Number of players in this game
	PlayerGuesses map[int][]int // Guesses made by each player (player ID -> []guesses)
}

// GameAnalytics stores and manages game statistics
type GameAnalytics struct {
	mu           sync.RWMutex
	gamesPlayed  int                  // Total number of games played
	gamesWon     int                  // Total number of games won
	gameHistory  []*GameStats         // History of all games
	secretCounts map[int]int          // Count of each secret code
	guessCounts  map[int]int          // Count of each guess made
	playerStats  map[int]*PlayerStats // Statistics by player ID
}

// PlayerStats tracks statistics for a specific player
type PlayerStats struct {
	GamesPlayed  int // Total games played
	GamesWon     int // Games won by this player
	TotalGuesses int // Total guesses made
	BestGame     int // Fewest guesses to win (0 if never won)
}

// NewGameAnalytics creates a new analytics tracker
func NewGameAnalytics() *GameAnalytics {
	return &GameAnalytics{
		gameHistory:  make([]*GameStats, 0),
		secretCounts: make(map[int]int),
		guessCounts:  make(map[int]int),
		playerStats:  make(map[int]*PlayerStats),
	}
}

// StartGame begins tracking a new game
func (ga *GameAnalytics) StartGame(secretCode int, playerCount int) *GameStats {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	// Increment the count for this secret code
	ga.secretCounts[secretCode]++
	ga.gamesPlayed++

	// Create new game stats
	stats := &GameStats{
		SecretCode:    secretCode,
		GuessCount:    0,
		Won:           false,
		StartTime:     time.Now(),
		PlayerCount:   playerCount,
		PlayerGuesses: make(map[int][]int),
	}

	// Add to history
	ga.gameHistory = append(ga.gameHistory, stats)
	return stats
}

// RecordGuess tracks a player's guess
func (ga *GameAnalytics) RecordGuess(stats *GameStats, playerID int, guess int) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	// Increment total guesses for this game
	stats.GuessCount++

	// Record player's guess
	if _, exists := stats.PlayerGuesses[playerID]; !exists {
		stats.PlayerGuesses[playerID] = make([]int, 0)
	}
	stats.PlayerGuesses[playerID] = append(stats.PlayerGuesses[playerID], guess)

	// Track guess frequency
	ga.guessCounts[guess]++

	// Initialize player stats if not exists
	if _, exists := ga.playerStats[playerID]; !exists {
		ga.playerStats[playerID] = &PlayerStats{
			GamesPlayed: 0,
			GamesWon:    0,
			BestGame:    0,
		}
	}

	// Update total guesses for this player
	ga.playerStats[playerID].TotalGuesses++
}

// EndGame completes tracking for a game
func (ga *GameAnalytics) EndGame(stats *GameStats, winnerID int) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	stats.EndTime = time.Now()
	stats.Won = (winnerID > 0) // If winnerID is 0, game was abandoned or lost

	if winnerID > 0 {
		ga.gamesWon++

		// Update player stats
		if _, exists := ga.playerStats[winnerID]; !exists {
			ga.playerStats[winnerID] = &PlayerStats{
				GamesPlayed: 0,
				GamesWon:    0,
				BestGame:    0,
			}
		}

		playerStats := ga.playerStats[winnerID]
		playerStats.GamesWon++

		// Calculate player's guesses in this game
		playerGuesses := 0
		if guesses, exists := stats.PlayerGuesses[winnerID]; exists {
			playerGuesses = len(guesses)
		}

		// Update best game if this is better or first win
		if playerStats.BestGame == 0 || playerGuesses < playerStats.BestGame {
			playerStats.BestGame = playerGuesses
		}
	}

	// Ensure each player who participated has their GamesPlayed incremented
	for playerID := range stats.PlayerGuesses {
		if _, exists := ga.playerStats[playerID]; !exists {
			ga.playerStats[playerID] = &PlayerStats{
				GamesPlayed: 0,
				GamesWon:    0,
				BestGame:    0,
			}
		}
		ga.playerStats[playerID].GamesPlayed++
	}
}

// GetHardestNumbers returns the top N hardest numbers to guess
func (ga *GameAnalytics) GetHardestNumbers(n int) []struct {
	Number     int
	AvgGuesses float64
	Frequency  int
} {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	type numberStats struct {
		number     int
		avgGuesses float64
		frequency  int
	}

	// Map to store number -> total guesses and frequency
	numberData := make(map[int]struct {
		totalGuesses int
		frequency    int
	})

	// Compile data for each secret code
	for _, game := range ga.gameHistory {
		if game.Won {
			number := game.SecretCode
			data := numberData[number]
			data.totalGuesses += game.GuessCount
			data.frequency++
			numberData[number] = data
		}
	}

	// Convert to slice and calculate average guesses
	stats := make([]numberStats, 0, len(numberData))
	for number, data := range numberData {
		if data.frequency > 0 {
			avg := float64(data.totalGuesses) / float64(data.frequency)
			stats = append(stats, numberStats{
				number:     number,
				avgGuesses: avg,
				frequency:  data.frequency,
			})
		}
	}

	// Sort by average guesses (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].avgGuesses > stats[j].avgGuesses
	})

	// Take top N
	if n > len(stats) {
		n = len(stats)
	}
	stats = stats[:n]

	// Convert to return format
	result := make([]struct {
		Number     int
		AvgGuesses float64
		Frequency  int
	}, len(stats))

	for i, stat := range stats {
		result[i] = struct {
			Number     int
			AvgGuesses float64
			Frequency  int
		}{
			Number:     stat.number,
			AvgGuesses: stat.avgGuesses,
			Frequency:  stat.frequency,
		}
	}

	return result
}

// GetMostCommonGuesses returns the top N most common guesses
func (ga *GameAnalytics) GetMostCommonGuesses(n int) []struct {
	Guess     int
	Frequency int
} {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	// Convert map to slice
	guesses := make([]struct {
		guess     int
		frequency int
	}, 0, len(ga.guessCounts))

	for guess, count := range ga.guessCounts {
		guesses = append(guesses, struct {
			guess     int
			frequency int
		}{
			guess:     guess,
			frequency: count,
		})
	}

	// Sort by frequency (descending)
	sort.Slice(guesses, func(i, j int) bool {
		return guesses[i].frequency > guesses[j].frequency
	})

	// Take top N
	if n > len(guesses) {
		n = len(guesses)
	}
	guesses = guesses[:n]

	// Convert to return format
	result := make([]struct {
		Guess     int
		Frequency int
	}, len(guesses))

	for i, g := range guesses {
		result[i] = struct {
			Guess     int
			Frequency int
		}{
			Guess:     g.guess,
			Frequency: g.frequency,
		}
	}

	return result
}

// GetOverallStats returns overall game statistics
func (ga *GameAnalytics) GetOverallStats() struct {
	GamesPlayed       int
	GamesWon          int
	AvgGuessesPerGame float64
	AvgGuessesPerWin  float64
	TotalPlayers      int
	AvgPlayersPerGame float64
	TotalDuration     time.Duration
	AvgGameDuration   time.Duration
} {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	result := struct {
		GamesPlayed       int
		GamesWon          int
		AvgGuessesPerGame float64
		AvgGuessesPerWin  float64
		TotalPlayers      int
		AvgPlayersPerGame float64
		TotalDuration     time.Duration
		AvgGameDuration   time.Duration
	}{
		GamesPlayed: ga.gamesPlayed,
		GamesWon:    ga.gamesWon,
	}

	// Calculate other stats
	totalGuesses := 0
	winGuesses := 0
	totalPlayers := 0
	totalDuration := time.Duration(0)
	playersSet := make(map[int]struct{})

	for _, game := range ga.gameHistory {
		totalGuesses += game.GuessCount
		if game.Won {
			winGuesses += game.GuessCount
		}

		// Count unique players
		for playerID := range game.PlayerGuesses {
			playersSet[playerID] = struct{}{}
		}

		// Add to player count
		totalPlayers += game.PlayerCount

		// Add duration if game is completed
		if !game.EndTime.IsZero() {
			duration := game.EndTime.Sub(game.StartTime)
			totalDuration += duration
		}
	}

	// Calculate averages
	if ga.gamesPlayed > 0 {
		result.AvgGuessesPerGame = float64(totalGuesses) / float64(ga.gamesPlayed)
		result.AvgPlayersPerGame = float64(totalPlayers) / float64(ga.gamesPlayed)
		result.AvgGameDuration = totalDuration / time.Duration(ga.gamesPlayed)
	}

	if ga.gamesWon > 0 {
		result.AvgGuessesPerWin = float64(winGuesses) / float64(ga.gamesWon)
	}

	result.TotalPlayers = len(playersSet)
	result.TotalDuration = totalDuration

	return result
}

// GetPlayerStats returns statistics for a specific player
func (ga *GameAnalytics) GetPlayerStats(playerID int) *PlayerStats {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	if stats, exists := ga.playerStats[playerID]; exists {
		return stats
	}
	return nil
}

// GetTopPlayers returns the top N players by win rate
func (ga *GameAnalytics) GetTopPlayers(n int) []struct {
	PlayerID int
	WinRate  float64
	GamesWon int
} {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	type playerStat struct {
		id       int
		winRate  float64
		gamesWon int
	}

	// Get players with at least one game
	playerStats := make([]playerStat, 0, len(ga.playerStats))
	for id, stats := range ga.playerStats {
		if stats.GamesPlayed > 0 {
			winRate := float64(stats.GamesWon) / float64(stats.GamesPlayed)
			playerStats = append(playerStats, playerStat{
				id:       id,
				winRate:  winRate,
				gamesWon: stats.GamesWon,
			})
		}
	}

	// Sort by win rate (descending)
	sort.Slice(playerStats, func(i, j int) bool {
		if playerStats[i].winRate == playerStats[j].winRate {
			return playerStats[i].gamesWon > playerStats[j].gamesWon
		}
		return playerStats[i].winRate > playerStats[j].winRate
	})

	// Take top N
	if n > len(playerStats) {
		n = len(playerStats)
	}
	playerStats = playerStats[:n]

	// Convert to return format
	result := make([]struct {
		PlayerID int
		WinRate  float64
		GamesWon int
	}, len(playerStats))

	for i, p := range playerStats {
		result[i] = struct {
			PlayerID int
			WinRate  float64
			GamesWon int
		}{
			PlayerID: p.id,
			WinRate:  p.winRate,
			GamesWon: p.gamesWon,
		}
	}

	return result
}

// GetAnalyticsReport generates a formatted analytics report
func (ga *GameAnalytics) GetAnalyticsReport() string {
	overallStats := ga.GetOverallStats()
	hardestNumbers := ga.GetHardestNumbers(5)
	mostCommonGuesses := ga.GetMostCommonGuesses(5)
	topPlayers := ga.GetTopPlayers(5)

	report := "=== CODE BREAKER GAME ANALYTICS ===\n\n"

	// Overall stats
	report += fmt.Sprintf("OVERALL STATISTICS:\n")
	report += fmt.Sprintf("Games Played: %d\n", overallStats.GamesPlayed)
	report += fmt.Sprintf("Games Won: %d (%.1f%%)\n", overallStats.GamesWon, float64(overallStats.GamesWon)/float64(overallStats.GamesPlayed)*100)
	report += fmt.Sprintf("Average Guesses Per Game: %.2f\n", overallStats.AvgGuessesPerGame)
	report += fmt.Sprintf("Average Guesses Per Win: %.2f\n", overallStats.AvgGuessesPerWin)
	report += fmt.Sprintf("Total Unique Players: %d\n", overallStats.TotalPlayers)
	report += fmt.Sprintf("Average Players Per Game: %.2f\n", overallStats.AvgPlayersPerGame)
	report += fmt.Sprintf("Average Game Duration: %s\n\n", overallStats.AvgGameDuration.Round(time.Second))

	// Hardest numbers
	report += fmt.Sprintf("TOP 5 HARDEST NUMBERS TO GUESS:\n")
	if len(hardestNumbers) == 0 {
		report += "No data available yet\n"
	} else {
		for i, num := range hardestNumbers {
			report += fmt.Sprintf("%d. Number %d - %.2f guesses on average (appeared %d times)\n",
				i+1, num.Number, num.AvgGuesses, num.Frequency)
		}
	}
	report += "\n"

	// Most common guesses
	report += fmt.Sprintf("TOP 5 MOST COMMON GUESSES:\n")
	if len(mostCommonGuesses) == 0 {
		report += "No data available yet\n"
	} else {
		for i, guess := range mostCommonGuesses {
			report += fmt.Sprintf("%d. %d - guessed %d times\n",
				i+1, guess.Guess, guess.Frequency)
		}
	}
	report += "\n"

	// Top players
	report += fmt.Sprintf("TOP 5 PLAYERS BY WIN RATE:\n")
	if len(topPlayers) == 0 {
		report += "No data available yet\n"
	} else {
		for i, player := range topPlayers {
			report += fmt.Sprintf("%d. Player %d - %.1f%% win rate (%d wins)\n",
				i+1, player.PlayerID, player.WinRate*100, player.GamesWon)
		}
	}

	return report
}
