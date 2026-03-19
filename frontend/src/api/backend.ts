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

const POLL_TIME = 500;

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
        const raw = await fetch(nameServer.url + "/matchmakers")
        const servers = await raw.json() as BackendServer[];

        this.servers = servers;
        console.log(servers);
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

    private async connectSession(session: Session, user: string) {
        await fetch(`${(await this.server()).url}/queue/add`, {
            body: JSON.stringify({
                username: user,
            }),
            headers: {
                'Content-Type': 'application/json',
            },
            method: "POST"
        });
        // TODO: check for successful addition to queue.
        // Waiting for final queuing logic.

        session.interval = 0;
        let attempting = false;

        const attemptConnection = async () => {
            console.log("Checking if game was found!");
            if(attempting) return;
            attempting = true;
            const raw = await fetch(`${(await this.server()).url}/queue/poll?username=${user}`, {
                headers: {
                    'Content-Type': 'application/json',
                },
                method: "GET"
            });
            const result = await raw.json();

            console.log(result);

            // TODO: We assume the shape of the resulting json 
            // object for now.

            const inGame = result.type === "success";

            if(inGame) {
                const wsUrl = result.url;
                clearInterval(session.interval);
                // setup websocket connection
                const ws = new WebSocket(`${wsUrl}/ws`)

                console.log(`Attempting connection to backend game server at ${wsUrl}`)

                ws.addEventListener('close', (ev: CloseEvent) => {
                    if (!ev.wasClean) {
                        this.connectSession(session, user)
                    }
                })

                ws.addEventListener('error', (err) => {
                    console.log('Error Callback', err.type)
                    ws.close()
                })

                ws.addEventListener('open', () => {
                    session.connect(ws, user)
                })
            }

            attempting = false;
            console.log("exit");
        }

        session.interval = setInterval(attemptConnection, POLL_TIME);
    }

    public createSession(user: string): Session {
        const session = new Session()

        this.findServers();

        this.connectSession(session, user);

        return session;
    }
}
