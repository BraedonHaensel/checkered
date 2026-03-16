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
	"os"
	"strings"
)

type registerServerRequest struct {
	Address string `json:"address"`
}

var addr = flag.String("addr", ":9000", "http service address")

// Log the address the server is running on
func logAddress() {
	if strings.HasPrefix(*addr, ":") {
		// Port provided, so determine the local IP address
		// source: https://gosamples.dev/local-ip-address/
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		localAddress := conn.LocalAddr().(*net.UDPAddr)
		log.Printf("Name Server running on %s%s\n", localAddress.IP, *addr)
	} else {
		log.Println("Name Server running on", *addr)
	}
}

// Opens a file and returns a list of its lines
func readFileLines(name string) []string {
	contents, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	lines := []string{}
	for line := range strings.Lines(string(contents)) {
		trimmedLine := strings.TrimSpace(line)
		lines = append(lines, trimmedLine)
	}
	return lines
}

// Parses JSON data from a request. Returns an error and HTTP status code if the
// parsing fails.
func pasreJsonRequestData[T any](req *http.Request) (T, error, int) {
	var data T;
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
	servers := readFileLines("./matchmaking-servers.txt")

	jsonData, err := json.Marshal(servers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /matchmakers -", string(jsonData))
	w.Write(jsonData)
}

// GET /game-servers - Get the list of known Game Servers
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	servers := readFileLines("./game-servers.txt")

	jsonData, err := json.Marshal(servers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /game-servers -", string(jsonData))
	w.Write(jsonData)
}

// POST /register/matchmaker - Handle registering a Matchmaker server.
// Name Server
func handleRegisterMatchmaker(w http.ResponseWriter, req *http.Request) {
	data, err, errStatus := pasreJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	log.Println("POST /register/matchmaker -", data.Address)

	// Send response to client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker added successfully."))
}

// POST /register/game-server Handle registering a Game Server.
func handleRegisterGameServer(w http.ResponseWriter, req *http.Request) {
	data, err, errStatus := pasreJsonRequestData[registerServerRequest](req)
	if err != nil {
		http.Error(w, err.Error(), errStatus)
		return
	}
	log.Println("POST /register/game-server -", data.Address)

	// Send response to client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Matchmaker added successfully."))
}

// Run the Name Server
func main() {
	flag.Parse()
	logAddress()

	// Request handlers
	http.HandleFunc("/matchmakers", handleGetMatchmakers)
	http.HandleFunc("/game-servers", handleGetGameServers)
	http.HandleFunc("POST /register/matchmaker", handleRegisterMatchmaker)
	http.HandleFunc("POST /register/game-server", handleRegisterGameServer)

	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
