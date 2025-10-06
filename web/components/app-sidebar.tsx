'use client'

import {
    ChartArea,
    GitBranch,
    Home,
    MonitorDown,
    MonitorPlay,
    Settings,
    Wrench,
} from 'lucide-react'

import { useEffect, useState } from 'react'

import Link from 'next/link'

import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/components/ui/sidebar'

const items = [
    {
        title: 'Overview',
        url: '/',
        icon: Home,
    },
    {
        title: 'Dashboard',
        url: '/dashboard',
        icon: ChartArea,
    },
    {
        title: 'Downloads',
        url: '/downloads',
        icon: MonitorDown,
    },
    {
        title: 'Stream',
        url: '/stream',
        icon: MonitorPlay,
    },
    /*    {
        title: 'Channels',
        url: '/channels',
        icon: Users,
    },*/
    {
        title: 'Tools',
        url: '/tools',
        icon: Wrench,
    },
    {
        title: 'Settings',
        url: '/settings',
        icon: Settings,
    },
]

export function AppSidebar() {
    const [version, setVersion] = useState<string | null>(null)

    useEffect(() => {
        const fetchVersion = async () => {
            try {
                const response = await fetch('/api/info')
                const data = await response.json()
                setVersion(data.info.version)
            } catch (error) {
                console.error('Failed to fetch version:', error)
            }
        }

        fetchVersion()
    }, [])

    return (
        <Sidebar>
            <SidebarHeader>
                <SidebarGroupLabel>
                    <b className={'text-xl text-white'}>Video Archiver</b>
                </SidebarGroupLabel>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {items.map((item) => (
                                <SidebarMenuItem key={item.title}>
                                    <SidebarMenuButton asChild>
                                        <a href={item.url}>
                                            <item.icon />
                                            <span>{item.title}</span>
                                        </a>
                                    </SidebarMenuButton>
                                </SidebarMenuItem>
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarGroupLabel className={'gap-x-4'}>
                    <p className={'text-md'}>version {version}-BETA</p>{' '}
                    <Link
                        target={'_blank'}
                        href={'https://github.com/Fx64b/video-archiver'}
                    >
                        <GitBranch size={20} />
                    </Link>
                </SidebarGroupLabel>
            </SidebarFooter>
        </Sidebar>
    )
}
