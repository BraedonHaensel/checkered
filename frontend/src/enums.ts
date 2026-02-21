// Player colors
export const PlayerColor = {
  RED: 'red',
  BLACK: 'black',
}
export type PlayerColor = (typeof PlayerColor)[keyof typeof PlayerColor]

// Tile states (piece presence)
export const TileState = {
  EMPTY: 0,
  RED_STANDARD_PIECE: 1,
  RED_KING_PIECE: 2,
  BLACK_STANDARD_PIECE: 3,
  BLACK_KING_PIECE: 4,
}
export type TileState = (typeof TileState)[keyof typeof TileState]
