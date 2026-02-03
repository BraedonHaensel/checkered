const express = require('express');
const http = require('http');
const WebSocket = require('ws');
const path = require('path');

// Server port
const PORT = 3000;

let waitingPlayers = [];
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
    waitingPlayers.push(webSocket);
});

// Start listening for server connections
server.listen(PORT, () => {
    console.log(`Server listening on http://localhost:${PORT}.`);
});
