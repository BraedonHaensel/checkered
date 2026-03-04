import type { PlayerColor, TileState } from './enums'

export type GameStatus = 'SEARCHING' | 'IN_GAME' | 'GAME_OVER'

export type PreviousMove = {
    sourceIndex: number
    destIndex: number
}

export interface GameState {
    status: GameStatus
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
    previousMove: PreviousMove | undefined
}
