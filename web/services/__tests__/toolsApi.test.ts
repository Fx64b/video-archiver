import {
    ToolInput,
    cancelToolJob,
    listToolJobs,
    submitTool,
} from '@/services/toolsApi'

describe('toolsApi', () => {
    const originalFetch = global.fetch

    afterEach(() => {
        global.fetch = originalFetch
        jest.restoreAllMocks()
    })

    function mockFetch(impl: jest.Mock) {
        global.fetch = impl as unknown as typeof fetch
    }

    it('submits to the correct URL and unwraps the message envelope', async () => {
        const fetchMock = jest.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: { id: 'job-1', operation_type: 'trim' },
            }),
        })
        mockFetch(fetchMock)

        const inputs: ToolInput[] = [{ id: 'v1', type: 'video' }]
        const job = await submitTool('trim', inputs, {
            start_time: '0',
            end_time: '10',
        })

        expect(job.id).toBe('job-1')
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/tools/trim')
        expect(opts.method).toBe('POST')
        const body = JSON.parse(opts.body)
        expect(body.input_files).toEqual(['v1'])
        expect(body.input_type).toBe('videos')
        expect(body.parameters).toEqual({ start_time: '0', end_time: '10' })
    })

    it('passes a playlist as a single parent ID', async () => {
        const fetchMock = jest.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ message: { id: 'job-2' } }),
        })
        mockFetch(fetchMock)

        const inputs: ToolInput[] = [{ id: 'p1', type: 'playlist' }]
        await submitTool('concat', inputs, { output_format: 'mp4' })

        const body = JSON.parse(fetchMock.mock.calls[0][1].body)
        expect(body.input_files).toEqual(['p1'])
        expect(body.input_type).toBe('playlist')
    })

    it('throws with the server error text on failure', async () => {
        const fetchMock = jest.fn().mockResolvedValue({
            ok: false,
            status: 400,
            text: async () => 'end_time must be greater than start_time',
        })
        mockFetch(fetchMock)

        await expect(
            submitTool('trim', [{ id: 'v1', type: 'video' }], {})
        ).rejects.toThrow('end_time must be greater than start_time')
    })

    it('cancels a job via DELETE', async () => {
        const fetchMock = jest.fn().mockResolvedValue({ ok: true })
        mockFetch(fetchMock)

        await cancelToolJob('job-9')
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/tools/jobs/job-9')
        expect(opts.method).toBe('DELETE')
    })

    it('lists jobs and unwraps the paginated message', async () => {
        const fetchMock = jest.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: {
                    items: [{ id: 'j1' }],
                    total_count: 1,
                    page: 1,
                    limit: 20,
                    total_pages: 1,
                },
            }),
        })
        mockFetch(fetchMock)

        const result = await listToolJobs(1, 20)
        expect(result.total_count).toBe(1)
        expect(result.items).toHaveLength(1)
    })
})
