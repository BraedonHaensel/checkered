import { cn } from '../lib/utils'

type Props = {
    reverse?: boolean
    className?: string
}

// Labels for displaying the row numbers
export const RowLabels = ({ reverse = false, className = '' }: Props) => {
    let numbers = Array.from({ length: 8 }).map((_, i: number) => i + 1)
    if (!reverse) numbers = numbers.reverse()

    return (
        <div className={cn('grid grid-rows-8', className)}>
            {numbers.map((number) => (
                <span className="m-auto ml-0 text-xl">{number}</span>
            ))}
        </div>
    )
}
