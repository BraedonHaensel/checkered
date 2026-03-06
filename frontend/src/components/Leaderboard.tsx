import type { GetLeaderboardResponse } from "../api/request";

export const LeaderboardTable = ({leaderboard}: {leaderboard: GetLeaderboardResponse}) => {
    return <table className="w-full text-left table-auto min-w-max">
        <thead>
            <tr className="border-b">
                <th>Username</th>
                <th>Wins</th>
                <th>Losses</th>
            </tr>  
        </thead>
        <tbody>
            {(leaderboard.board || []).map((row, i) => <tr key={i}>
                <td>{row.username}</td>
                <td>{row.wins}</td>
                <td>{row.losses}</td>
            </tr>)}
        </tbody>
    </table>
}
