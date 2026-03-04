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

// Move directions
export const MoveDirection = {
    UP_LEFT: 0,
    UP_RIGHT: 1,
    DOWN_LEFT: 2,
    DOWN_RIGHT: 3,
}
export type MoveDirection = (typeof MoveDirection)[keyof typeof MoveDirection]

export const Page = {
    HOME: 0,
    GAME: 1,
    LOGIN: 2,
    LEADERBOARD: 3,
}
export type Page = (typeof Page)[keyof typeof Page]
