package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var addr = flag.String("addr", ":9000", "http service address")

// Log the address the server is running on
func logAddress() {
	if (strings.HasPrefix(*addr, ":")) {
		// Port provided. Determine the local IP address
		// Source: https://gosamples.dev/local-ip-address/
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		localAddress := conn.LocalAddr().(*net.UDPAddr)
		
		log.Printf("Name Server running on %s%s\n", localAddress.IP, *addr,)
	} else {
		log.Println("Name Server running on", *addr)
	}
}

// GET /matchmaking-servers
func handleGetMatchmakingServers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "No matchmaking servers\n")
}

// GET /game-servers
func handleGetGameServers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "No game servers\n")
}

func main() {
	flag.Parse()
	logAddress()

	http.HandleFunc("/matchmaking-servers", handleGetMatchmakingServers)
	http.HandleFunc("/game-servers", handleGetGameServers)

	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
