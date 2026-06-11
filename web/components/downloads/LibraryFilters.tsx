import { listTags } from '@/services/libraryApi'
import { Tag } from '@/types'
import { ChevronDown, Search, TagIcon, X } from 'lucide-react'

import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'

interface LibraryFiltersProps {
    search: string
    tag: string
    onSearchChange: (value: string) => void
    onTagChange: (value: string) => void
}

const SEARCH_DEBOUNCE_MS = 300

/**
 * Search box and tag filter for the downloads library. Search input is
 * debounced before it reaches the URL (and triggers a fetch); the tag list is
 * loaded from the backend catalog with usage counts.
 */
export function LibraryFilters({
    search,
    tag,
    onSearchChange,
    onTagChange,
}: LibraryFiltersProps) {
    const [draft, setDraft] = useState(search)
    const [tags, setTags] = useState<Tag[]>([])
    const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

    // Keep the input in sync when the URL changes from elsewhere
    // (e.g. back navigation).
    useEffect(() => {
        setDraft(search)
    }, [search])

    useEffect(() => {
        listTags()
            .then(setTags)
            .catch((err) => console.error('Failed to load tags:', err))
    }, [])

    const handleInput = (value: string) => {
        setDraft(value)
        if (debounceRef.current) {
            clearTimeout(debounceRef.current)
        }
        debounceRef.current = setTimeout(
            () => onSearchChange(value.trim()),
            SEARCH_DEBOUNCE_MS
        )
    }

    const clearSearch = () => {
        if (debounceRef.current) {
            clearTimeout(debounceRef.current)
        }
        setDraft('')
        onSearchChange('')
    }

    return (
        <div className="mb-6 flex flex-col gap-2 sm:flex-row sm:items-center">
            <div className="relative flex-1 sm:max-w-sm">
                <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                <Input
                    value={draft}
                    onChange={(e) => handleInput(e.target.value)}
                    placeholder="Search by title or channel…"
                    className="pr-8 pl-9"
                    aria-label="Search downloads"
                />
                {draft && (
                    <button
                        type="button"
                        aria-label="Clear search"
                        onClick={clearSearch}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2.5 -translate-y-1/2"
                    >
                        <X className="h-4 w-4" />
                    </button>
                )}
            </div>

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        variant={tag ? 'secondary' : 'outline'}
                        className="flex items-center gap-2"
                    >
                        <TagIcon className="h-4 w-4" />
                        <span className="max-w-32 truncate">
                            {tag || 'All tags'}
                        </span>
                        <ChevronDown className="h-4 w-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="start"
                    className="max-h-80 w-56 overflow-y-auto"
                >
                    <DropdownMenuLabel>Filter by tag</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                        className={!tag ? 'bg-accent' : ''}
                        onClick={() => onTagChange('')}
                    >
                        All tags
                    </DropdownMenuItem>
                    {tags.map((t) => (
                        <DropdownMenuItem
                            key={t.id}
                            className={
                                tag.toLowerCase() === t.name.toLowerCase()
                                    ? 'bg-accent'
                                    : ''
                            }
                            onClick={() => onTagChange(t.name)}
                        >
                            <span className="flex-1 truncate">{t.name}</span>
                            <span className="text-muted-foreground ml-2 text-xs">
                                {t.count}
                            </span>
                        </DropdownMenuItem>
                    ))}
                    {tags.length === 0 && (
                        <DropdownMenuItem disabled>
                            No tags yet
                        </DropdownMenuItem>
                    )}
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    )
}
