import { SelectedInput, countSelectedVideos } from '../toolsState'

describe('countSelectedVideos', () => {
    const video = (id: string): SelectedInput => ({
        id,
        type: 'video',
        title: id,
    })

    it('counts individual videos as one each', () => {
        expect(countSelectedVideos([video('a'), video('b')])).toBe(2)
    })

    it('counts a playlist as its video count', () => {
        const playlist: SelectedInput = {
            id: 'p1',
            type: 'playlist',
            title: 'Playlist',
            videoCount: 12,
        }
        expect(countSelectedVideos([playlist])).toBe(12)
    })

    it('counts a channel as its video count', () => {
        const channel: SelectedInput = {
            id: 'c1',
            type: 'channel',
            title: 'Channel',
            videoCount: 7,
        }
        expect(countSelectedVideos([channel])).toBe(7)
    })

    it('assumes at least two videos when the count is unknown', () => {
        const playlist: SelectedInput = {
            id: 'p1',
            type: 'playlist',
            title: 'Playlist',
        }
        expect(countSelectedVideos([playlist])).toBe(2)
    })

    it('sums mixed selections', () => {
        const playlist: SelectedInput = {
            id: 'p1',
            type: 'playlist',
            title: 'Playlist',
            videoCount: 3,
        }
        expect(countSelectedVideos([video('a'), playlist])).toBe(4)
    })

    it('returns zero for an empty selection', () => {
        expect(countSelectedVideos([])).toBe(0)
    })
})
