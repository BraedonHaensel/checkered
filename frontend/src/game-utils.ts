import { MoveDirection, PlayerColor, TileState } from './enums'

/**
 * Get the starting tile states for each of the 32 playable dark tiles in a
 * checkers board.
 *
 * @returns List of TileState values for a new checkers board.
 */
export const getNewBoardTileStates = (): TileState[] => {
    // Only 32 tiles are playable in checkers. The first 12 start with black
    // pieces, the middle 8 are empty, and the last 12 start with red pieces
    const tileStates = [
        ...Array(12).fill(TileState.RED_STANDARD_PIECE),
        ...Array(8).fill(TileState.EMPTY),
        ...Array(12).fill(TileState.BLACK_STANDARD_PIECE),
    ]
    return tileStates
}

/**
 * Checks if the given tileState is for a black piece (standard or king).
 *
 * @param tileState TileState of the piece to check.
 * @returns True if the tileState is for a black piece, otherwise false.
 */
export const isBlackPiece = (tileState: TileState): boolean => {
    return [
        TileState.BLACK_STANDARD_PIECE,
        TileState.BLACK_KING_PIECE,
    ].includes(tileState)
}

/**
 * Checks if the given tileState is for a red piece (standard or king).
 *
 * @param tileState TileState of the piece to check.
 * @returns True if the tileState is for a red piece, otherwise false.
 */
export const isRedPiece = (tileState: TileState): boolean => {
    return [
        TileState.BLACK_STANDARD_PIECE,
        TileState.BLACK_KING_PIECE,
    ].includes(tileState)
}

/**
 * Checks if the given tileState is for a standard piece.
 *
 * @param tileState TileState of the piece to check.
 * @returns True if the tileState is for a standard piece, otherwise false.
 */
export const isStandardPiece = (tileState: TileState): boolean => {
    return [
        TileState.BLACK_STANDARD_PIECE,
        TileState.RED_STANDARD_PIECE,
    ].includes(tileState)
}

/**
 * Checks if the given tileState is for a king piece.
 *
 * @param tileState TileState of the piece to check.
 * @returns True if the tileState is for a king piece, otherwise false.
 */
export const isKingPiece = (tileState: TileState): boolean => {
    return [TileState.BLACK_KING_PIECE, TileState.RED_KING_PIECE].includes(
        tileState
    )
}

/**
 * Converts a tile index to a row index.
 *
 * @param tileIndex Tile index to convert.
 * @returns The row index of the given tileIndex.
 */
export const tileIndexToRow = (tileIndex: number): number => {
    // There are 4 playable tile indices per row of a checkers board
    return Math.floor(tileIndex / 4)
}

/**
 * Checks if the row index is for an offset row. An offset row is one where all
 * the dark tiles are shifted right by one tile.
 *
 * @param row Row index to check.
 * @returns True if the row index is for an offset row.
 */
export const isOffsetRow = (row: number): boolean => {
    // The first row is an offset row, as well as every second row after that.
    return row % 2 === 0
}

/**
 * Converts a tile index to a column index.
 *
 * @param tileIndex Tile index to convert.
 * @returns The column index of the given tileIndex.
 */
export const tileIndexToCol = (tileIndex: number): number => {
    // There are 4 playable tiles per row of a checkers board. Each of these
    // tiles is spaced 2 columns apart.
    const col = (tileIndex % 4) * 2
    // Add 1 to the column index if the tile is in an offset row.
    const offset = isOffsetRow(tileIndexToRow(tileIndex)) ? 1 : 0
    return col + offset
}

/**
 * Checks if a move direction is upward, otherwise it is downward.
 *
 * @param moveDirection Move direction to check.
 * @returns True if the move direction is upward, otherwise false (downward).
 */
export const isUpwardMoveDirection = (
    moveDirection: MoveDirection
): boolean => {
    // An upwards move is either diagonally up left or up right
    return [MoveDirection.UP_LEFT, MoveDirection.UP_RIGHT].includes(
        moveDirection
    )
}

