/**
 * Backend endpoints, baked in at build time via Vite env vars.
 * Defaults match the local docker-compose setup.
 */
export const SERVER_URL =
    import.meta.env.VITE_SERVER_URL ?? 'http://localhost:8080'

export const SERVER_URL_WS =
    import.meta.env.VITE_SERVER_URL_WS ?? 'ws://localhost:8081'
