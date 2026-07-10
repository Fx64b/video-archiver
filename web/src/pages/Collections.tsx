import {
    createCollection,
    deleteCollection,
    listCollections,
    updateCollection,
} from '@/services/collectionsApi'
import { Collection } from '@/types'
import {
    AlertCircle,
    FolderOpen,
    MoreVertical,
    Pencil,
    Plus,
    Trash2,
} from 'lucide-react'
import { toast } from 'sonner'

import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { CollectionFormDialog } from '@/components/collections/CollectionFormDialog'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'

export default function CollectionsPage() {
    const navigate = useNavigate()
    const [collections, setCollections] = useState<Collection[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [formOpen, setFormOpen] = useState(false)
    const [editTarget, setEditTarget] = useState<Collection | null>(null)
    const [deleteTarget, setDeleteTarget] = useState<Collection | null>(null)

    const refresh = useCallback(async () => {
        try {
            setCollections(await listCollections())
            setError(null)
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : 'Failed to load collections'
            )
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        refresh()
    }, [refresh])

    const handleSubmit = async (name: string, description: string) => {
        if (editTarget) {
            await updateCollection(editTarget.id, name, description)
            toast.success('Collection updated')
        } else {
            await createCollection(name, description)
            toast.success('Collection created')
        }
        await refresh()
    }

    const handleDelete = async () => {
        if (!deleteTarget) return
        try {
            await deleteCollection(deleteTarget.id)
            toast.success('Collection deleted')
            await refresh()
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to delete')
        }
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-6 p-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="mb-2 text-3xl font-bold">Collections</h1>
                    <p className="text-muted-foreground">
                        Custom playlists of your downloaded videos — process
                        them like a playlist in the tools section
                    </p>
                </div>
                <Button
                    className="gap-2"
                    onClick={() => {
                        setEditTarget(null)
                        setFormOpen(true)
                    }}
                >
                    <Plus className="h-4 w-4" />
                    New Collection
                </Button>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            {loading ? (
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                        <Card key={i}>
                            <CardContent className="p-0">
                                <Skeleton className="aspect-video w-full rounded-t-lg" />
                                <div className="space-y-2 p-4">
                                    <Skeleton className="h-4 w-full" />
                                    <Skeleton className="h-3 w-20" />
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            ) : collections.length === 0 ? (
                <div className="text-muted-foreground py-16 text-center">
                    <FolderOpen className="mx-auto mb-4 h-12 w-12" />
                    <p className="mb-1 font-medium">No collections yet</p>
                    <p className="text-sm">
                        Create a collection here, or add videos from their
                        detail page.
                    </p>
                </div>
            ) : (
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {collections.map((collection) => (
                        <Card
                            key={collection.id}
                            className="hover:border-primary cursor-pointer overflow-hidden border pt-0 transition-all hover:shadow-md"
                            onClick={() =>
                                navigate(`/collections/${collection.id}`)
                            }
                        >
                            <CardContent className="p-0">
                                <div className="bg-muted relative aspect-video">
                                    {collection.thumbnail ? (
                                        <img
                                            src={collection.thumbnail}
                                            alt={collection.name}
                                            className="absolute inset-0 h-full w-full object-cover"
                                        />
                                    ) : (
                                        <div className="absolute inset-0 flex items-center justify-center">
                                            <FolderOpen className="text-muted-foreground h-8 w-8" />
                                        </div>
                                    )}
                                    <div className="absolute right-2 bottom-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
                                        {collection.video_count} video
                                        {collection.video_count === 1
                                            ? ''
                                            : 's'}
                                    </div>
                                </div>
                                <div className="flex items-start justify-between gap-2 p-4">
                                    <div className="min-w-0">
                                        <h3 className="truncate font-semibold">
                                            {collection.name}
                                        </h3>
                                        {collection.description && (
                                            <p className="text-muted-foreground mt-1 line-clamp-2 text-sm">
                                                {collection.description}
                                            </p>
                                        )}
                                    </div>
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 shrink-0"
                                                onClick={(e) =>
                                                    e.stopPropagation()
                                                }
                                            >
                                                <MoreVertical className="h-4 w-4" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent
                                            align="end"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <DropdownMenuItem
                                                onClick={() => {
                                                    setEditTarget(collection)
                                                    setFormOpen(true)
                                                }}
                                            >
                                                <Pencil className="mr-2 h-4 w-4" />
                                                Edit
                                            </DropdownMenuItem>
                                            <DropdownMenuItem
                                                className="text-destructive focus:text-destructive"
                                                onClick={() =>
                                                    setDeleteTarget(collection)
                                                }
                                            >
                                                <Trash2 className="mr-2 h-4 w-4" />
                                                Delete
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            <CollectionFormDialog
                open={formOpen}
                onOpenChange={setFormOpen}
                collection={editTarget}
                onSubmit={handleSubmit}
            />
            <ConfirmDialog
                open={deleteTarget !== null}
                onOpenChange={(open) => !open && setDeleteTarget(null)}
                title="Delete this collection?"
                description="The collection is removed, but the videos in it stay in your library."
                onConfirm={handleDelete}
            />
        </div>
    )
}
