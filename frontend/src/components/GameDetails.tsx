import type { PreviousMove } from '../game-state'
import { isJumpMove } from '../game-utils'

export const IngameDetails = ({
    children,
    statusMessage,
    moves,
}: {
    children: any[] | any
    statusMessage: string
    moves?: PreviousMove[]
}) => {
    const combinedMoves: string[] = []

    const addNewMove = (move: PreviousMove, arr: string[]) => {
        const isJump = isJumpMove(move.sourceIndex, move.destIndex)
        const seperator = isJump ? 'x' : '-'

        arr.push(`${move.sourceIndex}${seperator}${move.destIndex}`)
    }

    const appendFollowupMove = (move: PreviousMove, arr: string[]) => {
        const isJump = isJumpMove(move.sourceIndex, move.destIndex)
        const seperator = isJump ? 'x' : '-'
        arr[arr.length - 1] =
            `${arr[arr.length - 1]}${seperator}${move.destIndex}`
    }

    if (moves && moves.length > 0) {
        addNewMove(moves[0], combinedMoves)

        for (let i = 1; i < moves.length; i++) {
            if (moves[i].sourceIndex === moves[i - 1].destIndex) {
                appendFollowupMove(moves[i], combinedMoves)
            } else {
                addNewMove(moves[i], combinedMoves)
            }
        }
    }

    return (
        <div className="h-full w-full p-5">
            <div className="grid h-full w-full grid-rows-[1fr_min-content] rounded-lg bg-neutral-900">
                <div className="flex h-full flex-col gap-5 p-2">
                    <h2 className="w-full text-center text-xl">
                        {statusMessage}
                    </h2>
                    <hr />
                    <div className="flrx-grow flex h-full max-h-[70vh] min-h-0 w-full flex-col items-center overflow-auto">
                        <div className="grid w-full grid-cols-2 items-center">
                            {moves && (
                                <>
                                    <b className="text-center">Black</b>
                                    <b className="text-center">Red</b>
                                </>
                            )}
                            {combinedMoves.map((move, i) => (
                                <span
                                    key={i}
                                    className="block text-center text-wrap wrap-break-word"
                                >
                                    {move}
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
                <div className="flex flex-row flex-wrap justify-between gap-2 p-2">
                    {children}
                </div>
            </div>
        </div>
    )
}
