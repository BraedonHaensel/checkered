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
	gameID      uuid.UUID
	redPlayer   *Client
	blackPlayer *Client
	GameState   *Game
	resultChan  chan GameResult
}

// checks if the game has finished
// returns (true, winner_username, loser_username)
// if the game is finished
// returns (false, "", "") otherwise
func (gameRoom *GameRoom) finishedGame() (bool, string, string) {
	// TODO: implement
	return false, "", ""
}

// updates the game state to play the gameMove
// checks that it is a valid move before playing
// and if it returns true then the move has been
// played and the gamestate has been updated,
// otherwise the move was not valid and the
// gamestate is still the same. If the game is over
// then the gameroom will send the result of the game
// to the server loop via resultChan
func (gameRoom *GameRoom) playMove(gameMove GameMove) bool {
	if !gameRoom.isValidMove(gameMove) {
		return false
	}
	// TODO: implement

	// check if the game is finished after playing the move
	is_finished, winner, loser := gameRoom.finishedGame()
	if is_finished {
		result := GameResult{
			gameID: gameRoom.gameID,
			winner: winner,
			loser:  loser,
		}
		gameRoom.resultChan <- result
		// set the players current game to null
		gameRoom.blackPlayer.currentGame = nil
		gameRoom.redPlayer.currentGame = nil

	}
	return true
}

// checks if the game move is a valid move
func (gameRoom *GameRoom) isValidMove(_ GameMove) bool {
	// TODO: implement
	return true
}

func (gameRoom *GameRoom) messageFromNewGame(playerKind string) FoundGame {
	return FoundGame{
		Kind: "start",
		Side: playerKind,
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
	Kind string `json:"type"`
	From int    `json:"source_index"`
	To   int    `json:"destination_index"`
}

type GameResult struct {
	gameID uuid.UUID
	// username of the winner
	winner string
	// username of the loser
	loser string
}
