import GameBoard from '../components/GameBoard'
import { IngameDetails } from '../components/GameDetails'
import { PlayerCard } from '../components/PlayerCard'
import { LeaderboardButton, SearchButton } from '../components/Buttons'
import { Page, PlayerColor } from '../enums'
import type { GameState } from '../game-state'
import { getNewBoardTileStates } from '../game-utils'

const Home = ({ setPage, user }: { setPage: (page: Page) => void, user: string }) => {
    // Array with the state of each of the 32 playable dark tiles in the checkers board
    const gameState: GameState = {
        tileStates: getNewBoardTileStates(),
        playerColor: PlayerColor.BLACK,
        isYourTurn: false,
        previousMove: undefined,
    }

    return <div className="h-lvh grid grid-rows-[min-content_1fr] justify-center items-center pt-5">
        <h1 className='w-full text-center'>CHECKERED</h1>
        <div className="w-[100vw] grid grid-cols-[1fr] grid-rows-[1fr_min-content] md:grid-cols-[2fr_1fr] md:grid-rows-[1fr] items-center h-full">
            <div className="flex flex-col items-center h-full w-full">
                <PlayerCard player="Opponent" color={PlayerColor.RED} captured={0} lost={0}/>
                <GameBoard gameState={gameState} onPieceMove={() => {}} />
                <PlayerCard player={user} color={PlayerColor.BLACK}  captured={0} lost={0} />
            </div>
            <IngameDetails statusMessage={`Welcome, ${user}`}>
                <SearchButton onClick={() => setPage(Page.GAME)} />
                <LeaderboardButton onClick={() => setPage(Page.LEADERBOARD)} />
            </IngameDetails>
        </div>
    </div>
}

export default Home
