import { FileVideo } from 'lucide-react'

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

export default function ConvertPage() {
    const { submit, isSubmitting, error } = useToolSubmit('convert')
    const [outputFormat, setOutputFormat] = useState('mp4')
    const [bitrate, setBitrate] = useState('')

    const handleSubmit = () =>
        submit({ output_format: outputFormat, bitrate: bitrate.trim() })

    return (
        <ToolPageShell
            title="Convert Format"
            description="Convert videos between container formats"
            icon={<FileVideo className="h-6 w-6" />}
            submitLabel="Start Converting"
            isSubmitting={isSubmitting}
            error={error}
            onSubmit={handleSubmit}
        >
            <div className="space-y-2">
                <Label>Output Format</Label>
                <Select value={outputFormat} onValueChange={setOutputFormat}>
                    <SelectTrigger>
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="mp4">MP4</SelectItem>
                        <SelectItem value="webm">WebM</SelectItem>
                        <SelectItem value="mkv">MKV</SelectItem>
                        <SelectItem value="avi">AVI</SelectItem>
                        <SelectItem value="mov">MOV</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            <div className="space-y-2">
                <Label htmlFor="bitrate">Video Bitrate (optional)</Label>
                <Input
                    id="bitrate"
                    placeholder="e.g. 2M, 5M"
                    value={bitrate}
                    onChange={(e) => setBitrate(e.target.value)}
                />
            </div>
        </ToolPageShell>
    )
}
