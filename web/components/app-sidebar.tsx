import {Home, MonitorPlay, Settings, GitBranch, ListVideo, Users, Wrench, ChartArea} from "lucide-react"

import {
    Sidebar,
    SidebarContent, SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel, SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from "@/components/ui/sidebar"
import Link from "next/link";

// Menu items.
const items = [
    {
        title: "Overview",
        url: "/",
        icon: Home,
    },
    {
        title: "Dashboard",
        url: "/dashboard",
        icon: ChartArea,
    },
    {
        title: "Videos",
        url: "/videos",
        icon: MonitorPlay,
    },
    {
        title: "Playlists",
        url: "/playlists",
        icon: ListVideo,
    },
    {
        title: "Channels",
        url: "/channels",
        icon: Users,
    },
    {
        title: "Tools",
        url: "/tools",
        icon: Wrench,
    },
    {
        title: "Settings",
        url: "/settings",
        icon: Settings,
    },
]

export function AppSidebar() {
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

                    <p className={'text-md'}>v0.0.0</p> <Link target={'_blank'} href={'https://github.com/Fx64b/video-archiver'}><GitBranch size={20} /></Link>
                </SidebarGroupLabel>
            </SidebarFooter>
        </Sidebar>
    )
}
