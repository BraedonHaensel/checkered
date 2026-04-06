package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	Checkered "github.com/akeuben/checkered"
	"github.com/joho/godotenv"
)

// Setting local addresses including the default nameserver
var (
	addr          = flag.String("addr", "", "http service address")
	nameServerURL = flag.String("ns", "", "full Name Server URL")
)

func main() {
	// Reading command line
	flag.Parse()
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")
	godotenv.Load("../../../.env")

	addr := Checkered.ParseStringOption(*addr, "APP_MATCHMAKER_BIND", ":4000")
	nameServerURL := Checkered.ParseStringOption(*nameServerURL, "APP_NAMESERVER_URL", "http://localhost:9000")

	log.Printf("Using name server located at %s\n", nameServerURL)

	url := Checkered.GetFullURL(addr)

	// Instantiating a new matchmaker object
	matchmaker := Checkered.NewMatchmaker(url, nameServerURL)

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

	http.HandleFunc("POST /internal/new-leader-data-sync", func(w http.ResponseWriter, r *http.Request) {
		matchmaker.HandleNewLeaderDataSyncRequest(w, r)
	})

	// Register with the Name Server
	log.Println("Matchmaker running on", url)
	matchmaker.Register(url)

	// Start a leader election
	go matchmaker.InitiateElection()

	// Start listening for requests
	err := http.ListenAndServe(addr, LeaderMiddleware(matchmaker, Checkered.CORSMiddleware(http.DefaultServeMux)))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Middleware for the initial handling of requests received by this Matchmaker
func LeaderMiddleware(matchmaker *Checkered.Matchmaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.Contains(r.URL.Path, "internal") {
			// Internal request (from direct cross-Matchmaker communication).
			// Handle the request locally
			log.Println("Handling internal request for endpoint:", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Client request, wait until it is safe to proceed
		matchmaker.AcceptingClientRequestsMu.RLock()

		if matchmaker.IsLeader() {
			// Is leader, handle the request locally
			log.Println("Handling client request for endpoint:", r.URL.Path)
			next.ServeHTTP(w, r)
			matchmaker.AcceptingClientRequestsMu.RUnlock()
			return
		}

		// Client request needs to be redirected to the leader. Release the lock
		// as it will not be processed here.
		matchmaker.AcceptingClientRequestsMu.RUnlock()

		// Get the leader URL so we can redirect the client request to the
		// leader server, using a HTTP 307, Temporary Redirect.
		newUrl, err := url.Parse(matchmaker.Leader.URL)
		if err != nil {
			log.Println("Error, failed to parse leader url:", matchmaker.Leader.URL)
			return
		}

		// Set up a proxy to redirect the client request to the leader
		proxy := httputil.NewSingleHostReverseProxy(newUrl)

		// Need to deep clone the request so that we do not
		// double read if we handle locally.
		// https://stackoverflow.com/questions/62017146/http-request-clone-is-not-deep-clone
		r2 := r.Clone(r.Context())
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				// Something went really wrong. We can't clone the data for whatevery reason,
				// assume we are out of memory and crash
				panic("Failed to clone request body")
			}
			r2.Body = io.NopCloser(bytes.NewReader(body))
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		// Handle errors from redirecting client requests to the leader
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Println("Failed to contact leader, initiating election...")
			matchmaker.AcceptingClientRequestsMu.TryRLock()
			matchmaker.AcceptingClientRequestsMu.RUnlock()
			matchmaker.InitiateElection()
			LeaderMiddleware(matchmaker, next).ServeHTTP(w, r)
		}

		// Redirect the client request to the leader
		proxy.ServeHTTP(w, r2)
	})
}
