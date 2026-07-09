import {
    addJobTags,
    deleteDownload,
    getPlaybackInfo,
    listTags,
    removeJobTag,
    requestTranscode,
} from '@/services/libraryApi'

describe('libraryApi', () => {
    const originalFetch = global.fetch

    afterEach(() => {
        global.fetch = originalFetch
        vi.restoreAllMocks()
    })

    function mockFetch(impl: ReturnType<typeof vi.fn>) {
        global.fetch = impl as unknown as typeof fetch
    }

    it('lists tags and unwraps the message envelope', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: [{ id: 1, name: 'music', count: 3 }],
            }),
        })
        mockFetch(fetchMock)

        const tags = await listTags()

        expect(tags).toEqual([{ id: 1, name: 'music', count: 3 }])
        expect(fetchMock.mock.calls[0][0]).toContain('/tags')
    })

    it('adds tags with a POST to the job tags endpoint', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: [{ id: 1, name: 'music', source: 'user' }],
            }),
        })
        mockFetch(fetchMock)

        const tags = await addJobTags('job-1', ['music'])

        expect(tags).toHaveLength(1)
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/job/job-1/tags')
        expect(opts.method).toBe('POST')
        expect(JSON.parse(opts.body)).toEqual({ tags: ['music'] })
    })

    it('removes a tag with a DELETE request', async () => {
        const fetchMock = vi.fn().mockResolvedValue({ ok: true })
        mockFetch(fetchMock)

        await removeJobTag('job-1', 42)

        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/job/job-1/tags/42')
        expect(opts.method).toBe('DELETE')
    })

    it('deletes a download and throws the server error text on failure', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: false,
            status: 404,
            text: async () => 'Job not found',
        })
        mockFetch(fetchMock)

        await expect(deleteDownload('missing')).rejects.toThrow('Job not found')
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/job/missing')
        expect(opts.method).toBe('DELETE')
    })

    it('fetches playback info for a video', async () => {
        const info = {
            container: 'mp4',
            video_codec: 'vp9',
            audio_codec: 'opus',
            browser_safe: false,
        }
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ message: info }),
        })
        mockFetch(fetchMock)

        const result = await getPlaybackInfo('job-1')

        expect(result).toEqual(info)
        expect(fetchMock.mock.calls[0][0]).toContain(
            '/video/job-1/playback-info'
        )
    })

    it('requests a transcode with a POST and unwraps the job state', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: { job_id: 'tj-1', status: 'pending', progress: 0 },
            }),
        })
        mockFetch(fetchMock)

        const result = await requestTranscode('job-1')

        expect(result.job_id).toBe('tj-1')
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/video/job-1/transcode')
        expect(opts.method).toBe('POST')
    })
})
