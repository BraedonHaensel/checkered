import { useEffect, useState } from "react"
import type { GetLeaderboardResponse } from "../api/request"
import { getLeaderboard } from "../api/api";
import { LeaderboardTable } from "../components/Leaderboard";
import { Page } from "../enums";
import { DashboardButton } from "../components/Buttons";
import { IngameDetails } from "../components/GameDetails";

const Leaderboard = ({setPage, user}: {setPage: (page: Page) => void, user: string}) => {
    const [leaderboard, setLeaderboard] = useState<GetLeaderboardResponse | null>(null);

    useEffect(() => {
        getLeaderboard().then(result => setLeaderboard(result))
    }, [setLeaderboard])


    return <div className="h-lvh grid grid-rows-[min-content_1fr] justify-center items-center pt-5">
        <h1 className='w-full text-center'>CHECKERED</h1>
        <div className="w-[100vw] grid grid-cols-[1fr] grid-rows-[auto_min-content] md:grid-cols-[2fr_1fr] md:grid-rows-[1fr] items-center md:h-full">
            <div className="flex flex-col min-h-[67lvh] items-center h-full w-full">
                <h2 className="text-3xl">Leaderboard</h2>
                {leaderboard == null ? <>
                    <p>Please wait...</p>
                </> : <>
                        <LeaderboardTable leaderboard={leaderboard}/>
                    </>}
            </div>
            <IngameDetails statusMessage={`Welcome, ${user}`}>
                <DashboardButton onClick={() => setPage(Page.HOME)} exitsGame={false} />
            </IngameDetails>
        </div>
    </div>
}

export default Leaderboard;
