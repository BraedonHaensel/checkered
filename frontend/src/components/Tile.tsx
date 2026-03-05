import React from 'react'
import { PlayerColor, TileState } from '../enums'
import { isBlackPiece, isKingPiece } from '../game-utils'
import Piece from './Piece'

type LightTileProps = {
    tileIndex?: never
    tileState?: never
    playerColor?: never
    canSelectPiece?: never
    isPieceSelected?: never
    isMoveDestination?: never
    showPreviousMoveHighlight?: never
    onClick?: never
}

type DarkTileProps = {
    tileIndex: number
    tileState: TileState
    playerColor: PlayerColor
    canSelectPiece: boolean
    isPieceSelected: boolean
    isMoveDestination: boolean
    showPreviousMoveHighlight: boolean
    onClick: ((tileIndex: number) => void) | undefined
}

type Props = LightTileProps | DarkTileProps

// Individual tile in a checkers board
const Tile = React.memo(
    ({
        tileIndex,
        tileState,
        playerColor,
        canSelectPiece,
        isPieceSelected,
        isMoveDestination,
        showPreviousMoveHighlight,
        onClick,
    }: Props) => {
        if (tileIndex === undefined) {
            // Light tiles are not playable in a game of checkers, no further logic required
            return <div className="bg-[#FFE4C4]"></div>
        }

        // Check if the tile contains a piece
        const hasPiece = tileState !== TileState.EMPTY
        if (hasPiece && isMoveDestination) {
            throw new Error('Tile has both a piece and a move destination')
        }

        return (
            <div
                className={
                    'flex items-center justify-center bg-[#A0522D] ' +
                    // The board is rotated 180 for the red player's POV, so counter-rotate each tile back 180
                    `${playerColor === PlayerColor.RED && 'rotate-180'}`
                }
                onClick={() => onClick?.(tileIndex)}
            >
                {/* Previous move indicator */}
                <div
                    className={
                        'flex h-full w-full items-center justify-center ' +
                        `${showPreviousMoveHighlight ? 'bg-[#DFFFC4]/35' : ''}`
                    }
                >
                    {/* Display a checkers piece */}
                    {hasPiece && (
                        <Piece
                            isBlack={isBlackPiece(tileState)}
                            isKing={isKingPiece(tileState)}
                            showSelectableHighlight={canSelectPiece}
                            showSelectedHighlight={isPieceSelected}
                        />
                    )}
                    {/* Display a move destination indicator */}
                    {isMoveDestination && (
                        <div className="h-1/3 w-1/3 rounded-full bg-gray-300 opacity-70"></div>
                    )}
                    <i>{tileIndex}</i>
                </div>
            </div>
        )
    }
)

export default Tile