/**
 * Checks if a move direction is leftward, otherwise it is rightward.
 *
 * @param moveDirection Move direction to check.
 * @returns True if the move direction is leftward, otherwise false (rightward).
 */
export const isLeftwardMoveDirection = (
    moveDirection: MoveDirection
): boolean => {
    // A leftward move is either diagonally up left or down left
    return [MoveDirection.UP_LEFT, MoveDirection.DOWN_LEFT].includes(
        moveDirection
    )
}

/**
 * Checks if a move from the sourceIndex to destIndex is a jump move.
 *
 * @param sourceIndex Source tile index of the move.
 * @param destIndex Destination tile index of the move.
 * @returns True if the move is a jump move, otherwise false.
 */
export const isJumpMove = (sourceIndex: number, destIndex: number): boolean => {
    // Only jump moves can change a tile index by 7 or more
    return Math.abs(destIndex - sourceIndex) >= 7
}

/**
 * Gets the destination tile index for a move.
 *
 * @param sourceIndex Source tile index of the move.
 * @param moveDirection Move direction to check.
 * @param isJump Whether the move is a jump move.
 * @returns The destinations tile index for the move.
 */
const getDestIndexForMove = (
    sourceIndex: number,
    moveDirection: MoveDirection,
    isJump: boolean = false
): number => {
    // Check if the piece is on an offset row
    const isOffset = isOffsetRow(tileIndexToRow(sourceIndex))
    if (isUpwardMoveDirection(moveDirection)) {
        if (isLeftwardMoveDirection(moveDirection)) {
            // Up left move
            const change = isJump ? -9 : isOffset ? -4 : -5
            return sourceIndex + change
        }
        // Up right move
        const change = isJump ? -7 : isOffset ? -3 : -4
        return sourceIndex + change
    }
    if (isLeftwardMoveDirection(moveDirection)) {
        // Down left move
        const change = isJump ? 7 : isOffset ? 4 : 3
        return sourceIndex + change
    }
    // Down right move
    const change = isJump ? 9 : isOffset ? 5 : 4
    return sourceIndex + change
}

/**
 * Get the jumped tile index for a jump move from the sourceIndex to destIndex.
 * @param sourceIndex Source tile index of the move.
 * @param destIndex Destination tile index of the move.
 * @returns The jumped tile index of the jump move.
 */
export const getJumpedTileIndex = (
    sourceIndex: number,
    destIndex: number
): number => {
    if (!isJumpMove(sourceIndex, destIndex))
        throw new Error(
            `Moving from ${sourceIndex} to ${destIndex} is not a jump move!`
        )

    // Get the move direction of the jump
    const moveAmount = destIndex - sourceIndex
    const moveDirection =
        moveAmount === -9
            ? MoveDirection.UP_LEFT
            : moveAmount === -7
              ? MoveDirection.UP_RIGHT
              : moveAmount === 7
                ? MoveDirection.DOWN_LEFT
                : MoveDirection.DOWN_RIGHT

    // Return the index of the jumped tile (i.e., the destination tile index of
    // a normal move in this direction)
    return getDestIndexForMove(sourceIndex, moveDirection)
}

/**
 * Gets the destination tile index for a move in the given direction, or
 * undefined if there are no valid moves in the given direction.
 *
 * @param sourceIndex Source tile index to start the move from.
 * @param tileStates List of TileState values for the game board.
 * @param moveDirection Move direction to check.
 * @returns Destination tile index for a move in the given direction, or
 * undefined if there are no valid moves in the given direction.
 */
