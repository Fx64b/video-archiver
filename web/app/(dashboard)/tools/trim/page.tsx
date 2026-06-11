'use client'

import { Scissors } from 'lucide-react'

import { useState } from 'react'

import { useToolSubmit } from '@/hooks/useToolSubmit'

import ToolPageShell from '@/components/tools/ToolPageShell'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'

const TIME_RE = /^(\d+:)?\d{1,2}:\d{2}(\.\d+)?$|^\d+(\.\d+)?$/

export default function TrimPage() {
    const { submit, isSubmitting, error, setError } = useToolSubmit('trim')
    const [startTime, setStartTime] = useState('00:00:00')
    const [endTime, setEndTime] = useState('00:00:10')
    const [reEncode, setReEncode] = useState(false)

    const handleSubmit = () => {
        if (!TIME_RE.test(startTime) || !TIME_RE.test(endTime)) {
            setError('Times must be HH:MM:SS or seconds')
            return
        }
        submit({
            start_time: startTime,
            end_time: endTime,
            re_encode: reEncode,
        })
    }

    return (
        <ToolPageShell
            title="Trim Video"
            description="Cut videos to a specific time range"
            icon={<Scissors className="h-6 w-6" />}
            submitLabel="Start Trimming"
            isSubmitting={isSubmitting}
            error={error}
            onSubmit={handleSubmit}
            tips={
                <Card className="bg-muted/50 border-muted">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-sm">Tips</CardTitle>
                    </CardHeader>
                    <CardContent className="text-muted-foreground space-y-1.5 text-xs">
                        <p>
                            • Re-encode for frame-accurate cuts at exact
                            timestamps
                        </p>
                        <p>
                            • Without re-encode, cuts snap to the nearest
                            keyframe (faster)
                        </p>
                    </CardContent>
                </Card>
            }
        >
            <div className="space-y-2">
                <Label htmlFor="start-time">Start Time</Label>
                <Input
                    id="start-time"
                    placeholder="00:00:00"
                    value={startTime}
                    onChange={(e) => setStartTime(e.target.value)}
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="end-time">End Time</Label>
                <Input
                    id="end-time"
                    placeholder="00:00:10"
                    value={endTime}
                    onChange={(e) => setEndTime(e.target.value)}
                />
            </div>
            <div className="flex items-center justify-between space-x-2 pt-2">
                <Label htmlFor="re-encode" className="flex flex-col gap-1">
                    <span>Re-encode</span>
                    <span className="text-muted-foreground text-xs font-normal">
                        Enable for precise cutting
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
