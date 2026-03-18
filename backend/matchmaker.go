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
)

const LEADER_ELECTION_TIMEOUT_SEC = 5 * time.Second

// Responsible for handling the queue for new players finding a game, as well as
// maintaining the leaderboard and handling all leaderboard requests
type Matchmaker struct {
	ID int
	// Fully qualified URL of this Matchmaker
	URL string

	mu_leaderboard sync.Mutex
	leaderboard    Leaderboard
	queue          Queue[*Client]

	// URL of the Name Server
	nameServerURL string

	// Other Matchmakers in the network
	otherMatchmakers   []Server
	otherMatchmakersMu sync.Mutex
	// Whether this server is running in the current leader election
	runningInElection   bool
	runningInElectionMu sync.Mutex
	// ID of the current leader server
	leaderID   int
	leaderIDMu sync.Mutex
	// Timer and chan to wait for and detect receiving a bully() response
	bullyTimer     *time.Timer
	bullyTimerChan chan struct{}
	// Timer and chan to wait for and detect receiving a leader(i) response
	leaderTimer     *time.Timer
	leaderTimerChan chan struct{}
}

func NewMatchmaker(url, nameServerURL string) Matchmaker {
	queue := Queue[*Client]{}
	InitQueue(&queue, 100)
	return Matchmaker{
		URL:               url,
		mu_leaderboard:    sync.Mutex{},
		leaderboard:       Leaderboard{},
		queue:             queue,
		nameServerURL:     nameServerURL,
		runningInElection: false,
		otherMatchmakers:  []Server{},
	}
}

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
	return m.ID == m.leaderID;
}

func (m *Matchmaker) GetLeader() *Server {
	for _, other := range m.otherMatchmakers {
		if other.ID == m.leaderID {
			return &other;
		}
	}

	panic("No leader found!");
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
		m.leaderID = m.ID
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
		m.leaderID = m.ID
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
	res, err := http.Post(otherMatchmaker.URL+"/leader-election/election", "application/json", bytes.NewBuffer(jsonData))
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
	res, err := http.Post(otherMatchmaker.URL+"/leader-election/leader", "application/json", bytes.NewBuffer(jsonData))
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
	res, err := http.Post(otherMatchmaker.URL+"/leader-election/bully", "application/json", nil)
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
	id := otherMatchmaker.ID

	// Check if the message is from an unknown Matchmaker
	m.otherMatchmakersMu.Lock()
	if !slices.Contains(m.otherMatchmakers, otherMatchmaker) {
		m.otherMatchmakers = append(m.otherMatchmakers, otherMatchmaker)
	}
	m.otherMatchmakersMu.Unlock()

	// Set the new leader
	m.leaderIDMu.Lock()
	log.Println("Received a new leader ID:", id)
	m.leaderID = id
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

func (m *Matchmaker) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(m.leaderboard)
	if err != nil {
		errorStr := fmt.Errorf("getLeaderboard error: %v", err)
		fmt.Println(errorStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for a request to join the queue
func (m *Matchmaker) AddToQueue(w http.ResponseWriter, r *http.Request) {
}

// Handler for a request for a users status on the queue will respond with still
// in the queue or will respond with a new found match
func (m *Matchmaker) QueuePollRequest(w http.ResponseWriter, r *http.Request) {
}

// Handler for a request to leave the queue
func (m *Matchmaker) LeaveQueueRequest(w http.ResponseWriter, r *http.Request) {
}
