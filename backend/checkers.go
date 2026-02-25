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

type GameMove struct {
	from int `json:"from"`
	to   int `json:"to"`
}

type GameResult struct {
	gameID string
	// username of the winner
	winner string
	// username of the loser
	loser string
}
