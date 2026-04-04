import { LoaderCircle } from 'lucide-react'
import type { PreviousMove } from '../game-state'
import { isJumpMove, tileIndexToCoordinate } from '../game-utils'
import type { ReactNode } from 'react'

type Props = {
    children: ReactNode
    statusMessage: string
    isSearching?: boolean
    moves?: PreviousMove[]
}

export const GameDetails = ({
    children,
    statusMessage,
    isSearching = false,
    moves,
}: Props) => {
    const moveNotations: string[] = []

    // Gets the coordiante notation for a move
    const moveToCoordinateNotation = (move: PreviousMove) => {
        const isJump = isJumpMove(move.sourceIndex, move.destIndex)
        const seperator = isJump ? 'x' : '-'

        const sourceCoord = tileIndexToCoordinate(move.sourceIndex)
        const destCoord = tileIndexToCoordinate(move.destIndex)

        return `${sourceCoord}${seperator}${destCoord}`
    }

    // Append an extra jump to the previous move notation for multijumps
    const appendMultiJumpMove = (destIndex: number) => {
        const destCoord = tileIndexToCoordinate(destIndex)
        moveNotations[moveNotations.length - 1] =
            `${moveNotations[moveNotations.length - 1]}x${destCoord}`
    }

    if (moves && moves.length > 0) {
        moveNotations.push(moveToCoordinateNotation(moves[0]))

        for (let i = 1; i < moves.length; i++) {
            if (moves[i].sourceIndex === moves[i - 1].destIndex) {
                appendMultiJumpMove(moves[i].destIndex)
            } else {
                moveNotations.push(moveToCoordinateNotation(moves[i]))
            }
        }
    }

    return (
        <div className="h-full max-h-100 w-full md:max-h-full">
            <div className="grid h-full w-full grid-rows-[1fr_min-content] rounded-lg bg-neutral-900">
                <div className="flex h-full min-h-0 flex-col gap-5 p-2">
                    <h2 className="w-full text-center text-xl">
                        {statusMessage}
                    </h2>
                    <hr />
                    {isSearching ? (
                        <div className="m-auto">
                            <LoaderCircle
                                size={60}
                                className="m-auto animate-spin"
                            />
                        </div>
                    ) : (
                        <div className="flex w-full flex-col items-center overflow-y-auto">
                            <div className="grid w-full grid-cols-2 items-center">
                                {moves && (
                                    <>
                                        <b className="text-center">Black</b>
                                        <b className="text-center">Red</b>
                                    </>
                                )}
                                {moveNotations.map((move, i) => (
                                    <span
                                        key={i}
                                        className="block text-center text-wrap wrap-break-word"
                                    >
                                        {move}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
                <div className="flex flex-row flex-wrap justify-between gap-2 p-2">
                    {children}
                </div>
            </div>
        </div>
    )
}
