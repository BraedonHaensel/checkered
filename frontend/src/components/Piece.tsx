import { Crown } from 'lucide-react'

type Props = {
  isBlack: boolean
  isKing: boolean
  showMovableHighlight: boolean
  showSelectedHighlight: boolean
}

// Checkers piece
const Piece = ({
  isBlack,
  isKing,
  showMovableHighlight,
  showSelectedHighlight,
}: Props) => {
  return (
    <div
      className={
        `flex h-3/4 w-3/4 items-center justify-center rounded-full ` +
        `${isBlack ? 'bg-black' : 'bg-red-500'} ` +
        `${showSelectedHighlight ? 'border-4 border-gray-300' : showMovableHighlight && 'border-3 border-yellow-500'}`
      }
    >
      {isKing && <Crown size={32} color="gold" />}
    </div>
  )
}

export default Piece
