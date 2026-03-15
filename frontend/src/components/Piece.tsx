import { cn } from '../lib/utils'
import style from './Piece.module.css'
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
                className={cn(
                    'flex h-3/4 w-3/4 items-center justify-center rounded-full',
                    isBlack ? 'bg-black' : 'bg-red-500',
                    style.piece,
                    isBlack ? style.black : style.red,
                    isKing && style.king,
                    showSelectedHighlight
                        ? style.selected
                        : showSelectableHighlight && style.movable
                )}
            />
        )
    }
)

export default Piece
