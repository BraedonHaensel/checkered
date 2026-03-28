import { sleep } from '../lib/utils'
import type { Request, Response } from './request'
import { Session } from './session'

export type BackendServer = {
    url: string
    id: number
}

const getNameServerAddress = async (): Promise<string> => {
    const ns = import.meta.env.APP_NAMESERVER_URL;
    console.log("Using nameserver located at:", ns)
    if(ns)
        return ns
    return "http://localhost:9000"
}

const nameServer: BackendServer = {
    url: await getNameServerAddress(),
    id: 0,
}

// Join queue response from Matchmaker
type JoinQueueResponse = {
    type: 'SUCCESS' | 'ALREADY_IN_QUEUE'
}

// Leave queue response from Matchmaker.
type LeaveQueueResponse = {
    type: 'SUCCESS' | 'ALREADY_NOT_IN_QUEUE'
}

// Interval to poll for the Matchmaker queue status
const QUEUE_POLL_INTERVAL = 2000

// Poll response type from Matchmaker
type PollResponse = {
    type: 'IN_GAME' | 'IN_QUEUE' | 'NOT_IN_QUEUE'
    game_server?: BackendServer
    match_id?: string
}

export class Backend {
    private static _instance: Backend | null = null

    private servers: BackendServer[] = []
    private current: BackendServer | null = null

    static instance(): Backend {
        if (Backend._instance === null) {
            Backend._instance = new Backend()
        }

        return Backend._instance
    }

    private constructor() {}

    public async get<RequestType extends Request['type']>(
        type: RequestType,
        payload: Omit<Extract<Request, { type: RequestType }>, 'type'>
    ): Promise<Extract<Response, { type: RequestType }>> {
        const url = `${(await this.server()).url}/${type}`
        let raw: globalThis.Response | undefined = undefined
        try {
            raw = await fetch(url, {
                body:
                    Object.keys(payload).length > 0
                        ? JSON.stringify({ ...payload, type })
                        : null,
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'GET',
            })
        } catch {
            this.handleServerError()
            return this.get(type, payload)
        }

        return (await raw.json()) as Extract<Response, { type: RequestType }>
    }

    public async post<RequestType extends Request['type']>(
        type: RequestType,
        payload: Omit<Extract<Request, { type: RequestType }>, 'type'>
    ): Promise<Extract<Response, { type: RequestType }>> {
        const url = `${(await this.server()).url}/${type}`
        let raw: globalThis.Response | undefined = undefined
        try {
            raw = await fetch(url, {
                body:
                    Object.keys(payload).length > 0
                        ? JSON.stringify({ ...payload, type })
                        : null,
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'POST',
            })
        } catch {
            this.handleServerError()
            return this.post(type, payload)
        }

        return (await raw.json()) as Extract<Response, { type: RequestType }>
    }

    // Get the list of Matchmakers from the Name Server
    private async populateServerList() {
        try {
            const raw = await fetch(nameServer.url + '/matchmakers')
            const matchmakers = (await raw.json()) as BackendServer[]
            this.servers = matchmakers
            console.log('Matchmakers:', matchmakers)
        } catch (e) {
            if (e instanceof TypeError) {
                // Name Server connection down, try again
                console.error('Name Server connection failed:', e)
                await sleep(2000)
                await this.populateServerList()
            } else {
                console.error('Failed to populate the Matchmakers list:', e)
            }
        }
    }

    // Get a Matchmaker server
    public async server(): Promise<BackendServer> {
        if (this.current !== null) {
            return this.current
        }

        // Populate the list of Matchmakers
        if (this.servers.length === 0) {
            console.log('Populating the Matchmakers list...')
            await this.populateServerList()
            while (this.servers.length === 0) {
                console.error('No Matchmakers found')
                await sleep(2000)
                await this.populateServerList()
            }
        }

        const best = this.servers.shift()

        if (best === undefined) {
            throw new Error('Could not find any backend servers!')
        }

        this.current = best

        return this.current
    }

    public handleServerError() {
        // Stop using the current backend server
        this.current = null
    }

    /**
     * Sends a request to the Matchmaker for a new Game Server.
     * @param gameServerUrl URL of the original Game Server that has crashed.
     * @param matchId ID of the player's match.
     */
    private async sendNewGameServerRequest(
        gameServer: BackendServer,
        matchId: string
    ) {
        try {
            console.log('Requesting a new Game Server...')
            const raw = await fetch(
                `${(await this.server()).url}/match/request-new-game-server`,
                {
                    body: JSON.stringify({
                        old_game_server: gameServer,
                        match_id: matchId,
                    }),
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    method: 'POST',
                }
            )
            if (!raw.ok) {
                console.error(
                    'Failed to request a new Game Server:',
                    await raw.text()
                )
                return
            }
            const newGameServerRes = await raw.json()
            console.log('New Game Server response:', newGameServerRes)
        } catch (e) {
            console.error('Failed to request a new Game Server:', e)
        }
    }

