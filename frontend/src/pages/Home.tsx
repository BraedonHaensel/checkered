import GameBoard from '../components/GameBoard'
import SearchButton from '../components/SearchButton'
import { Page, PlayerColor } from '../enums'
import { getNewBoardTileStates } from '../utils'

const Home = ({ setPage }: { setPage: (page: Page) => void }) => {
    // Array with the state of each of the 32 playable dark tiles in the checkers board
    const tileStates = getNewBoardTileStates()

    return (
        <div className="space-y-6">
            <h1>CHECKERED</h1>
            <GameBoard
                tileStates={tileStates}
                playerColor={PlayerColor.BLACK}
                isYourTurn={false}
                onPieceMove={() => {}}
            />
            <SearchButton onClick={() => setPage(Page.GAME)} />
        </div>
    )
}

export default Home
