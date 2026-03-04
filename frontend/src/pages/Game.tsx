import { useState } from 'react'
import GameBoard from '../components/GameBoard'
import { PlayerColor, TileState } from '../enums'
import {
    getNewBoardTileStates,
    isJumpMove,
    isStandardPiece,
    getJumpedIndex,
    tileIndexToRow,
    isBlackPiece,
    getMoveDestinations,
    containsJumpMove,
} from '../game-utils'
import { useSession } from '../api/session'
import type { GameState } from '../game-state'

const Game = ({ user }: { user: string }) => {
    const [gameState, setGameState] = useState<GameState>({
        status: 'SEARCHING',
        tileStates: getNewBoardTileStates(),
        playerColor: PlayerColor.BLACK,
        isYourTurn: false,
        previousMove: undefined,
    })
    const { status, tileStates, playerColor, isYourTurn } = gameState
    const session = useSession(user)

    const updateGameState = (updates: Partial<GameState>) => {
        setGameState((prev) => ({ ...prev, ...updates }))
    }

    // Updates tileStates for a piece move
    const updateTileStatesForPieceMove = (
        sourceIndex: number,
        destIndex: number
    ) => {
        let newDestTileState = tileStates[sourceIndex]
        if (isStandardPiece(tileStates[sourceIndex])) {
            // Check for a promotion from a standard to crown piece
            const isBlack = isBlackPiece(newDestTileState)
            const destRow = tileIndexToRow(destIndex)
            if (isBlack && destRow === 0) {
                newDestTileState = TileState.BLACK_KING_PIECE
            } else if (!isBlack && destRow === 7) {
                newDestTileState = TileState.RED_KING_PIECE
            }
        }

        setGameState((prev) => {
            const newTileStates = [...prev.tileStates]

            // Move piece from source to dest index
            newTileStates[destIndex] = newDestTileState
            newTileStates[sourceIndex] = TileState.EMPTY

            // If this is a jump move, remove the jumped piece and potentially continue jumping
            if (isJumpMove(sourceIndex, destIndex)) {
                const jumpedIndex = getJumpedIndex(sourceIndex, destIndex)
                newTileStates[jumpedIndex] = TileState.EMPTY

                const opponentColor =
                    playerColor === PlayerColor.BLACK
                        ? PlayerColor.RED
                        : PlayerColor.BLACK

                // Check the new move destinations for a jump move for the current player
                const newMoveDestinations = getMoveDestinations(
                    newTileStates,
                    isYourTurn ? playerColor : opponentColor,
                    true
                )
                if (
                    !containsJumpMove(destIndex, newMoveDestinations[destIndex])
                ) {
                    // Piece can't continue jumping, change the current player's turn
                    updateGameState({ isYourTurn: !isYourTurn })
                }
            } else {
                // Normal move, change the current player's turn
                updateGameState({ isYourTurn: !isYourTurn })
            }

            return {
                ...prev,
                tileStates: newTileStates,
                previousMove: { sourceIndex, destIndex },
            }
        })
    }

    session.on('start', (message) => {
        console.log(`Game started: ${message.player_color}`)
        updateGameState({
            status: 'IN_GAME',
            playerColor: message.player_color,
            isYourTurn: message.player_color === PlayerColor.BLACK,
        })
    })

    session.on('move', (message) => {
        console.log(
            `Received move: ${message.source_index} to ${message.destination_index}`
        )
        // TODO simple implementation, will likely need more logic
        updateTileStatesForPieceMove(
            message.source_index,
            message.destination_index
        )
    })

    const handlePieceMove = (
        source_index: number,
        destination_index: number
    ) => {
        console.info(
            `Moving piece from ${source_index} to ${destination_index}`
        )
        // TODO update when we formalize the format of move messages
        session.send({
            type: 'move',
            source_index,
            destination_index,
        })
        // TODO simple implementation, will likely need more logic
        updateTileStatesForPieceMove(source_index, destination_index)
    }

    return (
        <div className="space-y-6">
            <h1>CHECKERED</h1>
            <GameBoard gameState={gameState} onPieceMove={handlePieceMove} />
            {status === 'SEARCHING' && <p>Searching for an opponent...</p>}
            {status === 'IN_GAME' && (
                <p className="text-2xl">
                    {isYourTurn ? 'Your turn!' : "Opponent's turn..."}
                </p>
            )}
        </div>
    )
}

export default Game
