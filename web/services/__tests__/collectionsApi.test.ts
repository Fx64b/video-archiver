import {
    addVideosToCollection,
    createCollection,
    deleteCollection,
    getCollectionVideos,
    listCollections,
    listCollectionsForVideo,
    removeVideoFromCollection,
    updateCollection,
} from '@/services/collectionsApi'

describe('collectionsApi', () => {
    const originalFetch = global.fetch

    afterEach(() => {
        global.fetch = originalFetch
        vi.restoreAllMocks()
    })

    function mockFetch(impl: ReturnType<typeof vi.fn>) {
        global.fetch = impl as unknown as typeof fetch
    }

    it('lists collections and unwraps the message envelope', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: [{ id: 'c1', name: 'Watch Later', video_count: 2 }],
            }),
        })
        mockFetch(fetchMock)

        const collections = await listCollections()

        expect(collections).toEqual([
            { id: 'c1', name: 'Watch Later', video_count: 2 },
        ])
        expect(fetchMock.mock.calls[0][0]).toContain('/collections')
    })

    it('returns an empty array when the list message is null', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ message: null }),
        })
        mockFetch(fetchMock)

        expect(await listCollections()).toEqual([])
        expect(await getCollectionVideos('c1')).toEqual([])
        expect(await listCollectionsForVideo('v1')).toEqual([])
    })

    it('creates a collection with a POST', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: { id: 'c1', name: 'Mix', video_count: 0 },
            }),
        })
        mockFetch(fetchMock)

        const collection = await createCollection('Mix', 'my favorites')

        expect(collection.id).toBe('c1')
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/collections')
        expect(opts.method).toBe('POST')
        expect(JSON.parse(opts.body)).toEqual({
            name: 'Mix',
            description: 'my favorites',
        })
    })

    it('updates a collection with a PUT', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: { id: 'c1', name: 'Renamed', video_count: 0 },
            }),
        })
        mockFetch(fetchMock)

        await updateCollection('c1', 'Renamed')

        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/collections/c1')
        expect(opts.method).toBe('PUT')
    })

    it('adds videos with a POST to the collection videos endpoint', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                message: { id: 'c1', name: 'Mix', video_count: 2 },
            }),
        })
        mockFetch(fetchMock)

        const collection = await addVideosToCollection('c1', ['v1', 'v2'])

        expect(collection.video_count).toBe(2)
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/collections/c1/videos')
        expect(opts.method).toBe('POST')
        expect(JSON.parse(opts.body)).toEqual({ video_ids: ['v1', 'v2'] })
    })

    it('removes a video with a DELETE request', async () => {
        const fetchMock = vi.fn().mockResolvedValue({ ok: true })
        mockFetch(fetchMock)

        await removeVideoFromCollection('c1', 'v1')

        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/collections/c1/videos/v1')
        expect(opts.method).toBe('DELETE')
    })

    it('deletes a collection and throws the server error text on failure', async () => {
        const fetchMock = vi.fn().mockResolvedValue({
            ok: false,
            status: 404,
            text: async () => 'Collection not found',
        })
        mockFetch(fetchMock)

        await expect(deleteCollection('missing')).rejects.toThrow(
            'Collection not found'
        )
        const [url, opts] = fetchMock.mock.calls[0]
        expect(url).toContain('/collections/missing')
        expect(opts.method).toBe('DELETE')
    })
})
