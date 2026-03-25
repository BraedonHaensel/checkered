package checkered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// -------------------- INITIAL ADMIN SET UP	 --------------------

const LEADER_ELECTION_TIMEOUT_SEC = 5 * time.Second

// Responsible for handling the queue for new players finding a game, as well as
// maintaining the leaderboard and handling all leaderboard requests
type Matchmaker struct {
	ID int
	// games that new clients are in
	matches        map[uuid.UUID]*Match
	mu_matches     sync.Mutex
	queue          Queue[string]
	mu_queue       sync.Mutex
	leaderboard    *Leaderboard
	mu_leaderboard sync.Mutex

	// Fully qualified URL of this Matchmaker
	URL string

	// URL of the Name Server
	nameServerURL string

	// Other Matchmakers in the network
	otherMatchmakers   []Server
	otherMatchmakersMu sync.Mutex
	// Whether this server is running in the current leader election
	runningInElection   bool
	runningInElectionMu sync.Mutex
	// ID of the current leader server
	Leader     Server
	leaderIDMu sync.Mutex
	// Timer and chan to wait for and detect receiving a bully() response
	bullyTimer     *time.Timer
	bullyTimerChan chan struct{}
	// Timer and chan to wait for and detect receiving a leader(i) response
	leaderTimer     *time.Timer
	leaderTimerChan chan struct{}
}

func NewMatchmaker(url, nameServerURL string) *Matchmaker {
	queue := Queue[string]{}
	InitQueue(&queue, 100)
	matchmaker := Matchmaker{
		matches: make(map[uuid.UUID]*Match),
		queue:   queue,
		leaderboard: &Leaderboard{
			Board: make([]LeaderboardEntry, 0),
		},
		mu_leaderboard: sync.Mutex{},

		URL:               url,
		nameServerURL:     nameServerURL,
		runningInElection: false,
		otherMatchmakers:  []Server{},
	}
	return &matchmaker
}

// -------------------- DISTRIBUTED SYSTEMS / LEADER ELECTION USING BULLY ALGORITHM --------------------

// Register with the Name Server
func (m *Matchmaker) Register(url string) {
	id, err := SendRegistrationRequest(url, m.nameServerURL+"/register/matchmaker")
	if err != nil {
		log.Fatal(err)
	}
	m.ID = id
	log.Println("Registered with ID:", m.ID)
	log.SetPrefix(fmt.Sprintf("[%d] ", m.ID))
}

func (m *Matchmaker) IsLeader() bool {
	return m.ID == m.Leader.ID
}

// Refreshes and sets the list of known other Matchmakers.
func (m *Matchmaker) RefreshOtherMatchmakersList() {
	// Get the list of all Matchmakers from the Name Server
	matchmakers, err := SendServerListRequest(m.nameServerURL + "/matchmakers")
	if err != nil {
		log.Println(err)
	}

	// Filter itself out of the list
	otherMatchmakers := []Server{}
	foundInList := false
	for i, matchmaker := range matchmakers {
		if matchmaker.ID == m.ID {
			otherMatchmakers = append(matchmakers[:i],
				matchmakers[i+1:]...)
			foundInList = true
			break
		}
	}

	if !foundInList {
		// Failed to find itself in the Name Server's list, something went wrong
		log.Fatalf("Matchmaker %d failed to find itself in the Name Server's list of Matchmakers", m.ID)
	}

	// Set the list of known other Matchmakers
	m.otherMatchmakersMu.Lock()
	log.Println("Refreshed the list of other Matchmakers")
	m.otherMatchmakers = otherMatchmakers
	m.otherMatchmakersMu.Unlock()
}

// Deregisters another Matchmaker from the Name Server.
func (m *Matchmaker) DeregisterOtherMatchmaker(otherMatchmakerID int) {
	// Create the registration request
	log.Println("Deregistering Matchmaker", otherMatchmakerID)
	body := fmt.Appendf(nil, `{"id": %d}`, otherMatchmakerID)
	res, err := http.Post(m.nameServerURL+"/deregister/matchmaker", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to deregister Matchmaker %d: %v", otherMatchmakerID, err)
	}
	defer res.Body.Close()

	// Filter the Matchmaker out of the list of other Matchmakers
	m.otherMatchmakersMu.Lock()
	for i, matchmaker := range m.otherMatchmakers {
		if matchmaker.ID == otherMatchmakerID {
			m.otherMatchmakers = append(m.otherMatchmakers[:i],
				m.otherMatchmakers[i+1:]...)
			break
		}
	}
	m.otherMatchmakersMu.Unlock()
}

