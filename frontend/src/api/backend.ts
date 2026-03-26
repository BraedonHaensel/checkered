import type { Request, Response } from './request'
import { Session } from './session'

export type BackendServer = {
    url: string
    id: number
}

const nameServer: BackendServer = {
    url: 'http://localhost:9000/',
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

    private async findServers() {
        const raw = await fetch(nameServer.url + '/matchmakers')
        const servers = (await raw.json()) as BackendServer[]

        this.servers = servers
        console.log(servers)
    }

    public async server(): Promise<BackendServer> {
        if (this.current !== null) {
            return this.current
        }

        if (this.servers.length == 0) {
            await this.findServers()
        }

        const best = this.servers.shift()

        if (best === undefined) {
            throw new Error('Could not find any backend servers!')
        }

        this.current = best

        return this.current
    }

    public handleServerError() {
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
            if (joinQueueRes.type === 'ALREADY_IN_QUEUE') {
                console.log('Already in the queue')
            } else if (joinQueueRes.type === 'SUCCESS') {
                console.log('Successfully joined the queue!')
            }
        } catch (e) {
            console.error('Failed to join the queue:', e)
            // TODO Matchmaker down, try a new one? For now just return
            return
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
            console.log('Polling the Matchmaker...')
            pollInProgress = true
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
            console.log('Poll response:', pollRes)

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

            const joinQueueRes: LeaveQueueResponse = await raw.json()
            if (joinQueueRes.type === 'ALREADY_NOT_IN_QUEUE') {
                console.log('Already not in the queue')
            } else if (joinQueueRes.type === 'SUCCESS') {
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

    public createSession(user: string): Session {
        const session = new Session()

        this.findServers()

        this.connectSession(session, user)

        return session
    }
}
