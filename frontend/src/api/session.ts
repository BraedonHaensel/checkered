import { useEffect, useRef } from 'react'
import type { Message } from './message'
import { Backend } from './backend'

export class Session {
    private handlers: Record<string, (msg: any) => void> = {
        registered: () => {
            this.wsConnected = true
            for (const message of this.queuedMessages) this.send(message)
        },
    }
    private ws: WebSocket | null = null
    private wsConnected: boolean = false
    private queuedMessages: Message[] = []

    public interval?: number

    public constructor() {}

    private onmessage(event: MessageEvent) {
        const data = JSON.parse(event.data) as Message

        const type = data.type

        const handler = this.handlers[type] || ((_: any) => {})

        handler(data)
    }

    public on<MessageType extends Message['type']>(
        type: MessageType,
        handler: (msg: Extract<Message, { type: MessageType }>) => void
    ): void {
        this.handlers[type] = handler
    }

    public connect(ws: WebSocket, user: string) {
        this.ws = ws
        this.ws.addEventListener('message', (event: MessageEvent) => {
            this.onmessage(event)
        })
        this.ws.send(
            JSON.stringify({
                type: 'join',
                user: user,
            })
        )
    }

    public send(msg: Message): void {
        if (!this.ws || !this.wsConnected) {
            this.queuedMessages.push(msg)
            return
        }
        if (this.wsConnected) this.ws.send(JSON.stringify(msg))
    }

    public end(): void {
        if (!this.ws) {
            return
        }
        this.ws.close()
        console.log('WebSocket session closed')
        if (this.interval) clearInterval(this.interval)
    }

    public connected(): boolean {
        return this.ws !== null && this.ws.readyState === this.ws.OPEN
    }
}

export const useSession = (
    username: string,
    onCreate?: (session: Session) => void | ((session: Session) => void)
): [Session, () => void, () => void] => {
    const sessionRef = useRef<Session | null>(null)

    const createSession = () => {
        console.log('Creating a new session')
        sessionRef.current = Backend.instance().createSession(username)
    }

    // eslint-disable-next-line react-hooks/refs
    if (!sessionRef.current) {
        // eslint-disable-next-line react-hooks/refs
        createSession()
    }

    useEffect(() => {
        const session = sessionRef.current!
        const onDisconnect = onCreate?.(session)

        return () => {
            if (session.connected()) {
                onDisconnect?.(session)
                session.end()
                sessionRef.current = null
            }
        }
    }, [onCreate])

    // Ends the current session.
    const closeSession = () => {
        const session = sessionRef.current!
        const onDisconnect = onCreate?.(session)
        if (session.connected()) {
            console.log('Ending the current session')
            onDisconnect?.(session)
            session.end()
        }
    }

    // Resets to a new session.
    const resetSession = () => {
        // End the previous sesesion
        closeSession()

        // Create a new session
        createSession()
    }

    // eslint-disable-next-line react-hooks/refs
    return [sessionRef.current!, closeSession, resetSession]
}
