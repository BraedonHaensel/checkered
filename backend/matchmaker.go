package checkered

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Responsible for handling
// the queue for new players
// finding a game, as well as
// maintaining the leaderboard
// and handling all leaderboard requests
type Matchmaker struct {
	mu_leaderboard sync.Mutex
	leaderboard    Leaderboard
	queue          Queue[*Client]
}

func NewMatchmaker() Matchmaker {
	queue := Queue[*Client]{}
	InitQueue(&queue, 100)
	return Matchmaker{
		mu_leaderboard: sync.Mutex{},
		leaderboard:    Leaderboard{},
		queue:          queue,
	}
}

func (m *Matchmaker) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(m.leaderboard)
	if err != nil {
		errorStr := fmt.Errorf("getLeaderboard error: %s", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for a request to join the queue
func (m *Matchmaker) AddToQueue(w http.ResponseWriter, r *http.Request) {

}

// Handler for a request for a users status on the queue
// will respond with still in the queue or will respond with
// a new found match
func (m *Matchmaker) QueuePollRequest(w http.ResponseWriter, r *http.Request) {

}

// Handler for a request to leave the queue
func (m *Matchmaker) LeaveQueueRequest(w http.ResponseWriter, r *http.Request) {

}
