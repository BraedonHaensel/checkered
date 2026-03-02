import { useEffect, useRef, useState } from "react";
import type { Message } from "./message";
import { Backend } from "./backend";

export class Session {

    private handlers: Record<string, Array<(msg: any) => void>> = {};
    private ws: WebSocket

    public constructor(ws: WebSocket) {
        this.ws = ws;
        this.ws.addEventListener("error", (err) => {
            console.error("An error occurred in the websocket connection!", err);

        })
        this.ws.addEventListener("message", (event: MessageEvent<any>) => {
            this.onmessage(event);
        })
    }

    private onmessage(event: MessageEvent<any>) {
        const data = JSON.parse(event.data) as Message;

        const type = data.type;

        const handlers = this.handlers[type] || [];

        handlers.forEach(handler => handler(data))
    }

    public on<MessageType extends Message["type"]>(type: MessageType, handler: (msg: Extract<Message, {type: MessageType}>) => void): void {
        if(!Object.keys(this.handlers).includes(type)) {
            this.handlers[type] = [];
        }

        this.handlers[type].push(handler);
    }

    public send(msg: Message): void {
        this.ws.send(JSON.stringify(msg));
    }

    public end(): void {
        this.ws.close();
    }

    public connected(): boolean {
        return this.ws.readyState === this.ws.OPEN;
    }
}

export const useSession = (
    onCreate?: (session: Session) => void | ((session: Session) => void)
): Session => {
    const sessionRef = useRef<Session | null>(null);

    if (!sessionRef.current) {
        // safe because ref survives strict-mode replays
        sessionRef.current = Backend.instance().createSession();
    }

    useEffect(() => {
        const session = sessionRef.current!;
        const onDisconnect = onCreate?.(session);

        return () => {
            if(sessionRef.current?.connected()) {
                onDisconnect?.(session);
                session.end();
                sessionRef.current = null;
            }
        };
    }, []);

    return sessionRef.current!;
};
