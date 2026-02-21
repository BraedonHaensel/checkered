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

export const isBlackPiece = (tileState: TileState) => {
  return [TileState.BLACK_STANDARD_PIECE, TileState.BLACK_KING_PIECE].includes(
    tileState
  )
}
