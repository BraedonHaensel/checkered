import Piece from './Piece'

type Props = {
  isDark: boolean
}

const Tile = ({ isDark }: Props) => {
  if (!isDark) {
    // Light tiles are never interacted with in a game of checkers
    return <div className="bg-[#FFE4C4]"></div>
  }

  return (
    <div className="flex items-center justify-center bg-[#A0522D]">
      <Piece isRed />
    </div>
  )
}

export default Tile