// Initiates a leader election using the Bully algorithm.
func (m *Matchmaker) InitiateElection() {
	m.runningInElectionMu.Lock()
	log.Println("Initiating a leader election")
	m.runningInElection = true
	m.runningInElectionMu.Unlock()
	m.RefreshOtherMatchmakersList()

	// Check which Matchmakers have a higher ID than this one
	higherIDMatchmakers := []Server{}
	m.otherMatchmakersMu.Lock()
	for _, otherMatchmaker := range m.otherMatchmakers {
		if otherMatchmaker.ID > m.ID {
			higherIDMatchmakers = append(higherIDMatchmakers, otherMatchmaker)
		}
	}
	m.otherMatchmakersMu.Unlock()

	if len(higherIDMatchmakers) == 0 {
		// This server has the highest ID, declare itself leader
		m.leaderIDMu.Lock()
		log.Println("Declaring itself leader as it has the highest ID:", m.ID)
		m.Leader = Server{
			ID:  m.ID,
			URL: m.URL,
		}
		m.leaderIDMu.Unlock()
		m.runningInElectionMu.Lock()
		m.runningInElection = false
		m.runningInElectionMu.Unlock()
		m.otherMatchmakersMu.Lock()
		log.Printf("Sending leader(%d) messages\n", m.ID)
		downMatchmakers := []Server{}
		for _, otherMatchmaker := range m.otherMatchmakers {
			if !m.sendLeaderMessage(otherMatchmaker) {
				downMatchmakers = append(downMatchmakers, otherMatchmaker)
			}
		}
		m.otherMatchmakersMu.Unlock()

		// Deregister any Matchmakers that did not respond
		for _, otherMatchmaker := range downMatchmakers {
			m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
		}
		return
	}

	// Start the bully response timeout before sending the election(i) messages to avoid race conditions
	m.bullyTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
	// The chan is used to interrupt waiting for the timer when a bully() is received
	m.bullyTimerChan = make(chan struct{})

	// Send election(i) to those with a higher ID
	log.Printf("Sending election(%d) messages to servers with higher IDs\n", m.ID)
	downMatchmakers := []Server{}
	for _, otherMatchmaker := range higherIDMatchmakers {
		if !m.sendElectionMessage(otherMatchmaker) {
			downMatchmakers = append(downMatchmakers, otherMatchmaker)
		}
	}

	// Deregister any Matchmakers that did not respond
	for _, otherMatchmaker := range downMatchmakers {
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
	}

	// Wait for a bully() response
	log.Printf("Waiting up to %dms for a bully() response\n",
		LEADER_ELECTION_TIMEOUT_SEC.Milliseconds())

	select {
	case <-m.bullyTimer.C:
		// Timer fired, so no bully() responses received in time. Declare itself leader
		m.leaderIDMu.Lock()
		log.Println("No bully() responses received. Declaring itself leader with ID:", m.ID)
		m.Leader = Server{
			ID:  m.ID,
			URL: m.URL,
		}
		m.leaderIDMu.Unlock()
		m.runningInElectionMu.Lock()
		m.runningInElection = false
		m.runningInElectionMu.Unlock()
		m.otherMatchmakersMu.Lock()
		log.Printf("Sending leader(%d) messages\n", m.ID)
		downMatchmakers := []Server{}
		for _, otherMatchmaker := range m.otherMatchmakers {
			if !m.sendLeaderMessage(otherMatchmaker) {
				downMatchmakers = append(downMatchmakers, otherMatchmaker)
			}
		}
		m.otherMatchmakersMu.Unlock()

		// Deregister any Matchmakers that did not respond
		for _, otherMatchmaker := range downMatchmakers {
			m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
		}
		return

	case <-m.bullyTimerChan:
		// Bullied before the timer fired. Wait for a leader(i) message
		log.Printf("Received a bully() response. Waiting up to %dms for a leader(i) message\n",
			LEADER_ELECTION_TIMEOUT_SEC.Milliseconds())
		m.leaderTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
		m.leaderTimerChan = make(chan struct{})
		select {
		case <-m.leaderTimer.C:
			// Timer fired, so no leader(i) received. Something went wrong, so initiate
			// a new election
			log.Println("No leader(i) responses received")
			m.InitiateElection()
		case <-m.leaderTimerChan:
			// Received a leader(i) message. The message is handled by the leader(i)
			// message handler, so return
			return
		}

	}
}

