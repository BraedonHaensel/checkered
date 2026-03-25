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

	http.HandleFunc("POST /queue/add", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.AddToQueue(w, r)
	})

	http.HandleFunc("POST /queue/poll", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.QueuePollRequest(w, r)
	})

	http.HandleFunc("/queue/leave", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.LeaveQueueRequest(w, r)
	})

	http.HandleFunc("POST /match/request-new-game-server", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.RequestNewGameServer(w, r)
	})

	http.HandleFunc("POST /match/updateleaderboard", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.UpdateLeaderboard(w, r)
	})

	http.HandleFunc("POST /match/end", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.EndMatch(w, r)
	})

	// --------------------------------------------

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

	// Replication endpoints

	http.HandleFunc("POST /internal/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.SetLeaderboard(w, r)
	})
	http.HandleFunc("POST /internal/queue", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.SetQueue(w, r)
	})
	http.HandleFunc("POST /internal/matches", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.SetMatches(w, r)
	})

	http.HandleFunc("GET /internal/request-data", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.PackageData(w, r)
	})

	// Register with the Name Server
	log.Println("Matchmaker running on", url)
	matchmaker.Register(url)

	// Start a leader election
	matchmaker.InitiateElection()

	// Request most recent data 
	if matchmaker.Leader.ID == matchmaker.ID {
		// We have become the new leader. We need to request the data from some other server.
		// Since requests are not closed until everything has been synchronized, we can pull 
		// from any of the other servers, if one exists.
		other, err := matchmaker.ChooseRandomOtherServer()
		if err == nil {
			matchmaker.RequestUpdatedData(*other)
		}
	} else {
		// There is another authorative leader. We send the message to them explicitly
		matchmaker.RequestUpdatedData(matchmaker.Leader)
	}

	err := http.ListenAndServe(*addr, LeaderMiddleware(matchmaker, Checkered.CORSMiddleware(http.DefaultServeMux)))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func LeaderMiddleware(matchmaker *Checkered.Matchmaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// If this is the leader server, or an internal route, that is, a route that is
		// destined for this server specifically (for cross-matchmaker communication),
		// then we handle the request locally.
		if strings.Contains(r.URL.Path, "internal") || matchmaker.IsLeader() {
			log.Println("Handling request locally for endpoint:", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we redirect the request to the leader server, using a HTTP 307, Temporary Redirect.
		newUrl, err := url.Parse(matchmaker.Leader.URL)

		if err != nil {
			println("Error, could not parse leader url")
		}

		proxy := httputil.NewSingleHostReverseProxy(newUrl)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Println("Failed to contact leader, initiating election...")
			matchmaker.InitiateElection()
			LeaderMiddleware(matchmaker, next).ServeHTTP(w, r)
		}

		proxy.ServeHTTP(w, r)
	})
}
