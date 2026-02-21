type Props = {
  isRed: boolean
}

const Piece = ({ isRed }: Props) => {
  return (
    <div
      className={`h-3/4 w-3/4 rounded-full ${isRed ? 'bg-red-500' : 'bg-black'}`}
    ></div>
  )
}

export default Piece
