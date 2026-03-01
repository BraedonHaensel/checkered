import { MoveDirection, PlayerColor, TileState } from './enums'

/**
 * Get the starting tile states for each of the 32 playable dark tiles in a checkers board.
 */
export const getNewBoardTileStates = () => {
  // Only 32 tiles are playable in checkers. The first 12 start with black pieces, the middle 8
  // are empty, and the last 12 start with red pieces
  const tileStates = [
    ...Array(12).fill(TileState.RED_STANDARD_PIECE),
    ...Array(8).fill(TileState.EMPTY),
    ...Array(12).fill(TileState.BLACK_STANDARD_PIECE),
  ]
  return tileStates
}

/**
 * Checks if the given tileState is for a black piece (standard or king).
 */
export const isBlackPiece = (tileState: TileState) => {
  return [TileState.BLACK_STANDARD_PIECE, TileState.BLACK_KING_PIECE].includes(
    tileState
  )
}

/**
 * Checks if the given tileState is for a king piece.
 */
export const isKingPiece = (tileState: TileState) => {
  return [TileState.BLACK_KING_PIECE, TileState.RED_KING_PIECE].includes(
    tileState
  )
}

/**
 * Returns the row number given a piece's tile index.
 */
export const tileIndexToRow = (tileIndex: number) => {
  // 4 playable tiles per row
  return Math.floor(tileIndex / 4)
}

/**
 * Returns the column number given a piece's tile index.
 */
export const tileIndexToCol = (tileIndex: number) => {
  // 4 playable columns per row, columns spaced 2 tiles apart
  const col = (tileIndex % 4) * 2
  // Playable columns are offset one tile to the right every second row
  const offset = tileIndexToRow(tileIndex) % 2 === 0 ? 1 : 0
  return col + offset
}

// Check if a move is upward, otherwise it is downward.
export const isUpwardMoveDirection = (moveDirection: MoveDirection) => {
  return [MoveDirection.UP_LEFT, MoveDirection.UP_RIGHT].includes(moveDirection)
}

// Check if a move is leftward, otherwise it is upward.
export const isLeftwardMoveDirection = (moveDirection: MoveDirection) => {
  return [MoveDirection.UP_LEFT, MoveDirection.DOWN_LEFT].includes(
    moveDirection
  )
}

// Checks if a move from the source to dest index is a jump move
export const isJumpMove = (sourceIndex: number, destIndex: number) => {
  // Only jump moves can change a tile index by 7 or more
  return Math.abs(destIndex - sourceIndex) >= 7
}

// Gets the tile index change amount required to move a piece in the given direction.
const getMoveAmountForDirection = (
  direction: MoveDirection,
  isOffsetRow: boolean,
  isDoubleJump: boolean = false
) => {
  if (isUpwardMoveDirection(direction)) {
    if (isLeftwardMoveDirection(direction)) {
      // Up left move
      return isDoubleJump ? -9 : isOffsetRow ? -4 : -5
    }
    // Up right move
    return isDoubleJump ? -7 : isOffsetRow ? -3 : -4
  }
  if (isLeftwardMoveDirection(direction)) {
    // Down left move
    return isDoubleJump ? 7 : isOffsetRow ? 4 : 3
  }
  // Down right move
  return isDoubleJump ? 9 : isOffsetRow ? 5 : 4
}

