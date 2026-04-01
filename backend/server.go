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
	gamesLock  sync.Mutex
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
	server := GameServer{
		clients:       make(map[string]*Client),
		games:         make(map[uuid.UUID]*Game),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		gameResults:   make(chan GameResult, 10),
		nameServerURL: nameServerURL,
	}
	InitQueue(&server.readyQueue, QUEUE_SIZE)
	return &server
}

type TakeOverRequest struct {
	GameID uuid.UUID
}

func (server *GameServer) TakeOverGame(w http.ResponseWriter, r *http.Request) {
	req, err, status := parseJsonRequestData[TakeOverRequest](r)
	if err != nil {
		log.Printf("Failed to parse takeover request %s (%d)", err, status)
		return
	}
	log.Printf("Taking over game: %s", req.GameID)
	server.gamesLock.Lock()
	defer server.gamesLock.Unlock()
	game, exists := server.games[req.GameID]
	if !exists {
		w.WriteHeader(404)
		w.Write([]byte("Game not found"))
		return
	}
	game.gameServer = server.ID
	server.broadcastGameState(game.CreateSnapshot(false))
	w.WriteHeader(200)
}

func (server *GameServer) HandleGameStateUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received broadcast")
	gameSnapshot, err, _ := parseJsonRequestData[GameSnapshot](r)
	if err != nil {
		log.Printf("Invalid snapshot received: %s", err)
		w.WriteHeader(400)
		w.Write([]byte("Invalid snapshot"))
		return
	}
	log.Printf("Received snapshot for game %s", gameSnapshot.GameID)
	server.applyUpdate(gameSnapshot)
	w.WriteHeader(202)
}

func (server *GameServer) applyUpdate(gameSnapshot GameSnapshot) {

	server.gamesLock.Lock()
	defer server.gamesLock.Unlock()
	gameId := uuid.MustParse(gameSnapshot.GameID)
	game, exists := server.games[gameId]
	if !exists {
		gameRoom := Game{
			resultChan: server.gameResults,
			mu:         sync.Mutex{},
		}
		game = &gameRoom
		server.games[gameId] = game
		log.Printf("Created game %s", gameId)
	}
	if gameSnapshot.Delete {
		delete(server.games, gameId)
		log.Printf("Removed game %s", gameId)
	} else {
		game.ApplySnapshot(gameSnapshot)
		log.Printf("Updated game %s", gameId)
	}
}

type GameSnapshots struct {
	Snapshots []GameSnapshot
}

func (server *GameServer) GetOwnedSnapshots(w http.ResponseWriter, r *http.Request) {
	ownedGames := make([]GameSnapshot, 0)
	for _, game := range server.games {
		if game.gameServer == server.ID {
			ownedGames = append(ownedGames, game.CreateSnapshot(false))
		}
	}
	ownedGamesBytes, err := json.Marshal(GameSnapshots{Snapshots: ownedGames})
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Write(ownedGamesBytes)
}

