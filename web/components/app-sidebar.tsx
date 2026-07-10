import {
    ChartArea,
    FolderOpen,
    GitBranch,
    Home,
    MonitorDown,
    Settings,
    Wrench,
} from 'lucide-react'

import { Link, useLocation } from 'react-router-dom'

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
        title: 'Collections',
        url: '/collections',
        icon: FolderOpen,
    },
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
    const { pathname } = useLocation()

    const isActive = (url: string) =>
        url === '/' ? pathname === '/' : pathname.startsWith(url)

    return (
        <Sidebar>
            <SidebarHeader>
                <SidebarGroupLabel>
                    <b className="text-foreground text-xl">Video Archiver</b>
                </SidebarGroupLabel>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {items.map((item) => (
                                <SidebarMenuItem key={item.title}>
                                    <SidebarMenuButton
                                        asChild
                                        isActive={isActive(item.url)}
                                    >
                                        <Link to={item.url}>
                                            <item.icon />
                                            <span>{item.title}</span>
                                        </Link>
                                    </SidebarMenuButton>
                                </SidebarMenuItem>
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarGroupLabel className={'gap-x-4'}>
                    <p className={'text-md'}>version {__APP_VERSION__}-BETA</p>
                    <a
                        target="_blank"
                        rel="noreferrer"
                        href="https://github.com/Fx64b/video-archiver"
                    >
                        <GitBranch size={20} />
                    </a>
                </SidebarGroupLabel>
            </SidebarFooter>
        </Sidebar>
    )
}
