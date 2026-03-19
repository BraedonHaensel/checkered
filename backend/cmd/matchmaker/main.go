package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	http.HandleFunc("POST /queue/add", func(w http.ResponseWriter, r *http.Request) {
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

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Otherwise, we redirect the request to the leader server, using a HTTP 307, Temporary Redirect.
		newUrl, err := url.Parse(r.URL.Scheme + matchmaker.Leader.URL)

		if err != nil {
			println("Error, could not parse leader url")
		}
		
		proxy := httputil.NewSingleHostReverseProxy(newUrl)

		proxy.ErrorHandler = func (w http.ResponseWriter, r *http.Request, err error) {
			log.Println("Failed to contact leader, initiating election...");
			matchmaker.InitiateElection();
			// TODO: Resend the request that failed to reach the leader.
		};

		proxy.ServeHTTP(w, r);
	})
}
