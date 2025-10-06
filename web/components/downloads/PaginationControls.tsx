import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from '@/components/ui/pagination'

interface PaginationControlsProps {
    currentPage: number
    totalPages: number
    onPageChange: (page: number) => void
}

export function PaginationControls({
    currentPage,
    totalPages,
    onPageChange,
}: PaginationControlsProps) {
    if (totalPages <= 1) return null

    const maxPagesToShow = 5
    const startPage = Math.max(
        1,
        currentPage - Math.floor(maxPagesToShow / 2)
    )
    const endPage = Math.min(totalPages, startPage + maxPagesToShow - 1)

    const pages = []

    // Add first page if not included in the range
    if (startPage > 1) {
        pages.push(
            <PaginationItem key="first">
                <PaginationLink onClick={() => onPageChange(1)}>
                    1
                </PaginationLink>
            </PaginationItem>
        )

        if (startPage > 2) {
            pages.push(
                <PaginationItem key="ellipsis-start">
                    <PaginationEllipsis />
                </PaginationItem>
            )
        }
    }

    // Add pages in the calculated range
    for (let i = startPage; i <= endPage; i++) {
        pages.push(
            <PaginationItem key={i}>
                <PaginationLink
                    isActive={currentPage === i}
                    onClick={() => onPageChange(i)}
                >
                    {i}
                </PaginationLink>
            </PaginationItem>
        )
    }

    // Add last page if not included in the range
    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            pages.push(
                <PaginationItem key="ellipsis-end">
                    <PaginationEllipsis />
                </PaginationItem>
            )
        }

        pages.push(
            <PaginationItem key="last">
                <PaginationLink onClick={() => onPageChange(totalPages)}>
                    {totalPages}
                </PaginationLink>
            </PaginationItem>
        )
    }

    return (
        <Pagination className="mt-8">
            <PaginationContent>
                <PaginationItem>
                    <PaginationPrevious
                        onClick={() =>
                            currentPage > 1 && onPageChange(currentPage - 1)
                        }
                        className={
                            currentPage === 1
                                ? 'pointer-events-none opacity-50'
                                : ''
                        }
                    />
                </PaginationItem>

                {pages}

                <PaginationItem>
                    <PaginationNext
                        onClick={() =>
                            currentPage < totalPages &&
                            onPageChange(currentPage + 1)
                        }
                        className={
                            currentPage === totalPages
                                ? 'pointer-events-none opacity-50'
                                : ''
                        }
                    />
                </PaginationItem>
            </PaginationContent>
        </Pagination>
    )
}
