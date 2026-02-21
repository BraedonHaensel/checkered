import { TileState } from '../enums'
import { isBlackPiece, isKingPiece } from '../utils'
import Piece from './Piece'

type Props = {
  isDark: boolean
  tileIndex?: number
  tileState: TileState
  onPieceClick: (pieceTileIndex: number) => void
  canMovePiece: boolean
  isCurrentlySelected: boolean
  isMoveDestination: boolean
  onDestinationClick: (destTileIndex: number) => void
}

// Individual tile in a checkers board
const Tile = ({
  isDark,
  tileIndex,
  tileState,
  onPieceClick,
  canMovePiece,
  isCurrentlySelected,
  isMoveDestination,
  onDestinationClick,
}: Props) => {
  if (!isDark) {
    // Light tiles are not playable in a game of checkers, no further logic required
    return <div className="bg-[#FFE4C4]"></div>
  }
  if (tileIndex === undefined) {
    throw new Error('tileIndex must be defined for light tiles')
  }

  return (
    <div
      className="flex items-center justify-center bg-[#A0522D]"
      onClick={() => {
        if (canMovePiece) {
          // Selected the piece on this tile
          onPieceClick(tileIndex)
        } else if (isMoveDestination) {
          // Selected this tile as the move destination
          onDestinationClick(tileIndex)
        }
      }}
    >
      {tileState !== TileState.EMPTY && (
        <Piece
          isBlack={isBlackPiece(tileState)}
          isKing={isKingPiece(tileState)}
          showMovableHighlight={canMovePiece}
          showSelectedHighlight={isCurrentlySelected}
        />
      )}
      {isMoveDestination && (
        <div className="h-1/3 w-1/3 rounded-full bg-gray-300 opacity-70"></div>
      )}
    </div>
  )
}

export default Tile
