import type { Request, Response } from './request'
import { Session } from './session'

export type BackendServer = {
    apiUrl: string
    wsUrl: string
    priority: number
}

const defaultTestingServer: BackendServer = {
    apiUrl: 'http://localhost:8080/api',
    wsUrl: 'ws://localhost:8080/ws',
    priority: 1,
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
        const url = `${this.server().apiUrl}/${type}`
        let raw: globalThis.Response = undefined as any
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
        } catch (e) {
            this.handleServerError()
            return this.get(type, payload)
        }

        return (await raw.json()) as Extract<Response, { type: RequestType }>
    }

    public async post<RequestType extends Request['type']>(
        type: RequestType,
        payload: Omit<Extract<Request, { type: RequestType }>, 'type'>
    ): Promise<Extract<Response, { type: RequestType }>> {
        const url = `${this.server().apiUrl}/${type}`
        let raw: globalThis.Response = undefined as any
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
        } catch (e) {
            this.handleServerError()
            return this.post(type, payload)
        }

        return (await raw.json()) as Extract<Response, { type: RequestType }>
    }

    private findServers() {
        this.servers = [defaultTestingServer]
    }

    public server(): BackendServer {
        if (this.current !== null) {
            return this.current
        }

        if (this.servers.length == 0) {
            this.findServers()
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

    private connectSession(session: Session, user: string) {
        const ws = new WebSocket(this.server().wsUrl)

        console.log('Attempting connection to backend server')

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
            session.connect(ws)
            session.send({
                type: 'join',
                user: user,
            })
        })
    }

    public createSession(user: string): Session {
        const session = new Session()

        this.connectSession(session, user)

        return session
    }
}
