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

	// ---------- Not sure if we'll use ----------

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

	// TODO get list of servers from Name Server before/at the start of the election
	// And have a server refresh loop
	go matchmaker.ServerLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Checkered.MatchmakerWs(matchmaker, w, r)
	})

	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