//
const getMoveDestinationInDirection = (
    sourceIndex: number,
    tileStates: TileState[],
    moveDirection: MoveDirection
): number | undefined => {
    // Check row boundaries
    const row = tileIndexToRow(sourceIndex)
    const isUpward = isUpwardMoveDirection(moveDirection)
    if (isUpward && row === 0) return undefined // Can't move upward from the top row
    if (!isUpward && row === 7) return undefined // Can't move downward from the bottom row

    // Check column boundaries
    const col = tileIndexToCol(sourceIndex)
    const isLeftward = isLeftwardMoveDirection(moveDirection)
    if (isLeftward && col === 0) return undefined // Can't move left from the leftmost column
    if (!isLeftward && col === 7) return undefined // Can't move right from the rightmost column

    // Get the destination tile index for a normal move in this direction
    let destIndex = getDestIndexForMove(sourceIndex, moveDirection)

    // Check if the destination tile is empty
    if (tileStates[destIndex] === TileState.EMPTY) {
        // Move is valid, return the destination tile index
        return destIndex
    }
    // Tile is populated, check if a jump move can be made over the piece

    // Check row boundaries again for jump moves
    if (isUpward && row <= 1) return undefined // Can't jump upward from the top 2 rows
    if (!isUpward && row >= 6) return undefined // Can't jump downward from the botttom 2 rows

    // Check column boundaries again for jump moves
    if (isLeftward && col <= 1) return undefined // Can't move left from the leftmost 2 columns
    if (!isLeftward && col >= 6) return undefined // Can't move right from the rightmost 2 columns

    // Check the piece colors involved in the jump move
    const isMovingBlackPiece = isBlackPiece(tileStates[sourceIndex])
    const isJumpingBlackPiece = isBlackPiece(tileStates[destIndex])

    // Can't jump your own piece
    if (isMovingBlackPiece && isJumpingBlackPiece) return undefined
    if (!isMovingBlackPiece && !isJumpingBlackPiece) return undefined

    // Get the destination tile index for a jump move in this direction
    destIndex = getDestIndexForMove(sourceIndex, moveDirection, true)

    // Check if the destination tile is empty
    if (tileStates[destIndex] === TileState.EMPTY) return destIndex

    // No possible moves
    return undefined
}

/**
 * Gets a list of the possible move destinations for the piece at the source
 * tile index.
 *
 * @param sourceIndex Source tile index to start the moves from.
 * @param playerColor Color of the player making the move.
 * @param tileStates List of TileState values for the game board.
 * @returns List of possible move destinations for the piece at the source tile
 * index.
 */
export const getPieceMoveDestinations = (
    sourceIndex: number,
    playerColor: PlayerColor,
    tileStates: TileState[]
): number[] => {
    // Get the state of the source tile
    const tileState = tileStates[sourceIndex]

    // Source tile must not be empty
    if (tileState === TileState.EMPTY) return []

    // Player must own the piece
    if (playerColor === PlayerColor.BLACK) {
        if (!isBlackPiece(tileState)) return []
    } else {
        if (isBlackPiece(tileState)) return []
    }

    // Get the move directions to try
    const moveDirections: MoveDirection[] = []
    if (isBlackPiece(tileState) || isKingPiece(tileState)) {
        // Try moving upwards
        moveDirections.push(MoveDirection.UP_LEFT, MoveDirection.UP_RIGHT)
    }
    if (!isBlackPiece(tileState) || isKingPiece(tileState)) {
        // Try moving downwards
        moveDirections.push(MoveDirection.DOWN_LEFT, MoveDirection.DOWN_RIGHT)
    }

    // Get the valid move destinations for each direction
    const moveDestinations = []
    for (const moveDirection of moveDirections) {
        const destIndex = getMoveDestinationInDirection(
            sourceIndex,
            tileStates,
            moveDirection
        )
        if (destIndex !== undefined) moveDestinations.push(destIndex)
    }
    return moveDestinations
}

/**
 * Checks if the list of move destinations for a piece contains a jump move.
 *
 * @param sourceIndex Source tile index to start the move from.
 * @param moveDestinations List of move destinations for the piece.
 * @returns True if the list of move destinations contains a jump move,
 * otherwise false.
 */
