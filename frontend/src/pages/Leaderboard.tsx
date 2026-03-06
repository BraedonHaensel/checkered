import { useEffect, useState } from "react"
import type { GetLeaderboardResponse } from "../api/request"
import { getLeaderboard } from "../api/api";
import { LeaderboardTable } from "../components/Leaderboard";
import { ReturnToDashboardButton } from "../components/ReturnToDashboardButton";
import type { Page } from "../enums";

const Leaderboard = ({setPage}: {setPage: (page: Page) => void}) => {
    const [leaderboard, setLeaderboard] = useState<GetLeaderboardResponse | null>(null);

    useEffect(() => {
        getLeaderboard().then(result => setLeaderboard(result))
    }, [setLeaderboard])

    return <>
        <h1>CHECKERED</h1>
        <h2 className="text-3xl">Leaderboard</h2>
        {leaderboard == null ? <>
            <p>Please wait...</p>
        </> : <>
            <LeaderboardTable leaderboard={leaderboard}/>
        </>}
        <ReturnToDashboardButton setPage={setPage} />
    </>
}

export default Leaderboard;
