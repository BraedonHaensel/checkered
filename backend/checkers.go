package checkered

import (
	"encoding/json"
	"log"
	"math"
	"sync"

	"github.com/google/uuid"
)

type TileState int
type PieceColor int
type MoveDirection int

const (
	Empty = iota
	RedStandardPiece
	RedKingPiece
	BlackStandardPiece
	BlackKingPiece
)

const (
	Black = iota
	Red
)

const (
	UpLeft = iota
	UpRight
	DownLeft
	DownRight
)

type Game struct {
	gameID              uuid.UUID
	gameServer          int
	redPlayer           *Client
	blackPlayer         *Client
	redPlayerUsername   string
	blackPlayerUsername string
	tileStates          []TileState
	turn                PieceColor
	previousMoves       []GameMove
	resultChan          chan GameResult
	blackWantsDraw      bool
	redWantsDraw        bool
	mu                  sync.Mutex
	snapshotId          int
	updateCallback      func(*Game)
}

type GameSnapshot struct {
	GameID              string      `json:"gameID"`
	GameServer          int         `json:"gameServer"`
	RedPlayerUsername   string      `json:"redPlayerUsername"`
	BlackPlayerUsername string      `json:"blackPlayerUsername"`
	TileStates          []TileState `json:"tileStates"`
	Turn                PieceColor  `json:"turn"`
	PreviousMoves       []GameMove  `json:"previousMoves"`
	BlackWantsDraw      bool        `json:"blackWantsDraw"`
	RedWantsDraw        bool        `json:"redWantsDraw"`
	Delete              bool        `json:"delete"`
	SnapshotId          int         `json:"snapshotId"`
}

func (game *Game) CreateSnapshot(delete bool) GameSnapshot {
	return GameSnapshot{
		GameID:              game.gameID.String(),
		GameServer:          game.gameServer,
		RedPlayerUsername:   game.redPlayerUsername,
		BlackPlayerUsername: game.blackPlayerUsername,
		TileStates:          game.tileStates,
		Turn:                game.turn,
		PreviousMoves:       game.previousMoves,
		BlackWantsDraw:      game.blackWantsDraw,
		RedWantsDraw:        game.redWantsDraw,
		Delete:              delete,
		SnapshotId:          game.snapshotId,
	}
}

func (game *Game) ApplySnapshot(snapshot GameSnapshot) {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.gameID = uuid.MustParse(snapshot.GameID)
	game.gameServer = snapshot.GameServer
	game.redPlayerUsername = snapshot.RedPlayerUsername
	game.blackPlayerUsername = snapshot.BlackPlayerUsername
	game.tileStates = snapshot.TileStates
	game.turn = snapshot.Turn
	game.previousMoves = snapshot.PreviousMoves
	game.blackWantsDraw = snapshot.BlackWantsDraw
	game.redWantsDraw = snapshot.RedWantsDraw
	game.snapshotId = snapshot.SnapshotId
}

func (game *Game) RegisterUpdate() {
	game.snapshotId++
	game.updateCallback(game)
}

func generateInitialTileStates() []TileState {
	tileStates := make([]TileState, 32)

	for i := range tileStates {
		if i < 12 {
			tileStates[i] = RedStandardPiece
		} else if i < 20 {
			tileStates[i] = Empty
		} else {
			tileStates[i] = BlackStandardPiece
		}
	}

	return tileStates
}

func isBlackPiece(state TileState) bool {
	switch state {
	case BlackStandardPiece:
		return true
	case BlackKingPiece:
		return true
	default:
		return false
	}
}

func isRedPiece(state TileState) bool {
	switch state {
	case RedStandardPiece:
		return true
	case RedKingPiece:
		return true
	default:
		return false
	}
}

func isStandardPiece(state TileState) bool {
	switch state {
	case RedStandardPiece:
		return true
	case BlackStandardPiece:
		return true
	default:
		return false
	}
}

func isKingPiece(state TileState) bool {
	switch state {
	case RedKingPiece:
		return true
	case BlackKingPiece:
		return true
	default:
		return false
	}
}

