package main

import (
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
	clients     map[uuid.UUID]*Client
	games       map[uuid.UUID]*GameRoom
	InQueue     []*Client
	leaderboard *Leaderboard

	register   chan *Client
	unregister chan *Client
}

var addr = flag.String("addr", ":8080", "http service address")

func InitServer() *Server {
	return &Server{}
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
	client := NewClient(conn)

	// let the server know and handle the new client
	server.register <- &client

	go client.readThread()
	go client.writeThread()
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *Server) serverLoop() {
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
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
