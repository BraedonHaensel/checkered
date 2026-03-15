import { useCallback, useEffect, useMemo, useState } from 'react'
import Tile from './Tile'
import { PlayerColor } from '../enums'
import { getMoveDestinations, isBlackPiece } from '../game-utils'
import type { GameState } from '../game-state'
import { ColumnLabels } from './ColumnLabels'
import { RowLabels } from './RowLabels'
import { cn } from '../lib/utils'

type Props = {
    gameState: GameState
    onPieceMove: (sourceIndex: number, destIndex: number) => void
}

// Checkers game board
const GameBoard = ({ gameState, onPieceMove }: Props) => {
    const [selectedPieceIndex, setSelectedPieceIndex] = useState<number>()

    const { tileStates, playerColor, isYourTurn, previousMoves } = gameState
    const previousMove = previousMoves[previousMoves.length - 1]

    useEffect(() => {
        if (previousMove !== undefined && isYourTurn) {
            // Check the color of the previously moved piece
            const isPreviousMoveBlack = isBlackPiece(
                tileStates[previousMove.destIndex]
            )
            if (
                (isPreviousMoveBlack && playerColor === PlayerColor.BLACK) ||
                (!isPreviousMoveBlack && playerColor === PlayerColor.RED)
            ) {
                // Player owns the previously moved piece and it's still their turn,
                // so select the piece again for a multi-jump
                // eslint-disable-next-line react-hooks/set-state-in-effect
                setSelectedPieceIndex(previousMove.destIndex)
                return
            }
        }

        // Clear selected piece after a board state change
        setSelectedPieceIndex(undefined)
    }, [isYourTurn, playerColor, previousMove, tileStates])

    // Array of arrays of valid move destinations for the piece on each tile
    const moveDestinations = useMemo(() => {
        return getMoveDestinations(
            tileStates,
            playerColor,
            isYourTurn,
            previousMove?.destIndex
        )
    }, [tileStates, playerColor, isYourTurn, previousMove])

    // Array of valid move destinations for the currently selected piece
    const selectedPieceMoveDestinations = useMemo(
        () =>
            selectedPieceIndex !== undefined
                ? moveDestinations[selectedPieceIndex]
                : [],
        [selectedPieceIndex, moveDestinations]
    )

    // Handle clicking on a tile or piece
    const handleTileClick = useCallback(
        (tileIndex: number) => {
            if (moveDestinations[tileIndex].length > 0) {
                // Clicked on a movable piece
                if (selectedPieceIndex === tileIndex) {
                    // Deselect the currently selected piece
                    setSelectedPieceIndex(undefined)
                } else {
                    // Select the piece to move
                    setSelectedPieceIndex(tileIndex)
                }
                return
            }

            if (
                selectedPieceIndex !== undefined &&
                selectedPieceMoveDestinations.includes(tileIndex)
            ) {
                // Destination tile clicked. Handle moving the piece in the parent
                onPieceMove(selectedPieceIndex, tileIndex)
            }
        },
        [
            selectedPieceIndex,
            moveDestinations,
            selectedPieceMoveDestinations,
            onPieceMove,
        ]
    )

    // Populate the checkers board tiles
    const tiles = useMemo(() => {
        const result = []

        // Index for the 32 playable dark tiles in a checkers board
        let tileIndex = 0
        for (let row = 0; row < 8; row++) {
            for (let col = 0; col < 8; col++) {
                // Key to track each tile in the list
                const tileKey = `${row}-${col}`

                // Check if this is a light tile (unplayable) or dark tile (playable)
                const isDark = (row + col) % 2 === 1
                if (!isDark) {
                    // Add a light tile
                    result.push(<Tile key={tileKey} />)
                    continue
                }

                // Calculate tile props
                const canSelectPiece = moveDestinations[tileIndex].length > 0
                const isMoveDestination =
                    selectedPieceMoveDestinations.includes(tileIndex)

                // Controls when the previous move indicator should be displayed
                const showPreviousMoveHighlight =
                    !selectedPieceIndex &&
                    (previousMove?.sourceIndex === tileIndex ||
                        previousMove?.destIndex === tileIndex)

                // Add a dark tile
                result.push(
                    <Tile
                        key={tileKey}
                        tileIndex={tileIndex}
                        tileState={tileStates[tileIndex]}
                        playerColor={playerColor}
                        canSelectPiece={canSelectPiece}
                        isPieceSelected={selectedPieceIndex === tileIndex}
                        isMoveDestination={isMoveDestination}
                        showPreviousMoveHighlight={showPreviousMoveHighlight}
                        onClick={
                            canSelectPiece || isMoveDestination
                                ? handleTileClick
                                : undefined
                        }
                    />
                )

                // Increment the tile index
                tileIndex++
            }
        }
        return result
    }, [
        tileStates,
        playerColor,
        moveDestinations,
        selectedPieceIndex,
        selectedPieceMoveDestinations,
        previousMove,
        handleTileClick,
    ])

    return (
        <div className="flex aspect-square h-full">
            <RowLabels
                className="w-10 pb-10"
                reverse={playerColor === PlayerColor.RED}
            />

            <div className="flex h-full w-full flex-col">
                <div
                    className={cn(
                        'grid h-full w-full grid-cols-8 grid-rows-8 overflow-hidden rounded-xl',
                        playerColor === PlayerColor.RED && 'rotate-180'
                    )}
                >
                    {tiles}
                </div>
                <ColumnLabels
                    className="h-10"
                    reverse={playerColor === PlayerColor.RED}
                />
            </div>
        </div>
    )
}

export default GameBoard
