import GameBoard from '../components/GameBoard'
import { GameDetails } from '../components/GameDetails'
import { PlayerCard } from '../components/PlayerCard'
import { LeaderboardButton, SearchButton } from '../components/Buttons'
import { Page, PlayerColor } from '../enums'
import type { GameState } from '../game-state'
import { getNewBoardTileStates } from '../game-utils'
import MaxSquare from '../components/MaxSquare'

const Home = ({
    setPage,
    username,
}: {
    setPage: (page: Page) => void
    username: string
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
        <div className="grid h-lvh grid-rows-[min-content_1fr] gap-3 px-5 pt-5 md:gap-5">
            <h1 className="text-center text-5xl md:text-6xl">CHECKERED</h1>
            <div className="grid grid-cols-1 grid-rows-[auto-min_content] gap-3 pb-5 md:min-h-130 md:grid-cols-[3fr_2fr] md:grid-rows-1 md:gap-5 lg:grid-cols-[2fr_1fr]">
                <div className="grid min-h-120 grid-rows-[min-content_3fr_min-content] gap-3 md:min-h-full md:gap-5">
                    <PlayerCard
                        player="Opponent"
                        color={PlayerColor.RED}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                    <MaxSquare>
                        <GameBoard
                            gameState={gameState}
                            onPieceMove={() => {}}
                        />
                    </MaxSquare>
                    <PlayerCard
                        player={username}
                        color={PlayerColor.BLACK}
                        captured={0}
                        lost={0}
                        turn={false}
                    />
                </div>
                <GameDetails statusMessage={`Welcome, ${username}`}>
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
