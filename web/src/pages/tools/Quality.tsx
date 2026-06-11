import { Settings2 } from 'lucide-react'

import { useState } from 'react'

import { useToolSubmit } from '@/hooks/useToolSubmit'

import ToolPageShell from '@/components/tools/ToolPageShell'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

export default function QualityPage() {
    const { submit, isSubmitting, error, setError } =
        useToolSubmit('adjust_quality')
    const [resolution, setResolution] = useState('720p')
    const [crf, setCrf] = useState('23')
    const [bitrate, setBitrate] = useState('')
    const [twoPass, setTwoPass] = useState(false)

    const handleSubmit = () => {
        if (twoPass && !bitrate.trim()) {
            setError('Two-pass encoding requires a target bitrate')
            return
        }
        submit({
            resolution: resolution === 'original' ? '' : resolution,
            crf: parseInt(crf, 10) || 0,
            bitrate: bitrate.trim(),
            two_pass: twoPass,
        })
    }

    return (
        <ToolPageShell
            title="Adjust Quality"
            description="Change resolution, bitrate or CRF"
            icon={<Settings2 className="h-6 w-6" />}
            submitLabel="Start Processing"
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
                            • Lower CRF means higher quality and bigger files
                            (18–28 is typical)
                        </p>
                        <p>
                            • Two-pass gives more accurate bitrate targeting but
                            is slower
                        </p>
                    </CardContent>
                </Card>
            }
        >
            <div className="space-y-2">
                <Label>Resolution</Label>
                <Select value={resolution} onValueChange={setResolution}>
                    <SelectTrigger>
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="original">Keep original</SelectItem>
                        <SelectItem value="2160p">2160p (4K)</SelectItem>
                        <SelectItem value="1440p">1440p</SelectItem>
                        <SelectItem value="1080p">1080p</SelectItem>
                        <SelectItem value="720p">720p</SelectItem>
                        <SelectItem value="480p">480p</SelectItem>
                        <SelectItem value="360p">360p</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            <div className="space-y-2">
                <Label htmlFor="crf">CRF (0–51)</Label>
                <Input
                    id="crf"
                    type="number"
                    min={0}
                    max={51}
                    value={crf}
                    onChange={(e) => setCrf(e.target.value)}
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="bitrate">Target Bitrate (optional)</Label>
                <Input
                    id="bitrate"
                    placeholder="e.g. 4M"
                    value={bitrate}
                    onChange={(e) => setBitrate(e.target.value)}
                />
            </div>
            <div className="flex items-center justify-between space-x-2 pt-2">
                <Label htmlFor="two-pass" className="flex flex-col gap-1">
                    <span>Two-pass encoding</span>
                    <span className="text-muted-foreground text-xs font-normal">
                        Requires a target bitrate
                    </span>
                </Label>
                <Switch
                    id="two-pass"
                    checked={twoPass}
                    onCheckedChange={setTwoPass}
                />
            </div>
        </ToolPageShell>
    )
}
