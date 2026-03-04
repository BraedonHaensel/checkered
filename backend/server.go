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
)

type Server struct {
	// we map uuid to client
	clients map[uuid.UUID]*Client
	// games that new clients are in
	games       map[uuid.UUID]*GameRoom
	InQueue     []*Client
	leaderboard *Leaderboard

	register     chan *Client
	unregister   chan *Client
	newGame      chan *GameRoom
	moveReciever chan GameMove
}

var addr = flag.String("addr", ":8080", "http service address")

func InitServer() *Server {
	return &Server{
		leaderboard: &Leaderboard{},
	}
}

var upgrader = websocket.Upgrader{}

func serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	// upgrade to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// create a new client for this request
	client := NewClient(conn, server.moveReciever)

	// let the server know and handle the new client
	server.register <- &client

	go client.readThread()
	go client.writeThread()
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *Server) serverLoop() {
	for {
		select {
		case new_client := <-server.register:
			server.clients[new_client.Uuid] = new_client
		case unregister := <-server.unregister:
			delete(server.clients, unregister.Uuid)
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
		case gameMove := <-server.moveReciever:
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
