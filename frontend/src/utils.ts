import { PlayerColor, TileState } from './enums'

/**
 * Get the starting tile states for each of the 32 playable dark tiles in a checkers board.
 */
export const getNewBoardTileStates = () => {
  // Only 32 tiles are playable in checkers. The first 12 start with black pieces, the middle 8
  // are empty, and the last 12 start with red pieces
  const tileStates = [
    ...Array(12).fill(TileState.RED_STANDARD_PIECE),
    ...Array(8).fill(TileState.EMPTY),
    ...Array(12).fill(TileState.BLACK_STANDARD_PIECE),
  ]
  return tileStates
}

/**
 * Checks if the given tileState is for a black piece (standard or king).
 */
export const isBlackPiece = (tileState: TileState) => {
  return [TileState.BLACK_STANDARD_PIECE, TileState.BLACK_KING_PIECE].includes(
    tileState
  )
}

/**
 * Checks if the given tileState is for a king piece.
 */
export const isKingPiece = (tileState: TileState) => {
  return [TileState.BLACK_KING_PIECE, TileState.RED_KING_PIECE].includes(
    tileState
  )
}

/**
 * Returns the row number given a piece's tile index.
 */
export const tileIndexToRow = (tileIndex: number) => {
  // 4 playable tiles per row
  return Math.floor(tileIndex / 4)
}

/**
 * Returns the column number given a piece's tile index.
 */
export const tileIndexToCol = (tileIndex: number) => {
  // 4 playable columns per row, columns spaced 2 tiles apart
  const col = (tileIndex % 4) * 2
  // Playable columns are offset one tile to the right every second row
  const offset = tileIndexToRow(tileIndex) % 2 === 0 ? 1 : 0
  return col + offset
}

/**
 * Get an array of the possible move destinations for the piece at the given tile index.
 */
export const getPieceMoveDestinations = (
  tileIndex: number,
  playerColor: PlayerColor,
  tileStates: Array<TileState>
): Array<number> => {
  const tileState = tileStates[tileIndex]

  // Tile must not be empty
  if (tileState === TileState.EMPTY) return []

  // Player must own the piece
  if (playerColor === PlayerColor.BLACK) {
    if (!isBlackPiece(tileState)) return []
  } else {
    if (isBlackPiece(tileState)) return []
  }

  // Check the piece's vertical move directions
  const canMoveUp = isBlackPiece(tileState) || isKingPiece(tileState)
  const canMoveDown = !isBlackPiece(tileState) || isKingPiece(tileState)

  const possibleMoves = []

  // Check for upwards moves
  if (canMoveUp) {
    // Must be below the top row to move up
    const row = tileIndexToRow(tileIndex)
    if (row > 0) {
      // Check if pieces are offset one column to the right in this row
      const isOffsetRow = row % 2 === 0
      const col = tileIndexToCol(tileIndex)
      if (col > 0) {
        // Can move up-left
        possibleMoves.push(isOffsetRow ? -4 : -5)
      }
      if (col < 7) {
        // Can move up-right
        possibleMoves.push(isOffsetRow ? -3 : -4)
      }
    }
  }

  // Check for downwards moves
  if (canMoveDown) {
    // Must be above the bottom row to move down
    const row = tileIndexToRow(tileIndex)
    if (row < 7) {
      // Check if pieces are offset one column to the right in this row
      const isOffsetRow = row % 2 === 0
      const col = tileIndexToCol(tileIndex)
      if (col > 0) {
        // Can move down-left
        possibleMoves.push(isOffsetRow ? 4 : 3)
      }
      if (col < 7) {
        // Can move down-right
        possibleMoves.push(isOffsetRow ? 5 : 4)
      }
    }
  }

  // Verify the move destinations are empty
  const moveDestinations = []
  for (const moveAmount of possibleMoves) {
    const destIndex = tileIndex + moveAmount
    if (tileStates[destIndex] === TileState.EMPTY) {
      // Destination is empty, can move piece
      moveDestinations.push(destIndex)
    }
  }

  // Return the valid moves
  return moveDestinations
}

/**
 * Gets an array of arrays of valid move destinations for each tile/piece.
 */
export const getMoveDestinations = (
  tileStates: Array<TileState>,
  playerColor: PlayerColor,
  isYourTurn: boolean
): Array<Array<number>> => {
  if (!isYourTurn)
    // Not your turn so no possible moves, fill with empty arrays
    return Array.from({ length: tileStates.length }, () => [])
  const moveDestinations = []
  for (let tileIndex = 0; tileIndex < tileStates.length; tileIndex++) {
    moveDestinations.push(
      getPieceMoveDestinations(tileIndex, playerColor, tileStates)
    )
  }
  return moveDestinations
}
