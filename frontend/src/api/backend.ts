import type { Request, Response } from "./request";
import { Session } from "./session";

export type BackendServer = {
    apiUrl: string,
    wsUrl: string,
    priority: number,
}

const defaultTestingServer: BackendServer = {
    apiUrl: "http://localhost:3000/api",
    wsUrl: "ws://localhost:3000/ws",
    priority: 1,
}

export class Backend {
    private static _instance: Backend | null = null;

    private servers: BackendServer[] = [];
    private current: BackendServer | null = null;

    static instance(): Backend {
        if(Backend._instance === null) {
            Backend._instance = new Backend();
        }

        return Backend._instance;
    }

    private constructor() {
    }

    public async get<RequestType extends Request>(type: RequestType["type"], payload: RequestType): Promise<Response & {type: RequestType["type"]}>  {
        const url = `${this.server().apiUrl}/${type}`;
        const raw = await fetch(url, {
            body: Object.keys(payload).length > 0 ? JSON.stringify(payload) : null,
            headers: {
                "Content-Type": "application/json",
            },
            method: "GET",
        });

        return await raw.json() as Response;
    }

    public async post<RequestType extends Request>(type: RequestType["type"], payload: RequestType): Promise<Response & {type: RequestType["type"]}> {
        const url = `${this.server().apiUrl}/${type}`;
        const raw = await fetch(url, {
            body: Object.keys(payload).length > 0 ? JSON.stringify(payload) : null,
            headers: {
                "Content-Type": "application/json",
            },
            method: "POST",
        });

        return await raw.json() as Response;
    }

    private findServers() {
        this.servers = [
            defaultTestingServer
        ]
    }

    public server(): BackendServer {
        if(this.current !== null) {
            return this.current;
        }

        if(this.servers.length == 0) {
            this.findServers();
        }

        const best = this.servers.shift();

        if(best === undefined) {
            throw new Error("Could not find any backend servers!");
        }
        
        this.current = best;

        return this.current;
    }


    public handleServerError() {
        this.current = null;
        
    }

    private connectSession(session: Session) {
        const ws = new WebSocket(this.server().wsUrl)

        console.log("Attempting connection to backend server")

        ws.addEventListener("close", (ev: CloseEvent) => {
            if(!ev.wasClean) {
                this.connectSession(session)
            }
        })

        ws.addEventListener("error", (err) => {
            console.log("Error Callback", err.type);
            ws.close();
        })
        
        ws.addEventListener("open", () => {
            session.connect(ws);
        })
    } 

    public createSession(): Session {
        const session = new Session();

        this.connectSession(session);

        return session;
    }
}
