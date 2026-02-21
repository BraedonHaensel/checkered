import { TileState } from './enums'

export const isBlackPiece = (tileState: TileState) => {
  return [TileState.BLACK_STANDARD_PIECE, TileState.BLACK_KING_PIECE].includes(
    tileState
  )
}
