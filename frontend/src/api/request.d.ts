/**
    Requests are messages that are sent over HTTP
    and correspond to an /api/xxx endpoint
*/
interface GenericRequest<T> {
    type: T
};

export interface GetLeaderboardRequest extends GenericRequest<"get_leaderboard"> {}

export interface GetLeaderboardResponse extends GenericRequest<"get_leaderboard"> {
    leaderboard: {name: string, wl_ratio: number}[]
}

export type Request = GetLeaderboardRequest;
export type Response = GetLeaderboardResponse;
