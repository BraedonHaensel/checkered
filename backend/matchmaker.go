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

const LEADER_ELECTION_TIMEOUT_SEC = 10 * time.Second
const LEADER_ELECTION_BULLY_TIMEOUT_SEC = 3 * time.Second

const MATCHMAKER_HEARTBEAT_INTERVAL = 3 * time.Second
const MATCHMAKER_HEARTBEAT_CONSECUTIVE_MISS_LIMIT = 3

const QUEUE_HEARTBEAT_TIMEOUT = 5 * time.Second

// Responsible for handling the queue for new players finding a game, as well as
// maintaining the leaderboard and handling all leaderboard requests
type Matchmaker struct {
	ID int

	// Current version for tracking the latest states among Matchmakers
	syncVersion   int
	syncVersionMu sync.Mutex

	// games that new clients are in
	matches           map[uuid.UUID]*Match
	matchesMu         sync.Mutex
	queue             Queue[string]
	queueMu           sync.Mutex
	queueHeartbeats   map[string]chan struct{}
	queueHeartbeatsMu sync.Mutex
	leaderboard       *Leaderboard
	leaderboardMu     sync.Mutex

	// Fully qualified URL of this Matchmaker
	URL string

	// URL of the Name Server
	nameServerURL string

	// Other Matchmakers in the network (does not include itself)
	otherMatchmakers   []Server
	otherMatchmakersMu sync.Mutex

	// Matchmaker heartbeat ticker
	matchmakerHeartbeatTicker *time.Ticker

	// Track if a recent Matchmaker heartbeat has been received
	matchmakerHeartbeatReceived   bool
	matchmakerHeartbeatReceivedMu sync.Mutex

	// Track the number of consecutively missed Matchmaker heartbeats
	matchmakerHeartbeatMisses   int
	matchmakerHeartbeatMissesMu sync.Mutex

	// Whether this server is running in the current leader election
	runningInElection   bool
	runningInElectionMu sync.Mutex
	// ID of the current leader server
	Leader   Server
	leaderMu sync.Mutex
	// Timer and chan to wait for and detect receiving a bully() response
	bullyTimer     *time.Timer
	bullyTimerChan chan struct{}
	// Timer and chan to wait for and detect receiving a leader response
	leaderTimer     *time.Timer
	leaderTimerChan chan struct{}
	// Timer and chan to wait for and detect receiving a leader response during a data sync
	syncLeaderTimer     *time.Timer
	syncLeaderTimerChan chan struct{}

	// Locks to pause client requests during synchronizations during a leader election
	AcceptingClientRequestsMu sync.RWMutex // multiple request handlers acquire read locks
	isClientRequestsLocked    bool         // check if write lock is already acquired
	isClientRequestsLockedMu  sync.Mutex   // lock to safely check the above bool
}

// All data that gets synchronized between Matchmakers
type AllSyncedData struct {
	SyncVersion int                  `json:"sync_version"`
	Leaderboard Leaderboard          `json:"leaderboard"`
	Queue       Queue[string]        `json:"queue"`
	Matches     map[uuid.UUID]*Match `json:"matches"`
}

// Leader message, includes the latest Matchmaker data to synchronize with
type LeaderMessage struct {
	Leader     Server        `json:"leader"`
	LatestData AllSyncedData `json:"latest_data"`
}

// Matchmaker heartbeat message, sent from the leader to non-leader Matchmakers
type Heartbeat struct {
	LeaderID int `json:"leader_id"`
}

// Matchmaker constructor
func NewMatchmaker(url, nameServerURL string) *Matchmaker {
	queue := Queue[string]{}
	InitQueue(&queue, 100)
	matchmaker := Matchmaker{
		matches:         make(map[uuid.UUID]*Match),
		queue:           queue,
		queueHeartbeats: make(map[string]chan struct{}),
		leaderboard: &Leaderboard{
			Board: make([]LeaderboardEntry, 0),
		},
		leaderboardMu: sync.Mutex{},

		URL:           url,
		nameServerURL: nameServerURL,

		matchmakerHeartbeatTicker:   time.NewTicker(MATCHMAKER_HEARTBEAT_INTERVAL),
		matchmakerHeartbeatReceived: false,
		matchmakerHeartbeatMisses:   0,

		runningInElection:      false,
		otherMatchmakers:       []Server{},
		syncVersion:            0,
		isClientRequestsLocked: false,
	}

	// Start the handler for Matchmaker heartbeat ticks
	go matchmaker.startMatchmakerHeartbeatTickHandler()

	return &matchmaker
}

