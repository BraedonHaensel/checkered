import type { GetLeaderboardResponse } from '../api/request'

export const LeaderboardTable = ({
    leaderboard,
}: {
    leaderboard: GetLeaderboardResponse['board'] | null
}) => {
    return (
        <table className="text-left">
            <thead>
                <tr className="border-b">
                    <th className="w-1/3">Username</th>
                    <th>Wins</th>
                    <th>Losses</th>
                </tr>
            </thead>
            <tbody>
                {(leaderboard || []).map((row, i) => (
                    <tr key={i}>
                        <td>{row.username}</td>
                        <td>{row.wins}</td>
                        <td>{row.losses}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    )
}
