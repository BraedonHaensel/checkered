package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
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

type Server struct {
	// we map username to client
	clients map[string]*Client
	// games that new clients are in
	games       map[uuid.UUID]*Game
	readyQueue  Queue[*Client]
	leaderboard *Leaderboard

	register     chan *Client
	unregister   chan *Client
	moveReceiver chan GameMove
	gameResults  chan GameResult
}

var addr = flag.String("addr", ":8080", "http service address")

func InitServer() *Server {
	server := Server{
		clients:     make(map[string]*Client),
		games:       make(map[uuid.UUID]*Game),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		leaderboard: &Leaderboard{},
		gameResults: make(chan GameResult, 10),
	}
	InitQueue(&server.readyQueue, QUEUE_SIZE)
	return &server
}

type RegisterMessage struct {
	Kind     string `json:"type"`
	Username string `json:"user"`
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin}

func serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
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
	client := NewClient(registerMessage.Username, conn, server)

	go client.readThread()
	go client.writeThread()

	// let the server know and handle the new client
	server.register <- &client

}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *Server) serverLoop() {
	log.Println("Server Running...")
	for {
		select {
		case new_client := <-server.register:
			server.clients[new_client.username] = new_client
			log.Printf(
				"New client \"%s\"",
				new_client.username,
			)
			// queue the new client into a game
			Enqueue(&server.readyQueue, new_client)
		case unregister := <-server.unregister:
			delete(server.clients, unregister.username)
			RemoveValue(&server.readyQueue, unregister)
			log.Printf(
				"Client \"%s\" deregistered",
				unregister.username,
			)
			// TODO: remove client from game rooms
		case gameResult := <-server.gameResults:
			server.leaderboard.UpdateLeaderboard(gameResult)
		default:
			if server.readyQueue.size < 2 {
				break
			}
			redPlayer := Dequeue(&server.readyQueue)
			blackPlayer := Dequeue(&server.readyQueue)
			log.Printf("New game created (red: %s, black: %s)\n", redPlayer.username, blackPlayer.username)
			gameRoom := Game{
				gameID:       	uuid.New(),
				redPlayer:    	redPlayer,
				blackPlayer:  	blackPlayer,
				tileStates:   	generateInitialTileStates(),
				turn:         	Red,
				previousMoves: 	make([]GameMove, 0),
				resultChan:   server.gameResults,
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

func main() {
	flag.Parse()
	server := InitServer()
	go server.serverLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(server, w, r)
	})
	http.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		server.getLeaderboard(w, r)
	})
	err := http.ListenAndServe(*addr, CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
