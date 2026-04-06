package checkered

import (
	"log"
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

// / return a json payload of the current leaderboard
// func (s *GameServer) GetLeaderboard(w http.ResponseWriter, _ *http.Request) {
// 	err := json.NewEncoder(w).Encode(s.leaderboard)
// 	if err != nil {
// 		errorStr := fmt.Errorf("getLeaderboard error: %s", err)
// 		fmt.Println(errorStr)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
