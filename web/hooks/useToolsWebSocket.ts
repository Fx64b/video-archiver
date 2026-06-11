import useWebSocketStore from '@/services/websocket'
import useToolsState from '@/store/toolsState'
import { ToolsProgressUpdate } from '@/types'
import { toast } from 'sonner'

import { useEffect } from 'react'

/**
 * Hook to integrate WebSocket tools progress updates with the tools state.
 * Subscribes to 'tools-progress' messages, updates the store and surfaces
 * completion/failure to the user via toasts.
 */
export function useToolsWebSocket() {
    const { updateJobProgress } = useToolsState()
    const subscribe = useWebSocketStore((state) => state.subscribe)

    useEffect(() => {
        const unsubscribe = subscribe(
            'tools-progress',
            (data: ToolsProgressUpdate) => {
                // Only toast on the transition into a terminal state, not on
                // every repeated update carrying the same status.
                const previous = useToolsState
                    .getState()
                    .jobProgress.get(data.jobID)

                updateJobProgress(data)

                if (previous?.status === data.status) return
                if (data.status === 'complete') {
                    toast.success(
                        'Processing complete — the file is ready in Processed Files.'
                    )
                } else if (data.status === 'failed') {
                    toast.error(
                        data.error
                            ? `Processing failed: ${data.error}`
                            : 'Processing failed.'
                    )
                }
            }
        )

        return () => {
            unsubscribe()
        }
    }, [subscribe, updateJobProgress])
}