// Sends a request to another Matchmaker for the latest data in order to synchronize
// this Matchmaker's state before announcing itself as the new leader.
// Returns (*data, isSuccess, isAlive)
func (m *Matchmaker) SendNewLeaderDataSyncRequest(otherMatchmaker Server) (*AllSyncedData, bool, bool) {
	// Send new leader data sync request
	res, err := http.Post(otherMatchmaker.URL+"/internal/new-leader-data-sync", "application/json", nil)
	if err != nil {
		log.Printf("Failed to send a new leader data sync request to Matchmaker %d, assuming it is down", otherMatchmaker.ID)
		return nil, false, false // failed, server is down
	}
	defer res.Body.Close()

	data, err := ParseJsonResponseData[AllSyncedData](res)
	if err != nil {
		log.Println("Failed to decode response for /internal/new-leader-data-sync", err)
		return nil, false, true // failed, server is alive
	}
	return &data, true, true // success
}

// Gets all data synchronized between Matchmakers.
func (m *Matchmaker) GetAllSyncedData() AllSyncedData {
	// Bundle all of this Matchmaker's data that gets synchronized
	m.syncVersionMu.Lock()
	m.leaderboardMu.Lock()
	m.queueMu.Lock()
	m.matchesMu.Lock()

	data := AllSyncedData{
		SyncVersion: m.syncVersion,
		Leaderboard: *m.leaderboard,
		Queue:       m.queue,
		Matches:     m.matches,
	}

	m.matchesMu.Unlock()
	m.queueMu.Unlock()
	m.leaderboardMu.Unlock()
	m.syncVersionMu.Unlock()

	return data
}

// Sets all data synchronized between Matchmakers.
func (m *Matchmaker) SetAllSyncedData(data AllSyncedData) {
	// Unpack and update state from the data
	m.syncVersionMu.Lock()
	m.leaderboardMu.Lock()
	m.queueMu.Lock()
	m.matchesMu.Lock()

	m.syncVersion = data.SyncVersion
	m.leaderboard = &data.Leaderboard
	m.queue = data.Queue
	m.matches = data.Matches

	log.Println("Matchmaker data updated to sync version:", m.syncVersion)

	m.matchesMu.Unlock()
	m.queueMu.Unlock()
	m.leaderboardMu.Unlock()
	m.syncVersionMu.Unlock()
}

// Handles a request from another Matchmaker for the latest data in order to synchronize
// the requester's state before the requester announcing itself as the new leader.
func (m *Matchmaker) HandleNewLeaderDataSyncRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a data sync request from a new leader")
	// Pause client requests until the synchronization is complete and the leader(i, latestData)
	// message has been received
	m.PauseClientRequests()

	// Get this Matchmaker's data
	data := m.GetAllSyncedData()

	// Start a timer to wait for a leader message
	m.syncLeaderTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
	m.syncLeaderTimerChan = make(chan struct{})

	// Send this Matchmaker's data
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		panic("Failed to encode json data for package request")
	}

	go func() {
		// Wait for a leader message after the data sync
		log.Printf("Waiting up to %dms for a leader message\n", LEADER_ELECTION_TIMEOUT_SEC.Milliseconds())
		select {
		case <-m.syncLeaderTimer.C:
			// Timer fired, so no leader received. Something went wrong, so initiate
			// a new election
			log.Println("No leader responses received")
			m.InitiateElection()
		case <-m.syncLeaderTimerChan:
			// Received a leader message. The message is handled by the leader
			// message handler, so return
			return
		}
	}()
}

// Starts the Matchmaker heartbeat tick handler.
func (m *Matchmaker) startMatchmakerHeartbeatTickHandler() {
	defer m.matchmakerHeartbeatTicker.Stop()

	// Handle each Matchmaker heartbeat tick
	for range m.matchmakerHeartbeatTicker.C {
		m.leaderMu.Lock()
		currentLeader := m.Leader
		m.leaderMu.Unlock()

		if currentLeader.URL == "" {
			// No leader chosen yet, skip heartbeat
			continue
		}

		if currentLeader.ID == m.ID {
			// This is the leader. Send heartbeats to all other Matchmakers
			m.sendHeartbeatsFromLeaderMatchmaker()
			continue
		}

		// This is not the leader. Check if a heartbeat was received
		m.matchmakerHeartbeatReceivedMu.Lock()
		m.matchmakerHeartbeatMissesMu.Lock()
		if m.matchmakerHeartbeatReceived {
			// Heartbeat received. Reset for the next interval
			m.matchmakerHeartbeatReceived = false
			m.matchmakerHeartbeatMisses = 0
		} else {
			// Heartbeat missed, increment the consecutive misses counter
			m.matchmakerHeartbeatMisses++
			log.Printf("Matchmaker leader heartbeat missed (x%d)", m.matchmakerHeartbeatMisses)
			if m.matchmakerHeartbeatMisses >= MATCHMAKER_HEARTBEAT_CONSECUTIVE_MISS_LIMIT {
				// Missed too many heartbeats in a row, assume the leader is down
				log.Printf("Failed to receive %d consecutive Matchmaker leader heartbeats, assuming it is down",
					MATCHMAKER_HEARTBEAT_CONSECUTIVE_MISS_LIMIT)
				m.DeregisterOtherMatchmaker(currentLeader.ID)
				m.InitiateElection()
				m.matchmakerHeartbeatMisses = 0
			}
		}
		m.matchmakerHeartbeatMissesMu.Unlock()
		m.matchmakerHeartbeatReceivedMu.Unlock()
	}
}

