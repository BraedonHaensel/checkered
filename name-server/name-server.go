package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Name Server address flag
var addr = flag.String("addr", ":9000", "http service address")

// Incrementing IDs to assign to the next registered server
var (
	nextMatchmakerID = 0
	nextGameServerID = 0
	matchmakerMu     sync.Mutex
	gameServerMu     sync.Mutex
)

// Each registered server has an ID and a URL
type Server struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// Lists of active server URLs
var (
	matchmakers = []Server{}
	gameServers = []Server{}
)

// Request format to register a Matchmaker or Game Server
type registerServerRequest struct {
	URL string `json:"url"`
}

// Request format to deregister a Matchmaker or Game Server
type deregisterServerRequest struct {
	ID int `json:"id"`
}

// Gets and increments the next Matchmaker ID.
func getNextMatchmakerID() int {
	matchmakerMu.Lock()
	defer matchmakerMu.Unlock()
	id := nextMatchmakerID
	nextMatchmakerID++
	return id
}

// Gets and increments the next Game Server ID.
func getNextGameServerID() int {
	gameServerMu.Lock()
	defer gameServerMu.Unlock()
	id := nextGameServerID
	nextGameServerID++
	return id
}

// Gets the fully qualified URL for the Name Server address.
func getFullURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		// Port provided, so determine the local IP address
		// Source: https://gosamples.dev/local-ip-address/
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		localAddress := conn.LocalAddr().(*net.UDPAddr)
		return "http://" + localAddress.IP.String() + addr
	} else {
		// Add "http://" to the address
		return "http://" + addr
	}
}

// Parses JSON data from a request. Returns an error and HTTP status code if the
// parsing fails.
func parseJsonRequestData[T any](req *http.Request) (T, error, int) {
	var data T
	// Parse the request body
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return data, fmt.Errorf("failed to read the request body: %w", err), http.StatusInternalServerError
	}

	// Parse the JSON request data
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&data)
	if err != nil {
		return data, fmt.Errorf("invalid request data: %w", err), http.StatusBadRequest
	}
	return data, nil, 0
}

// GET /matchmakers - Gets the list of known Matchmaker servers.
func handleGetMatchmakers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(matchmakers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /matchmakers -", string(jsonData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// GET /game-servers - Gets the list of known Game Servers.
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(gameServers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /game-servers -", string(jsonData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// POST /register/matchmaker - Handle registering a Matchmaker.
func handleRegisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.URL
	log.Println("POST /register/matchmaker -", url)

	// Get the ID to assign to the server
	id := getNextMatchmakerID()
	matchmakerMu.Lock()
	// Remove any old server entries from the same URL
	for i, matchmaker := range matchmakers {
		if matchmaker.URL == url {
			matchmakers = append(matchmakers[:i],
				matchmakers[i+1:]...)
			break
		}
	}
	// Add the new server entry
	matchmakers = append(matchmakers, Server{id, url})
	matchmakerMu.Unlock()

	// Send a success response to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": %d}`, id)
}

// POST /register/game-server - Handle registering a Game Server.
func handleRegisterGameServer(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.URL
	log.Println("POST /register/game-server -", url)

	// Get the ID to assign to the server
	id := getNextGameServerID()
	gameServerMu.Lock()
	// Remove any old server entries from the same URL
	for i, gameServer := range gameServers {
		if gameServer.URL == url {
			gameServers = append(gameServers[:i],
				gameServers[i+1:]...)
			break
		}
	}
	// Add the new server entry
	gameServers = append(gameServers, Server{id, url})
	gameServerMu.Unlock()

	// Send a success response to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id": %d}`, id)
}

// POST /deregister/matchmaker - Handle deregistering a Matchmaker.
func handleDeregisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	// Parse the server ID to deregister
	data, err, errStatus := parseJsonRequestData[deregisterServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	id := data.ID
	log.Println("POST /deregister/matchmaker -", id)

	// Remove the URL from the list of active servers
	matchmakerMu.Lock()
	for i, server := range matchmakers {
		if server.ID == id {
			matchmakers = append(matchmakers[:i],
				matchmakers[i+1:]...)
			break
		}
	}
	matchmakerMu.Unlock()

	// Send a success response to the client (regardless of whether the server
	// was already registered with the Name Server)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker deregistered successfully."))
}

// POST /deregister/game-server Handle deregistering a Game Server.
func handleDeregisterGameServer(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to deregister
	data, err, errStatus := parseJsonRequestData[deregisterServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	id := data.ID
	log.Println("POST /deregister/game-server -", id)

	// Remove the URL from the list of active servers
	gameServerMu.Lock()
	for i, server := range gameServers {
		if server.ID == id {
			gameServers = append(gameServers[:i],
				gameServers[i+1:]...)
			break
		}
	}
	gameServerMu.Unlock()

	// Send a success response to the client (regardless of whether the server
	// was already registered with the Name Server)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Game Server deregistered successfully."))
}

// Run the Name Server.
func main() {
	flag.Parse()

	// Request handlers
	http.HandleFunc("/matchmakers", handleGetMatchmakers)
	http.HandleFunc("/game-servers", handleGetGameServers)
	http.HandleFunc("POST /register/matchmaker", handleRegisterMatchmaker)
	http.HandleFunc("POST /register/game-server", handleRegisterGameServer)
	http.HandleFunc("POST /deregister/matchmaker", handleDeregisterMatchmaker)
	http.HandleFunc("POST /deregister/game-server", handleDeregisterGameServer)

	url := getFullURL(*addr)
	log.Println("Name Server running on", url)
	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
