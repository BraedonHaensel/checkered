package checkered

import (
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
	}
	InitQueue(&server.readyQueue, QUEUE_SIZE)
	return &server
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
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *GameServer) ServerLoop() {
	for {
		select {
		case client := <-server.register:
			var inGameAlready bool
			for _, game := range server.games {
				if game.redPlayer.username == client.username {
					// client is already in a game
					inGameAlready = true
				}
				if game.blackPlayer.username == client.username {
					// client is already in a game
					inGameAlready = true
				}
			}
			if inGameAlready {
				break
			}
			if client.enqueued {
				log.Printf("Client \"%s\" is already enqueued", client.username)
			} else {
				log.Printf("Client \"%s\" has joined the queue", client.username)
				// queue the new client into a game
				client.enqueued = true
				server.readyQueue.enqueue(client)
				log.Println("Queue:")
				server.readyQueue.forEach(func(c *Client, i int) {
					log.Printf("\t%d: %s", i, c.username)
				})
				log.Println("=================")
			}
		case unregister := <-server.unregister:
			delete(server.clients, unregister.username)
			RemoveValue(&server.readyQueue, unregister)
			log.Printf(
				"Client \"%s\" deregistered",
				unregister.username,
			)
			// TODO: remove client from game rooms
		case gameResult := <-server.gameResults:
			game, exists := server.games[gameResult.gameID]

			if exists {
				server.Mu_leaderboard.Lock()
				server.leaderboard.UpdateLeaderboard(gameResult)
				server.Mu_leaderboard.Unlock()

				game.mu.Lock()
				if game.blackPlayer != nil {
					game.blackPlayer.currentGame = nil
				}
				if game.redPlayer != nil {
					game.redPlayer.currentGame = nil
				}
				game.mu.Unlock()
				delete(server.games, gameResult.gameID)
			}
		default:
			if server.readyQueue.size < 2 {
				break
			}
			redPlayer := server.readyQueue.dequeue()
			blackPlayer := server.readyQueue.dequeue()
			redPlayer.enqueued = false
			blackPlayer.enqueued = false
			log.Printf("New game created (red: %s, black: %s)\n", redPlayer.username, blackPlayer.username)
			gameRoom := Game{
				gameID:        uuid.New(),
				redPlayer:     redPlayer,
				blackPlayer:   blackPlayer,
				tileStates:    generateInitialTileStates(),
				turn:          Red,
				previousMoves: make([]GameMove, 0),
				resultChan:    server.gameResults,
				mu:            sync.Mutex{},
			}
			server.games[gameRoom.gameID] = &gameRoom
			// tell both servers about the new game
			redMessage := gameRoom.messageFromNewGame("red", blackPlayer.username)
			blackMessage := gameRoom.messageFromNewGame("black", redPlayer.username)
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
	}
}
