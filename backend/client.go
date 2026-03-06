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
	currentGame *Game

	// used to send messages to the client via websocket
	send chan []byte

	// the registration channel
	register func(*Client)
	enqueued bool
}

type ClientMessage struct {
	Payload interface{}
}

func parseMessage[T any](data []byte) (interface{}, error) {
	var msg T
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func decodePayload(data []byte) (interface{}, error) {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	// Determine the type of the message
	msg_type_bytes, has_type := raw["type"]
	if !has_type {
		// Invalid message
		return nil, fmt.Errorf("unknown payload type")
	}
	var msg_type string
	err = json.Unmarshal(msg_type_bytes, &msg_type)
	if err != nil {
		// Type field is of wrong type
		return nil, fmt.Errorf("unable to decode payload")
	}
	// Parse the message based on the type
	switch msg_type {
	case "move":
		return parseMessage[GameMove](data)
	case "enqueue":
		return parseMessage[EnqueueRequest](data)
	}

	// _, is_found_game := raw["game_id"]
	// // its a game move
	// if is_found_game {
	// 	var foundGame FoundGame
	// 	err := json.Unmarshal(data, &foundGame)
	// 	log.Printf("Game found")
	// 	if err != nil {
	// 		log.Printf("%s", err)
	// 		return nil, err
	// 	}
	// 	return foundGame, err
	// }

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
			// log.Printf("Received message (%s)\n", p.Kind)
			c.handleNewMove(p)
		case EnqueueRequest:
			// log.Printf("Received message (%s)\n", p.Kind)
			c.register(c) // Enqueue
		case FoundGame: // TODO: remove as is not an actual message the client can send
			// log.Printf("Received message (%s)\n", p.Kind)
			c.handleFoundGame(p)
		}
	}
}

func (c *Client) handleFoundGame(p FoundGame) {
	panic("unimplemented")
}

type GameStateUpdate struct {
	Kind         string      `json:"type"`
	TileStates   []TileState `json:"tile_states"`
	Turn         string      `json:"turn"`
	PreviousMove *GameMove   `json:"previous_move,omitempty"`
}

type GameEndMessage struct {
	Kind   string `json:"type"`
	Winner string `json:"winner"`
}

type EnqueueRequest struct {
	Kind string `json:"type"`
}

func (c *Client) handleNewMove(p GameMove) {

	validMove := c.currentGame.playMove(p)

	newState := GameStateUpdate{
		Kind:         "update_state",
		TileStates:   c.currentGame.tileStates,
		Turn:         "red",
		PreviousMove: c.currentGame.previousMove,
	}

	if c.currentGame.turn == Black {
		newState.Turn = "black"
	}

	gameStateBytes, err := json.Marshal(newState)
	if err != nil {
		log.Printf("Error at Marshalling\n")
		return
	}
	if !validMove {
		c.send <- gameStateBytes
	} else {
		otherPlayer(c.username, c.currentGame).send <- gameStateBytes
	}
}

func otherPlayer(username string, game *Game) *Client {
	if game.blackPlayer.username == username {
		return game.redPlayer
	}
	return game.blackPlayer
}

func NewClient(username string, connection *websocket.Conn, register func(*Client)) Client {
	c := Client{
		// TODO: figure out how to handle new usernames
		username:    username,
		conn:        connection,
		status:      IDLE,
		currentGame: nil,
		send:        make(chan []byte),
		register:    register,
		enqueued:    false,
	}
	return c
}
