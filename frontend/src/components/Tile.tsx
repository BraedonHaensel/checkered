import { TileState } from '../enums'
import { isBlackPiece } from '../utils'
import Piece from './Piece'

type Props = {
  isDark: boolean
  playableTileIndex?: number
  tileState: TileState
  onClick: (playableTileIndex: number) => void
}

const Tile = ({ isDark, playableTileIndex, tileState, onClick }: Props) => {
  if (!isDark) {
    // Light tiles are not playable in a game of checkers
    return <div className="bg-[#FFE4C4]"></div>
  }
  if (playableTileIndex === undefined) {
    throw new Error('playableTileIndex must be defined for light tiles')
  }

  return (
    <div
      className="flex items-center justify-center bg-[#A0522D]"
      onClick={() => onClick(playableTileIndex)}
    >
      {tileState !== TileState.EMPTY && (
        <Piece isBlack={isBlackPiece(tileState)} />
      )}
    </div>
  )
}

export default Tile
