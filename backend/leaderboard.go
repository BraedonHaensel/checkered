package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Leaderboard struct {
	board []LeaderboardEntry
}

type LeaderboardEntry struct {
	username string
	wins     int
	losses   int
}

func (lb *Leaderboard) AddPlayerToLeaderboard(username string) {
}

func (lb *Leaderboard) UpdateLeaderboard(result GameResult) {
}

// / return a json payload of the current leaderboard
func (s *Server) getLeaderboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s.leaderboard)
	if err != nil {
		errorStr := fmt.Errorf("getLeaderboard error: %s", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
