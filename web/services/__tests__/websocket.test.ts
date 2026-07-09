import useWebSocketStore from '@/services/websocket'

vi.mock('sonner', () => ({
    toast: vi.fn(),
}))

// Minimal WebSocket stand-in the store can drive.
class MockWebSocket {
    static instances: MockWebSocket[] = []
    static OPEN = 1
    static CONNECTING = 0

    url: string
    readyState = MockWebSocket.CONNECTING
    onopen: (() => void) | null = null
    onmessage: ((event: { data: string }) => void) | null = null
    onclose: (() => void) | null = null
    onerror: ((err: unknown) => void) | null = null

    constructor(url: string) {
        this.url = url
        MockWebSocket.instances.push(this)
    }

    open() {
        this.readyState = MockWebSocket.OPEN
        this.onopen?.()
    }

    receive(data: unknown) {
        this.onmessage?.({ data: JSON.stringify(data) })
    }

    close() {
        this.readyState = 3
        this.onclose?.()
    }
}

describe('websocket store', () => {
    const RealWebSocket = global.WebSocket

    beforeEach(() => {
        vi.useFakeTimers()
        MockWebSocket.instances = []
        global.WebSocket = MockWebSocket as unknown as typeof WebSocket
        useWebSocketStore.setState({
            socket: null,
            isConnected: false,
            reconnectTimer: null,
            reconnectAttempts: 0,
            isReconnecting: false,
            listeners: new Map(),
            onReconnectCallbacks: new Set(),
        })
    })

    afterEach(() => {
        useWebSocketStore.getState().disconnect()
        global.WebSocket = RealWebSocket
        vi.useRealTimers()
        vi.restoreAllMocks()
    })

    function connectAndOpen(): MockWebSocket {
        useWebSocketStore.getState().connect()
        const ws = MockWebSocket.instances.at(-1)!
        ws.open()
        return ws
    }

    it('routes messages by their type discriminator only', () => {
        const ws = connectAndOpen()

        const progress = vi.fn()
        const tools = vi.fn()
        useWebSocketStore.getState().subscribe('download-progress', progress)
        useWebSocketStore.getState().subscribe('tools-progress', tools)

        ws.receive({ type: 'download-progress', jobID: 'a', progress: 50 })
        ws.receive({ type: 'tools-progress', jobID: 'b', progress: 10 })
        // Untyped messages (old-style field sniffing bait) are dropped.
        ws.receive({ jobID: 'c', progress: 99 })

        expect(progress).toHaveBeenCalledTimes(1)
        expect(progress).toHaveBeenCalledWith(
            expect.objectContaining({ jobID: 'a' })
        )
        expect(tools).toHaveBeenCalledTimes(1)
        expect(tools).toHaveBeenCalledWith(
            expect.objectContaining({ jobID: 'b' })
        )
    })

    it('unsubscribe stops delivery', () => {
        const ws = connectAndOpen()
        const cb = vi.fn()
        const unsubscribe = useWebSocketStore
            .getState()
            .subscribe('download-progress', cb)

        ws.receive({ type: 'download-progress', jobID: 'a' })
        unsubscribe()
        ws.receive({ type: 'download-progress', jobID: 'a' })

        expect(cb).toHaveBeenCalledTimes(1)
    })

    it('reconnects with growing backoff and resets on success', () => {
        const ws = connectAndOpen()
        expect(useWebSocketStore.getState().isConnected).toBe(true)

        // First close → attempt 0 delay (~1s + jitter <= 1.5s)
        ws.close()
        expect(useWebSocketStore.getState().isConnected).toBe(false)
        vi.advanceTimersByTime(1_600)
        expect(MockWebSocket.instances).toHaveLength(2)

        // Second close before opening → attempt 1 delay (~2s). Not yet at 1.5s…
        MockWebSocket.instances.at(-1)!.close()
        vi.advanceTimersByTime(1_500)
        expect(MockWebSocket.instances).toHaveLength(2)
        // …but reconnects by 2.5s.
        vi.advanceTimersByTime(1_100)
        expect(MockWebSocket.instances).toHaveLength(3)

        // Successful open resets the attempt counter and fires callbacks.
        const cb = vi.fn()
        useWebSocketStore.getState().onReconnect(cb)
        MockWebSocket.instances.at(-1)!.open()
        expect(useWebSocketStore.getState().reconnectAttempts).toBe(0)
        expect(cb).toHaveBeenCalledTimes(1)
    })

    it('disconnect prevents any further reconnects', () => {
        connectAndOpen()
        useWebSocketStore.getState().disconnect()
        vi.advanceTimersByTime(60_000)
        expect(MockWebSocket.instances).toHaveLength(1)
    })
})
