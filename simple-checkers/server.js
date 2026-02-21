const express = require('express');
const http = require('http');
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

// Create HTTP server
const server = http.createServer(app);

// Create WebSocket server
const webSocketServer = new WebSocket.Server({ server });

// Handle WebSocket connections
webSocketServer.on('connection', (webSocket) => {
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
                    playerColor: player === game.redPlayer ? 'red' : 'black',
                }),
            );
        }
    }

    webSocket.on('close', () => {
        // Remove from the queue if present
        waitingPlayerSockets = waitingPlayerSockets.filter(
            (player) => player !== webSocket,
        );
    });
});

// Start listening for server connections
server.listen(PORT, () => {
    console.log(`Server listening on http://localhost:${PORT}.`);
});