// Sends a heartbeat from this leader Matchmaker to all others.
func (m *Matchmaker) sendHeartbeatsFromLeaderMatchmaker() {
	// Create the heartbeat message
	data := Heartbeat{LeaderID: m.ID}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to marshal heartbeat message: %v", err)
	}

	// Track any down Matchmakers that fail to respond
	downMatchmakers := []Server{}

	// Send the heartbeat message to all other Matchmakers
	m.otherMatchmakersMu.Lock()
	log.Println("Sending Matchmaker heartbeats")
	for _, otherMatchmaker := range m.otherMatchmakers {
		res, err := http.Post(otherMatchmaker.URL+"/internal/heartbeat", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to send heartbeat to Matchmaker %d, assuming it is down", otherMatchmaker.ID)
			downMatchmakers = append(downMatchmakers, otherMatchmaker)
			continue
		}
		res.Body.Close()
	}
	m.otherMatchmakersMu.Unlock()

	// Deregister any Matchmakers that failed to receive the heartbeat
	for _, otherMatchmaker := range downMatchmakers {
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
	}
}

// Endpoint to receive a Matchmaker heartbeat from the current leader
func (m *Matchmaker) HandleMatchmakerHeartbeat(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[Heartbeat](r)
	if err != nil {
		errMsg := fmt.Errorf("HandleMatchmakerHeartbeat error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}
	leaderID := data.LeaderID

	// Ignore heartbeats from an unexpected leader ID
	if leaderID != m.Leader.ID {
		return
	}

	// Mark that a heartbeat has been received during the current interval
	m.matchmakerHeartbeatReceivedMu.Lock()
	m.matchmakerHeartbeatReceived = true
	m.matchmakerHeartbeatReceivedMu.Unlock()
}

// -------------------- DISTRIBUTED SYSTEMS / LEADER ELECTION USING BULLY ALGORITHM --------------------

// Pauses client requests (unless already paused)
func (m *Matchmaker) PauseClientRequests() {
	// Check if client requests are already paused
	m.isClientRequestsLockedMu.Lock()
	if !m.isClientRequestsLocked {
		// Acquire the write lock to pause client requests
		m.isClientRequestsLocked = true
		m.AcceptingClientRequestsMu.Lock()
		log.Println("Paused client requests!")
	}
	m.isClientRequestsLockedMu.Unlock()
}

// Resumes client requests (unless already resumed)
func (m *Matchmaker) ResumeClientRequests() {
	// Check if client requests are already resumed
	m.isClientRequestsLockedMu.Lock()
	if m.isClientRequestsLocked {
		// Release the write lock to resume client requests
		m.isClientRequestsLocked = false
		m.AcceptingClientRequestsMu.Unlock()
		log.Println("Resumed client requests!")
	}
	m.isClientRequestsLockedMu.Unlock()
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
	log.Println("Matchmakers refresh")
	m.otherMatchmakers = otherMatchmakers
	m.otherMatchmakersMu.Unlock()
}

// Deregisters another Matchmaker from the Name Server.
func (m *Matchmaker) DeregisterOtherMatchmaker(otherMatchmakerID int) {
	// Create the deregistration request
	log.Println("Deregistering Matchmaker", otherMatchmakerID)
	body := fmt.Appendf(nil, `{"id": %d}`, otherMatchmakerID)
	res, err := http.Post(m.nameServerURL+"/deregister/matchmaker", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to deregister Matchmaker %d: %v", otherMatchmakerID, err)
	} else {
		defer res.Body.Close()
	}

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

// Deregisters a Game Server from the Name Server.
func (m *Matchmaker) DeregisterGameServer(gameServerID int) {
	// Create the deregistration request
	log.Println("Deregistering Game Server", gameServerID)
	body := fmt.Appendf(nil, `{"id": %d}`, gameServerID)
	res, err := http.Post(m.nameServerURL+"/deregister/game-server", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to deregister Game Server %d: %v", gameServerID, err)
		return
	}
	defer res.Body.Close()
}

// Gets the latest data from all Matchmakers.
func (m *Matchmaker) GetLatestMatchmakerData() AllSyncedData {
	log.Println("Getting the latest data from all Matchmakers")
	// Start by assuming the latest data is your own
	latestData := m.GetAllSyncedData()

	// Get the latest data from all Matchmakers
	m.otherMatchmakersMu.Lock()
	downMatchmakers := []Server{}
	for _, otherMatchmaker := range m.otherMatchmakers {
		data, isSuccess, isAlive := m.SendNewLeaderDataSyncRequest(otherMatchmaker)
		if !isSuccess {
			if !isAlive {
				// Matchmaker is down
				downMatchmakers = append(downMatchmakers, otherMatchmaker)
			}
			continue
		}

		if data.SyncVersion > latestData.SyncVersion {
			// Received newer data
			latestData = *data
		}
	}
	m.otherMatchmakersMu.Unlock()

	// Deregister any Matchmakers that did not respond
	for _, otherMatchmaker := range downMatchmakers {
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
	}

	return latestData
}

// Synchronizes with the latest data from all Matchmakers, then sends leader messages with it.
func (m *Matchmaker) SynchronizeDataAndSendLeaderMessages() {
	// Pause client requests while synchronizing data for the leader(i, latestData) messages
	m.PauseClientRequests()

	// Get the latest data from all Matchmakers.
	latestData := m.GetLatestMatchmakerData()

	if latestData.SyncVersion > m.syncVersion {
		// Update to the latest data
		m.SetAllSyncedData(latestData)
	}

	m.runningInElectionMu.Lock()
	m.runningInElection = false
	m.runningInElectionMu.Unlock()

	// Send leader(i, latestData) messages with the latest synced Matchmaker data
	m.otherMatchmakersMu.Lock()
	log.Println("Sending leader messages")
	downMatchmakers := []Server{}
	for _, otherMatchmaker := range m.otherMatchmakers {
		if !m.sendLeaderMessage(otherMatchmaker, latestData) {
			downMatchmakers = append(downMatchmakers, otherMatchmaker)
		}
	}
	m.otherMatchmakersMu.Unlock()

	// Deregister any Matchmakers that did not respond
	for _, otherMatchmaker := range downMatchmakers {
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
	}

	// Starting all heartbeat watcher go routines to detect unresponsive clients in queue as only the leader needs to have this
	m.queue.forEach(func(username string, i int) {
		m.queueHeartbeatsMu.Lock()
		_, alreadyWatching := m.queueHeartbeats[username]
		m.queueHeartbeatsMu.Unlock()
		if !alreadyWatching {
			m.startHeartbeatWatcher(username)
		}
	})

	// Election complete, resume handling client requests
	m.ResumeClientRequests()
}

// Initiates a leader election using the Bully algorithm.
func (m *Matchmaker) InitiateElection() {
	// Pause client requests during the election
	m.PauseClientRequests()

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
		m.leaderMu.Lock()
		log.Printf("Declaring itself leader (has highest ID [%d])", m.ID)
		m.Leader = Server{
			ID:  m.ID,
			URL: m.URL,
		}
		m.leaderMu.Unlock()

		// Get the latest data from all Matchmakers, then send leader messages with it
		m.SynchronizeDataAndSendLeaderMessages()
		return
	}

	// Start the bully response timeout before sending the election(i) messages to avoid race conditions
	m.bullyTimer = time.NewTimer(LEADER_ELECTION_BULLY_TIMEOUT_SEC)
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
		LEADER_ELECTION_BULLY_TIMEOUT_SEC.Milliseconds())

	select {
	case <-m.bullyTimer.C:
		// Timer fired, so no bully() responses received in time. Declare itself leader
		m.leaderMu.Lock()
		log.Println("No bully() responses received. Declaring itself leader with ID:", m.ID)
		m.Leader = Server{
			ID:  m.ID,
			URL: m.URL,
		}
		m.leaderMu.Unlock()

		// Get the latest data from all Matchmakers, then send leader messages with it
		m.SynchronizeDataAndSendLeaderMessages()

	case <-m.bullyTimerChan:
		// Bullied before the timer fired. Wait for a leader message
		log.Printf("Received a bully() response. Waiting up to %dms for a leader message\n",
			LEADER_ELECTION_TIMEOUT_SEC.Milliseconds())
		m.leaderTimer = time.NewTimer(LEADER_ELECTION_TIMEOUT_SEC)
		m.leaderTimerChan = make(chan struct{})
		select {
		case <-m.leaderTimer.C:
			// Timer fired, so no leader received. Something went wrong, so initiate
			// a new election
			log.Println("No leader responses received")
			m.InitiateElection()
		case <-m.leaderTimerChan:
			// Received a leader message. The message is handled by the leader
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

// Sends a leader(i, latestData) message to another Matchmaker. Returns true if
// the recipient is alive.
func (m *Matchmaker) sendLeaderMessage(otherMatchmaker Server, latestData AllSyncedData) bool {
	// Create the leader(i, latestData) message
	data := LeaderMessage{
		Leader: Server{
			ID:  m.ID,
			URL: m.URL,
		},
		LatestData: latestData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	// Send the leader(i, latestData) message
	res, err := http.Post(otherMatchmaker.URL+"/internal/leader-election/leader", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send a leader message to Matchmaker %d, assuming it is down", otherMatchmaker.ID)
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
		errMsg := fmt.Errorf("HandleElectionRequest error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
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
		// Pause client requests during the election
		m.PauseClientRequests()

		// Message received from a server with a lower ID, so bully them.
		log.Printf("Received election(%d). Bullying as this Matchmaker's ID [%d] is higher", id, m.ID)
		m.sendBullyMessage(otherMatchmaker)
		m.runningInElectionMu.Lock()
		defer m.runningInElectionMu.Unlock()
		if !m.runningInElection {
			// This server has a higher ID and isn't running yet, so start an election
			go m.InitiateElection()
		}
	}
}

// Handle an incoming leader(i, latestData) Bully leader election message.
func (m *Matchmaker) HandleLeaderRequest(w http.ResponseWriter, r *http.Request) {
	// If this server was waiting, interrupt the leader timer so it never fires
	if m.leaderTimer != nil && m.leaderTimer.Stop() {
		if m.leaderTimerChan != nil {
			// Close the leader timer chan to notify the thread to stop waiting for the timer
			close(m.leaderTimerChan)
		}
	}
	// If this server was waiting during a data sync, interrupt the leader timer so it never fires
	if m.syncLeaderTimer != nil && m.syncLeaderTimer.Stop() {
		if m.syncLeaderTimerChan != nil {
			// Close the leader timer chan to notify the thread to stop waiting for the timer
			close(m.syncLeaderTimerChan)
		}
	}

	// Parse the Matchmaker's ID from the request
	data, err, errStatus := parseJsonRequestData[LeaderMessage](r)
	if err != nil {
		errMsg := fmt.Errorf("HandleLeaderRequest error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}
	otherMatchmaker := data.Leader

	// Check if the message is from an unknown Matchmaker
	m.otherMatchmakersMu.Lock()
	if !slices.Contains(m.otherMatchmakers, otherMatchmaker) {
		m.otherMatchmakers = append(m.otherMatchmakers, otherMatchmaker)
	}
	m.otherMatchmakersMu.Unlock()

	// Set the new leader
	m.leaderMu.Lock()
	log.Println("Received a new leader ID:", otherMatchmaker.ID)
	m.Leader = otherMatchmaker
	m.leaderMu.Unlock()

	// Reset the Matchmaker heartbeat ticker interval
	m.matchmakerHeartbeatTicker.Reset(MATCHMAKER_HEARTBEAT_INTERVAL)

	// No longer the leader, shut down all queue heartbeat goroutines
	if !m.IsLeader() {
		m.stopAllQueueHeartbeatWatchers()
	}

	m.runningInElectionMu.Lock()
	m.runningInElection = false
	m.runningInElectionMu.Unlock()

	if data.LatestData.SyncVersion > m.syncVersion {
		// Synchronize with the new leader's data
		m.SetAllSyncedData(data.LatestData)
	} else {
		log.Println("Data already synchronized with sync version:", m.syncVersion)
	}

	// Resume Client requests
	m.ResumeClientRequests()
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
		log.Println(errorStr)
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
	Type       string     `json:"type"`
	GameServer *Server    `json:"game_server,omitempty"`
	MatchID    *uuid.UUID `json:"match_id,omitempty"`
}

type RequestNewGameServerRequest struct {
	OldGameServer Server    `json:"old_game_server"`
	MatchID       uuid.UUID `json:"match_id"`
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

type Match struct {
	MatchID     uuid.UUID  `json:"match_id"`
	RedPlayer   string     `json:"red"`
	BlackPlayer string     `json:"black"`
	GameServer  Server     `json:"server"`
	Mu          sync.Mutex `json:"-"` // Omit from JSON
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
		errMsg := fmt.Errorf("AddToQueue error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	// Extracting the username
	username := data.Username

	// Check if user is already in queue
	alreadyQueued := false
	m.queueMu.Lock()
	m.queue.forEach(func(u string, i int) {
		if u == username {
			alreadyQueued = true
		}
	})

	if alreadyQueued {
		m.queueMu.Unlock()
		log.Printf("Player \"%s\" already in queue", username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QueueResponse{Type: "ALREADY_IN_QUEUE"})
		return
	}

	// Enqueue the user
	log.Printf("Adding \"%s\" to queue", username)
	m.queue.enqueue(username)
	log.Printf("Added \"%s\" to queue", username)

	// Start heartbeat watcher for this player
	m.startHeartbeatWatcher(username)

	// Matchmaking logic
	if m.queue.Size >= 2 {
		redPlayer := m.queue.dequeue()
		blackPlayer := m.queue.dequeue()
		m.queueMu.Unlock()

		// Stop heartbeat watchers for matched players
		m.stopQueueHeartbeatWatcher(redPlayer)
		m.stopQueueHeartbeatWatcher(blackPlayer)

		match, err := m.NewMatch(redPlayer, blackPlayer)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to create match: %v", err)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		// Marshal the match to JSON
		matchBytes, err := json.Marshal(&match)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to marshal match: %v", err)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		// Send a POST request to the game server with the match details
		log.Printf("Sending new game request to %s", match.GameServer.URL)
		res, err := http.Post(match.GameServer.URL+"/newGame", "application/json", bytes.NewBuffer(matchBytes))
		if err != nil {
			errMsg := fmt.Sprintf("Failed to send match to game server [%d] %s: %s", match.GameServer.ID, match.GameServer.URL, err)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		log.Printf("Game created on %s", match.GameServer.URL)
		defer res.Body.Close()

		m.matchesMu.Lock()
		m.matches[match.MatchID] = &match
		m.matchesMu.Unlock()
		log.Printf("Match created: %s (red) vs %s (black)", redPlayer, blackPlayer)
	} else {
		m.queueMu.Unlock()
	}

	m.IncrementSyncVersion()
	m.broadcastQueueUpdate()

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
	m.matchesMu.Lock()

	// Reset heartbeat timer if player is in queue
	m.queueHeartbeatsMu.Lock()
	if ch, exists := m.queueHeartbeats[username]; exists {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	m.queueHeartbeatsMu.Unlock()

	// Check if the user is in a match
	for matchID, match := range m.matches {
		if match.RedPlayer == username || match.BlackPlayer == username {
			m.matchesMu.Unlock()
			// User has been matched, return the game server URL
			log.Printf("User \"%s\" has been matched, sending game server [%d] %s", username, match.GameServer.ID, match.GameServer.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PollResponse{Type: "IN_GAME", GameServer: &match.GameServer, MatchID: &matchID})
			return
		}
	}
	m.matchesMu.Unlock()

	// Check if user is already in queue
	m.queueMu.Lock()
	alreadyQueued := false
	m.queue.forEach(func(u string, i int) {
		if u == username {
			alreadyQueued = true
		}
	})
	m.queueMu.Unlock()

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
		errMsg := fmt.Errorf("LeaveQueueRequest error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	username := data.Username

	// Check if user is actually in the queue
	m.queueMu.Lock()
	inQueue := false
	m.queue.forEach(func(u string, i int) {
		if u == username {
			inQueue = true
		}
	})
	m.queueMu.Unlock()

	if !inQueue {
		log.Printf("User \"%s\" tried to leave queue but was not in it", username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QueueResponse{Type: "ALREADY_NOT_IN_QUEUE"})
		return
	}

	// Stopping and removing heartbeat watcher as well
	m.stopQueueHeartbeatWatcher(username)

	m.queueMu.Lock()
	RemoveStringValue(&m.queue, username)
	m.queueMu.Unlock()
	log.Printf("User \"%s\" left the queue", username)

	m.IncrementSyncVersion()
	m.broadcastQueueUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "SUCCESS"})

}

// Creating player go routine, it waits for either a reset signal or a timeout
func (m *Matchmaker) startHeartbeatWatcher(username string) {

	// Creating a buffered channel that can hold only 1 singal
	resetCh := make(chan struct{}, 1)

	// Inserting the channel into the heartbeats map. Of course we lock the map first
	m.queueHeartbeatsMu.Lock()
	m.queueHeartbeats[username] = resetCh
	m.queueHeartbeatsMu.Unlock()

	go func() {
		timer := time.NewTimer(QUEUE_HEARTBEAT_TIMEOUT)
		defer timer.Stop()
		for {
			select {
			case _, open := <-resetCh:
				if !open {
					// Channel closed meaning player was matched or left queue cleanly
					return
				}
				// Reset signal received meaning the client polled so restart the timer
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(QUEUE_HEARTBEAT_TIMEOUT)

			case <-timer.C:
				// Check if watcher was already stopped before doing anything
				m.queueHeartbeatsMu.Lock()
				_, stillActive := m.queueHeartbeats[username]
				m.queueHeartbeatsMu.Unlock()

				if !stillActive {
					return
				}

				// Safe to remove
				// Timer fired meaning client hasn't polled in time and we should remove them
				log.Printf("Player \"%s\" heartbeat timed out, removing from queue", username)
				m.stopQueueHeartbeatWatcher(username)
				m.queueMu.Lock()
				RemoveStringValue(&m.queue, username)
				m.queueMu.Unlock()

				// Boroadcast the update
				m.IncrementSyncVersion()
				m.broadcastQueueUpdate()

				return
			}
		}
	}()
}

// This method is called when a player is matched, leaves or times out
// Used to get rid of the go routine that kept a timer on the player that was removed from the queue
func (m *Matchmaker) stopQueueHeartbeatWatcher(username string) {
	m.queueHeartbeatsMu.Lock()
	defer m.queueHeartbeatsMu.Unlock()
	if ch, exists := m.queueHeartbeats[username]; exists {
		close(ch)
		delete(m.queueHeartbeats, username)
	}
}

// Stops all hearbeat watchers, mainly just to ensure the previous leader after losing an election stops all go routines keeping tabs of clients in queue
func (m *Matchmaker) stopAllQueueHeartbeatWatchers() {
	m.queueHeartbeatsMu.Lock()
	defer m.queueHeartbeatsMu.Unlock()
	for username, ch := range m.queueHeartbeats {
		close(ch)
		delete(m.queueHeartbeats, username)
	}
	log.Println("Stopped all queue heartbeat watchers (lost election)")
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
	data, err, errStatus := parseJsonRequestData[RequestNewGameServerRequest](r)
	if err != nil {
		errMsg := fmt.Errorf("RequestNewGameServer error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}
	oldGameServer := data.OldGameServer

	m.matchesMu.Lock()
	match := m.matches[data.MatchID]
	m.matchesMu.Unlock()

	// Lock the match in case the opponent is making the same request
	match.Mu.Lock()
	defer match.Mu.Unlock()

	if match.GameServer.ID != oldGameServer.ID {
		// A new Game Server has already been chosen for the match
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "SUCCESS"})
		return
	}

	// Check the health of the Game Server
	if m.checkServerHealth(data.OldGameServer.URL, 5*time.Second) {
		// Game Server is healthy, keep it
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "HEALTHY"})
		return
	}

	// Deregister the Game Server
	m.DeregisterGameServer(data.OldGameServer.ID)

	log.Println("Finding a replacement Game Server for match", match.MatchID)

	// Locate a new gameserver
	var gameServer *Server = nil
	// Get all available game servers
	gameServers, err := SendServerListRequest(m.nameServerURL + "/game-servers")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(gameServers) == 0 {
		log.Println("No game servers available")
		http.Error(w, "No game servers available", http.StatusInternalServerError)
		return
	}
	// Marshal the match to JSON
	matchBytes, err := json.Marshal(TakeOverRequest{
		GameID: match.MatchID,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal match: %v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	foundReplacement := false
	unresponsiveServers := make([]Server, 0)
	for i := range gameServers {
		gameServer = &gameServers[i]

		// Send a POST request to the game server with the match details
		res, err := http.Post(gameServer.URL+"/takeover", "application/json", bytes.NewBuffer(matchBytes))
		if err != nil {
			errMsg := fmt.Sprintf(
				"Failed to send match to game server [%d] %s: %s",
				gameServer.ID,
				gameServer.URL,
				err,
			)
			log.Println(errMsg)
			unresponsiveServers = append(unresponsiveServers, *gameServer)
			continue
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			continue
		}
		// Switch the match to the new Game Server
		match.GameServer = *gameServer
		// Broadcast the match change to the backup Matchmakers
		m.broadcastMatchesUpdate()
		foundReplacement = true
		break
	}
	for i := range unresponsiveServers {
		m.DeregisterGameServer(unresponsiveServers[i].ID)
	}
	if !foundReplacement {
		log.Printf("Failed to find replacement server for game %s", match.MatchID)
		// TODO Mark the game as aborted
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "FAILURE"})
		return
	}

	m.IncrementSyncVersion()
	// Broadcast the match change to the backup Matchmakers
	m.broadcastMatchesUpdate()

	log.Printf("Match %s moved to Game Server [%d] %s", match.MatchID, gameServer.ID, gameServer.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RequestNewGameServerResponse{Type: "SUCCESS"})
}

// Handler for a request to update the leaderboard
func (m *Matchmaker) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[GameResult](r)
	if err != nil {
		errMsg := fmt.Errorf("UpdateLeaderboard error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	// Update the leaderboard

	// if isNotADraw {
	// 	m.IncrementSyncVersion()
	// 	// Broadcast the new leaderboard if the game did not end in a draw
	// 	m.broadcastLeaderboardUpdate()
	// }
	log.Printf("Updating leaderboard (game=%s, winner=%s, loser=%s)", data.GameID, data.Winner, data.Loser)
	m.leaderboardMu.Lock()
	m.leaderboard.UpdateLeaderboard(data)
	m.leaderboardMu.Unlock()

	m.IncrementSyncVersion()
	m.broadcastLeaderboardUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "leaderboard updated"})
}

// Handler for a request to end a matchup
func (m *Matchmaker) EndMatch(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[EndMatchRequest](r)
	if err != nil {
		errMsg := fmt.Errorf("EndMatch error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	matchID := data.MatchID

	// Check if the match exists
	m.matchesMu.Lock()
	_, exists := m.matches[matchID]
	if !exists {
		m.matchesMu.Unlock()
		log.Printf("EndMatch: match %s not found", matchID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(EndMatchResponse{Type: "match not found"})
		return
	}

	// Remove the match
	delete(m.matches, matchID)
	m.matchesMu.Unlock()
	log.Printf("Match %s ended and removed", matchID)

	m.IncrementSyncVersion()
	m.broadcastMatchesUpdate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QueueResponse{Type: "match Removed"})
}

func (m *Matchmaker) SetLeaderboard(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[Leaderboard](r)
	if err != nil {
		errMsg := fmt.Errorf("SetLeaderboard error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	m.leaderboardMu.Lock()
	log.Printf("Leaderboard update received: %v", data)
	m.leaderboard = &data
	m.IncrementSyncVersion()
	m.leaderboardMu.Unlock()
}

func (m *Matchmaker) SetQueue(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[Queue[string]](r)
	if err != nil {
		errMsg := fmt.Errorf("SetQueue error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	m.queueMu.Lock()
	log.Println("Queue update received")
	m.queue = data
	m.IncrementSyncVersion()
	m.queueMu.Unlock()
}

func (m *Matchmaker) SetMatches(w http.ResponseWriter, r *http.Request) {
	data, err, errStatus := parseJsonRequestData[map[uuid.UUID]*Match](r)
	if err != nil {
		errMsg := fmt.Errorf("SetMatches error: %w", err)
		log.Println(errMsg)
		http.Error(w, errMsg.Error(), errStatus)
		return
	}

	m.matchesMu.Lock()
	log.Println("Matches update received")
	m.matches = data
	m.IncrementSyncVersion()
	m.matchesMu.Unlock()
}

// Increments the sync version used to track how recent a data state is
func (m *Matchmaker) IncrementSyncVersion() {
	m.syncVersionMu.Lock()
	m.syncVersion++
	log.Println("Sync version incremented to:", m.syncVersion)
	m.syncVersionMu.Unlock()
}

func (m *Matchmaker) broadcast(endpoint string, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	// Track any down Matchmakers that fail to respond
	downMatchmakers := []Server{}

	// Send the broadcast message to all other Matchmakers
	m.otherMatchmakersMu.Lock()
	for _, otherMatchmaker := range m.otherMatchmakers {
		res, err := http.Post(otherMatchmaker.URL+endpoint, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to send a broadcast message to Matchmaker %d, assuming it is down", otherMatchmaker.ID)
			downMatchmakers = append(downMatchmakers, otherMatchmaker)
			return
		}
		defer res.Body.Close()
	}
	m.otherMatchmakersMu.Unlock()

	// Deregister any Matchmakers that failed to receive the broadcast message
	for _, otherMatchmaker := range downMatchmakers {
		m.DeregisterOtherMatchmaker(otherMatchmaker.ID)
	}
}

// Broadcasts a leaderboard update to all other Matchmakers
func (m *Matchmaker) broadcastLeaderboardUpdate() {
	m.leaderboardMu.Lock()
	log.Println("Broadcasting leaderboard update to backups")
	m.broadcast("/internal/leaderboard", m.leaderboard)
	m.leaderboardMu.Unlock()
}

// Broadcasts a queue update to all other Matchmakers
func (m *Matchmaker) broadcastQueueUpdate() {
	m.queueMu.Lock()
	log.Println("Broadcasting queue update to backups")
	m.broadcast("/internal/queue", m.queue)
	m.queueMu.Unlock()
}

// Broadcasts a matches update to all other Matchmakers
func (m *Matchmaker) broadcastMatchesUpdate() {
	m.matchesMu.Lock()
	log.Println("Broadcasting matches update to backups")
	m.broadcast("/internal/matches", m.matches)
	m.matchesMu.Unlock()
	log.Println("Broadcast successful")
}
