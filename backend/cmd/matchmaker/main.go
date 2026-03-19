package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	Checkered "github.com/akeuben/checkered"
)

var (
	addr          = flag.String("addr", ":4000", "http service address")
	nameServerURL = flag.String("ns", "http://localhost:9000", "full Name Server URL")
)

func main() {
	flag.Parse()
	url := Checkered.GetFullURL(*addr)
	matchmaker := Checkered.NewMatchmaker(url, *nameServerURL)

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

	// Leader election endpoints
	http.HandleFunc("POST /internal/leader-election/election", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleElectionRequest(w, r)
	})
	http.HandleFunc("POST /internal/leader-election/bully", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleBullyRequest(w, r)
	})
	http.HandleFunc("POST /internal/leader-election/leader", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleLeaderRequest(w, r)
	})

	// Register with the Name Server
	log.Println("Matchmaker running on", url)
	matchmaker.Register(url)

	// Start a leader election
	go matchmaker.InitiateElection()

	// TODO get list of servers from Name Server before/at the start of the election
	// And have a server refresh loop

	err := http.ListenAndServe(*addr, LeaderMiddleware(&matchmaker, Checkered.CORSMiddleware(http.DefaultServeMux)))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func LeaderMiddleware(matchmaker *Checkered.Matchmaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If this is the leader server, or an internal route, that is, a route that is
		// destined for this server specifically (for cross-matchmaker communication), 
		// then we handle the request locally.
		if strings.Contains(r.URL.Path, "internal") || matchmaker.IsLeader() {
			log.Println("Handling reuqest locally for endpoint:", r.URL.Path);
			next.ServeHTTP(w,r);
			return;
		}
		
		// Otherwise, we redirect the request to the leader server, using a HTTP 307, Temporary Redirect.
		http.Redirect(w, r, r.URL.Scheme + matchmaker.Leader.URL + r.URL.Path, http.StatusTemporaryRedirect)
	})
}
