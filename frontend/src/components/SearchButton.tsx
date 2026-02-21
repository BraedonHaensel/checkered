type Props = {
  onClick: () => void
}

const SearchButton = ({ onClick }: Props) => {
  return <button onClick={onClick}>Search for Opponent</button>
}

export default SearchButton
