import { toast } from 'sonner'
import { create } from 'zustand'

/* eslint-disable @typescript-eslint/no-explicit-any */

interface WebSocketState {
    socket: WebSocket | null
    isConnected: boolean
    reconnectTimer: NodeJS.Timeout | null
    isReconnecting: boolean
    listeners: Map<string, Set<(data: any) => void>>

    connect: () => void
    disconnect: () => void
    subscribe: (type: string, callback: (data: any) => void) => () => void
}

const useWebSocketStore = create<WebSocketState>((set, get) => ({
    socket: null,
    isConnected: false,
    reconnectTimer: null,
    isReconnecting: false,
    listeners: new Map(),

    connect: () => {
        const { socket, disconnect } = get()

        if (socket?.readyState === WebSocket.OPEN) return

        // Close existing socket if it exists but isn't open
        if (socket) disconnect()

        const wsUrl = process.env.NEXT_PUBLIC_SERVER_URL_WS + '/ws'
        const newSocket = new WebSocket(wsUrl)

        newSocket.onopen = () => {
            console.log('WebSocket connected')

            const { isReconnecting } = get()
            set({ isConnected: true })

            // Clear any reconnect timer
            const { reconnectTimer } = get()
            if (reconnectTimer) {
                clearTimeout(reconnectTimer)
                set({ reconnectTimer: null })
            }

            if (isReconnecting) {
                toast('Reconnected to server successfully.')
                set({ isReconnecting: false })
            }
        }

        newSocket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data)
                const { listeners } = get()

                // Determine message type with a more robust check
                const type = data && 'metadata' in data
                    ? 'metadata'
                    : data && 'progress' in data
                        ? 'progress'
                        : 'unknown'

                if (type !== 'unknown') {
                    const typeListeners = listeners.get(type)
                    if (typeListeners) {
                        typeListeners.forEach((callback) => callback(data))
                    }

                    const allListeners = listeners.get('all')
                    if (allListeners) {
                        allListeners.forEach((callback) => callback(data))
                    }
                }
            } catch (error) {
                console.error('Error processing WebSocket message:', error)
            }
        }

        newSocket.onclose = () => {
            console.log('WebSocket connection closed')
            set({ isConnected: false })

            const { reconnectTimer } = get()
            if (!reconnectTimer) {
                const timer = setTimeout(() => {
                    console.log('Attempting to reconnect WebSocket...')
                    set({
                        reconnectTimer: null,
                        isReconnecting: true,
                    })
                    get().connect()
                }, 5000)

                set({ reconnectTimer: timer })
            }
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
            const currentListeners = get().listeners.get(type)
            if (currentListeners) {
                currentListeners.delete(callback)
            }
        }
    },
}))

// Automatically connect when the service is imported
if (typeof window !== 'undefined') {
    setTimeout(() => {
        useWebSocketStore.getState().connect()
    }, 0)
}

export default useWebSocketStore

/* eslint-enable @typescript-eslint/no-explicit-any */
