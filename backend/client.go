package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	// chosen username for user
	username string
	// connection
	conn *websocket.Conn

	// current context
	// either INGAME, QUEUING, SPECTATING or IDLE
	status string
	// the the current game that they are playing
	currentGame *GameRoom

	// used to send messages to the client via websocket
	send chan []byte
}

type ClientMessage struct {
	Payload interface{}
}

func decodePayload(data []byte) (interface{}, error) {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	_, is_game_move := raw["source_index"]
	// its a game move
	if is_game_move {
		var gameMove GameMove
		err := json.Unmarshal(data, &gameMove)
		if err != nil {
			return nil, err
		}
		return gameMove, err
	}
	_, is_found_game := raw["game_id"]
	// its a game move
	if is_found_game {
		var foundGame FoundGame
		err := json.Unmarshal(data, &foundGame)
		if err != nil {
			log.Printf("%s", err)
			return nil, err
		}
		return foundGame, err
	}

	return nil, fmt.Errorf("unknown payload type")
}

// message that is sent to the client when a game has been found
type FoundGame struct {
	Kind string `json:"type"`
	Side string `json:"player_color"`
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
		w.Close()
	}
}

func (c *Client) readThread() {
	// TODO: add logic for when a user sends a message back to the client
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// client closed the connection for some reason
			return
		}
		payload, err := decodePayload(message)
		if err != nil {
			// invalid json struct
			return
		}
		switch p := payload.(type) {
		case GameMove:
			c.handleNewMove(p)
		case FoundGame:
			c.handleFoundGame(p)
		}
	}
}

func (c *Client) handleFoundGame(p FoundGame) {
	panic("unimplemented")
}

func (c *Client) handleNewMove(p GameMove) {
	gameMoveBytes, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error at Marshalling\n")
		return
	}
	if c.currentGame.redPlayer.username == c.username {
		c.currentGame.blackPlayer.send <- gameMoveBytes
	} else {
		c.currentGame.redPlayer.send <- gameMoveBytes
	}
}

func NewClient(username string, connection *websocket.Conn) Client {
	c := Client{
		// TODO: figure out how to handle new usernames
		username:    username,
		conn:        connection,
		status:      IDLE,
		currentGame: nil,
		send:        make(chan []byte),
	}
	return c
}
