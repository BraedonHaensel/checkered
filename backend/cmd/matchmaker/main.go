package main

import (
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

// Setting local addresses including the default nameserver
var (
	addr          = flag.String("addr", ":4000", "http service address")
	nameServerURL = flag.String("ns", "http://localhost:9000", "full Name Server URL")
)

func main() {

	// Reading command line
	flag.Parse()
	url := Checkered.GetFullURL(*addr)

	// Instantiating a new matchmaker object
	matchmaker := Checkered.NewMatchmaker(url, *nameServerURL)

	// ---------------------------------------------

	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.GetLeaderboard(w, r)
	})

	http.HandleFunc("/queue/add", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.AddToQueue(w, r)
	})

	http.HandleFunc("/queue/poll", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.QueuePollRequest(w, r)
	})

	http.HandleFunc("/queue/leave", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.LeaveQueueRequest(w, r)
	})

	http.HandleFunc("POST /match/updateleaderboard", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.UpdateLeaderboard(w, r)
	})

	http.HandleFunc("POST /match/end", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.EndMatch(w, r)
	})

	// --------------------------------------------

	// Leader election endpoints
	http.HandleFunc("POST /leader-election/election", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleElectionRequest(w, r)
	})
	http.HandleFunc("POST /leader-election/bully", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleBullyRequest(w, r)
	})
	http.HandleFunc("POST /leader-election/leader", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleLeaderRequest(w, r)
	})

	// Register with the Name Server
	log.Println("Matchmaker running on", url)
	matchmaker.Register(url)

	// Start a leader election
	go matchmaker.InitiateElection()

	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
