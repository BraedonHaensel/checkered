import { Backend } from "./backend"
import type { GetLeaderboardResponse } from "./request"

export const getLeaderboard = async (): Promise<GetLeaderboardResponse> => {
    return await Backend.instance().get("leaderboard", {})
}
