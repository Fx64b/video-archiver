# Video Archiver — Web Frontend

A Vite + React + TypeScript single-page app using shadcn/ui, Tailwind CSS 4, Zustand and React Router.

## Getting Started

```bash
pnpm install
pnpm dev
```

The dev server runs on [http://localhost:3000](http://localhost:3000) and expects the backend on
`http://localhost:8080` (REST) and `ws://localhost:8081` (WebSocket). Override with:

- `VITE_SERVER_URL` — backend API base URL
- `VITE_SERVER_URL_WS` — WebSocket base URL

Both are baked into the bundle at build time (see `lib/env.ts`).

## Scripts

| Script               | Description                        |
| -------------------- | ---------------------------------- |
| `pnpm dev`           | Start the Vite dev server          |
| `pnpm build`         | Typecheck and build for production |
| `pnpm preview`       | Serve the production build locally |
| `pnpm test`          | Run the Vitest test suite          |
| `pnpm test:coverage` | Run tests with coverage            |
| `pnpm lint`          | Run ESLint                         |
| `pnpm format`        | Format with Prettier               |

## Structure

- `index.html`, `src/main.tsx` — app entry
- `src/App.tsx` — layout shell and routes
- `src/pages/` — one component per route
- `components/` — shared components (`components/ui/` is shadcn/ui)
- `store/` — Zustand stores
- `services/` — API/WebSocket clients
- `types/index.ts` — generated from Go structs via tygo (do not edit by hand)

## Docker

The production image builds the static bundle and serves it with nginx on port 3000
(see `Dockerfile` and `nginx.conf`). The SPA fallback rewrites unknown paths to `index.html`.
