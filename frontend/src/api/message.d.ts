import type { PlayerColor, TileState } from '../enums'

interface GenericMessage<T> {
    type: T
}

export interface JoinMessage extends GenericMessage<'join'> {
    user: string
}

export interface StartMessage extends GenericMessage<'start'> {
    player_color: PlayerColor
}

export interface MoveMessage extends GenericMessage<'move'> {
    source_index: number
    destination_index: number
}

export interface UpdateStateMessage extends GenericMessage<"update_state"> {
    tile_states: TileState[]
    turn: PlayerColor
    previous_move?: {
        source_index: number,
        destination_index: number,
    }
}

export interface GameEndMessage extends GenericMessage<"game_end"> {
    winner: PlayerColor | "draw"
}

export type Message = JoinMessage | StartMessage | MoveMessage | UpdateStateMessage | GameEndMessage
