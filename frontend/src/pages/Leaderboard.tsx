import { useEffect, useState } from 'react'
import type { GetLeaderboardResponse } from '../api/request'
import { getLeaderboard } from '../api/api'
import { LeaderboardTable } from '../components/Leaderboard'
import { Page } from '../enums'
import { HomeButton } from '../components/Buttons'
import { GameDetails } from '../components/GameDetails'

const Leaderboard = ({
    setPage,
    username,
}: {
    setPage: (page: Page) => void
    username: string
}) => {
    const [leaderboard, setLeaderboard] =
        useState<GetLeaderboardResponse | null>(null)

    useEffect(() => {
        getLeaderboard().then((result) => setLeaderboard(result))
    }, [setLeaderboard])

    return (
        <div className="grid h-lvh grid-rows-[min-content_1fr] gap-5 p-5">
            <h1 className="text-center">CHECKERED</h1>
            <div className="grid grid-cols-1 grid-rows-[auto_min-content] gap-5 lg:grid-cols-[2fr_1fr] lg:grid-rows-1">
                <div className="flex h-full min-h-[67lvh] w-full flex-col">
                    <h2 className="mx-auto text-3xl">Leaderboard</h2>
                    {leaderboard == null ? (
                        <>
                            <p>Please wait...</p>
                        </>
                    ) : (
                        <>
                            <LeaderboardTable leaderboard={leaderboard} />
                        </>
                    )}
                </div>
                <GameDetails statusMessage={`Welcome, ${username}`}>
                    <HomeButton onClick={() => setPage(Page.HOME)} />
                </GameDetails>
            </div>
        </div>
    )
}

export default Leaderboard
