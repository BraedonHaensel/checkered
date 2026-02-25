package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	// uuid for client
	uuid uuid.UUID
	// chosen username for user
	username string
	// connection
	conn *websocket.Conn

	// current context
	// either INGAME, QUEUING, SPECTATING or IDLE
	status string
	// the gameID of the current game that they are in game
	// or are spectating
	gameId string

	// used to send messages to the client via websocket
	send chan []byte
}

type ClientMessage struct {
	kind    string     `json:"kind"`
	NewGame *FoundGame `json:"found_game,omitempty"`
	newMove *GameMove  `json:new_move,omitempty`
}

// message that is sent to the client when a game has been found
type FoundGame struct {
}

func (c *Client) writeThread() {
	defer c.conn.Close()
	for message := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)
	}
}

func (c *Client) readThread() {
	// TODO: add logic for when a user sends a message back to the client
}

func NewClient(connection *websocket.Conn) Client {
	c := Client{
		uuid: uuid.New(),
		// TODO: figure out how to handle new usernames
		username: "",
		conn:     connection,
		status:   IDLE,
		gameId:   "",
		send:     make(chan []byte),
	}
	return c
}
