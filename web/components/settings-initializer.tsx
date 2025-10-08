'use client'

import useSettingsState from '@/store/settingsState'
import { useEffect } from 'react'

export function SettingsInitializer() {
    const { settings, fetchSettings } = useSettingsState()

    useEffect(() => {
        fetchSettings()
    }, [fetchSettings])

    useEffect(() => {
        if (settings) {
            // Apply theme
            useSettingsState.getState().setTheme(settings.theme)
        }
    }, [settings])

    return null
}
