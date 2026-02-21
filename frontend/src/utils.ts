import { TileState } from './enums'

/**
 * Get the starting tile states for each of the 32 playable dark tiles in a checkers board.
 */
export const getNewBoardPlayableTileStates = () => {
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
 * Returns the row number given a piece's playable tile index.
 */
export const playableTileIndexToRow = (playableTileIndex: number) => {
  // 4 playable tiles per row
  return Math.floor(playableTileIndex / 4)
}

/**
 * Returns the column number given a piece's playable tile index.
 */
export const playableTileIndexToCol = (playableTileIndex: number) => {
  // 4 playable columns per row, columns spaced 2 tiles apart
  const col = (playableTileIndex % 4) * 2
  // Playable columns are offset one tile to the right every second row
  const offset = playableTileIndexToRow(playableTileIndex) % 2 === 0 ? 1 : 0
  return col + offset
}

/**
 * Get the possible move destinations for a piece.
 */
export const getPieceMoveDestinations = (
  playableTileIndex: number,
  playableTileStates: Array<number>
) => {
  // Must be a piece
  const tileState = playableTileStates[playableTileIndex]
  if (tileState === TileState.EMPTY) return []

  const possibleMoves = []
  const canMoveUp = isBlackPiece(tileState) || isKingPiece(tileState)
  const canMoveDown = !isBlackPiece(tileState) || isKingPiece(tileState)

  // Check if the piece can move upwards
  if (canMoveUp) {
    // Must be below the top row to move down
    const row = playableTileIndexToRow(playableTileIndex)
    if (row > 0) {
      // Check if pieces are offset one column to the right in this row
      const isOffsetRow = row % 2 === 0
      const col = playableTileIndexToCol(playableTileIndex)
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

  // Check if the piece can move downwards
  if (canMoveDown) {
    // Must be above the bottom row to move up
    const row = playableTileIndexToRow(playableTileIndex)
    if (row < 7) {
      // Check if pieces are offset one column to the right in this row
      const isOffsetRow = row % 2 === 0
      const col = playableTileIndexToCol(playableTileIndex)
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

  // Verify the destinations are empty
  const moveDestinations = []
  for (const moveAmount of possibleMoves) {
    const destIndex = playableTileIndex + moveAmount
    if (playableTileStates[destIndex] === TileState.EMPTY) {
      // Destination is empty, can move piece
      moveDestinations.push(destIndex)
    }
  }

  // Return the remaining moves
  return moveDestinations
}

/**
 * Checks if the player can move the piece at the given index provided it is currently the palyer's
 * turn and they own the piece.
 */
export const canMoveOwnedPieceOnTurn = (
  playableTileIndex: number,
  playableTileStates: Array<number>
) => {
  return (
    getPieceMoveDestinations(playableTileIndex, playableTileStates).length > 0
  )
}
