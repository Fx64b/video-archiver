import { ToolsJob, ToolsProgressUpdate } from '@/types'
import { create } from 'zustand'

interface SelectedInput {
    id: string
    type: 'video' | 'playlist' | 'channel'
    title: string
    thumbnail?: string
}

interface ToolsState {
    // Active jobs and progress
    activeJobs: Map<string, ToolsJob>
    jobProgress: Map<string, ToolsProgressUpdate>

    // Job history
    jobHistory: ToolsJob[]

    // Selected inputs for processing
    selectedInputs: SelectedInput[]

    // UI state
    isProcessing: boolean
    currentOperation: string | null

    // Actions - Active jobs
    addActiveJob: (job: ToolsJob) => void
    updateJobProgress: (update: ToolsProgressUpdate) => void
    removeActiveJob: (jobId: string) => void
    getActiveJob: (jobId: string) => ToolsJob | undefined

    // Actions - Job history
    addToHistory: (job: ToolsJob) => void
    clearHistory: () => void

    // Actions - Input selection
    addSelectedInput: (input: SelectedInput) => void
    removeSelectedInput: (id: string) => void
    clearSelectedInputs: () => void
    isInputSelected: (id: string) => boolean

    // Actions - UI state
    setIsProcessing: (value: boolean) => void
    setCurrentOperation: (operation: string | null) => void
}

const useToolsState = create<ToolsState>((set, get) => ({
    // Initial state
    activeJobs: new Map<string, ToolsJob>(),
    jobProgress: new Map<string, ToolsProgressUpdate>(),
    jobHistory: [],
    selectedInputs: [],
    isProcessing: false,
    currentOperation: null,

    // Active jobs actions
    addActiveJob: (job) =>
        set((state) => {
            const newActiveJobs = new Map(state.activeJobs)
            newActiveJobs.set(job.id, job)
            return {
                activeJobs: newActiveJobs,
                isProcessing: newActiveJobs.size > 0
            }
        }),

    updateJobProgress: (update) =>
        set((state) => {
            const newJobProgress = new Map(state.jobProgress)
            newJobProgress.set(update.jobID, update)

            // Also update the job status if it exists in activeJobs
            const newActiveJobs = new Map(state.activeJobs)
            const existingJob = newActiveJobs.get(update.jobID)
            if (existingJob) {
                newActiveJobs.set(update.jobID, {
                    ...existingJob,
                    status: update.status,
                    progress: update.progress,
                })
            }

            return {
                jobProgress: newJobProgress,
                activeJobs: newActiveJobs,
            }
        }),

    removeActiveJob: (jobId) =>
        set((state) => {
            const newActiveJobs = new Map(state.activeJobs)
            const newJobProgress = new Map(state.jobProgress)

            // Move to history before removing
            const job = newActiveJobs.get(jobId)
            if (job) {
                const newHistory = [job, ...state.jobHistory].slice(0, 50) // Keep last 50
                newActiveJobs.delete(jobId)
                newJobProgress.delete(jobId)

                return {
                    activeJobs: newActiveJobs,
                    jobProgress: newJobProgress,
                    jobHistory: newHistory,
                    isProcessing: newActiveJobs.size > 0,
                }
            }

            newActiveJobs.delete(jobId)
            newJobProgress.delete(jobId)
            return {
                activeJobs: newActiveJobs,
                jobProgress: newJobProgress,
                isProcessing: newActiveJobs.size > 0,
            }
        }),

    getActiveJob: (jobId) => get().activeJobs.get(jobId),

    // Job history actions
    addToHistory: (job) =>
        set((state) => ({
            jobHistory: [job, ...state.jobHistory].slice(0, 50), // Keep last 50
        })),

    clearHistory: () => set({ jobHistory: [] }),

    // Input selection actions
    addSelectedInput: (input) =>
        set((state) => {
            // Prevent duplicates
            if (state.selectedInputs.some((i) => i.id === input.id)) {
                return state
            }
            return {
                selectedInputs: [...state.selectedInputs, input],
            }
        }),

    removeSelectedInput: (id) =>
        set((state) => ({
            selectedInputs: state.selectedInputs.filter((i) => i.id !== id),
        })),

    clearSelectedInputs: () => set({ selectedInputs: [] }),

    isInputSelected: (id) => get().selectedInputs.some((i) => i.id === id),

    // UI state actions
    setIsProcessing: (value) => set({ isProcessing: value }),

    setCurrentOperation: (operation) => set({ currentOperation: operation }),
}))

export default useToolsState
