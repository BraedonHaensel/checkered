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
                {(leaderboard || [])
                    .sort((a, b) => {
                        // Sort by most wins
                        if (a.wins != b.wins) return b.wins - a.wins
                        // If tied, sort by least losses
                        return a.losses - b.losses
                    })
                    .map((row, i) => (
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
