# Build stage: compile the Vite app to static files.
# Pin pnpm to the 10.x line that matches the lockfile (lockfileVersion 9.0).
FROM node:24-alpine AS builder
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate
WORKDIR /app

# Install dependencies first to keep them cached between source changes
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .

# Backend endpoints are baked into the static bundle at build time
ARG VITE_SERVER_URL=http://localhost:8080
ARG VITE_SERVER_URL_WS=ws://localhost:8081
ENV VITE_SERVER_URL=$VITE_SERVER_URL
ENV VITE_SERVER_URL_WS=$VITE_SERVER_URL_WS

RUN pnpm build

# Runtime stage: serve the static bundle with nginx
FROM nginx:alpine AS runner
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
