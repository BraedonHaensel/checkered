package checkered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const LEADER_ELECTION_TIMEOUT_SEC = 5 * time.Second

// The information known about the other Matchmakers in the network
type OtherMatchmaker struct {
	ID  int
	URL string
}

type ElectionMessage struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}
type LeaderMessage = ElectionMessage

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
	otherMatchmakers []OtherMatchmaker
	// Whether this server is running in the current leader election
	runningInElection bool
	// ID of the current leader server
	leaderID int
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
		otherMatchmakers:  []OtherMatchmaker{},
	}
}

// Initiates a leader election using the Bully algorithm.
func (m *Matchmaker) InitiateElection() {
	log.Println("Initiating a leader election")
	m.runningInElection = true

	// Check which Matchmakers have a higher ID than this one
	higherIDMatchmakers := []OtherMatchmaker{}
	for _, otherMatchmaker := range m.otherMatchmakers {
		if otherMatchmaker.ID > m.ID {
			higherIDMatchmakers = append(higherIDMatchmakers, otherMatchmaker)
		}
	}

	if len(higherIDMatchmakers) == 0 {
		// This server has the highest ID, declare itself leader
		log.Println("Declaring itself leader as it has the highest ID:", m.ID)
		m.leaderID = m.ID
		for _, otherMatchmaker := range m.otherMatchmakers {
			m.sendLeaderMessage(otherMatchmaker.URL)
		}
		return
	}

	// Send election(i) to those with a higher ID
	log.Printf("Sending election(%d) messages to servers with higher IDs\n", m.ID)
	for _, otherMatchmaker := range higherIDMatchmakers {
		m.sendElectionMessage(otherMatchmaker.URL)
	}

	// Wait for a bully() response
	log.Printf("Waiting up to %dms for a bully() response\n", LEADER_ELECTION_TIMEOUT_SEC)
	m.bullyTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
	// The chan is used to interrupt waiting for the timer when a bully() is received
	m.bullyTimerChan = make(chan struct{})
	select {
	case <-m.bullyTimer.C:
		// Timer fired, so no bully() responses received in time. Declare itself leader
		log.Println("No bully() responses received. Declaring itself leader with ID:", m.ID)
		m.leaderID = m.ID
		for _, otherMatchmaker := range m.otherMatchmakers {
			m.sendLeaderMessage(otherMatchmaker.URL)
		}
		return

	case <-m.bullyTimerChan:
		// Bullied before the timer fired. Wait for a leader(i) message
		log.Println("Received a bully() response. Waiting for a leader(i) message")
		m.leaderTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
		m.leaderTimerChan = make(chan struct{})
		select {
		case <-m.leaderTimer.C:
			// Timer fired, so no leader(i) received. Something went wrong, so initiate
			// a new election
			m.InitiateElection()
		case <-m.leaderTimerChan:
			// Received a leader(i) message. The message is handled by the leader(i)
			// message handler, so return
			return
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

// Register with the Name Server
func (m *Matchmaker) Register(url string) {
	m.ID = SendRegistrationRequest(url, m.nameServerURL+"/register/matchmaker")
	log.Println("Registered with ID:", m.ID)
}

// Sends a leader(i) message to another Matchmaker.
func (m *Matchmaker) sendLeaderMessage(otherMatchmakerURL string) {
	// Create the leader(i) message data
	data := OtherMatchmaker{
		ID:  m.ID,
		URL: m.URL,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	// Send a leader(i) message
	res, err := http.Post(otherMatchmakerURL+"/leader-election/leader", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to send a leader(%d) message to %s. This is expected if the Matchmaker "+
			"is down. Error: %v", m.ID, otherMatchmakerURL, err)
		return
	}
	defer res.Body.Close()
}

// Sends an election(i) message to another Matchmaker.
func (m *Matchmaker) sendElectionMessage(otherMatchmakerURL string) {
	// Create the election(i) message data
	data := OtherMatchmaker{
		ID:  m.ID,
		URL: m.URL,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	// Send an election(i) message
	res, err := http.Post(otherMatchmakerURL+"/leader-election/election", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to send an election(%d) message to %s. This is expected if the Matchmaker "+
			"is down. Error: %v", m.ID, otherMatchmakerURL, err)
		return
	}
	defer res.Body.Close()
}

// Sends a bully() message to another Matchmaker.
func (m *Matchmaker) sendBullyMessage(otherMatchmakerURL string) {
	// Send a bully() message
	res, err := http.Post(otherMatchmakerURL+"/leader-election/bully", "application/json", nil)
	if err != nil {
		log.Printf("failed to send a bully() message to %s. This is expected if the Matchmaker "+
			"is down. Error: %v", otherMatchmakerURL, err)
		return
	}
	defer res.Body.Close()
}

// Handle an election(i) Bully leader election message.
func (m *Matchmaker) HandleElectionRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the elected server's ID from the request
	data, err, errStatus := parseJsonRequestData[LeaderMessage](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	id := data.ID

	if id < m.ID {
		// Message received from a server with a lower ID, so bully them. Find
		// their URL to send the bully message to
		url := ""
		for _, matchmaker := range m.otherMatchmakers {
			if matchmaker.ID == id {
				url = matchmaker.URL
				break
			}
		}
		if url == "" {
			// Message received from an unknown leader ID, ignore

			// TODO Should URL be included in case it's a new server that joined?
			// Or do a refresh instead in case there's other out of date info?
			//
			//
			//
			return
		}
		log.Printf("Received election(%d). Bullying as its ID is higher: %d\n", id, m.ID)
		m.sendBullyMessage(url)
		if !m.runningInElection {
			// This server has a higher ID and isn't running yet, so start an election
			m.InitiateElection()
		}
	}
}

// Handle a bully() Bully leader election message.
func (m *Matchmaker) HandleBullyRequest(w http.ResponseWriter, r *http.Request) {
	// Interrupt the bully timer so it never fires
	if m.bullyTimer != nil && m.bullyTimer.Stop() {
		if m.bullyTimerChan != nil {
			// Close the bully timer chan to notify the thread to stop waiting for the timer
			close(m.bullyTimerChan)
		}
	}
}

// Handle a leader(i) Bully leader election message.
func (m *Matchmaker) HandleLeaderRequest(w http.ResponseWriter, r *http.Request) {
	// If this server was waiting, interrupt the leader timer so it never fires
	if m.leaderTimer != nil && m.leaderTimer.Stop() {
		if m.leaderTimerChan != nil {
			// Close the leader timer chan to notify the thread to stop waiting for the timer
			close(m.leaderTimerChan)
		}
	}

	// Parse the leader ID from the request
	data, err, errStatus := parseJsonRequestData[LeaderMessage](r)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	id := data.ID

	// TODO what if you don't know the leader? You won't have a URL entry for them in your database
	// Could send the url in the leader(i) message, or could refresh and then check if you
	// have it now
	//
	//
	//
	//

	// Set the new leader
	log.Println("Received new leader ID:", id)
	m.leaderID = id
	m.runningInElection = false
}
