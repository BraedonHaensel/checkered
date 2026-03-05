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

type GameStateUpdate struct {
	Kind         string      `json:"type"`
	TileStates   []TileState `json:"tile_states"`
	Turn         string      `json:"turn"`
	PreviousMove *GameMove   `json:"previous_move,omitempty"`
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
