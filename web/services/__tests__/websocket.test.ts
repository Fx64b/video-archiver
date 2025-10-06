/**
 * @jest-environment jsdom
 */

// Mock the toast from sonner
jest.mock('sonner', () => ({
    toast: jest.fn(),
}))

describe('websocket service', () => {
    let mockWebSocket: any
    let onOpenCallback: (() => void) | null = null
    let onMessageCallback: ((event: MessageEvent) => void) | null = null
    let onCloseCallback: (() => void) | null = null
    let onErrorCallback: ((error: Event) => void) | null = null

    beforeEach(() => {
        jest.clearAllMocks()
        jest.useFakeTimers()

        // Reset callbacks
        onOpenCallback = null
        onMessageCallback = null
        onCloseCallback = null
        onErrorCallback = null

        // Mock WebSocket
        mockWebSocket = {
            readyState: WebSocket.CONNECTING,
            close: jest.fn(),
            send: jest.fn(),
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
        }

        // @ts-ignore
        global.WebSocket = jest.fn().mockImplementation(() => {
            return mockWebSocket
        })

        // Capture event listeners
        Object.defineProperty(mockWebSocket, 'onopen', {
            set: (callback: () => void) => {
                onOpenCallback = callback
            },
        })

        Object.defineProperty(mockWebSocket, 'onmessage', {
            set: (callback: (event: MessageEvent) => void) => {
                onMessageCallback = callback
            },
        })

        Object.defineProperty(mockWebSocket, 'onclose', {
            set: (callback: () => void) => {
                onCloseCallback = callback
            },
        })

        Object.defineProperty(mockWebSocket, 'onerror', {
            set: (callback: (error: Event) => void) => {
                onErrorCallback = callback
            },
        })
    })

    afterEach(() => {
        jest.clearAllTimers()
        jest.useRealTimers()
    })

    it('should connect to websocket on initialization', async () => {
        // Import after mocking to trigger auto-connect
        jest.isolateModules(() => {
            require('../websocket')
        })

        jest.runOnlyPendingTimers()

        expect(global.WebSocket).toHaveBeenCalledWith(
            'ws://localhost:8081/ws'
        )
    })

    it('should set isConnected to true when websocket opens', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const store = useWebSocketStore.getState()
        store.connect()

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        expect(useWebSocketStore.getState().isConnected).toBe(true)
    })

    it('should handle metadata messages', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const callback = jest.fn()
        const store = useWebSocketStore.getState()

        store.connect()
        store.subscribe('metadata', callback)

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        const metadataMessage = {
            metadata: {
                _type: 'video',
                id: 'test-id',
                title: 'Test Video',
            },
        }

        onMessageCallback?.({
            data: JSON.stringify(metadataMessage),
        } as MessageEvent)

        expect(callback).toHaveBeenCalledWith(metadataMessage)
    })

    it('should handle progress messages', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const callback = jest.fn()
        const store = useWebSocketStore.getState()

        store.connect()
        store.subscribe('progress', callback)

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        const progressMessage = {
            jobID: 'job-1',
            progress: 50,
            currentVideoProgress: 75,
        }

        onMessageCallback?.({
            data: JSON.stringify(progressMessage),
        } as MessageEvent)

        expect(callback).toHaveBeenCalledWith(progressMessage)
    })

    it('should notify all listeners with "all" subscription', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const allCallback = jest.fn()
        const metadataCallback = jest.fn()
        const store = useWebSocketStore.getState()

        store.connect()
        store.subscribe('all', allCallback)
        store.subscribe('metadata', metadataCallback)

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        const metadataMessage = {
            metadata: {
                _type: 'video',
                id: 'test-id',
            },
        }

        onMessageCallback?.({
            data: JSON.stringify(metadataMessage),
        } as MessageEvent)

        expect(allCallback).toHaveBeenCalledWith(metadataMessage)
        expect(metadataCallback).toHaveBeenCalledWith(metadataMessage)
    })

    it('should unsubscribe listeners', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const callback = jest.fn()
        const store = useWebSocketStore.getState()

        store.connect()
        const unsubscribe = store.subscribe('metadata', callback)

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        const metadataMessage = {
            metadata: {
                _type: 'video',
                id: 'test-id',
            },
        }

        onMessageCallback?.({
            data: JSON.stringify(metadataMessage),
        } as MessageEvent)

        expect(callback).toHaveBeenCalledTimes(1)

        unsubscribe()

        onMessageCallback?.({
            data: JSON.stringify(metadataMessage),
        } as MessageEvent)

        // Should still be 1, not 2
        expect(callback).toHaveBeenCalledTimes(1)
    })

    it('should attempt to reconnect when connection closes', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )
        const { toast } = require('sonner')

        const store = useWebSocketStore.getState()
        store.connect()

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        expect(useWebSocketStore.getState().isConnected).toBe(true)

        // Simulate connection close
        onCloseCallback?.()

        expect(useWebSocketStore.getState().isConnected).toBe(false)

        // Fast-forward timer
        jest.advanceTimersByTime(5000)

        // Should attempt to reconnect
        expect(global.WebSocket).toHaveBeenCalledTimes(2)
    })

    it('should disconnect and cleanup', async () => {
        const { default: useWebSocketStore } = await import(
            '../websocket'
        )

        const store = useWebSocketStore.getState()
        store.connect()

        mockWebSocket.readyState = WebSocket.OPEN
        onOpenCallback?.()

        store.disconnect()

        expect(mockWebSocket.close).toHaveBeenCalled()
        expect(useWebSocketStore.getState().socket).toBe(null)
        expect(useWebSocketStore.getState().isConnected).toBe(false)
    })
})
