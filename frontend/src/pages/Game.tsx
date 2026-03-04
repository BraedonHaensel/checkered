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
    hasRemainingPieces,
    hasLegalMoves,
} from '../game-utils'
import { useSession } from '../api/session'
import type { GameState, GameStatus } from '../game-state'

const Game = ({ user }: { user: string }) => {
    const [gameState, setGameState] = useState<GameState>({
        tileStates: getNewBoardTileStates(),
        playerColor: PlayerColor.BLACK,
        isYourTurn: false,
        previousMove: undefined,
    })
    const [gameStatus, setGameStatus] = useState<GameStatus>({
        state: 'SEARCHING',
    })
    const session = useSession(user)

    const updateGameState = (updates: Partial<GameState>) => {
        setGameState((prev) => ({ ...prev, ...updates }))
    }

    // Performs a piece move and updates the game state
    const performPieceMove = (sourceIndex: number, destIndex: number) => {
        setGameState((prev) => {
            const newTileStates = [...prev.tileStates]

            // Check the state of the moved piece
            let newPieceState = newTileStates[sourceIndex]
            const isBlackMove = isBlackPiece(newPieceState)

            // Check for a promotion from a standard to crown piece
            if (isStandardPiece(newPieceState)) {
                const destRow = tileIndexToRow(destIndex)
                if (isBlackMove && destRow === 0) {
                    newPieceState = TileState.BLACK_KING_PIECE
                } else if (!isBlackMove && destRow === 7) {
                    newPieceState = TileState.RED_KING_PIECE
                }
            }

            // Move the piece from the source to dest index
            newTileStates[destIndex] = newPieceState
            newTileStates[sourceIndex] = TileState.EMPTY

            // Get the colors of the player making the current move, and the player that is waiting
            const currentPlayerColor = isBlackMove
                ? PlayerColor.BLACK
                : PlayerColor.RED
            const waitingPlayerColor = isBlackMove
                ? PlayerColor.RED
                : PlayerColor.BLACK

            // Check if the move is a jump
            const isJump = isJumpMove(sourceIndex, destIndex)
            if (isJump) {
                // Remove the jumped piece
                const jumpedIndex = getJumpedIndex(sourceIndex, destIndex)
                newTileStates[jumpedIndex] = TileState.EMPTY

                // Check if the game is finished due to the jumped waiting player being out of pieces
                if (!hasRemainingPieces(newTileStates, waitingPlayerColor)) {
                    setGameStatus({
                        state: 'FINISHED',
                        winner: currentPlayerColor,
                    })
                    return {
                        ...prev,
                        tileStates: newTileStates,
                        isYourTurn: false,
                        previousMove: {
                            sourceIndex,
                            destIndex,
                        },
                    }
                }
            }

            if (isJump) {
                // Get the move destinations for the new board state
                const newMoveDestinations = getMoveDestinations(
                    newTileStates,
                    currentPlayerColor,
                    true,
                    destIndex
                )
                if (
                    containsJumpMove(destIndex, newMoveDestinations[destIndex])
                ) {
                    // The previously moved piece can jump again. Keep the current player's turn
                    return {
                        ...prev,
                        tileStates: newTileStates,
                        previousMove: {
                            sourceIndex,
                            destIndex,
                        },
                    }
                }
            }

            // Waiting player's turn next. Check if they are out of moves
            if (!hasLegalMoves(newTileStates, waitingPlayerColor)) {
                // Game over, waiting player is out of moves
                setGameStatus({
                    state: 'FINISHED',
                    winner: currentPlayerColor,
                })
                return {
                    ...prev,
                    tileStates: newTileStates,
                    isYourTurn: false,
                    previousMove: {
                        sourceIndex,
                        destIndex,
                    },
                }
            }

            // Return the new game state with the next turn going to the waiting player
            return {
                ...prev,
                tileStates: newTileStates,
                isYourTurn: gameState.playerColor === waitingPlayerColor,
                previousMove: {
                    sourceIndex,
                    destIndex,
                },
            }
        })
    }

    session.on('start', (message) => {
        console.log(`Game started: ${message.player_color}`)
        setGameStatus({ state: 'IN_GAME' })
        updateGameState({
            playerColor: message.player_color,
            isYourTurn: message.player_color === PlayerColor.BLACK,
        })
    })

    session.on('move', (message) => {
        console.log(
            `Received move: ${message.source_index} to ${message.destination_index}`
        )
        // TODO simple implementation, will likely need more logic
        performPieceMove(message.source_index, message.destination_index)
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
        performPieceMove(source_index, destination_index)
    }

    return (
        <div className="space-y-6">
            <h1>CHECKERED</h1>
            <GameBoard gameState={gameState} onPieceMove={handlePieceMove} />
            {gameStatus.state === 'SEARCHING' && (
                <p>Searching for an opponent...</p>
            )}
            {gameStatus.state === 'IN_GAME' && (
                <p className="text-2xl">
                    {gameState.isYourTurn ? 'Your turn!' : "Opponent's turn..."}
                </p>
            )}
            {gameStatus.state === 'FINISHED' && (
                <p className="text-2xl">
                    {gameState.playerColor === gameStatus.winner
                        ? 'Your win!'
                        : 'You lose!'}
                </p>
            )}
        </div>
    )
}

export default Game
