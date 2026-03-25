package checkered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	INGAME     = "InGame"
	QUEUING    = "Queuing"
	SPECTATING = "Spectating"
	IDLE       = "Idle"
	writeWait  = 10 * time.Second
	QUEUE_SIZE = 100
)

type PendingGame struct {
	match       Match
	blackClient *Client
	redClient   *Client
	mu          sync.Mutex
}

type GameServer struct {
	ID int
	// we map username to client
	clients map[string]*Client
	// games that new clients are in
	games          map[uuid.UUID]*Game
	readyQueue     Queue[*Client]
	leaderboard    *Leaderboard
	Mu_leaderboard sync.Mutex

	register     chan *Client
	unregister   chan *Client
	moveReceiver chan GameMove
	gameResults  chan GameResult

	// URL of the Name Server
	nameServerURL string

	pendingGames map[uuid.UUID]*PendingGame

	// Consistency related attributes
	inQueue Queue[]
}

func InitServer(nameServerURL string) *GameServer {
	server := GameServer{
		clients:        make(map[string]*Client),
		games:          make(map[uuid.UUID]*Game),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		leaderboard:    &Leaderboard{},
		Mu_leaderboard: sync.Mutex{},
		gameResults:    make(chan GameResult, 10),
		nameServerURL:  nameServerURL,
		pendingGames:   make(map[uuid.UUID]*PendingGame),
	}
	InitQueue(&server.readyQueue, QUEUE_SIZE)
	return &server
}

func (server *GameServer) handleInternalMessage(w http.ResponseWriter, r *http.Request) {
	// TODO: Determine which type of message was recieved
	// If a game update then update the local game state and ack
	// If a game creation then create a game but do not start
	// If a game deletion then remove the game
}

func (server *GameServer) takeOverGame(w http.ResponseWriter, r *http.Request) {
	// TODO: Setup game for client connections
}

func (server *GameServer) broadcastMessage() {
	// TODO: setup message structure
	// TODO: broadcast message to all servers
	// TODO: wait for acks or timeouts
	// TODO: inform nameserver of any timeouts
}

// Register with the Name Server
func (server *GameServer) Register(url string) {
	id, err := SendRegistrationRequest(url, server.nameServerURL+"/register/game-server")
	if err != nil {
		log.Fatal(err)
	}
	server.ID = id
	log.Println("Registered with ID:", server.ID)
	log.SetPrefix(fmt.Sprintf("[%d] ", server.ID))
}

type RegisterMessage struct {
	Kind     string `json:"type"`
	Username string `json:"user"`
}

type ConfirmRegistration struct {
	Kind string `json:"type"`
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin}

func ServeWs(server *GameServer, w http.ResponseWriter, r *http.Request) {
	// upgrade to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// get a message from the client indicating who they are
	var registerMessage RegisterMessage
	err = conn.ReadJSON(&registerMessage)
	if err != nil {
		log.Println(err)
		return
	}
	// create a new client for this request
	client := NewClient(registerMessage.Username, conn, server.unregister, func(client *Client) {
		server.register <- client
	})
	server.clients[client.username] = &client

	go client.readThread()
	go client.writeThread()

	// Send confirmation of registration to client
	confirmation_msg := ConfirmRegistration{Kind: "registered"}
	confirmation, err := json.Marshal(confirmation_msg)
	client.send <- confirmation

	// let the server know and handle the new client
	// server.register <- &client
	log.Printf("Client \"%s\" connected", client.username)
	// Find the game the user is attached to
	var pendingGame *PendingGame = nil
	var opponent *Client = nil
	for _, pg := range server.pendingGames {
		if pg.match.BlackPlayer == client.username {
			pendingGame = pg
			pendingGame.mu.Lock()
			defer pendingGame.mu.Unlock()
			pendingGame.blackClient = &client
			opponent = pg.redClient
			break
		}
		if pg.match.RedPlayer == client.username {
			pendingGame = pg
			pendingGame.mu.Lock()
			defer pendingGame.mu.Unlock()
			pendingGame.redClient = &client
			opponent = pendingGame.blackClient
			break
		}
	}
	log.Printf("Game Identified %s\n", pendingGame.match.MatchID)
	if pendingGame == nil {
		log.Printf("Player %s is not in a pending game\n", client.username)
		return
	}

	// If the other player hasn't registered start a timeout
	if opponent == nil {
		// TODO: Start forfeit timeout
		log.Println("Waiting for opponent")
		return
	}
	// If the other player has registered start the game
	log.Printf("Starting game: %s", pendingGame.match.MatchID)
	server.StartGame(pendingGame)
}

