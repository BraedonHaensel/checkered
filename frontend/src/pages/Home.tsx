import { useState } from 'react'
import GameBoard from '../components/GameBoard'
import SearchButton from '../components/SearchButton'
import useWebSocket from 'react-use-websocket'
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

const SERVER_WEBSOCKET_URL = 'ws://localhost:3000'

// WebSocket message types
type StartMessage = { type: 'start'; playerColor: PlayerColor }
type MoveMessage = { type: 'move'; sourceIndex: number; destIndex: number }
type WebSocketMessage = StartMessage | MoveMessage

const Home = () => {
  // Array with the state of each of the 32 playable dark tiles in the checkers board
  const [tileStates, setTileStates] = useState<TileState[]>(
    getNewBoardTileStates()
  )
  // Start with the WebSocket connection disabled until the user searches for a game
  const [wsConnectionEnabled, setWsConnectionEnabled] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
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

  // Handle WebSocket messages
  const handleMessage = (message: WebSocketMessage) => {
    if (!message) return
    console.info(`Received message: ${JSON.stringify(message)}`)

    switch (message.type) {
      case 'start':
        // TODO update when we formalize the format of start messages
        // Game start message
        console.log(`Game started: ${message.playerColor}`)
        setIsSearching(false)
        setIsInGame(true)
        setPlayerColor(message.playerColor)
        updateIsYourTurn(message.playerColor === PlayerColor.BLACK)
        break
      case 'move':
        console.log(
          `Received move: ${message.sourceIndex} to ${message.destIndex}`
        )
        // TODO simple implementation, will likely need more logic
        updateTileStatesForPieceMove(message.sourceIndex, message.destIndex)
    }
  }

  // WebSocket setup
  const { sendJsonMessage } = useWebSocket<WebSocketMessage>(
    SERVER_WEBSOCKET_URL,
    {
      onOpen: () => {
        console.info('WebSocket connection established')
      },
      onMessage: (event) => {
        const message: WebSocketMessage = JSON.parse(event.data)
        handleMessage(message)
      },
    },
    wsConnectionEnabled // Delay connecting until the user searches for a game
  )

  // Handle clicking the "search for opponent" button
  const handleSearchForOpponent = () => {
    setIsSearching(true)
    setWsConnectionEnabled(true)
  }

  const handlePieceMove = (sourceIndex: number, destIndex: number) => {
    console.info(`Moving piece from ${sourceIndex} to ${destIndex}`)
    // TODO update when we formalize the format of move messages
    sendJsonMessage({
      type: 'move',
      sourceIndex,
      destIndex,
    })
    // TODO simple implementation, will likely need more logic
    updateTileStatesForPieceMove(sourceIndex, destIndex)
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
      {isSearching ? (
        <p>Searching for an opponent...</p>
      ) : (
        !isInGame && <SearchButton onClick={handleSearchForOpponent} />
      )}
      {statusMessage && <p className="text-2xl">{statusMessage}</p>}
    </div>
  )
}

export default Home
