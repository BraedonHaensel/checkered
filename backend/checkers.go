package main

import (
	"github.com/google/uuid"
)

type Game struct {
	gameID      string
	redPlayer   string
	blackPlayer string
}

type GameRoom struct {
	gameID      string
	redPlayer   *Client
	blackPlayer *Client
	GameState   *Game
}

func (gameRoom *GameRoom) playMove(gameMove GameMove) {
	panic("unimplemented")
}

func (gameRoom *GameRoom) isValidMove(gameMove GameMove) bool {
	panic("unimplemented")
}

func (gameRoom *GameRoom) messageFromNewGame(playerKind string) FoundGame {
	return FoundGame{
		GameID: gameRoom.gameID,
		Side:   playerKind,
	}
}

func otherPlayer(playerKind string) string {
	if playerKind == "red" {
		return "black"
	}
	if playerKind == "black" {
		return "red"
	}
	return "error"
}

type GameMove struct {
	GameID   uuid.UUID `json:"game_id"`
	Username string    `json:"user"`
	From     int       `json:"from"`
	To       int       `json:"to"`
}

type GameResult struct {
	gameID string
	// username of the winner
	winner string
	// username of the loser
	loser string
}
