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
	// we map uuid to client
	clients map[uuid.UUID]*Client
	// games that new clients are in
	games       map[uuid.UUID]*Game
	readyQueue  Queue[*Client]
	leaderboard *Leaderboard

	register     chan *Client
	unregister   chan *Client
	newGame      chan *Game
	moveReceiver chan GameMove
}

var addr = flag.String("addr", ":8080", "http service address")

func InitServer() *Server {
	server := Server{
		clients:      make(map[uuid.UUID]*Client),
		games:        make(map[uuid.UUID]*Game),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		newGame:      make(chan *Game),
		moveReceiver: make(chan GameMove),
		leaderboard:  &Leaderboard{},
	}
	InitQueue(&server.readyQueue, QUEUE_SIZE)
	return &server
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
	// create a new client for this request
	client := NewClient(conn, server.moveReceiver)

	go client.readThread()
	go client.writeThread()

	// let the server know and handle the new client
	server.register <- &client

	log.Printf("Client connected\n")

}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *Server) serverLoop() {
	log.Println("Server Running...")
	for {
		select {
		case new_client := <-server.register:
			server.clients[new_client.Uuid] = new_client
			log.Printf(
				"New client \"%s\" registered with UUID: \"%s\"",
				new_client.username,
				new_client.Uuid,
			)
		case unregister := <-server.unregister:
			delete(server.clients, unregister.Uuid)
			log.Printf(
				"Client \"%s\" with UUID: \"%s\" deregistered",
				unregister.username,
				unregister.Uuid,
			)
			// TODO: remove client from game rooms
		case newGame := <-server.newGame:
			// tell both servers about the new game
			redMessage := newGame.messageFromNewGame("red")
			blackMessage := newGame.messageFromNewGame("black")
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
			// send the message that they have found a game to both players
			server.clients[newGame.blackPlayer.Uuid].send <- blackBytes
			server.clients[newGame.redPlayer.Uuid].send <- redBytes
		case gameMove := <-server.moveReceiver:
			// TODO: check if the game exists, check if the user is using the same term
			gameState := server.games[gameMove.GameID]
			isValidMove := gameState.isValidMove(gameMove)
			if !isValidMove {
				// TODO: send a message back to the client that this is an invalid move
			}
			gameState.playMove(gameMove)
			playerUUID := gameMove.UserID
			message, err := json.Marshal(gameMove)
			if err != nil {
				log.Println(err)
				return
			}
			if gameState.blackPlayer.Uuid == playerUUID {
				// send the move to the red player
				server.clients[gameState.redPlayer.Uuid].send <- message
			}
			if gameState.redPlayer.Uuid == playerUUID {
				// send the move to the black player
				server.clients[gameState.blackPlayer.Uuid].send <- message
			}
		default:
			if server.readyQueue.size >= 2 {
				redPlayer := Dequeue(&server.readyQueue)
				blackPlayer := Dequeue(&server.readyQueue)
				log.Printf("New game created (red: %s, black: %s)\n", redPlayer.Uuid, blackPlayer.Uuid)
				gameRoom := Game{redPlayer: redPlayer, blackPlayer: blackPlayer}
				server.newGame <- &gameRoom
			}
		}
	}
}

func main() {
	flag.Parse()
	server := InitServer()
	server.leaderboard.AddPlayerToLeaderboard("akeuben")
	server.leaderboard.AddPlayerToLeaderboard("test")
	server.leaderboard.UpdateLeaderboard(GameResult{
		gameID: "test",
		winner: "akeuben",
		loser: "test",
	})
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
