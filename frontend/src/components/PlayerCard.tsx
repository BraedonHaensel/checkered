import { PlayerColor } from '../enums'
import { cn } from '../lib/utils'
import Piece from './Piece'

import style from './Piece.module.css'

export const PlayerCard = ({
    player,
    color,
    captured,
    lost,
    turn,
}: {
    player: string
    color: PlayerColor
    captured: number
    lost: number
    turn: boolean
}) => {
    const advantage = captured - lost
    return (
        <div className="box-border flex w-full gap-3 rounded-lg bg-neutral-900 p-3">
            <div
                className={cn(
                    'flex aspect-square h-full items-center justify-center rounded-xl border-5 bg-white',
                    turn && 'border-yellow-500'
                )}
            >
                <Piece
                    isBlack={color === PlayerColor.BLACK}
                    isKing={false}
                    showSelectedHighlight={false}
                    showSelectableHighlight={false}
                />
            </div>
            <div className="grid min-h-16 w-full grid-rows-2 items-center md:min-h-19">
                <p>
                    <b>{player}</b>&nbsp;
                    <i>
                        {advantage > 0
                            ? `(+${advantage})`
                            : advantage < 0
                              ? `(${advantage})`
                              : ''}
                    </i>
                </p>
                <div className="flex h-full w-full items-center gap-3 overflow-x-auto px-1">
                    {new Array(captured).fill(0).map((i) => (
                        <div key={i}>
                            <div
                                className={cn(
                                    'h-4 w-4 rounded-full',
                                    style.piece,
                                    color === PlayerColor.BLACK
                                        ? cn(style.red, 'bg-red-500')
                                        : cn(style.white, 'bg-white')
                                )}
                            />
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}
