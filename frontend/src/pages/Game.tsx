import { useReducer, useState } from 'react'
import GameBoard from '../components/GameBoard'
import { Page, PlayerColor, TileState } from '../enums'
import {
    getNewBoardTileStates,
    isJumpMove,
    isStandardPiece,
    getJumpedTileIndex,
    tileIndexToRow,
    isBlackPiece,
    getMoveDestinations,
    containsJumpMove,
    hasRemainingPieces,
    hasLegalMoves,
    pieceCount,
} from '../game-utils'
import { useSession } from '../api/session'
import type { GameState, GameStatus } from '../game-state'
import { PlayerCard } from '../components/PlayerCard'
import { IngameDetails } from '../components/GameDetails'
import {
    DashboardButton,
    DrawButton,
    SearchButton,
} from '../components/Buttons'

const DEFAULT_GAME_STATE: GameState = {
    tileStates: getNewBoardTileStates(),
    playerColor: PlayerColor.BLACK,
    isYourTurn: false,
    previousMoves: [],
    opponent: 'Opponent',
    draw_requested: false
}
const DEFAULT_GAME_STATUS: GameStatus = { state: 'SEARCHING' }

const Game = ({
    user,
    setPage,
}: {
    user: string
    setPage: (page: Page) => void
}) => {
    const session = useSession(user)
    const [gameState, setGameState] = useState<GameState>(DEFAULT_GAME_STATE)
    const [gameStatus, setGameStatus] = useReducer<
        GameStatus,
        [GameStatus],
        [GameStatus]
    >(
        (prev, ...arg) => {
            if (arg.length != 1) {
                console.error('Unexpected state update', arg)
                return prev
            }
            const newState = arg[0]
            if (prev.state != 'SEARCHING' && newState.state === 'SEARCHING') {
                console.log('Searching...')
                session.send({ type: 'enqueue' })
            }
            return newState
        },
        [DEFAULT_GAME_STATUS],
        (i) => {
            console.log('Searching...')
            session.send({ type: 'enqueue' })
            return i[0]
        }
    )

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
                const jumpedIndex = getJumpedTileIndex(sourceIndex, destIndex)
                newTileStates[jumpedIndex] = TileState.EMPTY

                // Check if the game is finished due to the jumped waiting player being out of pieces
                if (!hasRemainingPieces(newTileStates, waitingPlayerColor)) {
                    return {
                        ...prev,
                        tileStates: newTileStates,
                        isYourTurn: false,
                        draw_requested: false,
                        previousMoves: [
                            ...prev.previousMoves,
                            {
                                sourceIndex,
                                destIndex,
                            },
                        ],
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
                        draw_requested: false,
                        previousMoves: [
                            ...prev.previousMoves,
                            {
                                sourceIndex,
                                destIndex,
                            },
                        ],
                    }
                }
            }

            // Waiting player's turn next. Check if they are out of moves
            if (!hasLegalMoves(newTileStates, waitingPlayerColor)) {
                return {
                    ...prev,
                    tileStates: newTileStates,
                    isYourTurn: false,
                    draw_requested: false,
                    previousMoves: [
                        ...prev.previousMoves,
                        {
                            sourceIndex,
                            destIndex,
                        },
                    ],
                }
            }

            // Return the new game state with the next turn going to the waiting player
            return {
                ...prev,
                tileStates: newTileStates,
                isYourTurn: gameState.playerColor === waitingPlayerColor,
                draw_requested: false,
                previousMoves: [
                    ...prev.previousMoves,
                    {
                        sourceIndex,
                        destIndex,
                    },
                ],
            }
        })
    }

    const resetState = () => {
        updateGameState(DEFAULT_GAME_STATE)
        setGameStatus(DEFAULT_GAME_STATUS)
    }

    session.on('start', (message) => {
        console.log(`Game started: ${message.player_color}`)
        setGameStatus({ state: 'IN_GAME' })
        updateGameState({
            playerColor: message.player_color,
            isYourTurn: message.player_color === PlayerColor.BLACK,
            opponent: message.opponent,
        })
    })

    session.on('update_state', (message) => {
        console.log(`Update State`, message)
        setGameState((oldState: GameState) => ({
            ...oldState,
            draw_requested: false,
            tileStates: message.tile_states,
            previousMoves: message.previous_moves.map((move) => ({
                sourceIndex: move.source_index,
                destIndex: move.destination_index,
            })),
            isYourTurn: message.turn == oldState.playerColor,
        }))
    })

    session.on("request_draw", () => {
        setGameState(oldState => ({
            ...oldState,
            draw_requested: true,
        }))
    })

    session.on('game_end', (message) => {
        console.log('Game Over!', message)
        setGameStatus({
            state: 'FINISHED',
            winner: message.winner,
        })
        setGameState((oldState) => ({
            ...oldState,
            draw_requested: false,
            isYourTurn: false,
        }))
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

    let statusMessage = 'Unknown'

    switch (gameStatus.state) {
        case 'SEARCHING':
            statusMessage = 'Searching for Opponent...'
            break
        case 'IN_GAME':
            statusMessage = gameState.isYourTurn
                ? 'Your turn!'
                : "Opponent's turn..."
            break
        case 'FINISHED':
            statusMessage =
                gameStatus.winner === "draw" 
                    ? "Draw!"
                    : gameState.playerColor === gameStatus.winner
                        ? 'Your win!'
                        : 'You lose!'
            break
    }

    const forfeit = () => {
        session.send({
            type: 'forfeit',
        })
    }

    const draw = () => {
        session.send({type: "request_draw"})
        setGameState(old => ({...old, draw_requested: true}))
    }

    const opponentColor = gameState.playerColor === 'red' ? 'black' : 'red'

    const captured = 12 - pieceCount(gameState.tileStates, opponentColor)
    const lost = 12 - pieceCount(gameState.tileStates, gameState.playerColor)

    return (
        <div className="grid h-lvh grid-rows-[min-content_1fr] items-center justify-center pt-5">
            <h1 className="w-full text-center">CHECKERED</h1>
            <div className="grid w-screen grid-cols-[1fr] grid-rows-[auto_min-content] items-center md:h-full md:grid-cols-[2fr_1fr] md:grid-rows-[1fr]">
                <div className="flex h-full min-h-[67lvh] w-full flex-col items-center">
                    <PlayerCard
                        player={gameState.opponent}
                        color={opponentColor}
                        captured={lost}
                        lost={captured}
                        turn={
                            gameStatus.state === 'IN_GAME' &&
                            !gameState.isYourTurn
                        }
                    />
                    <GameBoard
                        gameState={gameState}
                        onPieceMove={handlePieceMove}
                    />
                    <PlayerCard
                        player={user}
                        color={gameState.playerColor}
                        captured={captured}
                        lost={lost}
                        turn={
                            gameStatus.state === 'IN_GAME' &&
                            !gameState.isYourTurn
                        }
                    />
                </div>
                <IngameDetails
                    statusMessage={statusMessage}
                    isSearching={gameStatus.state === 'SEARCHING'}
                    moves={
                        gameStatus.state === 'SEARCHING'
                            ? undefined
                            : gameState.previousMoves
                    }
                >
                    <DashboardButton
                        onClick={() =>
                            gameStatus.state === 'IN_GAME'
                                ? forfeit()
                                : setPage(Page.HOME)
                        }
                        exitsGame={gameStatus.state === 'IN_GAME'}
                    />
                    {gameStatus.state === 'IN_GAME' && (
                        <DrawButton onClick={() => draw()} requested={gameState.draw_requested} />
                    )}
                    {gameStatus.state === 'FINISHED' && (
                        <SearchButton onClick={resetState} />
                    )}
                </IngameDetails>
            </div>
        </div>
    )
}

export default Game
