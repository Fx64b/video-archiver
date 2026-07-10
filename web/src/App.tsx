import { Suspense, lazy } from 'react'
import { Route, Routes } from 'react-router-dom'

import { AppSidebar } from '@/components/app-sidebar'
import { SettingsInitializer } from '@/components/settings-initializer'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'

import ChannelDetail from './pages/ChannelDetail'
import CollectionDetail from './pages/CollectionDetail'
import Collections from './pages/Collections'
import Downloads from './pages/Downloads'
import Overview from './pages/Overview'
import PlaylistDetail from './pages/PlaylistDetail'
import Settings from './pages/Settings'
import VideoDetail from './pages/VideoDetail'
import Concat from './pages/tools/Concat'
import Convert from './pages/tools/Convert'
import ExtractAudio from './pages/tools/ExtractAudio'
import Quality from './pages/tools/Quality'
import Results from './pages/tools/Results'
import Rotate from './pages/tools/Rotate'
import ToolsHome from './pages/tools/ToolsHome'
import Trim from './pages/tools/Trim'
import Workflow from './pages/tools/Workflow'

// The dashboard pulls in recharts, by far the heaviest dependency, so it is
// split out of the main bundle and only loaded when visited.
const Dashboard = lazy(() => import('./pages/Dashboard'))

export default function App() {
    return (
        <>
            <SettingsInitializer />
            <SidebarProvider>
                <AppSidebar />
                <main className="w-full">
                    <SidebarTrigger />
                    <Routes>
                        <Route path="/" element={<Overview />} />
                        <Route
                            path="/dashboard"
                            element={
                                <Suspense fallback={null}>
                                    <Dashboard />
                                </Suspense>
                            }
                        />
                        <Route path="/downloads" element={<Downloads />} />
                        <Route
                            path="/downloads/video/:id"
                            element={<VideoDetail />}
                        />
                        <Route
                            path="/downloads/playlist/:id"
                            element={<PlaylistDetail />}
                        />
                        <Route
                            path="/downloads/channel/:id"
                            element={<ChannelDetail />}
                        />
                        <Route path="/collections" element={<Collections />} />
                        <Route
                            path="/collections/:id"
                            element={<CollectionDetail />}
                        />
                        <Route path="/settings" element={<Settings />} />
                        <Route path="/tools" element={<ToolsHome />} />
                        <Route path="/tools/results" element={<Results />} />
                        <Route path="/tools/trim" element={<Trim />} />
                        <Route path="/tools/concat" element={<Concat />} />
                        <Route path="/tools/convert" element={<Convert />} />
                        <Route
                            path="/tools/extract-audio"
                            element={<ExtractAudio />}
                        />
                        <Route path="/tools/quality" element={<Quality />} />
                        <Route path="/tools/rotate" element={<Rotate />} />
                        <Route path="/tools/workflow" element={<Workflow />} />
                    </Routes>
                </main>
            </SidebarProvider>
            <Toaster />
        </>
    )
}
