import { Collection } from '@/types'
import { Loader2 } from 'lucide-react'

import { useEffect, useState } from 'react'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface CollectionFormDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    /** Collection to edit; omit to create a new one. */
    collection?: Collection | null
    /** Called with the entered values. The dialog closes when it resolves. */
    onSubmit: (name: string, description: string) => Promise<void>
}

/**
 * Dialog for creating or renaming a collection. The same form serves both
 * flows; the presence of `collection` decides labels and initial values.
 */
export function CollectionFormDialog({
    open,
    onOpenChange,
    collection,
    onSubmit,
}: CollectionFormDialogProps) {
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const isEdit = Boolean(collection)

    // Re-seed the form whenever the dialog opens for a (different) target.
    useEffect(() => {
        if (open) {
            setName(collection?.name ?? '')
            setDescription(collection?.description ?? '')
            setError(null)
        }
    }, [open, collection])

    const handleSubmit = async () => {
        if (!name.trim()) {
            setError('A name is required')
            return
        }
        setBusy(true)
        setError(null)
        try {
            await onSubmit(name.trim(), description.trim())
            onOpenChange(false)
        } catch (err) {
            setError(
                err instanceof Error ? err.message : 'Failed to save collection'
            )
        } finally {
            setBusy(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={(o) => !busy && onOpenChange(o)}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>
                        {isEdit ? 'Edit Collection' : 'New Collection'}
                    </DialogTitle>
                    <DialogDescription>
                        {isEdit
                            ? 'Change the name or description of this collection.'
                            : 'Group downloaded videos into a custom playlist you can manage and process in the tools section.'}
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="collection-name">Name</Label>
                        <Input
                            id="collection-name"
                            placeholder="e.g. Watch Later"
                            value={name}
                            maxLength={100}
                            onChange={(e) => setName(e.target.value)}
                            onKeyDown={(e) =>
                                e.key === 'Enter' && handleSubmit()
                            }
                            autoFocus
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="collection-description">
                            Description (optional)
                        </Label>
                        <Input
                            id="collection-description"
                            placeholder="What belongs in this collection?"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            onKeyDown={(e) =>
                                e.key === 'Enter' && handleSubmit()
                            }
                        />
                    </div>
                    {error && (
                        <p className="text-destructive text-sm">{error}</p>
                    )}
                </div>
                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={busy}
                    >
                        Cancel
                    </Button>
                    <Button onClick={handleSubmit} disabled={busy}>
                        {busy && (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        )}
                        {isEdit ? 'Save' : 'Create'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
