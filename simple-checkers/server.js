const express = require('express');
const http = require('http');
const cors = require("cors");
const WebSocket = require('ws');
const path = require('path');

// Server port
const PORT = 3000;

let waitingPlayerSockets = [];
let activeGames = [];

// Create Express app
const app = express();

// Serve static files from the public directory
app.use(express.static(path.join(__dirname, 'public')));
app.use(cors());

// Create HTTP server
const server = http.createServer(app);

// Create WebSocket server
const webSocketServer = new WebSocket.Server({ server, path: "/ws" });

// Handle WebSocket connections
webSocketServer.on('connection', (webSocket) => {
    console.log('Received connection');

    // Add player to queue
    waitingPlayerSockets.push(webSocket);

    // Check for enough players to start a game
    if (waitingPlayerSockets.length >= 2) {
        // Get two players from the queue
        const player1 = waitingPlayerSockets.shift();
        const player2 = waitingPlayerSockets.shift();

        // Randomly select the black player
        const blackPlayer = Math.random() < 0.5 ? player1 : player2;

        // Initialize game data
        const game = {
            blackPlayer: blackPlayer,
            redPlayer: blackPlayer === player1 ? player2 : player1,
        };
        activeGames.push(game);

        // Notify players to begin
        for (const player of [game.redPlayer, game.blackPlayer]) {
            player.send(
                JSON.stringify({
                    type: 'start',
                    player_color: player === game.redPlayer ? 'red' : 'black',
                }),
            );
        }
    }

    webSocket.on('message', (message) => {
        const data = JSON.parse(message);
        console.log(`Received move: ${data.source} to ${data.destination}`);
        const game = activeGames.find((g) =>
            [g.blackPlayer, g.redPlayer].includes(webSocket),
        );

        if (data.type === 'move' && game && !game.gameOver) {
            // TODO simple demo, real implementation needs to do way more!
            if (game.blackPlayer === webSocket) {
                game.redPlayer.send(
                    JSON.stringify({
                        type: 'move',
                        source: data.source,
                        destination: data.destination,
                    }),
                );
            } else {
                game.blackPlayer.send(
                    JSON.stringify({
                        type: 'move',
                        source: data.source,
                        destination: data.destination,
                    }),
                );
            }
        }
    });

    webSocket.on('close', () => {
        // Remove from the queue if present
        waitingPlayerSockets = waitingPlayerSockets.filter(
            (player) => player !== webSocket,
        );
        console.log("Disconnected.")
    });
    webSocket.on('error', (err) => console.error('WebSocket error:', err));
});

// Start listening for server connections
server.listen(PORT, () => {
    console.log(`Server listening on http://localhost:${PORT}.`);
});

app.get("/api/get_leaderboard", (req, res) => {
    res.json({
        type: "get_leaderboard",
        leaderboard: [
            {
                "name": "Test",
                "wl_ratio": 0.5,
            }
        ]
    })
})
