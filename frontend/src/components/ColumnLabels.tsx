type Props = {
    reverse?: boolean
}

// Labels for displaying the column numbers
export const ColumnLabels = ({ reverse = false }: Props) => {
    let letters = Array.from({ length: 8 }).map((_, i: number) =>
        String.fromCharCode(65 + i)
    )
    if (reverse) letters = letters.reverse()

    return (
        <div className="grid w-full grid-cols-8 py-2">
            {letters.map((letter) => (
                <span className="text-center text-xl">{letter}</span>
            ))}
        </div>
    )
}
