import { PlayerColor } from '../enums'
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
        <div className="w-full p-5 lg:pr-0">
            <div className="box-border flex w-full gap-3 rounded-lg bg-neutral-900 p-3">
                <div
                    className={`flex aspect-square w-15 items-center justify-center rounded-xl bg-white ${turn ? 'border-5 border-yellow-500' : ''}`}
                >
                    <Piece
                        isBlack={color === PlayerColor.BLACK}
                        isKing={false}
                        showSelectedHighlight={false}
                        showSelectableHighlight={false}
                    />
                </div>
                <div className="grid w-full grid-rows-2 items-center">
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
                    <div className="flex h-full w-full flex-row items-center">
                        <div className="flex h-full w-full flex-row items-center gap-3 px-1">
                            {new Array(captured).fill(0).map(() => (
                                <div className="h-4 w-4">
                                    <div
                                        className={`h-full w-full rounded-[100vw] ${style.piece} ${color === PlayerColor.BLACK ? style.red + ' bg-red-500' : style.white + ' bg-white'}`}
                                    />
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
