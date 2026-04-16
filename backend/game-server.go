package checkered

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	INGAME                              = "InGame"
	QUEUING                             = "Queuing"
	SPECTATING                          = "Spectating"
	IDLE                                = "Idle"
	writeWait                           = 10 * time.Second
	QUEUE_SIZE                          = 100
	REFRESH_OTHER_GAME_SERVERS_INTERVAL = 20 * time.Second
)

type GameServer struct {
	ID int
	// we map username to client
	clients map[string]*Client
	// games that new clients are in
	games      map[uuid.UUID]*Game
	gamesMu    sync.Mutex
	readyQueue Queue[*Client]

	register     chan *Client
	unregister   chan *Client
	moveReceiver chan GameMove
	gameResults  chan GameResult

	// URL of the Name Server
	nameServerURL string

	// Other Game Servers in the network (does not include itself)
	otherGameServers   []Server
	otherGameServersMu sync.Mutex
}

func InitServer(nameServerURL string) *GameServer {
	gs := GameServer{
		clients:       make(map[string]*Client),
		games:         make(map[uuid.UUID]*Game),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		gameResults:   make(chan GameResult, 10),
		nameServerURL: nameServerURL,
	}
	InitQueue(&gs.readyQueue, QUEUE_SIZE)
	return &gs
}

type TakeOverRequest struct {
	GameID uuid.UUID
}

/*
Handles requests for this server to take ownership of a game
*/
func (gs *GameServer) TakeOverGame(w http.ResponseWriter, r *http.Request) {
	req, err, status := parseJsonRequestData[TakeOverRequest](r)
	if err != nil {
		log.Printf("Failed to parse takeover request %s (%d)", err, status)
		return
	}
	log.Printf("Taking over game: %s", req.GameID)
	// Get updated snapshots to make sure this server has the game
	gs.syncGames()
	log.Printf("Games Synced")
	// Find the game (should have it if any server has it)
	gs.gamesMu.Lock()
	defer gs.gamesMu.Unlock()
	game, exists := gs.games[req.GameID]
	if !exists {
		log.Printf("Unable to locate game %s", req.GameID)
		w.WriteHeader(404)
		w.Write([]byte("Game not found"))
		return
	}
	// Set this server as the owning server
	game.gameServer = gs.ID
	gs.broadcastGameState(game.CreateSnapshot(false))
	w.WriteHeader(200)
}

/*
Handle requests to update game state
*/
func (gs *GameServer) HandleGameStateUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received broadcast")
	gameSnapshot, err, _ := parseJsonRequestData[GameSnapshot](r)
	if err != nil {
		log.Printf("Invalid snapshot received: %s", err)
		w.WriteHeader(400)
		w.Write([]byte("Invalid snapshot"))
		return
	}
	log.Printf("Received snapshot for game %s", gameSnapshot.GameID)
	gs.applyUpdate(gameSnapshot)
	w.WriteHeader(202)
}

/*
Apply an update
Args:

	gameSnapshot (GameSnapshot) The snapshot of the update to apply
*/
func (gs *GameServer) applyUpdate(gameSnapshot GameSnapshot) {
	gameId := uuid.MustParse(gameSnapshot.GameID)
	gs.gamesMu.Lock()
	game, exists := gs.games[gameId]
	gs.gamesMu.Unlock()
	if !exists {
		game = gs.makeGame(
			gameId,
			gameSnapshot.RedPlayerUsername,
			gameSnapshot.BlackPlayerUsername,
		)
		log.Printf("Created game %s as it was previously unknown", gameId)
	}
	if gameSnapshot.Delete {
		delete(gs.games, gameId)
		log.Printf("Removed game %s", gameId)
	} else {
		game.ApplySnapshot(gameSnapshot)
		log.Printf("Received updated for game %s", gameId)
	}
}

type GameSnapshots struct {
	Snapshots []GameSnapshot `json:"snapshots"`
}

/*
Get a snapshot of all the games this server has knowledge of
*/
func (gs *GameServer) GetSnapshots(w http.ResponseWriter, r *http.Request) {
	ownedGames := make([]GameSnapshot, 0)
	for _, game := range gs.games {
		ownedGames = append(ownedGames, game.CreateSnapshot(false))
	}
	ownedGamesBytes, err := json.Marshal(GameSnapshots{Snapshots: ownedGames})
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Write(ownedGamesBytes)
}

