package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var addr = flag.String("addr", ":9000", "http service address")

// log the address the server is running on
func logAddress() {
	if strings.HasPrefix(*addr, ":") {
		// port provided, so determine the local IP address
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

// opens a file and returns a list of its lines
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

// GET /matchmaking-servers
func handleGetMatchmakingServers(w http.ResponseWriter, req *http.Request) {
	servers := readFileLines("./matchmaking-servers.txt")

	jsonData, err := json.Marshal(servers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /matchmaking-servers -", string(jsonData))
	w.Write(jsonData)
}

// GET /game-servers
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	servers := readFileLines("./game-servers.txt")

	jsonData, err := json.Marshal(servers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("GET /game-servers -", string(jsonData))
	w.Write(jsonData)
}

// run the Name Server
func main() {
	flag.Parse()
	logAddress()

	// request handlers
	http.HandleFunc("/matchmaking-servers", handleGetMatchmakingServers)
	http.HandleFunc("/game-servers", handleGetGameServers)

	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
