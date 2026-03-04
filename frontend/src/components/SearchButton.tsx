type Props = {
    onClick: () => void
}

// Search for opponent button
const SearchButton = ({ onClick }: Props) => {
    return <button onClick={onClick}>Search for Opponent</button>
}

export default SearchButton
