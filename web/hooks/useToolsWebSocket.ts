import { useEffect } from 'react'
import useWebSocketStore from '@/services/websocket'
import useToolsState from '@/store/toolsState'
import { ToolsProgressUpdate } from '@/types'

/**
 * Hook to integrate WebSocket tools progress updates with the tools state
 * Automatically subscribes to 'tools-progress' messages and updates the store
 */
export function useToolsWebSocket() {
    const { updateJobProgress } = useToolsState()
    const subscribe = useWebSocketStore((state) => state.subscribe)

    useEffect(() => {
        // Subscribe to tools progress updates
        const unsubscribe = subscribe('tools-progress', (data: ToolsProgressUpdate) => {
            updateJobProgress(data)
        })

        // Cleanup on unmount
        return () => {
            unsubscribe()
        }
    }, [subscribe, updateJobProgress])
}
