import { useMemo } from 'react'
import Tile from './Tile'
import { PlayerColor, TileState } from '../enums'

type Props = {
  playableTileStates: Array<number>
  playerColor: PlayerColor
  isYourTurn: boolean
}

const GameBoard = ({ playableTileStates, playerColor, isYourTurn }: Props) => {
  // Handle clicking on a playable tile
  const handleTileClick = (playableTileIndex: number) => {
    console.log(`Clicked tile: ${playableTileIndex}`)
  }

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
        result.push(
          <Tile
            key={`${row}-${col}`}
            isDark={isDark}
            playableTileIndex={isDark ? playableTileIndex : undefined}
            tileState={tileState}
            onClick={handleTileClick}
          />
        )
        if (isDark) playableTileIndex++
      }
    }
    return result
  }, [playableTileStates])

  return (
    <div
      className={`grid h-150 w-150 ${playerColor === PlayerColor.RED && 'rotate-180'} grid-cols-8 grid-rows-8 border-5 border-black`}
    >
      {tiles}
    </div>
  )
}

export default GameBoard
