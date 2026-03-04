import type { PlayerColor, TileState } from './enums'

export type GameStatus =
    | { state: 'SEARCHING' }
    | { state: 'IN_GAME' }
    | { state: 'FINISHED'; winner: PlayerColor | 'DRAW' }

export type PreviousMove = {
    sourceIndex: number
    destIndex: number
}

export interface GameState {
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
    previousMove: PreviousMove | undefined
}

export interface GameState {
    status: GameStatus
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
}
