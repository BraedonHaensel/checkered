import { useEffect, useRef } from 'react'
import type { Message } from './message'
import { Backend } from './backend'

export class Session {
    private handlers: Record<string, (msg: any) => void> = {}
    private ws: WebSocket | null = null

    public constructor() {}

    private onmessage(event: MessageEvent<any>) {
        const data = JSON.parse(event.data) as Message

        const type = data.type

        const handler = this.handlers[type] || ((_: any) => {});

        handler(data);
    }

    public on<MessageType extends Message['type']>(
        type: MessageType,
        handler: (msg: Extract<Message, { type: MessageType }>) => void
    ): void {
        this.handlers[type] = handler;
    }

    public connect(ws: WebSocket) {
        this.ws = ws
        this.ws.addEventListener('message', (event: MessageEvent<any>) => {
            this.onmessage(event)
        })
    }

    public send(msg: Message): void {
        if (!this.ws) {
            return
        }
        this.ws.send(JSON.stringify(msg))
    }

    public end(): void {
        if (!this.ws) {
            return
        }
        this.ws.close()
    }

    public connected(): boolean {
        return this.ws !== null && this.ws.readyState === this.ws.OPEN
    }
}

export const useSession = (
    user: string,
    onCreate?: (session: Session) => void | ((session: Session) => void)
): Session => {
    const sessionRef = useRef<Session | null>(null)

    if (!sessionRef.current) {
        // safe because ref survives strict-mode replays
        sessionRef.current = Backend.instance().createSession(user)
    }

    useEffect(() => {
        const session = sessionRef.current!
        const onDisconnect = onCreate?.(session)

        return () => {
            if (sessionRef.current?.connected()) {
                onDisconnect?.(session)
                session.end()
                sessionRef.current = null
            }
        }
    }, [])

    return sessionRef.current!
}
