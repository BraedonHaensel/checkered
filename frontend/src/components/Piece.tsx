type Props = {
  isBlack: boolean
}

const Piece = ({ isBlack }: Props) => {
  return (
    <div
      className={`h-3/4 w-3/4 rounded-full ${isBlack ? 'bg-black' : 'bg-red-500'}`}
    ></div>
  )
}

export default Piece
