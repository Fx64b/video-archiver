import React from 'react'

import type { Metadata } from 'next'
import localFont from 'next/font/local'

import { AppSidebar } from '@/components/app-sidebar'
import { SettingsInitializer } from '@/components/settings-initializer'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import { ThemeProvider } from '@/components/ui/theme-provider'

import './globals.css'

const geistSans = localFont({
    src: './fonts/GeistVF.woff',
    variable: '--font-geist-sans',
    weight: '100 900',
})
const geistMono = localFont({
    src: './fonts/GeistMonoVF.woff',
    variable: '--font-geist-mono',
    weight: '100 900',
})

export const metadata: Metadata = {
    title: 'Video Archiver',
    description:
        'A self-hosted YouTube video archiver with a modern web interface. Download, manage, and organize your YouTube videos, playlists, and channels locally.',
}

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode
}>) {
    return (
        <html lang="en" suppressHydrationWarning>
            <body
                className={`${geistSans.variable} ${geistMono.variable} antialiased`}
            >
                <ThemeProvider
                    attribute="class"
                    defaultTheme="system"
                    enableSystem
                    disableTransitionOnChange
                >
                    <SettingsInitializer />
                    <SidebarProvider>
                        <AppSidebar />
                        <main className={'w-full'}>
                            <SidebarTrigger />
                            {children}
                        </main>
                    </SidebarProvider>
                    <Toaster />
                </ThemeProvider>
            </body>
        </html>
    )
}