func (server *GameServer) broadcastGameState(message GameSnapshot) error {
	// Marshal the message
	log.Printf("Broadcasting snapshot for game: %s", message.GameID)
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	server.RefreshOtherGameServersList()
	for _, gs := range server.otherGameServers {
		url := gs.URL + "/internal"
		log.Printf("Sending request to: %s", url)
		ack, err := http.Post(url, "application/json", bytes.NewBuffer(messageBytes))
		if err != nil {
			// TODO: If no response update the nameserver
			return err
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
func (server *GameServer) Register(url string) {
	id, err := SendRegistrationRequest(url, server.nameServerURL+"/register/game-server")
	if err != nil {
		log.Fatal(err)
	}
	server.ID = id
	log.Println("Registered with ID:", server.ID)
	log.SetPrefix(fmt.Sprintf("[%d] ", server.ID))
	// Get current snapshots
	server.RefreshOtherGameServersList()
	for i := range server.otherGameServers {
		otherServer := server.otherGameServers[i]
		res, err := http.Get(otherServer.URL + "/snapshots")
		if err != nil {
			log.Println(err)
			continue
		}
		snapshots, err := ParseJsonResponseData[GameSnapshots](res)
		if err != nil {
			log.Println(err)
			continue
		}
		for j := range snapshots.Snapshots {
			server.applyUpdate(snapshots.Snapshots[j])
		}
	}
}

// Refreshes and sets the list of known other Game Servers.
func (server *GameServer) RefreshOtherGameServersList() {
	// Get the list of all Game Servers from the Name Server
	gameServers, err := SendServerListRequest(server.nameServerURL + "/game-servers")
	if err != nil {
		log.Println(err)
	}

	// Filter itself out of the list
	otherGameServers := []Server{}
	foundInList := false
	for i, gameServer := range gameServers {
		if gameServer.ID == server.ID {
			otherGameServers = append(gameServers[:i],
				gameServers[i+1:]...)
			foundInList = true
			break
		}
	}

	if !foundInList {
		// Failed to find itself in the Name Server's list, something went wrong
		log.Fatalf("Game Server %d failed to find itself in the Name Server's list of Game Servers", server.ID)
	}

	// Set the list of known other Game Servers
	server.otherGameServersMu.Lock()
	server.otherGameServers = otherGameServers
	server.otherGameServersMu.Unlock()
	log.Println("Refreshed the list of other Game Servers")
	for _, gs := range server.otherGameServers {
		log.Printf("    %d: %s", gs.ID, gs.URL)
	}
}

// Start the ticker to periodically refresh the list of other Game Servers.
func (server *GameServer) StartOtherGameServersRefreshTicker(ctx context.Context) {
	ticker := time.NewTicker(REFRESH_OTHER_GAME_SERVERS_INTERVAL)

	// Initial immediate refresh
	server.RefreshOtherGameServersList()

	// Periodic refresh routine
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Ticker cancelled, return
				return
			case <-ticker.C:
				// Ticker fired, refresh the list of other game servers
				server.RefreshOtherGameServersList()
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

func ServeWs(server *GameServer, w http.ResponseWriter, r *http.Request) {
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
		server.unregister,
		func(client *Client) {
			server.register <- client
		},
		func(g *Game) {
			server.broadcastGameState(g.CreateSnapshot(false))
		},
	)
	server.clients[client.username] = &client

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
	server.gamesLock.Lock()
	for _, pg := range server.games {
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
	server.gamesLock.Unlock()
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
			server.gamesLock.Lock()
			defer server.gamesLock.Unlock()
			game, exists := server.games[matchID]
			if exists {
				opponentUsername := game.blackPlayerUsername
				opponentClient := game.blackPlayer
				color := "red"
				if opponentUsername == client.username {
					opponentUsername = game.redPlayerUsername
					opponentClient = game.blackPlayer
					color = "black"
				}
				if opponentClient == nil {
					log.Println("Failed to find opponent in time. Ending match...")

					server.HandlePlayerDisconnect(matchID, client, color, opponentUsername)
				}
				return
			}
		})

		return
	}
	// If the other player has registered start the game
	log.Printf("Starting game: %s", pendingGame.gameID)
	server.StartGame(pendingGame.gameID)
}

// client is the client that is still connected
func (server *GameServer) HandlePlayerDisconnect(matchID uuid.UUID, client Client, clientColor string, opponentUsername string) {
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

	server.informMatchmakerGameEnded(gameResult)
}

func (server *GameServer) CreateGame(w http.ResponseWriter, r *http.Request) {
	match, err, status := parseJsonRequestData[Match](r)
	if err != nil {
		fmt.Printf("Failed to parse game creation request %s (%d)", err, status)
		return
	}
	uuid := match.MatchID
	pendingGame := Game{
		gameID:              uuid,
		gameServer:          server.ID,
		redPlayer:           nil,
		blackPlayer:         nil,
		redPlayerUsername:   match.RedPlayer,
		blackPlayerUsername: match.BlackPlayer,
		tileStates:          generateInitialTileStates(),
		turn:                Red,
		previousMoves:       make([]GameMove, 0),
		resultChan:          server.gameResults,
		mu:                  sync.Mutex{},
	}
	server.gamesLock.Lock()
	defer server.gamesLock.Unlock()
	server.games[uuid] = &pendingGame
	err = server.broadcastGameState(pendingGame.CreateSnapshot(false))
	if err != nil {
		log.Printf("Failed to broadcast game creation: %s", err)
	}
	log.Printf("Created pending game: %s", uuid)

}

func (server *GameServer) StartGame(gameID uuid.UUID) {
	server.gamesLock.Lock()
	defer server.gamesLock.Unlock()
	game := server.games[gameID]
	if game.gameServer != server.ID {
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
	server.clients[game.blackPlayer.username].currentGame = game
	server.clients[game.redPlayer.username].currentGame = game
	// send the message that they have found a game to both players
	server.clients[game.blackPlayer.username].send <- blackBytes
	server.clients[game.redPlayer.username].send <- redBytes

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
	server.clients[game.blackPlayer.username].send <- initialStateBytes
	server.clients[game.redPlayer.username].send <- initialStateBytes
}

// main idea of this loop is to register new active users and put them
// into the server struct
func (server *GameServer) ServerLoop() {
	for {
		select {
		case <-server.register:

		case <-server.unregister:
			log.Println("Attempted unregister")
			// TODO: remove client from game rooms
		case gameResult := <-server.gameResults:
			log.Printf(
				"Handling game result (game=%s)",
				gameResult.GameID,
			)
			server.gamesLock.Lock()
			game, exists := server.games[gameResult.GameID]

			if exists {
				server.broadcastGameState(game.CreateSnapshot(true))
				server.informMatchmakerGameEnded(gameResult)

				delete(server.games, gameResult.GameID)
			} else {
				log.Printf("Game %s does not exist\n", gameResult.GameID)
			}
			server.gamesLock.Unlock()
		}
	}
}

func (server *GameServer) informMatchmakerGameEnded(gameResult GameResult) {

	matchmakingServers, err := SendServerListRequest(
		server.nameServerURL + "/matchmakers",
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
