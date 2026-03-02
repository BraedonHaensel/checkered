import type { PlayerColor } from "../enums";

interface GenericMessage<T> {
    type: T,
}

export interface StartMessage extends GenericMessage<"start"> {
    player_colour: PlayerColor,
}

export interface MoveMessage extends GenericMessage<"move"> {
    source: number,
    destination: number
}

export type Message = StartMessage
    | MoveMessage
