type Props = {
  isRed: boolean
}

const Piece = ({ isRed }: Props) => {
  return (
    <div
      className={`h-3/4 w-3/4 rounded-full ${isRed ? 'bg-red-500' : 'bg-black-500'}`}
    ></div>
  )
}

export default Piece
