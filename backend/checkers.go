package main

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

// TODO: fill in the
type GameMove struct {
	fromX int `json:"from_x"`
	fromY int `json:"from_y"`
	toX   int `json:"to_x"`
	toY   int `json:"to_y"`
}

type GameResult struct {
	gameID string
	// username of the winner
	winner string
	// username of the loser
	loser string
}
