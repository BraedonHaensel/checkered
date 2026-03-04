import { Crown } from 'lucide-react'
import React from 'react'

type Props = {
    isBlack: boolean
    isKing: boolean
    showSelectableHighlight: boolean
    showSelectedHighlight: boolean
}

// Checkers piece
const Piece = React.memo(
    ({
        isBlack,
        isKing,
        showSelectableHighlight,
        showSelectedHighlight,
    }: Props) => {
        return (
            <div
                className={
                    `flex h-3/4 w-3/4 items-center justify-center rounded-full ` +
                    `${isBlack ? 'bg-black' : 'bg-red-500'} ` +
                    `${showSelectedHighlight ? 'border-4 border-gray-300' : showSelectableHighlight && 'border-3 border-yellow-500'}`
                }
            >
                {isKing && <Crown size={32} color="gold" />}
            </div>
        )
    }
)

export default Piece