func tileIndexToRow(index int) int {
	return int(math.Floor(float64(index) / 4.0))
}

func isOffsetRow(row int) bool {
	return row%2 == 0
}

func tileIndexToCol(index int) int {
	col := (index % 4) * 2
	offset := 0
	if tileIndexToRow(index)%2 == 0 {
		offset = 1
	}
	return col + offset
}

func isUpwardMoveDirection(direction MoveDirection) bool {
	switch direction {
	case UpLeft:
		return true
	case UpRight:
		return true
	default:
		return false
	}
}

func isLeftwardMoveDiretion(direction MoveDirection) bool {
	switch direction {
	case UpLeft:
		return true
	case DownLeft:
		return true
	default:
		return false
	}
}

func isJumpMove(from int, to int) bool {
	distance := to - from

	return int(math.Abs(float64(distance))) >= 7
}

func getMoveAmountForDirection(direction MoveDirection, isOffsetRow bool, isJump bool) int {
	if isUpwardMoveDirection(direction) {
		if isLeftwardMoveDiretion(direction) {
			// Up left move
			if isJump {
				return -9
			} else if isOffsetRow {
				return -4
			} else {
				return -5
			}
		}
		// Up right move
		if isJump {
			return -7
		} else if isOffsetRow {
			return -3
		} else {
			return -4
		}
	}
	if isLeftwardMoveDiretion(direction) {
		if isJump {
			return 7
		} else if isOffsetRow {
			return 4
		} else {
			return 3
		}
	}
	if isJump {
		return 9
	} else if isOffsetRow {
		return 5
	} else {
		return 4
	}
}

func getMoveDirectionByDistance(distance int) MoveDirection {
	if distance == -9 {
		return UpLeft
	}
	if distance == -7 {
		return UpRight
	}
	if distance == 7 {
		return DownLeft
	}
	return DownRight
}

func getJumpedIndex(from int, to int) int {
	if !isJumpMove(from, to) {
		panic("Invalid jump move")
	}

	distance := to - from
	direction := getMoveDirectionByDistance(distance)

	return (from + getMoveAmountForDirection(direction, isOffsetRow(tileIndexToRow(from)), false))
}

func getMoveDestinationInDirection(index int, color PieceColor, tiles []TileState, direction MoveDirection) *int {
	row := tileIndexToRow(index)
	upwards := isUpwardMoveDirection(direction)
	if upwards && row == 0 {
		return nil
	}
	if !upwards && row == 7 {
		return nil
	}

	col := tileIndexToCol(index)
	leftward := isLeftwardMoveDiretion(direction)
	if leftward && col == 0 {
		return nil
	}
	if !leftward && col == 7 {
		return nil
	}

	moveAmount := getMoveAmountForDirection(direction, isOffsetRow(row), false)
	dest := index + moveAmount

	if tiles[dest] == Empty {
		return &dest
	}

	if upwards && row <= 1 {
		return nil
	}
	if !upwards && row >= 6 {
		return nil
	}

	if leftward && col <= 1 {
		return nil
	}
	if !leftward && col >= 6 {
		return nil
	}

	isBlack := color == Black
	isJumpingBlack := isBlackPiece(tiles[dest])

	if isBlack == isJumpingBlack {
		return nil
	}

	moveAmount = getMoveAmountForDirection(direction, isOffsetRow(row), true)
	dest = index + moveAmount

	if tiles[dest] == Empty {
		return &dest
	}

	return nil
}

func getPieceMoveDestinations(index int, color PieceColor, tiles []TileState) []int {
	state := tiles[index]

	if state == Empty {
		return make([]int, 0)
	}

	if color == Black {
		if !isBlackPiece(state) {
			return make([]int, 0)
		}
	} else {
		if isBlackPiece(state) {
			return make([]int, 0)
		}
	}

	moveDirections := make([]MoveDirection, 0)

	if isBlackPiece(state) || isKingPiece(state) {
		moveDirections = append(moveDirections, UpLeft)
		moveDirections = append(moveDirections, UpRight)
	}
	if !isBlackPiece(state) || isKingPiece(state) {
		moveDirections = append(moveDirections, DownLeft)
		moveDirections = append(moveDirections, DownRight)
	}

	moveDestinations := make([]int, 0)

	for _, direction := range moveDirections {
		dest := getMoveDestinationInDirection(index, color, tiles, direction)
		if dest != nil {
			moveDestinations = append(moveDestinations, *dest)
		}
	}

	return moveDestinations
}

