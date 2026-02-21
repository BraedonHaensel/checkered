import { TileState } from '../enums'
import { isBlackPiece } from '../utils'
import Piece from './Piece'

type Props = {
  isDark: boolean
  tileState: TileState
}

const Tile = ({ isDark, tileState }: Props) => {
  if (!isDark) {
    // Light tiles are never interacted with in a game of checkers
    return <div className="bg-[#FFE4C4]"></div>
  }

  return (
    <div className="flex items-center justify-center bg-[#A0522D]">
      {tileState !== TileState.EMPTY && (
        <Piece isBlack={isBlackPiece(tileState)} />
      )}
    </div>
  )
}

export default Tile
