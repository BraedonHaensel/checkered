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
} from '../utils'
import { useSession } from '../api/session'

const Game = ({user}: {user: string}) => {
    // Array with the state of each of the 32 playable dark tiles in the checkers board
    const [tileStates, setTileStates] = useState<TileState[]>(
        getNewBoardTileStates()
    )
    const session = useSession(user);
    // Start with the WebSocket connection disabled until the user searches for a game
    const [isInGame, setIsInGame] = useState(false)
    const [statusMessage, setStatusMessage] = useState<string>()
    const [playerColor, setPlayerColor] = useState(PlayerColor.BLACK)
    const [isYourTurn, setIsYourTurn] = useState(false)

    const updateIsYourTurn = (isYourTurn: boolean) => {
        setIsYourTurn(isYourTurn)
        setStatusMessage(isYourTurn ? 'Your turn!' : "Opponent's turn...")
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

        setTileStates((prev) => {
            const newTileStates = [...prev]

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
                if (!containsJumpMove(destIndex, newMoveDestinations[destIndex])) {
                    // Piece can't continue jumping, change the current player's turn
                    updateIsYourTurn(!isYourTurn)
                }
            } else {
                // Normal move, change the current player's turn
                updateIsYourTurn(!isYourTurn)
            }

            return newTileStates
        })
    }

    session.on("start", (message) => {
        console.log(`Game started: ${message.player_colour}`)
        setIsInGame(true)
        setPlayerColor(message.player_colour)
        updateIsYourTurn(message.player_colour === PlayerColor.BLACK)
    })

    session.on("move", (message) => {
        console.log(
            `Received move: ${message.source} to ${message.destination}`
        )
        // TODO simple implementation, will likely need more logic
        updateTileStatesForPieceMove(message.source, message.destination)
    });

    const handlePieceMove = (source: number, destination: number) => {
        console.info(`Moving piece from ${source} to ${destination}`)
        // TODO update when we formalize the format of move messages
        session.send({
            type: 'move',
            source,
            destination,
        })
        // TODO simple implementation, will likely need more logic
        updateTileStatesForPieceMove(source, destination)
    }

    return (
        <div className="space-y-6">
            <h1>CHECKERED</h1>
            <GameBoard
                tileStates={tileStates}
                playerColor={playerColor}
                isYourTurn={isYourTurn}
                onPieceMove={handlePieceMove}
            />
            {!isInGame &&
                <p>Searching for an opponent...</p>
            }
            {statusMessage && <p className="text-2xl">{statusMessage}</p>}
        </div>
    )
}

export default Game
