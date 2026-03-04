package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Leaderboard struct {
	Board []LeaderboardEntry `json:"board"`
}

type LeaderboardEntry struct {
	Username string `json:"username"`
	Wins     int `json:"wins"`
	Losses   int `json:"losses"`
}

func (lb *Leaderboard) AddPlayerToLeaderboard(username string) {
	for i := range lb.Board {
		if lb.Board[i].Username == username {
			return
		}
	}

	lb.Board = append(lb.Board, LeaderboardEntry{
		Username: username,
		Wins: 0,
		Losses: 0,
	})
}

func (lb *Leaderboard) UpdateLeaderboard(result GameResult) {
	for i := range lb.Board {
		if lb.Board[i].Username == result.winner {
			lb.Board[i].Wins++
		}
		if lb.Board[i].Username == result.loser {
			lb.Board[i].Losses++
		}
	}
}

// / return a json payload of the current leaderboard
func (s *Server) getLeaderboard(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(s.leaderboard)
	if err != nil {
		errorStr := fmt.Errorf("getLeaderboard error: %s", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
