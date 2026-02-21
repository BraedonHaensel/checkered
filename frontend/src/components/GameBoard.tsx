import { useMemo } from 'react'
import Tile from './Tile'
import { PlayerColor, TileState } from '../enums'

type Props = {
  playerColor: PlayerColor
  isYourTurn: boolean
}

const GameBoard = ({ playerColor, isYourTurn }: Props) => {
  // Populate the checkers board
  const tiles = useMemo(() => {
    const result = []
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 8; col++) {
        const isDark = (row + col) % 2 === 1
        const tileState =
          row <= 2 && isDark
            ? TileState.RED_STANDARD_PIECE
            : row >= 5 && isDark
              ? TileState.BLACK_STANDARD_PIECE
              : TileState.EMPTY
        result.push(
          <Tile key={`${row}-${col}`} isDark={isDark} tileState={tileState} />
        )
      }
    }
    return result
  }, [])

  return (
    <div
      className={`grid h-150 w-150 ${playerColor === PlayerColor.RED && 'rotate-180'} grid-cols-8 grid-rows-8 border-5 border-black`}
    >
      {tiles}
    </div>
  )
}

export default GameBoard
