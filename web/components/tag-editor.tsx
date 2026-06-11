import { addJobTags, removeJobTag } from '@/services/libraryApi'
import { Tag } from '@/types'
import { Plus, Sparkles, X } from 'lucide-react'
import { toast } from 'sonner'

import { useState } from 'react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip'

interface TagEditorProps {
    jobId: string
    tags: Tag[]
    onChange: (tags: Tag[]) => void
}

/**
 * Editable tag list for a download. Auto-generated tags (from the video's
 * category, channel and keywords) are marked with a sparkle; user tags can be
 * added with the input and any tag can be removed.
 */
export function TagEditor({ jobId, tags, onChange }: TagEditorProps) {
    const [input, setInput] = useState('')
    const [busy, setBusy] = useState(false)

    const addTag = async () => {
        const name = input.trim()
        if (!name) return
        setBusy(true)
        try {
            const updated = await addJobTags(jobId, [name])
            onChange(updated)
            setInput('')
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : 'Failed to add tag'
            )
        } finally {
            setBusy(false)
        }
    }

    const removeTag = async (tag: Tag) => {
        try {
            await removeJobTag(jobId, tag.id)
            onChange(tags.filter((t) => t.id !== tag.id))
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : 'Failed to remove tag'
            )
        }
    }

    return (
        <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
                {tags.length === 0 && (
                    <p className="text-muted-foreground text-sm">
                        No tags yet — add one below
                    </p>
                )}
                {tags.map((tag) => (
                    <Badge
                        key={tag.id}
                        variant={
                            tag.source === 'auto' ? 'secondary' : 'default'
                        }
                        className="gap-1 pr-1 text-xs"
                    >
                        {tag.source === 'auto' && (
                            <TooltipProvider>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <Sparkles className="h-3 w-3" />
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        <p>Added automatically from metadata</p>
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>
                        )}
                        {tag.name}
                        <button
                            type="button"
                            aria-label={`Remove tag ${tag.name}`}
                            className="hover:bg-background/30 ml-0.5 rounded-full p-0.5"
                            onClick={() => removeTag(tag)}
                        >
                            <X className="h-3 w-3" />
                        </button>
                    </Badge>
                ))}
            </div>
            <div className="flex gap-2">
                <Input
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                            e.preventDefault()
                            addTag()
                        }
                    }}
                    placeholder="Add a tag…"
                    className="h-8 text-sm"
                    maxLength={50}
                />
                <Button
                    size="sm"
                    variant="outline"
                    onClick={addTag}
                    disabled={busy || !input.trim()}
                >
                    <Plus className="h-4 w-4" />
                    Add
                </Button>
            </div>
        </div>
    )
}