/*
Broadcast a snapshot of a game to all over known servers
*/
func (gs *GameServer) broadcastGameState(message GameSnapshot) error {
	// Marshal the message
	log.Printf("Broadcasting snapshot for game: %s", message.GameID)
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	gs.RefreshOtherGameServersList()
	for _, otherGameServer := range gs.otherGameServers {
		url := otherGameServer.URL + "/internal"
		log.Printf("Sending request to: %s", url)
		ack, err := http.Post(url, "application/json", bytes.NewBuffer(messageBytes))
		if err != nil {
			gs.DeregisterGameServer(otherGameServer.ID)
			log.Printf("Error sending update to %s: %s", url, err.Error())
			continue
		}
		ack.Body.Close()
		if ack.StatusCode != 202 {
			return fmt.Errorf(
				"Received improper ack (%d) from server \"%s\"",
				ack.StatusCode,
				url,
			)
		}
	}
	log.Printf("Message Broadcast Successful")

	return nil
}

// Register with the Name Server
func (gs *GameServer) Register(url string) {
	id, err := SendRegistrationRequest(url, gs.nameServerURL+"/register/game-server")
	if err != nil {
		log.Fatal(err)
	}
	gs.ID = id
	log.Println("Registered with ID:", gs.ID)
	log.SetPrefix(fmt.Sprintf("[%d] ", gs.ID))
	// Get current snapshots
	gs.syncGames()
}

/*
Request snapshots from all other servers and apply them
Note: After calling all games will have the most up to date snapshots applied
*/
func (gs *GameServer) syncGames() {
	log.Printf("Syncing Games")
	gs.RefreshOtherGameServersList()
	gs.otherGameServersMu.Lock()
	defer gs.otherGameServersMu.Unlock()
	for _, otherGameServer := range gs.otherGameServers {
		log.Printf("Requesting snapshots from %s", otherGameServer.URL)
		res, err := http.Get(otherGameServer.URL + "/snapshots")
		if err != nil {
			log.Println(err)
			continue
		}
		snapshots, err := ParseJsonResponseData[GameSnapshots](res)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Applying snapshots")
		for _, snapshot := range snapshots.Snapshots {
			log.Printf("\t%s (%d)", snapshot.GameID, snapshot.SnapshotId)
			gameID, err := uuid.Parse(snapshot.GameID)
			if err != nil {
				log.Println(err)
				continue
			}
			game, exists := gs.games[gameID]
			log.Printf("Game %s exists? %t", gameID, exists)
			if !exists || (game.snapshotId < snapshot.SnapshotId) {
				log.Printf("Applying snapshot %d for game %s", snapshot.SnapshotId, gameID)
				gs.applyUpdate(snapshot)
			} else if snapshot.SnapshotId < game.snapshotId {
				log.Printf(
					"Received out of date snapshot %d, expected %d+",
					snapshot.SnapshotId,
					game.snapshotId,
				)
			}
		}
	}
}

// Refreshes and sets the list of known other Game Servers.
func (gs *GameServer) RefreshOtherGameServersList() {
	// Get the list of all Game Servers from the Name Server
	gameServers, err := SendServerListRequest(gs.nameServerURL + "/game-servers")
	if err != nil {
		log.Println(err)
	}

	// Filter itself out of the list
	otherGameServers := []Server{}
	foundInList := false
	for i, gameServer := range gameServers {
		if gameServer.ID == gs.ID {
			otherGameServers = append(gameServers[:i],
				gameServers[i+1:]...)
			foundInList = true
			break
		}
	}

	if !foundInList {
		// Failed to find itself in the Name Server's list, something went wrong
		log.Fatalf("Game Server %d failed to find itself in the Name Server's list of Game Servers", gs.ID)
	}

	// Set the list of known other Game Servers
	gs.otherGameServersMu.Lock()
	gs.otherGameServers = otherGameServers
	gs.otherGameServersMu.Unlock()
	log.Println("Refreshed the list of other Game Servers")
	for _, gs := range gs.otherGameServers {
		log.Printf("    %d: %s", gs.ID, gs.URL)
	}
}

// Start the ticker to periodically refresh the list of other Game Servers.
func (gs *GameServer) StartOtherGameServersRefreshTicker(ctx context.Context) {
	ticker := time.NewTicker(REFRESH_OTHER_GAME_SERVERS_INTERVAL)

	// Initial immediate refresh
	gs.RefreshOtherGameServersList()

	// Periodic refresh routine
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Ticker cancelled, return
				return
			case <-ticker.C:
				// Ticker fired, refresh the list of other game servers
				gs.RefreshOtherGameServersList()
			}
		}
	}()
}

type RegisterMessage struct {
	Kind     string `json:"type"`
	Username string `json:"user"`
}

