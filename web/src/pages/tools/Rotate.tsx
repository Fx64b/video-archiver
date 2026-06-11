import { RotateCw } from 'lucide-react'

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

export default function RotatePage() {
    const { submit, isSubmitting, error, setError } = useToolSubmit('rotate')
    const [rotation, setRotation] = useState('90')
    const [flipH, setFlipH] = useState(false)
    const [flipV, setFlipV] = useState(false)

    const handleSubmit = () => {
        const degrees = parseInt(rotation, 10)
        if (degrees === 0 && !flipH && !flipV) {
            setError('Choose a rotation or a flip')
            return
        }
        submit({ rotation: degrees, flip_h: flipH, flip_v: flipV })
    }

    return (
        <ToolPageShell
            title="Rotate Video"
            description="Rotate or flip a video"
            icon={<RotateCw className="h-6 w-6" />}
            submitLabel="Start Rotating"
            isSubmitting={isSubmitting}
            error={error}
            onSubmit={handleSubmit}
        >
            <div className="space-y-2">
                <Label>Rotation</Label>
                <Select value={rotation} onValueChange={setRotation}>
                    <SelectTrigger>
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="0">None</SelectItem>
                        <SelectItem value="90">90° clockwise</SelectItem>
                        <SelectItem value="180">180°</SelectItem>
                        <SelectItem value="270">270° clockwise</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            <div className="flex items-center justify-between space-x-2 pt-2">
                <Label htmlFor="flip-h">Flip horizontally</Label>
                <Switch
                    id="flip-h"
                    checked={flipH}
                    onCheckedChange={setFlipH}
                />
            </div>
            <div className="flex items-center justify-between space-x-2">
                <Label htmlFor="flip-v">Flip vertically</Label>
                <Switch
                    id="flip-v"
                    checked={flipV}
                    onCheckedChange={setFlipV}
                />
            </div>
        </ToolPageShell>
    )
}
