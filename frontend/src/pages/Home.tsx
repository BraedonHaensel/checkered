import GameBoard from '../components/GameBoard'
import { GameDetails } from '../components/GameDetails'
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
        draw_requested: false,
    }

    return (
        <div className="grid h-lvh grid-rows-[min-content_1fr] gap-5 p-5 lg:min-h-180">
            <h1 className="text-center">CHECKERED</h1>
            <div className="grid grid-cols-1 grid-rows-[auto-min_content] gap-5 md:grid-cols-[3fr_2fr] md:grid-rows-1 lg:grid-cols-[2fr_1fr]">
                <div className="grid grid-rows-[min-content_3fr_min-content] gap-5">
                    <PlayerCard
                        player="Opponent"
                        color={PlayerColor.RED}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                    <div className="m-auto flex aspect-square max-h-90 w-full justify-center md:max-h-full md:max-w-full lg:aspect-auto lg:h-full">
                        <GameBoard
                            gameState={gameState}
                            onPieceMove={() => {}}
                        />
                    </div>
                    <PlayerCard
                        player={user}
                        color={PlayerColor.BLACK}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                </div>
                <GameDetails statusMessage={`Welcome, ${user}`}>
                    <SearchButton onClick={() => setPage(Page.GAME)} />
                    <LeaderboardButton
                        onClick={() => setPage(Page.LEADERBOARD)}
                    />
                </GameDetails>
            </div>
        </div>
    )
}

export default Home
