import { Layers } from 'lucide-react'

import { useState } from 'react'

import { useToolSubmit } from '@/hooks/useToolSubmit'

import ToolPageShell from '@/components/tools/ToolPageShell'
import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

export default function ConcatPage() {
    const { submit, isSubmitting, error } = useToolSubmit('concat')
    const [outputFormat, setOutputFormat] = useState('mp4')
    const [reEncode, setReEncode] = useState(false)

    const handleSubmit = () =>
        submit({ output_format: outputFormat, re_encode: reEncode })

    return (
        <ToolPageShell
            title="Concatenate Videos"
            description="Merge multiple videos into a single file"
            icon={<Layers className="h-6 w-6" />}
            minSelection={2}
            submitLabel="Start Merging"
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
                        <SelectItem value="mkv">MKV</SelectItem>
                        <SelectItem value="webm">WebM</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            <div className="flex items-center justify-between space-x-2 pt-2">
                <Label htmlFor="re-encode" className="flex flex-col gap-1">
                    <span>Re-encode</span>
                    <span className="text-muted-foreground text-xs font-normal">
                        Required when clips have different codecs
                    </span>
                </Label>
                <Switch
                    id="re-encode"
                    checked={reEncode}
                    onCheckedChange={setReEncode}
                />
            </div>
        </ToolPageShell>
    )
}
