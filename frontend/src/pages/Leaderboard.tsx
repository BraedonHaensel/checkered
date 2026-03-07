import { useEffect, useState } from 'react'
import type { GetLeaderboardResponse } from '../api/request'
import { getLeaderboard } from '../api/api'
import { LeaderboardTable } from '../components/Leaderboard'
import { Page } from '../enums'
import { DashboardButton } from '../components/Buttons'
import { IngameDetails } from '../components/GameDetails'

const Leaderboard = ({
    setPage,
    user,
}: {
    setPage: (page: Page) => void
    user: string
}) => {
    const [leaderboard, setLeaderboard] =
        useState<GetLeaderboardResponse | null>(null)

    useEffect(() => {
        getLeaderboard().then((result) => setLeaderboard(result))
    }, [setLeaderboard])

    return (
        <div className="grid h-lvh grid-rows-[min-content_1fr] items-center justify-center pt-5">
            <h1 className="w-full text-center">CHECKERED</h1>
            <div className="grid w-screen grid-cols-[1fr] grid-rows-[auto_min-content] items-center lg:h-full lg:grid-cols-[2fr_1fr] lg:grid-rows-[1fr]">
                <div className="flex h-full min-h-[67lvh] w-full flex-col items-center p-5">
                    <h2 className="text-3xl">Leaderboard</h2>
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
                <IngameDetails statusMessage={`Welcome, ${user}`}>
                    <DashboardButton
                        onClick={() => setPage(Page.HOME)}
                        exitsGame={false}
                    />
                </IngameDetails>
            </div>
        </div>
    )
}

export default Leaderboard