func containsJumpMove(index int, moveDestinations []int) bool {
	for _, dest := range moveDestinations {
		if isJumpMove(index, dest) {
			return true
		}
	}
	return false
}

func getMoveDestinations(tiles []TileState, color PieceColor, prevMoveDestIndex *int) [][]int {
	if prevMoveDestIndex != nil {
		prevMoveColor := Red
		if isBlackPiece(tiles[*prevMoveDestIndex]) {
			prevMoveColor = Black
		}

		if PieceColor(prevMoveColor) == color {
			tmpPieceDestinations := getPieceMoveDestinations(*prevMoveDestIndex, color, tiles)
			pieceDestinations := make([]int, 0)

			for _, dest := range tmpPieceDestinations {
				if isJumpMove(*prevMoveDestIndex, dest) {
					pieceDestinations = append(pieceDestinations, dest)
				}
			}

			moveDestinations := make([][]int, len(tiles))
			moveDestinations[*prevMoveDestIndex] = pieceDestinations

			return moveDestinations
		}
	}

	moveDestinations := make([][]int, 0)
	for index := range tiles {
		moveDestinations = append(moveDestinations, getPieceMoveDestinations(index, color, tiles))
	}

	hasJump := false
	for index, moves := range moveDestinations {
		if containsJumpMove(index, moves) {
			hasJump = true
			break
		}
	}

	if hasJump {
		for i := range moveDestinations {
			old := moveDestinations[i]
			moveDestinations[i] = make([]int, 0)

			for _, dest := range old {
				if isJumpMove(i, dest) {
					moveDestinations[i] = append(moveDestinations[i], dest)
				}
			}
		}
	}

	return moveDestinations
}

func hasLegalMoves(tiles []TileState, color PieceColor) bool {
	playerMoveDestinations := getMoveDestinations(tiles, color, nil)

	for _, dests := range playerMoveDestinations {
		if len(dests) > 0 {
			return true
		}
	}

	return false
}

func hasRemainingPieces(tiles []TileState, color PieceColor) bool {
	for _, state := range tiles {
		if color == Black && isBlackPiece(state) {
			return true
		}
		if color == Red && isRedPiece(state) {
			return true
		}
	}
	return false
}

func (game *Game) playMove(gameMove GameMove) bool {
	if !game.isValidMove(gameMove) {
		return false
	}
	game.redWantsDraw = false
	game.blackWantsDraw = false
	newTileStates := make([]TileState, len(game.tileStates))
	copy(newTileStates, game.tileStates)

	// Check state of the moved piece
	newPieceState := newTileStates[gameMove.From]
	isBlackMove := isBlackPiece(newPieceState)

	// Check for a promotion from a standard to crown piece
	if isStandardPiece(newPieceState) {
		destRow := tileIndexToRow(gameMove.To)
		if isBlackMove && destRow == 0 {
			newPieceState = BlackKingPiece
		} else if !isBlackMove && destRow == 7 {
			newPieceState = RedKingPiece
		}
	}

	newTileStates[gameMove.To] = newPieceState
	newTileStates[gameMove.From] = Empty

	currentPlayerColor := Red
	waitingPlayerColor := Black
	if isBlackMove {
		currentPlayerColor = Black
		waitingPlayerColor = Red
	}

	isJump := isJumpMove(gameMove.From, gameMove.To)

	if isJump {
		jumpedIndex := getJumpedIndex(gameMove.From, gameMove.To)
		newTileStates[jumpedIndex] = Empty

		if !hasRemainingPieces(newTileStates, PieceColor(waitingPlayerColor)) {
			game.tileStates = newTileStates
			game.previousMoves = append(game.previousMoves, gameMove)

			game.handleGameEnd()
			return true
		}

		newMoveDestinations := getMoveDestinations(newTileStates, PieceColor(currentPlayerColor), &gameMove.To)

		if containsJumpMove(gameMove.To, newMoveDestinations[gameMove.To]) {
			game.tileStates = newTileStates
			game.previousMoves = append(game.previousMoves, gameMove)

			return true
		}
	}

	if !hasLegalMoves(newTileStates, PieceColor(waitingPlayerColor)) {
		game.tileStates = newTileStates
		game.previousMoves = append(game.previousMoves, gameMove)

		game.handleGameEnd()
		return true
	}

	game.tileStates = newTileStates
	game.turn = PieceColor(waitingPlayerColor)
	game.previousMoves = append(game.previousMoves, gameMove)

	return true
}

