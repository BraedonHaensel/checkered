import { useCallback, useEffect, useMemo, useState } from 'react'
import Tile from './Tile'
import { PlayerColor } from '../enums'
import { getMoveDestinations } from '../utils'

type Props = {
    tileStates: number[] // Array with the state of each of the 32 playable dark tiles in the checkers board
    playerColor: PlayerColor
    isYourTurn: boolean
    onPieceMove: (sourceIndex: number, destIndex: number) => void
}

// Checkers game board
const GameBoard = ({
    tileStates,
    playerColor,
    isYourTurn,
    onPieceMove,
}: Props) => {
    const [selectedPieceIndex, setSelectedPieceIndex] = useState<number>()

    useEffect(() => {
        // Clear selected piece after a board state change
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setSelectedPieceIndex(undefined)
    }, [tileStates])

    // Array of arrays of valid move destinations for the piece on each tile
    const moveDestinations = useMemo(() => {
        return getMoveDestinations(tileStates, playerColor, isYourTurn)
    }, [tileStates, playerColor, isYourTurn])

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
        handleTileClick,
    ])

    return (
        <div
            className={`grid h-150 w-150 ${playerColor === PlayerColor.RED && 'rotate-180'} grid-cols-8 grid-rows-8 border-5 border-black`}
        >
            {tiles}
        </div>
    )
}

export default GameBoard
