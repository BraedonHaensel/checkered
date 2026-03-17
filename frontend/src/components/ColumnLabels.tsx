import { cn } from '../lib/utils'

type Props = {
    reverse?: boolean
    className?: string
}

// Labels for displaying the column numbers
export const ColumnLabels = ({ reverse = false, className = '' }: Props) => {
    let letters = Array.from({ length: 8 }).map((_, i: number) =>
        String.fromCharCode(65 + i)
    )
    if (reverse) letters = letters.reverse()

    return (
        <div className={cn('grid grid-cols-8', className)}>
            {letters.map((letter) => (
                <span key={letter} className="m-auto mb-0 text-xl">
                    {letter}
                </span>
            ))}
        </div>
    )
}
