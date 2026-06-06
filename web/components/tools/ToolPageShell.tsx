'use client'

import useToolsState from '@/store/toolsState'
import { AlertCircle, ArrowLeft } from 'lucide-react'

import { ReactNode, useEffect } from 'react'

import Image from 'next/image'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'

interface ToolPageShellProps {
    title: string
    description: string
    icon: ReactNode
    /** Minimum number of selected items required to run the tool. */
    minSelection?: number
    submitLabel: string
    submittingLabel?: string
    isSubmitting: boolean
    error: string | null
    onSubmit: () => void
    /** Configuration form rendered in the left column. */
    children: ReactNode
    /** Optional tips card rendered under the configuration form. */
    tips?: ReactNode
}

/**
 * ToolPageShell provides the shared layout for every tool page: the header,
 * the selected-videos preview, error display and action buttons. Pages only
 * need to supply their configuration form and an onSubmit handler, which
 * removes the heavy duplication the previous per-page implementations had.
 */
export default function ToolPageShell({
    title,
    description,
    icon,
    minSelection = 1,
    submitLabel,
    submittingLabel = 'Starting...',
    isSubmitting,
    error,
    onSubmit,
    children,
    tips,
}: ToolPageShellProps) {
    const router = useRouter()
    const { selectedInputs, clearSelectedInputs } = useToolsState()

    // Send the user back to the dashboard if they arrive without a selection
    // (e.g. on a hard refresh, since the selection lives in memory).
    useEffect(() => {
        if (selectedInputs.length === 0) {
            router.push('/tools')
        }
    }, [selectedInputs.length, router])

    if (selectedInputs.length === 0) {
        return null
    }

    const notEnough = selectedInputs.length < minSelection

    const handleCancel = () => {
        clearSelectedInputs()
        router.push('/tools')
    }

    return (
        <div className="flex min-h-screen w-full flex-col gap-6">
            <div className="flex items-center gap-4">
                <Link href="/tools">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                </Link>
                <div className="flex items-center gap-3">
                    <div className="text-muted-foreground">{icon}</div>
                    <div>
                        <h1 className="text-2xl font-bold">{title}</h1>
                        <p className="text-muted-foreground text-sm">
                            {description}
                        </p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                <div className="space-y-6 lg:col-span-1">
                    <Card>
                        <CardHeader>
                            <CardTitle>Configuration</CardTitle>
                            <CardDescription>
                                Set the options for this tool
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {children}
                        </CardContent>
                    </Card>
                    {tips}
                </div>

                <div className="space-y-6 lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>
                                Selected Videos ({selectedInputs.length})
                            </CardTitle>
                            <CardDescription>
                                The operation is applied to all selected videos
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                {selectedInputs.map((input) => (
                                    <Card
                                        key={input.id}
                                        className="overflow-hidden"
                                    >
                                        <CardContent className="p-0">
                                            {input.thumbnail && (
                                                <div className="relative aspect-video">
                                                    <Image
                                                        src={input.thumbnail}
                                                        alt={input.title}
                                                        fill
                                                        className="object-cover"
                                                        unoptimized
                                                    />
                                                </div>
                                            )}
                                            <div className="p-3">
                                                <p className="line-clamp-2 text-sm font-medium">
                                                    {input.title}
                                                </p>
                                                <p className="text-muted-foreground mt-1 text-xs capitalize">
                                                    {input.type}
                                                </p>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    {notEnough && (
                        <Alert>
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>
                                This tool requires at least {minSelection}{' '}
                                selected videos.
                            </AlertDescription>
                        </Alert>
                    )}

                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    <div className="flex gap-3">
                        <Button
                            className="flex-1"
                            size="lg"
                            onClick={onSubmit}
                            disabled={isSubmitting || notEnough}
                        >
                            {isSubmitting ? submittingLabel : submitLabel}
                        </Button>
                        <Button
                            variant="outline"
                            size="lg"
                            onClick={handleCancel}
                            disabled={isSubmitting}
                        >
                            Cancel
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
