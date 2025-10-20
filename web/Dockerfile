# Base image with pnpm pre-installed
FROM node:25-alpine AS base
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# Install dependencies only when needed
FROM base AS deps
# Copy dependency manifests
COPY package.json pnpm-lock.yaml* ./
# Install all dependencies (including devDependencies for build)
RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM base AS builder
# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules
# Copy all source files
COPY . .
# Set environment to production
ENV NODE_ENV production
# Set build-time environment variables for Next.js
ARG NEXT_PUBLIC_SERVER_URL=http://backend:8080
ARG NEXT_PUBLIC_SERVER_URL_WS=ws://backend:8081
ENV NEXT_PUBLIC_SERVER_URL=$NEXT_PUBLIC_SERVER_URL
ENV NEXT_PUBLIC_SERVER_URL_WS=$NEXT_PUBLIC_SERVER_URL_WS
# Build the application
RUN pnpm build

# Production image, copy only necessary files
FROM base AS runner
ENV NODE_ENV production

# Copy package files
COPY package.json pnpm-lock.yaml* ./

# Install production dependencies + typescript (needed for next.config.ts)
RUN pnpm install --frozen-lockfile --prod && \
    pnpm add -D typescript

# Copy built application from builder
COPY --from=builder /app/next.config.ts ./
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next

# Expose the default Next.js port
EXPOSE 3000

# Start the Next.js application
CMD ["pnpm", "start"]
