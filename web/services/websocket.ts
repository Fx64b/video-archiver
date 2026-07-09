import { toast } from 'sonner'
import { create } from 'zustand'

import { serverWsUrl } from '@/lib/env'

/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * WebSocket connection store.
 *
 * Every message from the backend carries a `type` discriminator
 * (WSTypeDownloadProgress / WSTypeMetadataUpdate / WSTypeToolsProgress in the
 * generated types) and is routed to subscribers of exactly that type — no
 * shape sniffing. Reconnects use exponential backoff with jitter, and the
 * connection is opened explicitly from the app root, not as an import side
 * effect.
 */

const BASE_RECONNECT_DELAY_MS = 1_000
const MAX_RECONNECT_DELAY_MS = 30_000

function reconnectDelay(attempt: number): number {
    const exponential = Math.min(
        MAX_RECONNECT_DELAY_MS,
        BASE_RECONNECT_DELAY_MS * 2 ** attempt
    )
    // Jitter spreads clients out so they don't reconnect in lockstep.
    return exponential + Math.random() * 500
}

interface WebSocketState {
    socket: WebSocket | null
    isConnected: boolean
    reconnectTimer: ReturnType<typeof setTimeout> | null
    reconnectAttempts: number
    isReconnecting: boolean
    listeners: Map<string, Set<(data: any) => void>>
    onReconnectCallbacks: Set<() => void>

    connect: () => void
    disconnect: () => void
    subscribe: (type: string, callback: (data: any) => void) => () => void
    onReconnect: (callback: () => void) => () => void
}

const useWebSocketStore = create<WebSocketState>((set, get) => ({
    socket: null,
    isConnected: false,
    reconnectTimer: null,
    reconnectAttempts: 0,
    isReconnecting: false,
    listeners: new Map(),
    onReconnectCallbacks: new Set(),

    connect: () => {
        const { socket, disconnect } = get()

        if (
            socket?.readyState === WebSocket.OPEN ||
            socket?.readyState === WebSocket.CONNECTING
        ) {
            return
        }

        // Close existing socket if it exists but isn't open
        if (socket) disconnect()

        const newSocket = new WebSocket(serverWsUrl() + '/ws')

        newSocket.onopen = () => {
            console.log('WebSocket connected')

            const { isReconnecting, onReconnectCallbacks, reconnectTimer } =
                get()
            if (reconnectTimer) clearTimeout(reconnectTimer)
            set({
                isConnected: true,
                reconnectAttempts: 0,
                reconnectTimer: null,
            })

            if (isReconnecting) {
                toast('Reconnected to server successfully.')
                set({ isReconnecting: false })

                // Reload data that may have changed while disconnected.
                onReconnectCallbacks.forEach((callback) => callback())
            }
        }

        newSocket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data)
                const type = typeof data?.type === 'string' ? data.type : ''
                if (!type) {
                    console.warn('WebSocket message without type — dropped')
                    return
                }

                const { listeners } = get()
                listeners.get(type)?.forEach((callback) => callback(data))
                listeners.get('all')?.forEach((callback) => callback(data))
            } catch (error) {
                console.error('Error processing WebSocket message:', error)
            }
        }

        newSocket.onclose = () => {
            set({ isConnected: false })

            const { reconnectTimer, reconnectAttempts } = get()
            if (reconnectTimer) return

            const delay = reconnectDelay(reconnectAttempts)
            console.log(
                `WebSocket closed — reconnecting in ${Math.round(delay)}ms`
            )
            const timer = setTimeout(() => {
                set({
                    reconnectTimer: null,
                    reconnectAttempts: reconnectAttempts + 1,
                    isReconnecting: true,
                })
                get().connect()
            }, delay)

            set({ reconnectTimer: timer })
        }

        newSocket.onerror = (error) => {
            console.error('WebSocket error:', error)
            newSocket.close()
        }

        set({ socket: newSocket })
    },

    disconnect: () => {
        const { socket, reconnectTimer } = get()

        if (socket) {
            // Prevent the close handler from scheduling a reconnect.
            socket.onclose = null
            socket.close()
            set({ socket: null, isConnected: false })
        }

        if (reconnectTimer) {
            clearTimeout(reconnectTimer)
            set({ reconnectTimer: null })
        }
    },

    subscribe: (type: string, callback: (data: any) => void) => {
        const { listeners } = get()

        if (!listeners.has(type)) {
            listeners.set(type, new Set())
        }

        const typeListeners = listeners.get(type)!
        typeListeners.add(callback)

        return () => {
            get().listeners.get(type)?.delete(callback)
        }
    },

    onReconnect: (callback: () => void) => {
        const { onReconnectCallbacks } = get()
        onReconnectCallbacks.add(callback)

        return () => {
            get().onReconnectCallbacks.delete(callback)
        }
    },
}))

export default useWebSocketStore

/* eslint-enable @typescript-eslint/no-explicit-any */
