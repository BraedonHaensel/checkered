package checkered

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

type Leaderboard struct {
	Board []LeaderboardEntry `json:"board"`
}

type LeaderboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
}

// Adds a player to the leaderboard if they are not in it already
func (lb *Leaderboard) AddPlayerToLeaderboard(username string) {
	for i := range lb.Board {
		if lb.Board[i].Username == username {
			return
		}
	}

	lb.Board = append(lb.Board, LeaderboardEntry{
		Username: username,
		Wins:     0,
		Losses:   0,
	})
}

// Updates the leaderboard for a game result. Returns false if it's a draw, true otherwise
func (lb *Leaderboard) UpdateLeaderboard(result GameResult) bool {
	if result.IsDraw {
		// Game ended in a draw, no leaderboard update required
		log.Printf("Game Result: (game=%s, draw)", result.GameID)
		lb.AddPlayerToLeaderboard(result.Winner)
		lb.AddPlayerToLeaderboard(result.Loser)
		return false
	}

	winner := result.Winner
	loser := result.Loser
	log.Printf("Game Result: (game=%s, winner=%s, loser=%s)", result.GameID, winner, loser)

	// Ensure each player is in the leaderboard
	lb.AddPlayerToLeaderboard(result.Winner)
	lb.AddPlayerToLeaderboard(result.Loser)

	// Update the win/loss scores for each player
	for i := range lb.Board {
		if lb.Board[i].Username == result.Winner {
			lb.Board[i].Wins++
			log.Printf("New score: %s = w%d/l%d\n",
				result.Winner,
				lb.Board[i].Wins,
				lb.Board[i].Losses)
		}
		if lb.Board[i].Username == result.Loser {
			lb.Board[i].Losses++
			log.Printf("New score: %s = w%d/l%d\n",
				result.Loser,
				lb.Board[i].Wins,
				lb.Board[i].Losses)
		}
	}

	return true
}

// Saves the leaderboard to disk.
func (lb *Leaderboard) SaveBackupToDisk(filename string) {
	// Marshal to JSON
	jsonData, err := json.Marshal(lb)
	if err != nil {
		log.Fatal("Failed to marshal leaderboard:", err)
	}

	// Write to the backup file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Fatal("Failed to write leaderboard backup:", err)
	}
}

// Loads and returns the leaderboard from disk.
func (lb *Leaderboard) LoadBackupFromDisk(filename string) *Leaderboard {
	// Read from the backup file
	fileData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to load leaderboard backup:", err)
	}

	// Parse the JSON file data
	var leaderboardData Leaderboard
	decoder := json.NewDecoder(bytes.NewReader(fileData))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&leaderboardData)
	if err != nil {
		log.Fatal("Failed to decode leaderboard backup:", err)
	}

	return &leaderboardData
}
