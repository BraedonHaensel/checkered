package main

import (
	"encoding/json"
	"fmt"
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
	Payload interface{}
}

func decodePayload(data []byte) (interface{}, error) {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	_, is_game_move := raw["from"]
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
			return nil, err
		}
		return foundGame, err
	}

	return nil, fmt.Errorf("unknown payload type")
}

// message that is sent to the client when a game has been found
type FoundGame struct {
	GameID           string `json:"game_id"`
	StartingPosition string `json:"starting_position"`
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
	panic("unimplemented")
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
