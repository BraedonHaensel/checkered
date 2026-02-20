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

// Special index to mark unoccupiable tiles (every second tile)
const UNOCCUPIABLE_TILE_INDEX = -1;

// WebSocket instance for server connections
let webSocket;

// Tile occupation states (corresponds to piece positions)
let tileStates;

// Player color
let isBlackPlayer;

// Track if it's the current player's turn
let isYourTurn;

// Track the currently selected piece's tile index
let selectedPieceTileIndex;

/**
 * Gets the TILE_STATE for the player's pieces
 */
function getOwnedPieceState() {
    return isBlackPlayer ? TILE_STATE.BLACK_PIECE : TILE_STATE.RED_PIECE;
}

/**
 * Gets the tile at a given index
 */
function getTileFromIndex(tileIndex) {
    return document.querySelector(`.tile[data-index='${tileIndex}']`);
}

/**
 * Gets the piece at a given index
 */
function getPieceFromIndex(pieceIndex) {
    return document.querySelector(`.piece[data-index='${pieceIndex}']`);
}

/**
 * Returns the row number given a piece's index
 */
function indexToRow(pieceIndex) {
    // 4 occupiable tiles per row
    return Math.floor(pieceIndex / 4);
}

/**
 * Returns the column number given a piece's index
 */
function indexToCol(pieceIndex) {
    // 4 occupiable columns per row, cols spaced 2 tiles apart
    const col = (pieceIndex % 4) * 2;
    // Occupiable columns are offset one tile to the right every second row
    const offset = indexToRow(pieceIndex) % 2 === 0 ? 1 : 0;
    return col + offset;
}

/**
 * Get the possible move destinations for a pice
 */
function getPieceMoveDestinations(pieceTileIndex) {
    // Must be your turn to move a piece
    if (!isYourTurn) return [];

    // Must be your piece colour
    if (tileStates[pieceTileIndex] !== getOwnedPieceState()) return [];

    const possibleMoves = [];
    const canMoveUp = isBlackPlayer; // TODO or crown
    const canMoveDown = !isBlackPlayer; // TODO or crown

    // Check if the piece can move upwards
    if (canMoveUp) {
        // Must be below the top row to move down
        const row = indexToRow(pieceTileIndex);
        if (row > 0) {
            // Check if pieces are offset one column to the right in this row
            const isOffsetRow = row % 2 === 0;
            const col = indexToCol(pieceTileIndex);
            if (col > 0) {
                // Can move up-left
                possibleMoves.push(isOffsetRow ? -4 : -5);
            }
            if (col < 7) {
                // Can move up-right
                possibleMoves.push(isOffsetRow ? -3 : -4);
            }
        }
    }

    // Check if the piece can move downwards
    if (canMoveDown) {
        // Must be above the bottom row to move up
        const row = indexToRow(pieceTileIndex);
        if (row < 7) {
            // Check if pieces are offset one column to the right in this row
            const isOffsetRow = row % 2 === 0;
            const col = indexToCol(pieceTileIndex);
            if (col > 0) {
                // Can move down-left
                possibleMoves.push(isOffsetRow ? 4 : 3);
            }
            if (col < 7) {
                // Can move down-right
                possibleMoves.push(isOffsetRow ? 5 : 4);
            }
        }
    }

    // Verify the destinations are empty
    const moveDestinations = [];
    for (const moveAmount of possibleMoves) {
        const destIndex = pieceTileIndex + moveAmount;
        if (tileStates[destIndex] === TILE_STATE.EMPTY) {
            // Destination is empty, can move piece
            moveDestinations.push(destIndex);
        }
    }

    // Return the remaining moves
    return moveDestinations;
}

/**
 * Checks if the player can currently move the piece at a given index
 */
function canMovePiece(pieceTileIndex) {
    return getPieceMoveDestinations(pieceTileIndex).length > 0;
}

/**
 * Highlights all pieces with available moves
 */
function highlightMovablePieces() {
    for (let i = 0; i < tileStates.length; i++) {
        if (canMovePiece(i)) {
            const piece = getPieceFromIndex(i);
            piece.style.border = '2px solid yellow';
        }
    }
}

/**
 * Remove piece highlights
 */
function removePieceHighlights() {
    for (let i = 0; i < tileStates.length; i++) {
        const piece = getPieceFromIndex(i);
        if (piece) {
            piece.style.border = 'none';
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
 * Add move destination option indicators
 */
function addMoveDestinationOptionIndicators(pieceTileIndex) {
    const moveDestinations = getPieceMoveDestinations(pieceTileIndex);
    for (const moveDestination of moveDestinations) {
        // Get destination tile
        const destTile = getTileFromIndex(moveDestination);

        // Create move indicator dot
        const moveOptionIndicator = document.createElement('div');
        moveOptionIndicator.classList.add('move-option-indicator');

        // Add indicator to tile
        destTile.appendChild(moveOptionIndicator);
    }
}

/**
 * Remove previuos move destination option indicators
 */
function removeMoveDestinationOptionIndicators(pieceTileIndex) {
    const moveDestinations = getPieceMoveDestinations(pieceTileIndex);
    for (const moveDestination of moveDestinations) {
        // Get destination tile
        const destTile = getTileFromIndex(moveDestination);

        // Get the move indicator dot
        const moveOptionIndicator = destTile.querySelector(
            '.move-option-indicator',
        );

        // Remove the move indicator dot
        if (moveOptionIndicator) destTile.removeChild(moveOptionIndicator);
    }
}

/**
 * Move the currently selected piece to the destination tile index
 */
function handleMoveToDestination(destTileIndex) {
    console.log(
        `Moving piece from ${selectedPieceTileIndex} to ${destTileIndex}`,
    );

    // TODO possible issue: animating piece movement when they are bound to tiles
    // Could do no animations. How previous location and highlight last moved piece on opponent view?

    // TODO I should just start doing this in react, not keep prototyping in js
}

/**
 * Handle clicking on a tile
 */
function handleTileClick(e) {
    // Parse the clicked tile's index
    const tile = e.currentTarget;
    const tileIndex = parseInt(tile.dataset.index);

    // Check if this click is to select a piece destination
    const pieceMoveDestinations = getPieceMoveDestinations(
        selectedPieceTileIndex,
    );
    if (pieceMoveDestinations.includes(tileIndex)) {
        // Perform the piece move
        handleMoveToDestination(tileIndex);
        return;
    }

    if (selectedPieceTileIndex !== undefined) {
        // Remove previous move destination option indicators
        removeMoveDestinationOptionIndicators(selectedPieceTileIndex);

        // Reset movable piece highlights
        highlightMovablePieces();
    }

    if (tileIndex === UNOCCUPIABLE_TILE_INDEX || !canMovePiece(tileIndex)) {
        // Selecting a tile without a movable piece clears any previous selection
        selectedPieceTileIndex = undefined;
        return;
    }

    // Save the newly selected piece's tile index
    selectedPieceTileIndex = tileIndex;

    // Highlight the selected piece
    const piece = getPieceFromIndex(selectedPieceTileIndex);
    piece.style.border = '4px solid silver';

    // Add move destination option indicators
    addMoveDestinationOptionIndicators(selectedPieceTileIndex);
}

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
            } else {
                tile.dataset.index = UNOCCUPIABLE_TILE_INDEX;
            }

            // Add tile to board
            gameBoard.appendChild(tile);

            // Add click event listener
            tile.addEventListener('click', handleTileClick);
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
        const tile = getTileFromIndex(i);
        tile.appendChild(piece);
    }
}

// Initialize the checkers board
createBoardTiles();
updateTileStates(getNewBoardTileStates());
