import { PlayerColor } from "../enums";
import Piece from "./Piece";

import style from "./Piece.module.css"

export const PlayerCard = ({player, color, captured, lost, turn}: {player: string, color: PlayerColor, captured: number, lost: number, turn: boolean}) => {
    const advantage = captured-lost;
    return <div className="p-5 md:pr-0 w-full">
        <div className="flex w-full box-border p-3 gap-3 bg-neutral-900 rounded-lg">
            <div className={`w-15 aspect-square bg-white flex items-center justify-center rounded-xl ${turn ? "border-5 border-yellow-500" : ""}`}>
                <Piece isBlack={color === PlayerColor.BLACK} isKing={false} showSelectedHighlight={false} showSelectableHighlight={false} />
            </div>
            <div className="grid grid-rows-2 items-center w-full">
                <p>
                    <b>{player}</b>&nbsp;
                    <i>{advantage > 0 ? `(+${advantage})` : advantage < 0 ? `(${advantage})` : ""}</i>
                </p>
                <div className="flex flex-row items-center w-full h-full">
                    <div className="px-1 flex flex-row items-center gap-3 w-full h-full">
                        {
                            new Array(captured).fill(0).map(() => (
                                <div className="w-4 h-4">
                                    <div className={`rounded-[100vw] w-full h-full ${style.piece} ${color === PlayerColor.BLACK ? style.red + " bg-red-500" : style.white + " bg-white"}`} />
                                </div>
                            ))
                        } 
                    </div>
                </div>
            </div>
        </div>
    </div>
}
