import { getPlaybackInfo, requestTranscode } from '@/services/libraryApi'
import { getToolJob, toolOutputPreviewUrl } from '@/services/toolsApi'
import { PlaybackInfo, PlaybackTranscode, VideoMetadata } from '@/types'
import {
    Maximize2,
    Pause,
    Play,
    RotateCcw,
    Volume2,
    VolumeX,
    Wand2,
} from 'lucide-react'

import { useEffect, useRef, useState } from 'react'

import { SERVER_URL } from '@/lib/env'
import { formatSeconds } from '@/lib/utils'

import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'

interface VideoPlayerProps {
    jobId: string
    metadata?: VideoMetadata
    className?: string
    isAudio?: boolean
}

export default function VideoPlayer({
    jobId,
    metadata,
    className = '',
    isAudio = false,
}: VideoPlayerProps) {
    const videoRef = useRef<HTMLVideoElement>(null)
    const [isPlaying, setIsPlaying] = useState(false)
    const [currentTime, setCurrentTime] = useState(0)
    const [duration, setDuration] = useState(0)
    const [volume, setVolume] = useState(1)
    const [isMuted, setIsMuted] = useState(false)
    const [isControlsVisible, setIsControlsVisible] = useState(true)
    const [, setIsFullscreen] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [retryKey, setRetryKey] = useState(0)
    const [playback, setPlayback] = useState<PlaybackInfo | null>(null)
    const [transcode, setTranscode] = useState<PlaybackTranscode | null>(null)
    const [transcodeError, setTranscodeError] = useState<string | null>(null)

    // A finished transcode is a browser-safe mp4 served by the tools output
    // endpoint; otherwise play the original download.
    const usingTranscode = transcode?.status === 'complete'
    const videoUrl = usingTranscode
        ? toolOutputPreviewUrl(transcode.job_id)
        : `${SERVER_URL}/video/${jobId}`
    const transcodeRunning =
        transcode?.status === 'pending' || transcode?.status === 'processing'
    // Codecs the <video> element can't decode (e.g. VP9/Opus remuxed into
    // mp4): don't offer a player that will just error — offer a transcode.
    const needsTranscode =
        playback !== null && !playback.browser_safe && !usingTranscode

    useEffect(() => {
        let cancelled = false
        getPlaybackInfo(jobId)
            .then((info) => {
                if (cancelled) return
                setPlayback(info)
                setTranscode(info.transcode ?? null)
            })
            .catch(() => {
                // Playback info is advisory — without it the player simply
                // tries the original file as before.
            })
        return () => {
            cancelled = true
        }
    }, [jobId])

    const transcodeJobId = transcode?.job_id
    useEffect(() => {
        if (!transcodeRunning || !transcodeJobId) return
        const timer = setInterval(async () => {
            try {
                const job = await getToolJob(transcodeJobId)
                setTranscode({
                    job_id: job.id,
                    status: job.status,
                    progress: job.progress,
                })
            } catch {
                // Keep polling; transient fetch errors shouldn't kill the UI.
            }
        }, 3000)
        return () => clearInterval(timer)
    }, [transcodeRunning, transcodeJobId])

    useEffect(() => {
        const video = videoRef.current
        if (!video) return

        const updateTime = () => setCurrentTime(video.currentTime)
        const updateDuration = () => setDuration(video.duration)
        const handleEnded = () => setIsPlaying(false)
        const handleError = () => {
            setError(
                usingTranscode
                    ? 'Failed to play the compatible version. Your browser may not support h264/aac playback.'
                    : playback && !playback.browser_safe
                      ? `Your browser can't decode this video (${playback.video_codec}/${playback.audio_codec}).`
                      : 'Failed to load video. The video file might not be available.'
            )
        }

        video.addEventListener('timeupdate', updateTime)
        video.addEventListener('loadedmetadata', updateDuration)
        video.addEventListener('ended', handleEnded)
        video.addEventListener('error', handleError)

        return () => {
            video.removeEventListener('timeupdate', updateTime)
            video.removeEventListener('loadedmetadata', updateDuration)
            video.removeEventListener('ended', handleEnded)
            video.removeEventListener('error', handleError)
        }
        // videoUrl/retryKey remount the <video> element, so listeners must
        // rebind to the new element.
    }, [videoUrl, retryKey, playback, usingTranscode])

    const startTranscode = async () => {
        setTranscodeError(null)
        try {
            setTranscode(await requestTranscode(jobId))
        } catch (e) {
            setTranscodeError(
                e instanceof Error ? e.message : 'Failed to start transcode'
            )
        }
    }

    const retry = () => {
        setError(null)
        // Changing the key remounts the <video> element, forcing a fresh load.
        setRetryKey((k) => k + 1)
    }

    const togglePlayPause = () => {
        const video = videoRef.current
        if (!video) return

        if (video.paused) {
            video.play()
            setIsPlaying(true)
        } else {
            video.pause()
            setIsPlaying(false)
        }
    }

    const handleSeek = (newTime: number[]) => {
        const video = videoRef.current
        if (!video) return

        video.currentTime = newTime[0]
        setCurrentTime(newTime[0])
    }

    const handleVolumeChange = (newVolume: number[]) => {
        const video = videoRef.current
        if (!video) return

        const volume = newVolume[0]
        video.volume = volume
        setVolume(volume)
        setIsMuted(volume === 0)
    }

    const toggleMute = () => {
        const video = videoRef.current
        if (!video) return

        if (isMuted) {
            video.volume = volume > 0 ? volume : 0.5
            setIsMuted(false)
        } else {
            video.volume = 0
            setIsMuted(true)
        }
    }

    const toggleFullscreen = () => {
        const video = videoRef.current
        if (!video) return

        if (!document.fullscreenElement) {
            video.requestFullscreen?.()
            setIsFullscreen(true)
        } else {
            document.exitFullscreen?.()
            setIsFullscreen(false)
        }
    }

    const restart = () => {
        const video = videoRef.current
        if (!video) return

        video.currentTime = 0
        setCurrentTime(0)
    }

    // An error while playing the transcoded version must NOT fall into this
    // branch — re-offering "Create compatible version" there would loop.
    if (needsTranscode && !error) {
        return (
            <div
                className={`bg-muted flex aspect-video items-center justify-center rounded-lg ${className}`}
            >
                <div className="max-w-md px-6 text-center">
                    <p className="mb-1 font-medium">
                        This video can&apos;t be played in your browser
                    </p>
                    <p className="text-muted-foreground mb-4 text-sm">
                        It was saved with codecs your browser can&apos;t decode
                        ({playback?.video_codec || 'unknown'}/
                        {playback?.audio_codec || 'unknown'}). Create a
                        compatible version to watch it here — the original file
                        stays untouched.
                    </p>
                    {transcodeRunning ? (
                        <p className="text-muted-foreground text-sm">
                            Creating compatible version…{' '}
                            {Math.round(transcode?.progress ?? 0)}%
                        </p>
                    ) : (
                        <Button onClick={startTranscode}>
                            <Wand2 className="mr-2 h-4 w-4" />
                            Create compatible version
                        </Button>
                    )}
                    {transcode?.status === 'failed' && (
                        <p className="text-destructive mt-2 text-sm">
                            The last transcode failed — you can try again.
                        </p>
                    )}
                    {transcodeError && (
                        <p className="text-destructive mt-2 text-sm">
                            {transcodeError}
                        </p>
                    )}
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div
                className={`bg-muted flex aspect-video items-center justify-center rounded-lg ${className}`}
            >
                <div className="text-center">
                    <p className="text-muted-foreground mb-2">{error}</p>
                    <Button variant="outline" onClick={retry}>
                        Retry
                    </Button>
                </div>
            </div>
        )
    }

    return (
        <div
            className={`group relative aspect-video overflow-hidden rounded-lg bg-black ${className}`}
            onMouseEnter={() => setIsControlsVisible(true)}
            onMouseLeave={() => setIsControlsVisible(false)}
        >
            <video
                key={`${videoUrl}-${retryKey}`}
                ref={videoRef}
                src={videoUrl}
                className="h-full w-full object-contain"
                onClick={togglePlayPause}
                poster={metadata?.thumbnail}
            />

            {/* Audio-only files have no video track, so keep the thumbnail
                visible during playback instead of a black frame. */}
            {isAudio && metadata?.thumbnail && (
                <img
                    src={metadata.thumbnail}
                    alt=""
                    className="pointer-events-none absolute inset-0 h-full w-full object-contain"
                />
            )}

            {/* Controls overlay */}
            <div
                className={`absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent transition-opacity duration-300 ${
                    isControlsVisible ? 'opacity-100' : 'opacity-0'
                }`}
            >
                {/* Play button overlay */}
                <div className="absolute inset-0 flex items-center justify-center">
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-16 w-16 rounded-full bg-black/50 text-white hover:bg-black/70"
                        onClick={togglePlayPause}
                    >
                        {isPlaying ? (
                            <Pause className="h-8 w-8" />
                        ) : (
                            <Play className="h-8 w-8" />
                        )}
                    </Button>
                </div>

                {/* Bottom controls */}
                <div className="absolute right-0 bottom-0 left-0 p-4">
                    {/* Progress bar */}
                    <div className="mb-3">
                        <Slider
                            value={[currentTime]}
                            max={duration || 100}
                            step={1}
                            onValueChange={handleSeek}
                            className="w-full"
                        />
                    </div>

                    {/* Control buttons */}
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-white hover:bg-white/20"
                                onClick={togglePlayPause}
                            >
                                {isPlaying ? (
                                    <Pause className="h-4 w-4" />
                                ) : (
                                    <Play className="h-4 w-4" />
                                )}
                            </Button>

                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-white hover:bg-white/20"
                                onClick={restart}
                            >
                                <RotateCcw className="h-4 w-4" />
                            </Button>

                            <div className="flex items-center gap-2">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 text-white hover:bg-white/20"
                                    onClick={toggleMute}
                                >
                                    {isMuted ? (
                                        <VolumeX className="h-4 w-4" />
                                    ) : (
                                        <Volume2 className="h-4 w-4" />
                                    )}
                                </Button>
                                <div className="w-20">
                                    <Slider
                                        value={[isMuted ? 0 : volume]}
                                        max={1}
                                        step={0.1}
                                        onValueChange={handleVolumeChange}
                                        className="w-full"
                                    />
                                </div>
                            </div>

                            <span className="text-sm text-white">
                                {formatSeconds(currentTime)} /{' '}
                                {formatSeconds(duration)}
                            </span>
                        </div>

                        {!isAudio && (
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-white hover:bg-white/20"
                                onClick={toggleFullscreen}
                            >
                                <Maximize2 className="h-4 w-4" />
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
