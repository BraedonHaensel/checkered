import type { PlayerColor, TileState } from './enums'

export type GameStatus =
    | { state: 'SEARCHING' }
    | { state: 'IN_GAME' }
    | { state: 'FINISHED'; winner: PlayerColor | 'draw' }

export type PreviousMove = {
    sourceIndex: number
    destIndex: number
}

export interface GameState {
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
    previousMoves: PreviousMove[]
    opponent: string
    draw_requested: boolean
}