export const containsJumpMove = (
    sourceIndex: number,
    moveDestinations: number[]
): boolean => {
    return (
        moveDestinations.find((destIndex) =>
            isJumpMove(sourceIndex, destIndex)
        ) !== undefined
    )
}

/**
 * Gets the list of lists of valid move destinations for each tile/piece.
 *
 * @param tileStates List of TileState values for the game board.
 * @param playerColor Color of the player making the move.
 * @param isYourTurn Whether it is the player's turn.
 * @param previousMoveDestIndex Destination tile index of the previous move.
 * @returns List of lists of valid move destinations for each tile/piece.
 */
export const getMoveDestinations = (
    tileStates: TileState[],
    playerColor: PlayerColor,
    isYourTurn: boolean,
    previousMoveDestIndex: number | undefined
): number[][] => {
    if (!isYourTurn)
        // Not your turn so no possible moves, fill with empty arrays
        return Array.from({ length: tileStates.length }, () => [])

    if (previousMoveDestIndex !== undefined) {
        // Get the color of the player that performed the previous move
        const previousMovePlayerColor = isBlackPiece(
            tileStates[previousMoveDestIndex]
        )
            ? PlayerColor.BLACK
            : PlayerColor.RED
        if (previousMovePlayerColor === playerColor && isYourTurn) {
            // It's still the player's turn, so they must be performing an additional jump with the
            // piece at the previousMoveDestIndex
            let pieceDestinations = getPieceMoveDestinations(
                previousMoveDestIndex,
                playerColor,
                tileStates
            )
            // Filter for double jumps only
            pieceDestinations = pieceDestinations.filter((destIndex) =>
                isJumpMove(previousMoveDestIndex, destIndex)
            )
            const moveDestinations: number[][] = Array.from(
                { length: tileStates.length },
                () => []
            )
            // Return the move destinations where the previously moved piece must continue jumping
            moveDestinations[previousMoveDestIndex] = pieceDestinations
            return moveDestinations
        }
    }

    // Get the move destinations for each tile index
    let moveDestinations: number[][] = []
    for (let tileIndex = 0; tileIndex < tileStates.length; tileIndex++) {
        moveDestinations.push(
            getPieceMoveDestinations(tileIndex, playerColor, tileStates)
        )
    }

    // Force jump moves if available
    const containsJump =
        moveDestinations.find((destinations, tileIndex) =>
            containsJumpMove(tileIndex, destinations)
        ) !== undefined
    if (containsJump) {
        moveDestinations = moveDestinations.map((destinations, tileIndex) =>
            destinations.filter((destIndex) => isJumpMove(tileIndex, destIndex))
        )
    }

    return moveDestinations
}

/**
 * Checks if a player has any legal moves, assuming it is their turn.
 *
 * @param tileStates List of TileState values for the game board.
 * @param playerColor Color of the player to check.
 * @returns True if the player has any legal moves, otherwise false.
 */
export const hasLegalMoves = (
    tileStates: TileState[],
    playerColor: PlayerColor
): boolean => {
    // Get all of the player's moves
    const playerMoveDestinations = getMoveDestinations(
        tileStates,
        playerColor,
        true,
        undefined
    )
    // Check if the player has any moves
    const hasMove =
        playerMoveDestinations.find(
            (moveDestinations) => moveDestinations.length > 0
        ) !== undefined
    return hasMove
}

/**
 * Checks if a player has any pieces remaining.
 *
 * @param tileStates List of TileState values for the game board.
 * @param playerColor Color of the player to check.
 * @returns True if the player has any pieces remaining, otherwise false.
 */
export const hasRemainingPieces = (
    tileStates: TileState[],
    playerColor: PlayerColor
) => {
    // Get the piece ownership function to use
    const isOwnedPiece =
        playerColor === PlayerColor.BLACK ? isBlackPiece : isRedPiece
    // Check if the player owns any of the remaining pieces
    return tileStates.find((tileState) => isOwnedPiece(tileState)) !== undefined
}
