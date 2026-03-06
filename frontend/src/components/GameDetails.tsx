import type { PreviousMove } from "../game-state"
import { isJumpMove } from "../game-utils"

export const IngameDetails = ({children, statusMessage, moves}: {children: any[] | any, statusMessage: string, moves?: PreviousMove[]}) => {

    const combinedMoves: string[] = [];

    const addNewMove = (move: PreviousMove, arr: string[]) => {
        const isJump = isJumpMove(move.sourceIndex, move.destIndex)
        const seperator = isJump ? "x" : "-"

        arr.push(`${move.sourceIndex}${seperator}${move.destIndex}`)
    }

    const appendFollowupMove = (move: PreviousMove, arr: string[]) => {
        const isJump = isJumpMove(move.sourceIndex, move.destIndex)
        const seperator = isJump ? "x" : "-"
        arr[arr.length-1] = `${arr[arr.length-1]}${seperator}${move.destIndex}`
    }

    if(moves && moves.length > 0) {

        addNewMove(moves[0], combinedMoves)
        
        for(let i = 1; i < moves.length; i++) {
            if(moves[i].sourceIndex === moves[i-1].destIndex) {
                appendFollowupMove(moves[i], combinedMoves)
            } else {
                addNewMove(moves[i], combinedMoves)
            }
        }
    }
    
    return <div className="p-5 w-full h-full">
        <div className="bg-neutral-900 w-full h-full rounded-lg grid grid-rows-[1fr_min-content]">
            <div className="p-2 flex flex-col gap-5 h-full">
                <h2 className="w-full text-center text-xl">{statusMessage}</h2>
                <hr />
                <div className="flex flex-col items-center h-full w-full overflow-auto flrx-grow min-h-0 max-h-[70vh]">
                    <div className="grid grid-cols-2 items-center w-full">
                        <b className="text-center">Black</b>
                        <b className="text-center">Red</b>
                        {
                            combinedMoves.map((move, i) => (
                                <span key={i} className="text-center text-wrap block wrap-break-word">{move}</span>
                            ))
                        }
                    </div>
                </div>
            </div>
            <div className="flex flex-row flex-wrap justify-between p-2 gap-2">
                {children}
            </div>
        </div>
    </div>
}