    /**
     * Sends a request to add a user to the queue.
     * @param username Username of the user leaving the queue.
     */
    private async sendJoinQueueRequest(username: string) {
        try {
            console.log('Joining the queue...')
            const raw = await fetch(`${(await this.server()).url}/queue/add`, {
                body: JSON.stringify({
                    username: username,
                }),
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'POST',
            })
            if (!raw.ok) {
                console.error('Failed to join the queue:', await raw.text())
                return
            }

            const joinQueueRes: JoinQueueResponse = await raw.json()
            if (!joinQueueRes.type)
                throw Error('Invalid response: missing type')
            if (joinQueueRes.type === 'ALREADY_IN_QUEUE') {
                console.log('Already in the queue')
            } else if (joinQueueRes.type === 'SUCCESS') {
                console.log('Successfully joined the queue!')
            }
        } catch (e) {
            if (e instanceof TypeError) {
                // Matchmaker connection down, try again with a new server
                console.error('Matchmaker connection failed:', e)
                this.handleServerError()
                await sleep(2000)
                await this.sendJoinQueueRequest(username)
            } else {
                console.error('Failed to join the queue:', e)
            }
        }
    }

    /**
     * Sends a poll request for the user's current queueing status.
     * @param username Username of the user to poll.
     * @returns A promise that resolves to a poll response, or undefined if an
     * unexpected error occurs.
     */
    private async sendPollRequest(
        username: string
    ): Promise<PollResponse | undefined> {
        try {
            console.log('Polling the Matchmaker...')
            const raw = await fetch(
                `${(await this.server()).url}/queue/poll?username=${username}`,
                {
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    method: 'POST',
                }
            )
            const pollRes: PollResponse = await raw.json()
            if (!pollRes.type) throw Error('Invalid response: missing type')
            console.log('Poll response:', pollRes)
            return pollRes
        } catch (e) {
            if (e instanceof TypeError) {
                // Matchmaker connection down, try again with a new server
                console.error('Matchmaker connection failed:', e)
                this.handleServerError()
                await sleep(2000)
                return await this.sendPollRequest(username)
            } else {
                console.error(
                    "Failed to poll for the user's queueing status:",
                    e
                )
            }
        }
    }

    private async connectSession(
        session: Session,
        username: string,
        joinQueue = true
    ) {
        // Add the user to the matchmaking queue
        if (joinQueue) await this.sendJoinQueueRequest(username)

        session.interval = 0
        let pollInProgress = false

        // Function called periodically to poll the Matchmaker for the queueing status
        const pollMatchmaker = async () => {
            // Skip if a concurrent poll is in progress
            if (pollInProgress) return

            // Poll the matchmaker for the current queueing status
            pollInProgress = true
            const pollRes = await this.sendPollRequest(username)
            if (!pollRes) return

            // Check if the player was not in the queue or a game
            if (pollRes.type === 'NOT_IN_QUEUE') {
                console.log('Not in the queue! Rejoining...')
                this.connectSession(session, username)
                return
            }

            // Check if a game is found
            if (pollRes.type === 'IN_GAME') {
                // Stop the polling interval
                clearInterval(session.interval)

                // Game found, connect to the Game Server using a WebSocket
                const wsUrl = `${pollRes.game_server!.url}/ws`
                const ws = new WebSocket(wsUrl)
                console.log(
                    `Attempting WebSocket connection to Game Server: ${wsUrl}`
                )

                ws.addEventListener('close', async (ev: CloseEvent) => {
                    if (!ev.wasClean) {
                        console.log(
                            'Unclean socket close. Game Server likely crashed!'
                        )

                        // Request a new Game Server
                        await this.sendNewGameServerRequest(
                            pollRes.game_server!,
                            pollRes.match_id!
                        )

                        // Attempt a new connection
                        this.connectSession(session, username, false)
                    }
                })

                ws.addEventListener('error', (err) => {
                    console.log('Error Callback', err.type)
                    console.error(err)
                    ws.close()
                })

                ws.addEventListener('open', () => {
                    // Socket connection successful, continue connection setup
                    session.connect(ws, username)
                })
            }

            // Stop blocking concurrent polls
            pollInProgress = false
        }

        // Start periodically polling the Matchmaker for the queueing status
        session.interval = setInterval(pollMatchmaker, QUEUE_POLL_INTERVAL)
        pollMatchmaker()
    }

    /**
     * Sends a request to remove a user from the queue.
     * @param username Username of the user leaving the queue.
     */
    private async sendLeaveQueueRequest(username: string) {
        try {
            console.log('Leaving the queue...')
            const raw = await fetch(
                `${(await this.server()).url}/queue/leave`,
                {
                    body: JSON.stringify({
                        username: username,
                    }),
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    method: 'POST',
                }
            )
            if (!raw.ok) {
                console.error('Failed to leave queue:', await raw.text())
                return
            }

            const leaveQueueRes: LeaveQueueResponse = await raw.json()
            if (!leaveQueueRes.type)
                throw Error('Invalid response: missing type')
            if (leaveQueueRes.type === 'ALREADY_NOT_IN_QUEUE') {
                console.log('Already not in the queue')
            } else if (leaveQueueRes.type === 'SUCCESS') {
                console.log('Successfully left the queue!')
            }
        } catch (e) {
            console.error('Failed to leave the queue:', e)
            return
        }
    }

    /**
     * Cancels any ongoing queueing.
     * @param session Session to close.
     * @param username Username of the user for the session.
     */
    public cancelQueueing(session: Session, username: string) {
        // Cancel polling
        clearInterval(session.interval)
        console.log('Cancelled polling the queue')

        // Leave the queue
        this.sendLeaveQueueRequest(username)
    }

    // Creates and connects a session
    public createSession(user: string): Session {
        const session = new Session()
        this.connectSession(session, user)
        return session
    }
}
