import GameBoard from '../components/GameBoard'
import SearchButton from '../components/SearchButton'
import { Page, PlayerColor } from '../enums'
import type { GameState } from '../game-state'
import { getNewBoardTileStates } from '../game-utils'

const Home = ({ setPage }: { setPage: (page: Page) => void }) => {
    // Array with the state of each of the 32 playable dark tiles in the checkers board
    const gameState: GameState = {
        status: 'SEARCHING',
        tileStates: getNewBoardTileStates(),
        playerColor: PlayerColor.BLACK,
        isYourTurn: false,
    }

    return (
        <div className="space-y-6">
            <h1>CHECKERED</h1>
            <GameBoard gameState={gameState} onPieceMove={() => {}} />
            <SearchButton onClick={() => setPage(Page.GAME)} />
        </div>
    )
}

export default Home
