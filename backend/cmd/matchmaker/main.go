package main

import (
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

var addr = flag.String("addr", ":4000", "http service address")
var nameServerUrl = flag.String("ns", "http://localhost:9000", "full Name Server URL")

func main() {
	flag.Parse()
	matchmaker := Checkered.NewMatchmaker()
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

	// Register with the Name Server
	url := Checkered.GetFullUrl(*addr)
	Checkered.SendRegistrationRequest(url, *nameServerUrl+"/register/matchmaker")

	log.Println("Matchmaker running on", url)
	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
