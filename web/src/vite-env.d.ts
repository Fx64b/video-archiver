/// <reference types="vite/client" />

/** Injected at build time from package.json (see vite.config.ts). */
declare const __APP_VERSION__: string

interface ImportMetaEnv {
    readonly VITE_SERVER_URL?: string
    readonly VITE_SERVER_URL_WS?: string
}
