# Build stage: compile the Vite app to static files.
# Pin pnpm to the 10.x line that matches the lockfile (lockfileVersion 9.0).
FROM node:24-alpine AS builder
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate

WORKDIR /app

# Install dependencies first so they are cached independently of source
# changes. The cache mount persists the pnpm content-addressable store across
# image builds, so even lockfile changes only download new packages.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile

COPY . .

# No URLs are baked in: the bundle uses same-origin /api paths and nginx
# proxies them to the backend, so one image works on any host. Set
# VITE_SERVER_URL / VITE_SERVER_URL_WS here only to target a backend on a
# different origin.
RUN pnpm build

# Runtime stage: serve the static bundle with nginx
FROM nginx:1.27-alpine AS runner
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
