import { ChevronDown, SortAsc, SortDesc } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface SortOption {
    label: string
    value: string
}

interface SortControlsProps {
    sortOptions: SortOption[]
    currentSort: string
    currentOrder: 'asc' | 'desc'
    onSortChange: (value: string) => void
    onOrderChange: () => void
}

export function SortControls({
    sortOptions,
    currentSort,
    currentOrder,
    onSortChange,
    onOrderChange,
}: SortControlsProps) {
    const getCurrentSortLabel = () => {
        const option = sortOptions.find((opt) => opt.value === currentSort)
        return option?.label || 'Sort by'
    }

    return (
        <div className="mb-4 flex items-center gap-2">
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="flex items-center gap-2">
                        <span>{getCurrentSortLabel()}</span>
                        <ChevronDown className="h-4 w-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                    <DropdownMenuLabel>Sort by</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    {sortOptions.map((option) => (
                        <DropdownMenuItem
                            key={option.value}
                            className={
                                currentSort === option.value ? 'bg-accent' : ''
                            }
                            onClick={() => onSortChange(option.value)}
                        >
                            {option.label}
                        </DropdownMenuItem>
                    ))}
                </DropdownMenuContent>
            </DropdownMenu>

            <Button
                variant="outline"
                size="icon"
                onClick={onOrderChange}
                title={
                    currentOrder === 'desc'
                        ? 'Descending order'
                        : 'Ascending order'
                }
            >
                {currentOrder === 'desc' ? (
                    <SortDesc className="h-4 w-4" />
                ) : (
                    <SortAsc className="h-4 w-4" />
                )}
            </Button>
        </div>
    )
}