// Sends an election(i) message to another Matchmaker. Returns true if the
// recipient is alive.
func (m *Matchmaker) sendElectionMessage(otherMatchmaker Server) bool {
	// Create the election(i) message data
	data := Server{
		ID:  m.ID,
		URL: m.URL,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	// Send an election(i) message
	res, err := http.Post(otherMatchmaker.URL+"/internal/leader-election/election", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send an election(%d) message to Matchmaker %d, assuming it is down", m.ID, otherMatchmaker.ID)
		return false
	}
	defer res.Body.Close()
	return true
}

// Sends a leader(i) message to another Matchmaker. Returns true if the recipient
// is alive.
func (m *Matchmaker) sendLeaderMessage(otherMatchmaker Server) bool {
	// Create the leader(i) message data
	data := Server{
		ID:  m.ID,
		URL: m.URL,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	// Send a leader(i) message
	res, err := http.Post(otherMatchmaker.URL+"/internal/leader-election/leader", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send a leader(%d) message to Matchmaker %d, assuming it is down", m.ID, otherMatchmaker.ID)
		return false
	}
	defer res.Body.Close()
	return true
}

// Sends a bully() message to another Matchmaker.
func (m *Matchmaker) sendBullyMessage(otherMatchmaker Server) {
	// Send a bully() message
	res, err := http.Post(otherMatchmaker.URL+"/internal/leader-election/bully", "application/json", nil)
	if err != nil {
		log.Printf("Failed to send a bully() message to Matchmaker %d, assuming it is down", otherMatchmaker.ID)
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
		return
	}
	defer res.Body.Close()
}

// Handle an incoming election(i) Bully leader election message.
func (m *Matchmaker) HandleElectionRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the elected Matchmaker's ID from the request
	otherMatchmaker, err, errStatus := parseJsonRequestData[Server](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	id := otherMatchmaker.ID

	// Check if the message is from an unknown Matchmaker
	m.otherMatchmakersMu.Lock()
	if !slices.Contains(m.otherMatchmakers, otherMatchmaker) {
		m.otherMatchmakers = append(m.otherMatchmakers, otherMatchmaker)
	}
	m.otherMatchmakersMu.Unlock()

	if id < m.ID {
		// Message received from a server with a lower ID, so bully them.
		log.Printf("Received election(%d). Bullying as this Matchmaker's ID is higher: %d\n", id, m.ID)
		m.sendBullyMessage(otherMatchmaker)
		m.runningInElectionMu.Lock()
		defer m.runningInElectionMu.Unlock()
		if !m.runningInElection {
			// This server has a higher ID and isn't running yet, so start an election
			go m.InitiateElection()
		}
	}
}

// Handle an incoming leader(i) Bully leader election message.
func (m *Matchmaker) HandleLeaderRequest(w http.ResponseWriter, r *http.Request) {
	// If this server was waiting, interrupt the leader timer so it never fires
	if m.leaderTimer != nil && m.leaderTimer.Stop() {
		if m.leaderTimerChan != nil {
			// Close the leader timer chan to notify the thread to stop waiting for the timer
			close(m.leaderTimerChan)
		}
	}

	// Parse the Matchmaker's ID from the request
	otherMatchmaker, err, errStatus := parseJsonRequestData[Server](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}

	// Check if the message is from an unknown Matchmaker
	m.otherMatchmakersMu.Lock()
	if !slices.Contains(m.otherMatchmakers, otherMatchmaker) {
		m.otherMatchmakers = append(m.otherMatchmakers, otherMatchmaker)
	}
	m.otherMatchmakersMu.Unlock()

	// Set the new leader
	m.leaderIDMu.Lock()
	log.Println("Received a new leader ID:", otherMatchmaker.ID)
	m.Leader = otherMatchmaker
	m.leaderIDMu.Unlock()
	m.runningInElectionMu.Lock()
	m.runningInElection = false
	m.runningInElectionMu.Unlock()
}

// Handle an incoming bully() Bully leader election message.
func (m *Matchmaker) HandleBullyRequest(w http.ResponseWriter, r *http.Request) {
	// Interrupt the bully timer so it never fires
	if m.bullyTimer != nil && m.bullyTimer.Stop() {
		if m.bullyTimerChan != nil {
			// Close the bully timer chan to notify the thread to stop waiting for the timer
			close(m.bullyTimerChan)
		}
	}
}

// -------------------- MATCHMAKING SERVER LOGIC --------------------

func (m *Matchmaker) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(m.leaderboard)
	if err != nil {
		errorStr := fmt.Errorf("getLeaderboard error: %v", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type QueueRequest struct {
	Username string `json:"username"`
}

type QueueResponse struct {
	Type string `json:"type"`
}

type PollResponse struct {
	Type       string `json:"type"`
	GameServer Server `json:"game_server"`
}

type RequestNewGameServerRequest struct {
	OldGameServerID int `json:"old_game_server_id"`
}

type RequestNewGameServerResponse struct {
	Type string `json:"type"`
}

type EndMatchRequest struct {
	MatchID uuid.UUID `json:"match_id"`
}

type EndMatchResponse struct {
	Type string `json:"type"`
}

// Fix - export the fields and add json tags
type GameResultStruct struct {
	GameID uuid.UUID `json:"game_id"`
	Winner string    `json:"winner"`
	Loser  string    `json:"loser"`
}

type Match struct {
	MatchID     uuid.UUID `json:"match_id"`
	RedPlayer   string    `json:"red"`
	BlackPlayer string    `json:"black"`
	GameServer  Server    `json:"server"`
}

func (m *Matchmaker) NewMatch(redPlayer, blackPlayer string) (Match, error) {

	// Get all available game servers
	gameServers, err := SendServerListRequest(m.nameServerURL + "/game-servers")
	if err != nil {
		log.Printf("Failed to fetch game servers: %s", err)
		return Match{}, err
	}

	if len(gameServers) == 0 {
		log.Println("No game servers available")
		return Match{}, fmt.Errorf("no game servers available")
	}

	// Pick the first one
	gameServer := gameServers[0]
	log.Printf("Picked game server: %s (ID: %d)", gameServer.URL, gameServer.ID)

	return Match{
		MatchID:     uuid.New(),
		RedPlayer:   redPlayer,
		BlackPlayer: blackPlayer,
		GameServer:  gameServer,
	}, nil
}

// Handler for a request to join the queue
func (m *Matchmaker) AddToQueue(w http.ResponseWriter, r *http.Request) {

	// Reading the data in the request
	data, err, errStatus := parseJsonRequestData[QueueRequest](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}

	// Extracting the username
	username := data.Username

	// Check if user is already in queue
	alreadyQueued := false
	m.mu_queue.Lock()
	m.queue.forEach(func(u string, i int) {
		if u == username {
			alreadyQueued = true
		}
	})
	m.mu_queue.Unlock()

	if alreadyQueued {
		log.Printf("Player \"%s\" already in queue", username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QueueResponse{Type: "ALREADY_IN_QUEUE"})
		return
	}

	// Enqueue the user
	m.mu_queue.Lock()
	m.queue.enqueue(username)
	m.mu_queue.Unlock()
	log.Printf("Adding \"%s\" to queue", username)

	// Matchmaking logic
	if m.queue.Size >= 2 {
		m.mu_queue.Lock()
		redPlayer := m.queue.dequeue()
		blackPlayer := m.queue.dequeue()
		m.mu_queue.Unlock()
		match, err := m.NewMatch(redPlayer, blackPlayer)
		if err != nil {
			log.Printf("Failed to create match: %s", err)
			return
		}
		m.mu_matches.Lock()
		m.matches[match.MatchID] = &match
		m.mu_matches.Unlock()

		// Marshal the match to JSON
		matchBytes, err := json.Marshal(match)
		if err != nil {
			log.Printf("Failed to marshal match: %s", err)
			return
		}

		m.broadcastMatchesChanged()

		// Send a POST request to the game server with the match details
		res, err := http.Post(match.GameServer.URL+"/newGame", "application/json", bytes.NewBuffer(matchBytes))
		if err != nil {
			log.Printf("Failed to send match to game server [%d] %s: %s", match.GameServer.ID, match.GameServer.URL, err)
			return
		}
		defer res.Body.Close()

		log.Printf("Match created: %s (red) vs %s (black)", redPlayer, blackPlayer)
	}

	m.broadcastQueueChanged()

	// Let the user know they're in queue
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "SUCCESS"})

}

// Handler for a request for a users status on the queue will respond with still
// in the queue or will respond with a new found match
func (m *Matchmaker) QueuePollRequest(w http.ResponseWriter, r *http.Request) {

	// Reading the data in the request
	username := r.URL.Query().Get("username")
	m.mu_matches.Lock()

	// Check if the user is in a match
	for _, match := range m.matches {
		if match.RedPlayer == username || match.BlackPlayer == username {
			m.mu_matches.Unlock()
			// User has been matched, return the game server URL
			log.Printf("User \"%s\" has been matched, sending game server [%d] %s", username, match.GameServer.ID, match.GameServer.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PollResponse{Type: "IN_GAME", GameServer: match.GameServer})
			return
		}
	}
	m.mu_matches.Unlock()

	// Check if user is already in queue
	m.mu_queue.Lock()
	alreadyQueued := false
	m.queue.forEach(func(u string, i int) {
		if u == username {
			alreadyQueued = true
		}
	})
	m.mu_queue.Unlock()

	if alreadyQueued {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PollResponse{Type: "IN_QUEUE"})
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PollResponse{Type: "NOT_IN_QUEUE"})
		return
	}

}

