'use client'

import { VideoMetadata } from '@/types'
import {
    Maximize2,
    Pause,
    Play,
    RotateCcw,
    Volume2,
    VolumeX,
} from 'lucide-react'

import { useEffect, useRef, useState } from 'react'

import { formatSeconds } from '@/lib/utils'

import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'

interface VideoPlayerProps {
    jobId: string
    metadata?: VideoMetadata
    className?: string
}

export default function VideoPlayer({
    jobId,
    metadata,
    className = '',
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

    const videoUrl = `${process.env.NEXT_PUBLIC_SERVER_URL}/video/${jobId}`

    useEffect(() => {
        const video = videoRef.current
        if (!video) return

        const updateTime = () => setCurrentTime(video.currentTime)
        const updateDuration = () => setDuration(video.duration)
        const handleEnded = () => setIsPlaying(false)
        const handleError = () => {
            setError(
                'Failed to load video. The video file might not be available.'
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
    }, [])

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

    if (error) {
        return (
            <div
                className={`bg-muted flex aspect-video items-center justify-center rounded-lg ${className}`}
            >
                <div className="text-center">
                    <p className="text-muted-foreground mb-2">{error}</p>
                    <Button variant="outline" onClick={() => setError(null)}>
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
                ref={videoRef}
                src={videoUrl}
                className="h-full w-full object-contain"
                onClick={togglePlayPause}
                poster={metadata?.thumbnail}
            />

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

                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-white hover:bg-white/20"
                            onClick={toggleFullscreen}
                        >
                            <Maximize2 className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
