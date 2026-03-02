import type { Request, Response } from "./request";
import { Session } from "./session";



export class Backend {
    private static _instance: Backend | null = null;

    static instance(): Backend {
        if(Backend._instance === null) {
            Backend._instance = new Backend();
        }

        return Backend._instance;
    }

    private constructor() {
    }

    public async get<RequestType extends Request>(type: RequestType["type"], payload: RequestType): Promise<Response & {type: RequestType["type"]}>  {
        const raw = await fetch(`http://${this.url()}/api/${type}`, {
            body: Object.keys(payload).length > 0 ? JSON.stringify(payload) : null,
            headers: {
                "Content-Type": "application/json",
            },
            method: "GET",
        });

        return await raw.json() as Response;
    }

    public async post<RequestType extends Request>(type: RequestType["type"], payload: RequestType): Promise<Response & {type: RequestType["type"]}> {
        const raw = await fetch(`http://${this.url()}/api/${type}`, {
            body: Object.keys(payload).length > 0 ? JSON.stringify(payload) : null,
            headers: {
                "Content-Type": "application/json",
            },
            method: "POST",
        });

        return await raw.json() as Response;
    }

    public url() {
        return "localhost:3000";
    }

    public wsUrl() {
        return `ws://${this.url()}/ws`;
    }

    public createSession(): Session {
        const ws = new WebSocket(this.wsUrl())
        ws.addEventListener("close", (ev: CloseEvent) => {
            if(!ev.wasClean) {
                alert("TODO: Reconnect websocket")
                // Here we need to notify the session that we are attempting to reconenct the web socket.
                // We'll need to implement a message queue in the session 
                // for when the current connection is unavailable.
            }
        })

        const session = new Session(ws);

        return session;
    }
}