// helper function to LeaveQueueRequest to remove value from the queue
func RemoveStringValue(q *Queue[string], value string) {
	originalSize := q.Size
	for i := 0; i < originalSize; i++ {
		v := q.dequeue()
		if v != value {
			q.enqueue(v)
		}
	}
}

// Handler for a request to leave the queue
func (m *Matchmaker) LeaveQueueRequest(w http.ResponseWriter, r *http.Request) {

	// Reading the data in the reuest
	data, err, errStatus := parseJsonRequestData[QueueRequest](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}

	username := data.Username

	// Check if user is actually in the queue
	m.mu_queue.Lock()
	inQueue := false
	m.queue.forEach(func(u string, i int) {
		if u == username {
			inQueue = true
		}
	})
	m.mu_queue.Unlock()

	if !inQueue {
		log.Printf("User \"%s\" tried to leave queue but was not in it", username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QueueResponse{Type: "ALREADY_NOT_IN_QUEUE"})
		return
	}
	m.mu_queue.Lock()
	RemoveStringValue(&m.queue, username)
	m.mu_queue.Unlock()
	log.Printf("User \"%s\" left the queue", username)

	m.broadcastQueueChanged()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "SUCCESS"})

}

// Checks the health of a server. Returns true if healthy, false otherwise
func (m *Matchmaker) checkServerHealth(url string, timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Handler for a request for a new Game Server to host a match
func (m *Matchmaker) RequestNewGameServer(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Getting server to shut downnnnnnnnnnn")
	data, err, errStatus := parseJsonRequestData[RequestNewGameServerRequest](r)
	if err != nil {
		fmt.Println("Errrr", err)
		http.Error(w, err.Error(), errStatus)
		return
	}

	fmt.Println("Server id to SHUT DOWNNNNNN", data.OldGameServerID)

	if m.checkServerHealth("http://localhost:4000", 5*time.Second) {
		// Game Server is healthy, keep it
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "HEALTHY"})
		return
	}

	// Get all available game servers
	// gameServers, err := SendServerListRequest(m.nameServerURL + "/game-servers")
	// if err != nil {
	// 	log.Printf("Failed to fetch game servers: %s", err)
	// 	return Match{}, err
	// }

	// if len(gameServers) == 0 {
	// 	log.Println("No game servers available")
	// 	return Match{}, fmt.Errorf("no game servers available")
	// }

	// Pick a new matchmaker
	// Will temp code a random matchmaker from list
	// Taylor TODO redirect to one with the latest game state backup
	fmt.Println("TODO Processing new game server request...")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "SUCCESS"})
}

