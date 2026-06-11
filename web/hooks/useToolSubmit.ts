import { ToolOperation, submitTool } from '@/services/toolsApi'
import useToolsState from '@/store/toolsState'

import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'

/**
 * useToolSubmit encapsulates the shared submit flow for every tool page:
 * call the API, register the returned job for progress tracking, clear the
 * selection and return to the tools dashboard.
 */
export function useToolSubmit(operation: ToolOperation) {
    const navigate = useNavigate()
    const { selectedInputs, clearSelectedInputs, addActiveJob } =
        useToolsState()
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const submit = useCallback(
        async (parameters: Record<string, unknown>) => {
            if (selectedInputs.length === 0) {
                setError('Please select at least one video')
                return
            }

            setIsSubmitting(true)
            setError(null)
            try {
                const job = await submitTool(
                    operation,
                    selectedInputs.map((i) => ({ id: i.id, type: i.type })),
                    parameters
                )
                addActiveJob(job)
                clearSelectedInputs()
                navigate('/tools')
            } catch (err) {
                setError(
                    err instanceof Error ? err.message : 'An error occurred'
                )
            } finally {
                setIsSubmitting(false)
            }
        },
        [operation, selectedInputs, addActiveJob, clearSelectedInputs, navigate]
    )

    return { submit, isSubmitting, error, setError }
}
