import {
    cn,
    formatBytes,
    formatResolution,
    formatSeconds,
    formatSubscriberNumber,
} from '../utils'

describe('utils', () => {
    describe('cn', () => {
        it('should merge class names correctly', () => {
            expect(cn('foo', 'bar')).toBe('foo bar')
        })

        it('should handle conditional classes', () => {
            expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz')
        })

        it('should merge tailwind classes without conflicts', () => {
            expect(cn('px-2 py-1', 'px-4')).toBe('py-1 px-4')
        })
    })

    describe('formatSeconds', () => {
        it('should format seconds less than an hour (MM:SS)', () => {
            expect(formatSeconds(90)).toBe('01:30')
            expect(formatSeconds(3599)).toBe('59:59')
        })

        it('should format seconds greater than or equal to an hour (HH:MM:SS)', () => {
            expect(formatSeconds(3600)).toBe('01:00:00')
            expect(formatSeconds(3661)).toBe('01:01:01')
        })

        it('should handle string input', () => {
            expect(formatSeconds('120')).toBe('02:00')
        })

        it('should return empty string for null', () => {
            expect(formatSeconds(null)).toBe('')
        })

        it('should handle zero', () => {
            expect(formatSeconds(0)).toBe('00:00')
        })
    })

    describe('formatSubscriberNumber', () => {
        it('should return the number as string for values less than 1000', () => {
            expect(formatSubscriberNumber(0)).toBe('0')
            expect(formatSubscriberNumber(500)).toBe('500')
            expect(formatSubscriberNumber(999)).toBe('999')
        })

        it('should format thousands with K suffix', () => {
            expect(formatSubscriberNumber(1000)).toBe('1K')
            expect(formatSubscriberNumber(1500)).toBe('1.5K')
            expect(formatSubscriberNumber(10000)).toBe('10K')
            expect(formatSubscriberNumber(999999)).toBe('1000K')
        })

        it('should format millions with M suffix', () => {
            expect(formatSubscriberNumber(1000000)).toBe('1M')
            expect(formatSubscriberNumber(1500000)).toBe('1.5M')
            expect(formatSubscriberNumber(10000000)).toBe('10M')
        })

        it('should remove trailing zeros', () => {
            expect(formatSubscriberNumber(1100)).toBe('1.1K')
            expect(formatSubscriberNumber(1000000)).toBe('1M')
        })
    })

    describe('formatBytes', () => {
        it('should return "0 Bytes" for zero', () => {
            expect(formatBytes(0)).toBe('0 Bytes')
        })

        it('should format bytes correctly', () => {
            expect(formatBytes(500)).toBe('500 Bytes')
        })

        it('should format kilobytes correctly', () => {
            expect(formatBytes(1024)).toBe('1 KB')
            expect(formatBytes(1536)).toBe('1.5 KB')
        })

        it('should format megabytes correctly', () => {
            expect(formatBytes(1048576)).toBe('1 MB')
            expect(formatBytes(5242880)).toBe('5 MB')
        })

        it('should format gigabytes correctly', () => {
            expect(formatBytes(1073741824)).toBe('1 GB')
            expect(formatBytes(2147483648)).toBe('2 GB')
        })

        it('should respect decimal precision parameter', () => {
            expect(formatBytes(1536, 2)).toBe('1.5 KB')
            expect(formatBytes(1536, 0)).toBe('2 KB')
        })
    })

    describe('formatResolution', () => {
        it('should return "8K" for 7680x4320 and above', () => {
            expect(formatResolution('7680x4320')).toBe('8K')
            expect(formatResolution('8000x4500')).toBe('8K')
        })

        it('should return "4K" for 3840x2160 and above', () => {
            expect(formatResolution('3840x2160')).toBe('4K')
            expect(formatResolution('4000x2200')).toBe('4K')
        })

        it('should return "1440p" for 2560x1440 and above', () => {
            expect(formatResolution('2560x1440')).toBe('1440p')
        })

        it('should return "1080p" for 1920x1080 and above', () => {
            expect(formatResolution('1920x1080')).toBe('1080p')
        })

        it('should return "720p" for 1280x720 and above', () => {
            expect(formatResolution('1280x720')).toBe('720p')
        })

        it('should return "480p" for 720x480 and above', () => {
            expect(formatResolution('720x480')).toBe('480p')
        })

        it('should return "360p" for 640x360 and above', () => {
            expect(formatResolution('640x360')).toBe('360p')
        })

        it('should return "240p" for 426x240 and above', () => {
            expect(formatResolution('426x240')).toBe('240p')
        })

        it('should return raw resolution for non-standard sizes', () => {
            expect(formatResolution('320x240')).toBe('320x240')
            expect(formatResolution('100x100')).toBe('100x100')
        })

        it('should handle case-insensitive input', () => {
            expect(formatResolution('1920X1080')).toBe('1080p')
        })

        it('should throw error for invalid format', () => {
            expect(() => formatResolution('invalid')).toThrow(
                'Invalid resolution format'
            )
            expect(() => formatResolution('1920')).toThrow(
                'Invalid resolution format'
            )
        })
    })
})