// Handler for a request to update the leaderboard
func (m *Matchmaker) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {

	data, err, errStatus := parseJsonRequestData[GameResultStruct](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	log.Printf("Game Result: (game=%s, winner=%s, loser=%s)", data.GameID, data.Winner, data.Loser)

	// Update the leaderboard
	m.mu_leaderboard.Lock()
	m.leaderboard.AddPlayerToLeaderboard(data.Winner)
	m.leaderboard.AddPlayerToLeaderboard(data.Loser)
	for i := range m.leaderboard.Board {
		if m.leaderboard.Board[i].Username == data.Winner {
			m.leaderboard.Board[i].Wins++
		}
		if m.leaderboard.Board[i].Username == data.Loser {
			m.leaderboard.Board[i].Losses++
		}
	}
	m.mu_leaderboard.Unlock()

	m.broadcastLeaderboardChanged()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "leaderboard updated"})
}

// Handler for a request to end a matchup
func (m *Matchmaker) EndMatch(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[EndMatchRequest](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}

	matchID := data.MatchID

	// Check if the match exists
	m.mu_matches.Lock()
	_, exists := m.matches[matchID]
	if !exists {
		m.mu_matches.Unlock()
		log.Printf("EndMatch: match %s not found", matchID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(EndMatchResponse{Type: "match not found"})
		return
	}

	// Remove the match
	delete(m.matches, matchID)
	m.mu_matches.Unlock()
	log.Printf("Match %s ended and removed", matchID)

	m.broadcastMatchesChanged()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "match Removed"})
}