type ConfirmRegistration struct {
	Kind string `json:"type"`
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{CheckOrigin: checkOrigin}

func ServeWs(gs *GameServer, w http.ResponseWriter, r *http.Request) {
	// upgrade to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// get a message from the client indicating who they are
	var registerMessage RegisterMessage
	err = conn.ReadJSON(&registerMessage)
	if err != nil {
		log.Println(err)
		return
	}
	// create a new client for this request
	client := NewClient(
		registerMessage.Username,
		conn,
		gs.unregister,
		func(client *Client) {
			gs.register <- client
		},
	)
	gs.clients[client.username] = &client

	go client.readThread()
	go client.writeThread()

	// Send confirmation of registration to client
	confirmation_msg := ConfirmRegistration{Kind: "registered"}
	confirmation, err := json.Marshal(confirmation_msg)
	client.send <- confirmation

	// let the server know and handle the new client
	// server.register <- &client
	log.Printf("Client \"%s\" connected", client.username)
	// Find the game the user is attached to
	var pendingGame *Game = nil
	var opponent *Client = nil
	gs.gamesMu.Lock()
	for _, pg := range gs.games {
		if pg.blackPlayerUsername == client.username {
			pendingGame = pg
			pendingGame.mu.Lock()
			defer pendingGame.mu.Unlock()
			pendingGame.blackPlayer = &client
			opponent = pg.redPlayer
			break
		}
		if pg.redPlayerUsername == client.username {
			pendingGame = pg
			pendingGame.mu.Lock()
			defer pendingGame.mu.Unlock()
			pendingGame.redPlayer = &client
			opponent = pendingGame.blackPlayer
			break
		}
	}
	gs.gamesMu.Unlock()
	if pendingGame == nil {
		log.Printf("Player %s is not in a pending game\n", client.username)
		return
	}
	log.Printf("Game Identified %s\n", pendingGame.gameID)

	// If the other player hasn't registered start a timeout
	if opponent == nil {
		//Start forfeit timeout
		log.Println("Waiting for opponent")

		matchID := pendingGame.gameID

		time.AfterFunc(5*time.Second, func() {
			gs.gamesMu.Lock()
			defer gs.gamesMu.Unlock()
			game, exists := gs.games[matchID]
			if exists {
				opponentUsername := game.blackPlayerUsername
				opponentClient := game.blackPlayer
				color := "red"
				if opponentUsername == client.username {
					opponentUsername = game.redPlayerUsername
					opponentClient = game.redPlayer
					color = "black"
				}
				if opponentClient == nil {
					log.Println("Failed to find opponent in time. Ending match...")

					gs.HandlePlayerDisconnect(matchID, client, color, opponentUsername)
				}
				return
			}
		})

		return
	}
	// If the other player has registered start the game
	log.Printf("Starting game: %s", pendingGame.gameID)
	gs.StartGame(pendingGame.gameID)
}

// client is the client that is still connected
func (gs *GameServer) HandlePlayerDisconnect(matchID uuid.UUID, client Client, clientColor string, opponentUsername string) {
	gameResult := GameResult{
		GameID: matchID,
		Winner: client.username,
		Loser:  opponentUsername,
		IsDraw: false,
	}

	gameEndMessage := GameEndMessage{
		Kind:   "game_end",
		Winner: clientColor,
	}

	marshalled, err := json.Marshal(gameEndMessage)

	if err != nil {
		log.Println("Failed to marshal data for game end message to client")
	}
	defer (func() {
		log.Println("Informing client game has ended prematurely.")
		client.send <- marshalled
	})()

	gs.informMatchmakerGameEnded(gameResult)
}

func (gs *GameServer) makeGame(gameID uuid.UUID, redPlayer string, blackPlayer string) *Game {
	pendingGame := Game{
		gameID:              gameID,
		gameServer:          gs.ID,
		redPlayer:           nil,
		blackPlayer:         nil,
		redPlayerUsername:   redPlayer,
		blackPlayerUsername: blackPlayer,
		tileStates:          generateInitialTileStates(),
		turn:                Red,
		previousMoves:       make([]GameMove, 0),
		resultChan:          gs.gameResults,
		mu:                  sync.Mutex{},
		snapshotId:          0,
		updateCallback: func(g *Game) {
			gs.broadcastGameState(g.CreateSnapshot(false))
		},
	}
	gs.gamesMu.Lock()
	gs.games[gameID] = &pendingGame
	gs.gamesMu.Unlock()
	return &pendingGame
}

func (gs *GameServer) CreateGame(w http.ResponseWriter, r *http.Request) {
	match, err, status := parseJsonRequestData[Match](r)
	if err != nil {
		fmt.Printf("Failed to parse game creation request %s (%d)", err, status)
		return
	}
	uuid := match.MatchID
	pendingGame := gs.makeGame(uuid, match.RedPlayer, match.BlackPlayer)
	err = gs.broadcastGameState(pendingGame.CreateSnapshot(false))
	if err != nil {
		log.Printf("Failed to broadcast game creation: %s", err)
	}
	log.Printf("Created pending game: %s", uuid)
	w.WriteHeader(202)

}

func (gs *GameServer) StartGame(gameID uuid.UUID) {
	gs.gamesMu.Lock()
	defer gs.gamesMu.Unlock()
	game := gs.games[gameID]
	if game.gameServer != gs.ID {
		panic("Attempted to start game by non-owning server")
	}
	redClient := game.redPlayer
	blackClient := game.blackPlayer
	if redClient == nil {
		panic("Attempted to start game with nil red client")
	}
	if blackClient == nil {
		panic("Attempted to start game with nil black client")
	}
	// tell both servers about the new game
	redMessage := game.messageFromNewGame("red", game.blackPlayerUsername)
	blackMessage := game.messageFromNewGame("black", game.redPlayerUsername)
	redBytes, err := json.Marshal(redMessage)
	if err != nil {
		log.Println(err)
		return
	}
	blackBytes, err := json.Marshal(blackMessage)
	if err != nil {
		log.Println(err)
		return
	}
	gs.clients[game.blackPlayer.username].currentGame = game
	gs.clients[game.redPlayer.username].currentGame = game
	// send the message that they have found a game to both players
	gs.clients[game.blackPlayer.username].send <- blackBytes
	gs.clients[game.redPlayer.username].send <- redBytes

	initialState := GameStateUpdate{
		Kind:          "update_state",
		TileStates:    game.tileStates,
		Turn:          "red",
		PreviousMoves: game.previousMoves,
	}

	if game.turn == Black {
		initialState.Turn = "black"
	}
	initialStateBytes, err := json.Marshal(initialState)
	if err != nil {
		log.Println(err)
		return
	}
	gs.clients[game.blackPlayer.username].send <- initialStateBytes
	gs.clients[game.redPlayer.username].send <- initialStateBytes
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (gs *GameServer) ServerLoop() {
	for {
		select {
		case <-gs.register:
		case client := <-gs.unregister:
			log.Printf("Deregistered %s", client.username)
		case gameResult := <-gs.gameResults:
			log.Printf(
				"Handling game result (game=%s)",
				gameResult.GameID,
			)
			gs.gamesMu.Lock()
			game, exists := gs.games[gameResult.GameID]

			if exists {
				gs.broadcastGameState(game.CreateSnapshot(true))
				gs.informMatchmakerGameEnded(gameResult)

				delete(gs.games, gameResult.GameID)
			} else {
				log.Printf("Game %s does not exist\n", gameResult.GameID)
			}
			gs.gamesMu.Unlock()
		}
	}
}

func (gs *GameServer) informMatchmakerGameEnded(gameResult GameResult) {

	matchmakingServers, err := SendServerListRequest(
		gs.nameServerURL + "/matchmakers",
	)
	if err != nil {
		log.Printf("Failed to fetch game servers: %s", err)
		return
	}
	if len(matchmakingServers) == 0 {
		log.Println("No match making servers available")
		return
	}
	matchmakingServer := matchmakingServers[0]
	log.Printf(
		"Selected match making server: %s (ID: %d)",
		matchmakingServer.URL,
		matchmakingServer.ID,
	)

	gameResultBytes, err := json.Marshal(gameResult)
	if err != nil {
		log.Printf("Failed to marshal result: %s", err)
		return
	}
	res, err := http.Post(
		matchmakingServer.URL+"/match/updateleaderboard",
		"application/json",
		bytes.NewBuffer(gameResultBytes),
	)
	if err != nil {
		log.Printf("Failed to send game results to match making server: %s", err)
		return
	}
	log.Printf("Leaderboard updated!")
	defer res.Body.Close()
	endMatchRequest := EndMatchRequest{MatchID: gameResult.GameID}
	endMatchRequestBytes, err := json.Marshal(endMatchRequest)
	if err != nil {
		log.Printf("Failed to marshal end game request: %s", err)
		return
	}
	res, err = http.Post(
		matchmakingServer.URL+"/match/end",
		"application/json",
		bytes.NewBuffer(endMatchRequestBytes),
	)
	if err != nil {
		log.Printf("Failed to send end game %s", err)
	}
}

// Deregisters a Game Server from the Name Server.
func (gs *GameServer) DeregisterGameServer(gameServerID int) {
	// Create the deregistration request
	log.Println("Deregistering Game Server", gameServerID)
	body := fmt.Appendf(nil, `{"id": %d}`, gameServerID)
	res, err := http.Post(
		gs.nameServerURL+"/deregister/game-server",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Printf("Failed to deregister Game Server %d: %v", gameServerID, err)
	}
	defer res.Body.Close()
}