func (game *Game) isValidMove(gameMove GameMove) bool {
	color := Red
	if isBlackPiece(game.tileStates[gameMove.From]) {
		color = Black
	}
	var prevMove *int = nil
	if len(game.previousMoves) != 0 {
		prevMove = &game.previousMoves[len(game.previousMoves)-1].To
	}
	allValidMoves := getMoveDestinations(game.tileStates, PieceColor(color), prevMove)

	validMoves := allValidMoves[gameMove.From]

	// Check if the move that was made is in the list of valid moves starting
	// from the From index
	for _, dest := range validMoves {
		if dest == gameMove.To {
			return true
		}
	}

	return false
}

// assuming that the game has ended return the winner as
// a string
func (game *Game) currentWinner() string {
	if game.turn == Red {
		return "red"
	} else {
		return "black"
	}
}

func (game *Game) handleGameEnd() {
	// TODO: send end message to cluster and wait for acks
	winner := game.currentWinner()
	endMessage := GameEndMessage{
		Kind:   "game_end",
		Winner: winner,
	}
	endMessageJson, err := json.Marshal(endMessage)
	if err != nil {
		panic("Marshalling Error")
	}
	game.blackPlayer.send <- endMessageJson
	game.redPlayer.send <- endMessageJson
	var winnerUsername string
	var loserUsername string
	if winner == "red" {
		winnerUsername = game.redPlayer.username
		loserUsername = game.blackPlayer.username
	} else {
		winnerUsername = game.blackPlayer.username
		loserUsername = game.redPlayer.username
	}
	// tell the server the result to update the leaderboard
	gameResult := GameResult{
		GameID: game.gameID,
		Winner: winnerUsername,
		Loser:  loserUsername,
		IsDraw: false,
	}

	// Send the game result to the Matchmaker
	log.Printf("Sending game result to matchmaker: (game=%s, winner=%s, loser=%s)",
		game.gameID, winnerUsername, loserUsername)
	game.resultChan <- gameResult
}

func (game *Game) handleGameDraw() {
	endMessage := GameEndMessage{
		Kind:   "game_end",
		Winner: "draw",
	}
	endMessageJson, err := json.Marshal(endMessage)
	if err != nil {
		panic("Marshalling Error")
	}
	game.blackPlayer.send <- endMessageJson
	game.redPlayer.send <- endMessageJson

	// tell the server the result to update the leaderboard
	gameResult := GameResult{
		GameID: game.gameID,
		Winner: game.blackPlayerUsername,
		Loser:  game.redPlayerUsername,
		IsDraw: true,
	}

	// Send the game result to the Matchmaker
	log.Printf("Sending game result to matchmaker: (game=%s, draw)", game.gameID)
	game.resultChan <- gameResult
}

func (game *Game) messageFromNewGame(playerKind string, opponent string) FoundGame {
	return FoundGame{
		Kind:     "start",
		Side:     playerKind,
		Opponent: opponent,
	}
}

type GameMove struct {
	Kind string `json:"type"`
	From int    `json:"source_index"`
	To   int    `json:"destination_index"`
}

type GameResult struct {
	GameID uuid.UUID `json:"game_id"`
	// Winner username, or nil if it's a draw
	Winner string `json:"winner,omitempty"`
	// Loser username, or nil if it's a draw
	Loser  string `json:"loser,omitempty"`
	IsDraw bool   `json:"is_draw"`
}
