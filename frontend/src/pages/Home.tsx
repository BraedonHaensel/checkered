import GameBoard from '../components/GameBoard'
import { IngameDetails } from '../components/GameDetails'
import { PlayerCard } from '../components/PlayerCard'
import { LeaderboardButton, SearchButton } from '../components/Buttons'
import { Page, PlayerColor } from '../enums'
import type { GameState } from '../game-state'
import { getNewBoardTileStates } from '../game-utils'

const Home = ({
    setPage,
    user,
}: {
    setPage: (page: Page) => void
    user: string
}) => {
    // Array with the state of each of the 32 playable dark tiles in the checkers board
    const gameState: GameState = {
        tileStates: getNewBoardTileStates(),
        playerColor: PlayerColor.BLACK,
        isYourTurn: false,
        previousMoves: [],
        opponent: 'Opponent',
    }

    return (
        <div className="grid h-lvh grid-rows-[min-content_1fr] items-center justify-center pt-5">
            <h1 className="w-full text-center">CHECKERED</h1>
            <div className="grid w-[100vw] grid-cols-[1fr] grid-rows-[auto_min-content] items-center lg:h-full lg:grid-cols-[2fr_1fr] lg:grid-rows-[1fr]">
                <div className="flex h-full min-h-[67lvh] w-full flex-col items-center">
                    <PlayerCard
                        player="Opponent"
                        color={PlayerColor.RED}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                    <GameBoard gameState={gameState} onPieceMove={() => {}} />
                    <PlayerCard
                        player={user}
                        color={PlayerColor.BLACK}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                </div>
                <IngameDetails statusMessage={`Welcome, ${user}`}>
                    <SearchButton onClick={() => setPage(Page.GAME)} />
                    <LeaderboardButton
                        onClick={() => setPage(Page.LEADERBOARD)}
                    />
                </IngameDetails>
            </div>
        </div>
    )
}

export default Home
