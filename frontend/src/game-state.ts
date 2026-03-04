import type { PlayerColor, TileState } from './enums'

export type GameStatus =
    | { state: 'SEARCHING' }
    | { state: 'IN_GAME' }
    | { state: 'FINISHED'; winner: PlayerColor | 'DRAW' }

export interface GameState {
    tileStates: TileState[]
    playerColor: PlayerColor
    isYourTurn: boolean
    previousMove:
        | {
              sourceIndex: number
              destIndex: number
          }
        | undefined
}
