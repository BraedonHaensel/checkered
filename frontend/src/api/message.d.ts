import type { PlayerColor, TileState } from '../enums'

interface GenericMessage<T> {
    type: T
}

export interface JoinMessage extends GenericMessage<'join'> {
    user: string
}

export type ConfirmRegistrationMessage = GenericMessage<'registered'>

export interface StartMessage extends GenericMessage<'start'> {
    player_color: PlayerColor
    opponent: string
}

export type EnqueueMessage = GenericMessage<'enqueue'>

export interface MoveMessage extends GenericMessage<'move'> {
    source_index: number
    destination_index: number
}

export interface UpdateStateMessage extends GenericMessage<'update_state'> {
    tile_states: TileState[]
    turn: PlayerColor
    previous_moves: {
        source_index: number
        destination_index: number
    }[]
}

export interface GameEndMessage extends GenericMessage<'game_end'> {
    winner: PlayerColor | 'draw'
}

export type ForfeitMessage = GenericMessage<'forfeit'>

export type RequestDraw = GenericMessage<'request_draw'>

export type Message =
    | JoinMessage
    | ConfirmRegistrationMessage
    | StartMessage
    | MoveMessage
    | UpdateStateMessage
    | GameEndMessage
    | EnqueueMessage
    | ForfeitMessage
    | RequestDraw
