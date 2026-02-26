import { useState } from 'react'
import GameBoard from '../components/GameBoard'
import SearchButton from '../components/SearchButton'
import useWebSocket from 'react-use-websocket'
import { PlayerColor, TileState } from '../enums'
import { getNewBoardTileStates } from '../utils'

const SERVER_WEBSOCKET_URL = 'ws://localhost:3000'

// WebSocket message types
type StartMessage = { type: 'start'; playerColor: PlayerColor }
type MoveMessage = { type: 'move' }
type WebSocketMessage = StartMessage | MoveMessage

const Home = () => {
  // Array with the state of each of the 32 playable dark tiles in the checkers board
  const [tileStates, setTileStates] = useState<Array<TileState>>(
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

  // Handle WebSocket messages
  const handleMessage = (message: WebSocketMessage) => {
    if (!message) return
    console.info(`Received message: ${JSON.stringify(message)}`)

    switch (message.type) {
      case 'start':
        // Game start message
        console.log(`Game started: ${message.playerColor}`)
        setIsSearching(false)
        setIsInGame(true)
        setPlayerColor(message.playerColor)
        updateIsYourTurn(message.playerColor === PlayerColor.BLACK)
        break
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

  return (
    <div className="space-y-6">
      <h1>CHECKERED</h1>
      <GameBoard
        tileStates={tileStates}
        playerColor={playerColor}
        isYourTurn={isYourTurn}
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
