/**
 * Backend endpoints.
 *
 * By default the frontend talks to its own origin under /api — the Vite dev
 * server (development) and nginx (production image) both proxy /api to the
 * backend, so the same build works on any host with no baked-in URLs and no
 * CORS. Setting VITE_SERVER_URL / VITE_SERVER_URL_WS at build time overrides
 * this for setups where the backend is reached directly.
 */
export const SERVER_URL: string = import.meta.env.VITE_SERVER_URL || '/api'

/**
 * Absolute WebSocket base URL. The WebSocket constructor needs a ws:// or
 * wss:// URL, so the same-origin default is derived from window.location at
 * call time rather than baked in.
 */
export function serverWsUrl(): string {
    const explicit = import.meta.env.VITE_SERVER_URL_WS
    if (explicit) return explicit
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${scheme}//${window.location.host}/api`
}
