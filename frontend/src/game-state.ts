import type { PlayerColor, TileState } from './enums'

export type GameStatus = 'SEARCHING' | 'IN_GAME' | 'GAME_OVER'

export interface GameState {
    status: GameStatus
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
}
