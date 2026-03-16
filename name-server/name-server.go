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

// Lists of active server addresses
var matchmakerAddresses = []string{}
var gameServerAddresses = []string{}

// Request format for Matchmaker or Game Server registration
type registerServerRequest struct {
	Address string `json:"address"`
}

// Gets the fully qualified URL for the Name Server address
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

// GET /matchmakers - Get the list of known Matchmaker servers
func handleGetMatchmakers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(matchmakerAddresses)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /matchmakers -", string(jsonData))
	w.Write(jsonData)
}

// GET /game-servers - Get the list of known Game Servers
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	jsonData, err := json.Marshal(gameServerAddresses)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /game-servers -", string(jsonData))
	w.Write(jsonData)
}

// POST /register/matchmaker - Handle registering a Matchmaker server.
// Name Server
func handleRegisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	// Parse the server address to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	address := data.Address
	log.Println("POST /register/matchmaker -", address)

	// Add the server address to the list of active servers
	if !slices.Contains(matchmakerAddresses, address) {
		matchmakerAddresses = append(matchmakerAddresses, address)
	}

	// Send a response to the client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker added successfully."))
}

// POST /register/game-server Handle registering a Game Server.
func handleRegisterGameServer(w http.ResponseWriter, req *http.Request) {
	// Parse the server address to register
	data, err, errStatus := parseJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	address := data.Address
	log.Println("POST /register/game-server -", address)

	// If new, add the server address to the list of active servers
	if !slices.Contains(gameServerAddresses, address) {
		gameServerAddresses = append(gameServerAddresses, address)
	}

	// Send a success response to the client (regardless of whether the server
	// was already registered with the Name Server)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Game Server added successfully."))
}

// Run the Name Server
func main() {
	flag.Parse()

	// Request handlers
	http.HandleFunc("/matchmakers", handleGetMatchmakers)
	http.HandleFunc("/game-servers", handleGetGameServers)
	http.HandleFunc("POST /register/matchmaker", handleRegisterMatchmaker)
	http.HandleFunc("POST /register/game-server", handleRegisterGameServer)

	url := getFullUrl(*addr)
	log.Println("Game Server running on", url)
	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
