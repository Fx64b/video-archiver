/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { defineConfig } from 'vite'

const pkg = JSON.parse(readFileSync(resolve(__dirname, 'package.json'), 'utf8'))

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': resolve(__dirname, '.'),
        },
    },
    define: {
        __APP_VERSION__: JSON.stringify(pkg.version),
    },
    server: {
        port: 3000,
        // Same-origin backend in development: the app calls /api/<route> and
        // the dev server forwards it (prefix stripped) to the backend —
        // including the /api/ws WebSocket upgrade. No baked URLs, no CORS.
        // VITE_PROXY_TARGET points at the backend container in `run.sh dev`.
        proxy: {
            '/api': {
                target: process.env.VITE_PROXY_TARGET || 'http://localhost:8080',
                changeOrigin: true,
                ws: true,
                rewrite: (path) => path.replace(/^\/api/, ''),
            },
        },
    },
    preview: {
        port: 3000,
    },
    build: {
        rollupOptions: {
            output: {
                // Keep the React runtime in its own long-cached chunk;
                // recharts already splits out via the lazy Dashboard import.
                manualChunks: {
                    'react-vendor': ['react', 'react-dom', 'react-router-dom'],
                },
            },
        },
    },
    test: {
        environment: 'jsdom',
        globals: true,
        setupFiles: ['./vitest.setup.ts'],
        include: ['**/__tests__/**/*.test.{ts,tsx}'],
        exclude: ['node_modules', 'dist'],
        coverage: {
            provider: 'v8',
            include: [
                'lib/**/*.{ts,tsx}',
                'services/**/*.{ts,tsx}',
                'store/**/*.{ts,tsx}',
                'components/**/*.{ts,tsx}',
            ],
        },
    },
})
