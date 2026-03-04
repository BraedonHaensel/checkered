import { useEffect, useState } from "react"
import type { GetLeaderboardResponse } from "../api/request"
import { getLeaderboard } from "../api/api";
import { LeaderboardTable } from "../components/Leaderboard";

const Leaderboard = () => {
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
    </>
}

export default Leaderboard;
