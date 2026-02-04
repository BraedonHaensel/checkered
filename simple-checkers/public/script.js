const gameBoard = document.getElementById('game-board');
const searchButton = document.getElementById('search-button');
const statusMessage = document.getElementById('status-message');

// Server WebSocket URL
const SERVER_WEBSOCKET_URL = 'ws://localhost:3000';

// Game board dimensions
const BOARD_ROWS = 8;
const BOARD_COLS = 8;

// Tile occupation states
const TILE_STATE = {
    EMPTY: 0,
    RED_PIECE: 1,
    BLACK_PIECE: 2,
};

// WebSocket instance for server connections
let webSocket;

// Tile occupation states (corresponds to piece positions)
let tileStates;

// Player color
let isBlackPlayer;

// Track if it's the current player's turn
let isYourTurn;

/**
 * Highlights all pieces with available moves
 */
function highlightMovablePieces() {
    // TODO only apply to movable pieces
    console.log('running');
    const ownedPieceState = isBlackPlayer
        ? TILE_STATE.BLACK_PIECE
        : TILE_STATE.RED_PIECE;

    for (let i = 0; i < tileStates.length; i++) {
        if (tileStates[i] === ownedPieceState) {
            const piece = document.querySelector(`.piece[data-index='${i}']`);
            piece.style.border = '2px solid yellow';
        }
    }
}

/**
 * Set if it's the current player's turn
 */
function setCurrentTurn(newIsYourTurn) {
    isYourTurn = newIsYourTurn;

    // Set turn status text
    statusMessage.textContent = isYourTurn
        ? 'Your turn!'
        : "Opponent's turn...";

    if (isYourTurn) {
        highlightMovablePieces();
    }
}

// Search for opponent button
searchButton.addEventListener('click', () => {
    // Open a new WebSocket connection
    if (webSocket) webSocket.close();
    webSocket = new WebSocket(SERVER_WEBSOCKET_URL);

    // Hide search button
    searchButton.style.display = 'none';

    // Set searching message
    statusMessage.textContent = 'Searching for an opponent...';

    // Handle socket messages
    webSocket.onmessage = (event) => {
        const data = JSON.parse(event.data);

        // Handle game start events
        if (data.type === 'start') {
            // Parse player color
            isBlackPlayer = data.playerColor === 'black';

            // Set board rotation for player's perspective
            gameBoard.style.transform = `rotate(${isBlackPlayer ? '0' : '180'}deg)`;

            // Set current player's turn
            setCurrentTurn(isBlackPlayer);
        }
    };
});

/**
 * Create the checkers board tiles.
 */
function createBoardTiles() {
    // Reset and create empty checkers board
    gameBoard.innerHTML = '';

    // Index to assign to the next occupiable tile
    let tileIndex = 0;

    for (let row = 0; row < BOARD_ROWS; row++) {
        for (let col = 0; col < BOARD_COLS; col++) {
            // Create tile
            const tile = document.createElement('div');
            tile.classList.add('tile');

            // Calculate if this is a dark or light tile
            const isDarkTile = (row + col) % 2 === 1;

            // Add tile color
            tile.style.backgroundColor = isDarkTile ? 'sienna' : 'bisque';

            // Only dark tiles are occupiable
            if (isDarkTile) {
                tile.dataset.index = tileIndex;
                tileIndex++;
            }

            // Add tile to board
            gameBoard.appendChild(tile);
        }
    }
}

/**
 * Get the tile states for a starting checkers board.
 */
function getNewBoardTileStates() {
    // Only 32 tiles are occupiable in checkers. The first 12 start with black pieces, the middle 8
    // are empty, and the last 12 start with red pieces
    const board = [
        ...Array(12).fill(TILE_STATE.RED_PIECE),
        ...Array(8).fill(TILE_STATE.EMPTY),
        ...Array(12).fill(TILE_STATE.BLACK_PIECE),
    ];
    return board;
}

/**
 * Updates the occupation states of each tile (corresponds to piece positions).
 */
function updateTileStates(newTileStates) {
    tileStates = newTileStates;

    // Process each tile's new state
    for (let i = 0; i < newTileStates.length; i++) {
        const tileState = newTileStates[i];

        // Handle empty tiles
        if (tileState === TILE_STATE.EMPTY) {
            continue;
        }

        // Create piece
        const piece = document.createElement('div');
        piece.classList.add('piece');
        piece.dataset.index = i;

        // Add piece color
        piece.style.backgroundColor =
            tileState === TILE_STATE.RED_PIECE ? 'red' : 'black';

        // Add piece to tile
        const tile = document.querySelector(`.tile[data-index='${i}']`);
        tile.appendChild(piece);
    }
}

// Initialize the checkers board
createBoardTiles();
updateTileStates(getNewBoardTileStates());
