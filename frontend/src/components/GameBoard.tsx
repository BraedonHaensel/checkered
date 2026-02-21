import { useCallback, useMemo, useState } from 'react'
import Tile from './Tile'
import { PlayerColor, TileState } from '../enums'
import {
  canMoveOwnedPieceOnTurn,
  getPieceMoveDestinations,
  isBlackPiece,
} from '../utils'

type Props = {
  tileStates: Array<number> // Array with the state of each of the 32 playable dark tiles in the checkers board
  playerColor: PlayerColor
  isYourTurn: boolean
}

// Checkers game board
const GameBoard = ({ tileStates, playerColor, isYourTurn }: Props) => {
  const [selectedPieceTileIndex, setSelectedPieceTileIndex] = useState<number>()

  // Handle clicking on a piece
  const handlePieceClick = useCallback(
    (pieceTileIndex: number) => {
      if (selectedPieceTileIndex === pieceTileIndex) {
        // Deselect the currently selected piece
        setSelectedPieceTileIndex(undefined)
      } else {
        // Select the piece to move
        setSelectedPieceTileIndex(pieceTileIndex)
      }
    },
    [selectedPieceTileIndex]
  )

  // Handle clicking on a move destination tile
  const handleDestinationClick = useCallback(
    (destTileIndex: number) => {
      console.log(
        `Moving piece from ${selectedPieceTileIndex} to ${destTileIndex}`
      )
    },
    [selectedPieceTileIndex]
  )

  // Checks if the player can move the piece at a given tile index.
  const getCanMovePiece = useCallback(
    (tileIndex: number): boolean => {
      // Must be the player's turn
      if (!isYourTurn) return false

      // Player must own the piece
      const tileState = tileStates[tileIndex]
      if (playerColor === PlayerColor.BLACK) {
        if (!isBlackPiece(tileState)) return false
      } else {
        if (isBlackPiece(tileState)) return false
      }

      // Check if the piece can be moved
      return canMoveOwnedPieceOnTurn(tileIndex, tileStates)
    },
    [tileStates, playerColor, isYourTurn]
  )

  // Populate the checkers board
  const tiles = useMemo(() => {
    const result = []
    const moveDestinations =
      selectedPieceTileIndex !== undefined
        ? getPieceMoveDestinations(selectedPieceTileIndex, tileStates)
        : []
    // Index for the 32 playable dark tiles in a checkers board
    let tileIndex = 0
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 8; col++) {
        const isDark = (row + col) % 2 === 1
        // If this is a playable tile set the state based on the board's tileStates
        const tileState = isDark ? tileStates[tileIndex] : TileState.EMPTY
        const canMovePiece = isDark && getCanMovePiece(tileIndex)
        result.push(
          <Tile
            key={`${row}-${col}`}
            isDark={isDark}
            tileIndex={isDark ? tileIndex : undefined}
            tileState={tileState}
            onPieceClick={handlePieceClick}
            canMovePiece={canMovePiece}
            isCurrentlySelected={selectedPieceTileIndex === tileIndex}
            isMoveDestination={moveDestinations.includes(tileIndex)}
            onDestinationClick={handleDestinationClick}
          />
        )
        if (isDark) tileIndex++
      }
    }
    return result
  }, [
    tileStates,
    selectedPieceTileIndex,
    getCanMovePiece,
    handlePieceClick,
    handleDestinationClick,
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
