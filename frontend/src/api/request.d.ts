/**
    Requests are messages that are sent over HTTP
    and correspond to an /api/xxx endpoint
*/
interface GenericRequest<T> {
    type: T
}

export type GetLeaderboardRequest = GenericRequest<'leaderboard'>

export interface GetLeaderboardResponse extends GenericRequest<'leaderboard'> {
    board: { username: string; wins: number; losses: number }[]
}

export type Request = GetLeaderboardRequest
export type Response = GetLeaderboardResponse
