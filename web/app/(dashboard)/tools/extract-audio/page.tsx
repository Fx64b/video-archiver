'use client'

import { FileAudio } from 'lucide-react'

import { useState } from 'react'

import { useToolSubmit } from '@/hooks/useToolSubmit'

import ToolPageShell from '@/components/tools/ToolPageShell'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'

export default function ExtractAudioPage() {
    const { submit, isSubmitting, error } = useToolSubmit('extract_audio')
    const [format, setFormat] = useState('mp3')
    const [bitrate, setBitrate] = useState('192k')
    const [sampleRate, setSampleRate] = useState('44100')

    const handleSubmit = () =>
        submit({
            output_format: format,
            bitrate: bitrate.trim(),
            sample_rate: parseInt(sampleRate, 10) || 0,
        })

    // Lossless formats ignore bitrate.
    const lossless = format === 'flac' || format === 'wav'

    return (
        <ToolPageShell
            title="Extract Audio"
            description="Extract the audio track from a video"
            icon={<FileAudio className="h-6 w-6" />}
            submitLabel="Start Extracting"
            isSubmitting={isSubmitting}
            error={error}
            onSubmit={handleSubmit}
        >
            <div className="space-y-2">
                <Label>Audio Format</Label>
                <Select value={format} onValueChange={setFormat}>
                    <SelectTrigger>
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="mp3">MP3</SelectItem>
                        <SelectItem value="aac">AAC</SelectItem>
                        <SelectItem value="ogg">OGG</SelectItem>
                        <SelectItem value="flac">FLAC (lossless)</SelectItem>
                        <SelectItem value="wav">WAV (lossless)</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            {!lossless && (
                <div className="space-y-2">
                    <Label>Bitrate</Label>
                    <Select value={bitrate} onValueChange={setBitrate}>
                        <SelectTrigger>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="128k">128 kbps</SelectItem>
                            <SelectItem value="192k">192 kbps</SelectItem>
                            <SelectItem value="256k">256 kbps</SelectItem>
                            <SelectItem value="320k">320 kbps</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            )}
            <div className="space-y-2">
                <Label htmlFor="sample-rate">Sample Rate (Hz)</Label>
                <Input
                    id="sample-rate"
                    type="number"
                    placeholder="44100"
                    value={sampleRate}
                    onChange={(e) => setSampleRate(e.target.value)}
                />
            </div>
        </ToolPageShell>
    )
}
