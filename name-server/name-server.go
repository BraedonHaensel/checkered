package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"strings"
)

// Name Server address flag
var addr = flag.String("addr", ":9000", "http service address")

// Lists of active server URLs
var matchmakerUrls = []string{}
var gameServerUrls = []string{}

// Request format for Matchmaker or Game Server registration
type registerServerRequest struct {
	Url string `json:"url"`
}

// Gets the fully qualified URL for the Name Server address.
func getFullUrl(addr string) string {
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
	if err != nil {
		return data, errors.New("failed to read the request body"), http.StatusInternalServerError
	}
	defer req.Body.Close()

	// Parse the JSON request data
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&data)
	if err != nil {
		return data, errors.New("invalid request data"), http.StatusBadRequest
	}
	return data, nil, 0
}

// GET /matchmakers - Gets the list of known Matchmaker servers.
func handleGetMatchmakers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(matchmakerUrls)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /matchmakers -", string(jsonData))
	w.Write(jsonData)
}

// GET /game-servers - Gets the list of known Game Servers.
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(gameServerUrls)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /game-servers -", string(jsonData))
	w.Write(jsonData)
}

// POST /register/matchmaker - Handle registering a Matchmaker server.
func handleRegisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.Url
	log.Println("POST /register/matchmaker -", url)

	// Add the server URL to the list of active servers
	if !slices.Contains(matchmakerUrls, url) {
		matchmakerUrls = append(matchmakerUrls, url)
	}

	// Send a response to the client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker registered successfully."))
}

// POST /register/game-server Handle registering a Game Server.
func handleRegisterGameServer(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.Url
	log.Println("POST /register/game-server -", url)

	// If new, add the server URL to the list of active servers
	if !slices.Contains(gameServerUrls, url) {
		gameServerUrls = append(gameServerUrls, url)
	}

	// Send a success response to the client (regardless of whether the server
	// was already registered with the Name Server)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Game Server registered successfully."))
}

// POST /deregister/matchmaker - Handle deregistering a Matchmaker server.
func handleDeregisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to deregister
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.Url
	log.Println("POST /deregister/matchmaker -", url)

	// Remove the URL from the list of active servers
	for i, addr := range matchmakerUrls {
		if addr == url {
			matchmakerUrls = append(matchmakerUrls[:i],
				matchmakerUrls[i+1:]...)
			break
		}
	}

	// Send a success response to the client (regardless of whether the server
	// was already registered with the Name Server)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker deregistered successfully."))
}

// POST /deregister/game-server Handle deregistering a Game Server.
func handleDeregisterGameServer(w http.ResponseWriter, req *http.Request) {
	// Parse the server URL to deregister
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	url := data.Url
	log.Println("POST /deregister/game-server -", url)

	// Remove the URL from the list of active servers
	for i, addr := range gameServerUrls {
		if addr == url {
			gameServerUrls = append(gameServerUrls[:i],
				gameServerUrls[i+1:]...)
			break
		}
	}

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

	url := getFullUrl(*addr)
	log.Println("Game Server running on", url)
	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