// Gets the destination tile index for a move in the given direction, or undefined if the piece
// can't move in the given direction
const getMoveDestinationInDirection = (
  tileIndex: number,
  playerColor: PlayerColor,
  tileStates: TileState[],
  direction: MoveDirection
): number | undefined => {
  // Check row boundaries
  const row = tileIndexToRow(tileIndex)
  const isUpward = isUpwardMoveDirection(direction)
  if (isUpward && row === 0) return undefined // Can't move upward from the top row
  if (!isUpward && row === 7) return undefined // Can't move downward from the bottom row

  // Check column boundaries
  const col = tileIndexToCol(tileIndex)
  const isLeftward = isLeftwardMoveDirection(direction)
  if (isLeftward && col === 0) return undefined // Can't move left from the leftmost col
  if (!isLeftward && col === 7) return undefined // Can't move right from the rightmost col

  // Check if this is an offset row (rows where the dark tiles are shifted right by one tile).
  // The amount to change the tile index by depends on whether the piece is from an offset row
  const isOffsetRow = row % 2 === 0

  // Get the tile index change amount for a move in this direction
  let moveAmount = getMoveAmountForDirection(direction, isOffsetRow)
  let destIndex = tileIndex + moveAmount

  // Check if the destination tile is empty
  if (tileStates[destIndex] === TileState.EMPTY) return destIndex

  // Check the player and piece colors
  const isBlackPlayer = playerColor === PlayerColor.BLACK
  const isJumpingBlackPiece = isBlackPiece(tileStates[destIndex])

  // Can't jump your own piece
  if (isBlackPlayer && isJumpingBlackPiece) return undefined
  if (!isBlackPlayer && !isJumpingBlackPiece) return undefined

  // Get the tile index change amount for a double jump move in this direction
  moveAmount = getMoveAmountForDirection(direction, isOffsetRow, true)

  // Check if the destination tile is empty
  destIndex = tileIndex + moveAmount
  if (tileStates[destIndex] === TileState.EMPTY) return destIndex

  // No possible moves
  return undefined
}

/**
 * Get an array of the possible move destinations for the piece at the given tile index.
 */
export const getPieceMoveDestinations = (
  tileIndex: number,
  playerColor: PlayerColor,
  tileStates: TileState[]
): number[] => {
  const tileState = tileStates[tileIndex]

  // Source tile must not be empty
  if (tileState === TileState.EMPTY) return []

  // Player must own the piece
  if (playerColor === PlayerColor.BLACK) {
    if (!isBlackPiece(tileState)) return []
  } else {
    if (isBlackPiece(tileState)) return []
  }

  // Move directions to try
  const moveDirections: MoveDirection[] = []
  if (isBlackPiece(tileState) || isKingPiece(tileState)) {
    // Try moving upwards
    moveDirections.push(MoveDirection.UP_LEFT, MoveDirection.UP_RIGHT)
  }
  if (!isBlackPiece(tileState) || isKingPiece(tileState)) {
    // Try moving downwards
    moveDirections.push(MoveDirection.DOWN_LEFT, MoveDirection.DOWN_RIGHT)
  }

  // Get the valid move destinations
  const moveDestinations = []
  for (const moveDirection of moveDirections) {
    const destIndex = getMoveDestinationInDirection(
      tileIndex,
      playerColor,
      tileStates,
      moveDirection
    )
    if (destIndex !== undefined) moveDestinations.push(destIndex)
  }
  return moveDestinations
}

/**
 * Gets an array of arrays of valid move destinations for each tile/piece.
 */
export const getMoveDestinations = (
  tileStates: TileState[],
  playerColor: PlayerColor,
  isYourTurn: boolean
): number[][] => {
  if (!isYourTurn)
    // Not your turn so no possible moves, fill with empty arrays
    return Array.from({ length: tileStates.length }, () => [])

  // Get the move destinations for each tile index
  let moveDestinations: number[][] = []
  for (let tileIndex = 0; tileIndex < tileStates.length; tileIndex++) {
    moveDestinations.push(
      getPieceMoveDestinations(tileIndex, playerColor, tileStates)
    )
  }

  // Force jump moves if available
  const containsJumpMove =
    moveDestinations.find((destinations, tileIndex) =>
      destinations.find((destIndex) => isJumpMove(tileIndex, destIndex))
    ) !== undefined
  if (containsJumpMove) {
    moveDestinations = moveDestinations.map((destinations, tileIndex) =>
      destinations.filter((destIndex) => isJumpMove(tileIndex, destIndex))
    )
  }

  return moveDestinations
}
