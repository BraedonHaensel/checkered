package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

var (
	addr          = flag.String("addr", ":5000", "http service address")
	nameServerURL = flag.String("ns", "http://localhost:9000", "full Name Server URL")
)

func main() {
	// Reading command line
	flag.Parse()
	url := Checkered.GetFullURL(*addr)

	// Create the server object
	server := Checkered.InitServer(*nameServerURL)

	// Start the server loop and register handlers
	go server.ServerLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Checkered.ServeWs(server, w, r)
	})
	http.HandleFunc("POST /newGame", server.CreateGame)

	// Endpoint to check the health of the Game Server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Note: Uncomment the below "select {}" line to simulate an unresponsive Game Server
		// select {}
		w.WriteHeader(http.StatusOK)
	})

	// Register with the Name Server
	log.Println("Game Server running on", url)
	server.Register(url)

	// Start the ticker for the periodic refresh of the list of other Game Servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.StartOtherGameServersRefreshTicker(ctx)

	// Start the HTTP server
	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