func (m *Matchmaker) SetLeaderboard(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[Leaderboard](r)
	if err != nil {
		errorStr := fmt.Errorf("setLeaderboard error: %v", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), errStatus)
		return
	}

	m.mu_leaderboard.Lock()
	log.Printf("New leaderboard: %v", data)
	m.leaderboard = &data
	m.mu_leaderboard.Unlock()
}

func (m *Matchmaker) SetQueue(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[Queue[string]](r)
	if err != nil {
		errorStr := fmt.Errorf("setQueue error: %v", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), errStatus)
		return
	}

	m.mu_queue.Lock()
	m.queue = data
	m.mu_queue.Unlock()
}

func (m *Matchmaker) SetMatches(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[map[uuid.UUID]*Match](r)
	if err != nil {
		errorStr := fmt.Errorf("setQueue error: %v", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), errStatus)
		return
	}

	m.mu_matches.Lock()
	m.matches = data
	m.mu_matches.Unlock()
}

func (m *Matchmaker) broadcast(endpoint string, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	m.otherMatchmakersMu.Lock()
	for _, server := range m.otherMatchmakers {
		res, err := http.Post(server.URL+endpoint, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to send a broadcast message to Matchmaker %d, assuming it is down", server.ID)
			// TODO: Inform name server that this matchmaker is down
			return
		}
		defer res.Body.Close()
	}
	m.otherMatchmakersMu.Unlock()
}

func (m *Matchmaker) broadcastLeaderboardChanged() {
	m.mu_leaderboard.Lock()
	m.broadcast("/internal/leaderboard", m.leaderboard)
	m.mu_leaderboard.Unlock()
}

func (m *Matchmaker) broadcastQueueChanged() {
	m.mu_queue.Lock()
	m.broadcast("/internal/queue", m.queue)
	m.mu_queue.Unlock()
}

func (m *Matchmaker) broadcastMatchesChanged() {
	m.mu_matches.Lock()
	m.broadcast("/internal/matches", m.matches)
	m.mu_matches.Unlock()
}
