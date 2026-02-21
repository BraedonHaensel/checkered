import { useCallback, useMemo, useState } from 'react'
import Tile from './Tile'
import { PlayerColor, TileState } from '../enums'
import { canMoveOwnedPieceOnTurn, isBlackPiece } from '../utils'

type Props = {
  playableTileStates: Array<number>
  playerColor: PlayerColor
  isYourTurn: boolean
}

// Checkers game board
const GameBoard = ({ playableTileStates, playerColor, isYourTurn }: Props) => {
  const [selectedPieceIndex, setSelectedPieceIndex] = useState<number>()

  // Handle clicking on a playable tile
  const handleTileClick = useCallback(
    (playableTileIndex: number) => {
      if (selectedPieceIndex === playableTileIndex) {
        // Deselect the already selected piece
        setSelectedPieceIndex(undefined)
      } else {
        // Select the piece to move
        setSelectedPieceIndex(playableTileIndex)
      }
    },
    [selectedPieceIndex]
  )

  // Checks if the player can move the piece at a given playable tile index.
  const getCanMovePiece = useCallback(
    (playableTileIndex: number): boolean => {
      // Must be the player's turn
      if (!isYourTurn) return false

      // Player must own the piece
      const tileState = playableTileStates[playableTileIndex]
      if (playerColor === PlayerColor.BLACK) {
        if (!isBlackPiece(tileState)) return false
      } else {
        if (isBlackPiece(tileState)) return false
      }

      // Check if the piece can be moved
      return canMoveOwnedPieceOnTurn(playableTileIndex, playableTileStates)
    },
    [playableTileStates, playerColor, isYourTurn]
  )

  // Populate the checkers board
  const tiles = useMemo(() => {
    const result = []
    // Index for the 32 playable dark tiles in a checkers board
    let playableTileIndex = 0
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 8; col++) {
        const isDark = (row + col) % 2 === 1
        // If this is a playable tile set the state based on the board's playableTileStates
        const tileState = isDark
          ? playableTileStates[playableTileIndex]
          : TileState.EMPTY
        const canMovePiece = isDark && getCanMovePiece(playableTileIndex)
        result.push(
          <Tile
            key={`${row}-${col}`}
            isDark={isDark}
            playableTileIndex={isDark ? playableTileIndex : undefined}
            tileState={tileState}
            onClick={handleTileClick}
            canMovePiece={canMovePiece}
            isCurrentlySelected={selectedPieceIndex === playableTileIndex}
          />
        )
        if (isDark) playableTileIndex++
      }
    }
    return result
  }, [playableTileStates, selectedPieceIndex, getCanMovePiece, handleTileClick])

  return (
    <div
      className={`grid h-150 w-150 ${playerColor === PlayerColor.RED && 'rotate-180'} grid-cols-8 grid-rows-8 border-5 border-black`}
    >
      {tiles}
    </div>
  )
}

export default GameBoard