func (server *GameServer) CreateGame(w http.ResponseWriter, r *http.Request) {
	match, err, status := parseJsonRequestData[Match](r)
	if err != nil {
		fmt.Printf("Failed to parse game creation request %s (%d)", err, status)
		return
	}
	uuid := match.MatchID
	pendingGame := PendingGame{
		match:       match,
		redClient:   nil,
		blackClient: nil,
		mu:          sync.Mutex{},
	}
	server.pendingGames[uuid] = &pendingGame
	log.Printf("Created pending game: %s", uuid)
}

func (server *GameServer) StartGame(pendingGame *PendingGame) {
	gameID := pendingGame.match.MatchID
	delete(server.pendingGames, gameID)
	match := pendingGame.match
	redClient := pendingGame.redClient
	blackClient := pendingGame.blackClient
	if redClient == nil {
		panic("Attempted to start game with nil red client")
	}
	if blackClient == nil {
		panic("Attempted to start game with nil black client")
	}
	gameRoom := Game{
		gameID:        gameID,
		redPlayer:     redClient,
		blackPlayer:   blackClient,
		tileStates:    generateInitialTileStates(),
		turn:          Red,
		previousMoves: make([]GameMove, 0),
		resultChan:    server.gameResults,
		mu:            sync.Mutex{},
	}
	server.games[gameRoom.gameID] = &gameRoom
	log.Println("Game room created")
	// tell both servers about the new game
	redMessage := gameRoom.messageFromNewGame("red", match.BlackPlayer)
	blackMessage := gameRoom.messageFromNewGame("black", match.RedPlayer)
	redBytes, err := json.Marshal(redMessage)
	if err != nil {
		log.Println(err)
		return
	}
	blackBytes, err := json.Marshal(blackMessage)
	if err != nil {
		log.Println(err)
		return
	}
	server.clients[gameRoom.blackPlayer.username].currentGame = &gameRoom
	server.clients[gameRoom.redPlayer.username].currentGame = &gameRoom
	// send the message that they have found a game to both players
	server.clients[gameRoom.blackPlayer.username].send <- blackBytes
	server.clients[gameRoom.redPlayer.username].send <- redBytes
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *GameServer) ServerLoop() {
	for {
		select {
		case client := <-server.register:
			// Find the game the user is attached to
			var pendingGame *PendingGame = nil
			var opponent *Client = nil
			for _, pg := range server.pendingGames {
				if pg.match.BlackPlayer == client.username {
					pendingGame = pg
					opponent = pg.redClient
					break
				}
				if pg.match.RedPlayer == client.username {
					pendingGame = pg
					opponent = pendingGame.blackClient
					break
				}
			}
			if pendingGame == nil {
				log.Printf("Player %s is not in a pending game\n", client.username)
				break
			}

			// If the other player hasn't registered start a timeout
			if opponent == nil {
				// TODO: Start forfeit timeout
				break
			}
			// If the other player has registered start the game
			server.StartGame(pendingGame)
		case <-server.unregister:
			log.Println("Attempted unregister")
			// TODO: remove client from game rooms
		case gameResult := <-server.gameResults:
			_, exists := server.games[gameResult.gameID]

			if exists {
				matchmakingServers, err := SendServerListRequest(
					server.nameServerURL + "/matchmakers",
				)
				if err != nil {
					log.Printf("Failed to fetch game servers: %s", err)
					break
				}
				if len(matchmakingServers) == 0 {
					log.Println("No match making servers available")
					break
				}
				matchmakingServer := matchmakingServers[0]
				log.Printf(
					"Selected match making server: %s (ID: %d)",
					matchmakingServer.URL,
					matchmakingServer.ID,
				)

				gameResultMessage := GameResultStruct{
					GameID: gameResult.gameID,
					Winner: *gameResult.winner,
					Loser:  *gameResult.loser,
				}
				gameResultBytes, err := json.Marshal(gameResultMessage)
				if err != nil {
					log.Printf("Failed to marshal result: %s", err)
					break
				}
				res, err := http.Post(
					matchmakingServer.URL+"/match/updateleaderboard",
					"application/json",
					bytes.NewBuffer(gameResultBytes),
				)
				if err != nil {
					log.Printf("Failed to send game results to match making server: %s", err)
					break
				}
				log.Printf("Leaderboard updated!")
				defer res.Body.Close()
				endMatchRequest := EndMatchRequest{MatchID: gameResult.gameID}
				endMatchRequestBytes, err := json.Marshal(endMatchRequest)
				if err != nil {
					log.Printf("Failed to marshal end game request: %s", err)
					return
				}
				res, err = http.Post(
					matchmakingServer.URL+"/match/end",
					"application/json",
					bytes.NewBuffer(endMatchRequestBytes),
				)
				if err != nil {
					log.Printf("Failed to send end game %s", err)
				}

				delete(server.games, gameResult.gameID)
			} else {
				log.Printf("Game %s does not exist\n", gameResult.gameID)
			}
		}
	}
}
