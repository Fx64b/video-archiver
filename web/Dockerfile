# Install dependencies only when needed
FROM node:24-alpine AS deps
WORKDIR /app

# Install pnpm
RUN npm install -g pnpm

# Only copy package.json and pnpm-lock.yaml to prevent cache busting
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM node:24-alpine AS builder
WORKDIR /app

# Install pnpm in the builder stage
RUN npm install -g pnpm

# Copy all project files
COPY . .

# Copy installed dependencies
COPY --from=deps /app/node_modules ./node_modules

# Set environment to production
ENV NODE_ENV production

# Build the application
RUN pnpm build

# Final image to serve the app
FROM node:24-alpine AS runner
WORKDIR /app

# Install pnpm
RUN npm install -g pnpm

# Copy package.json and lock file
COPY package.json pnpm-lock.yaml* ./

# Install production dependencies only
RUN pnpm install --frozen-lockfile --prod

# Copy necessary files from the builder
COPY --from=builder /app/next.config.ts ./
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next

# Expose the default Next.js port
EXPOSE 3000

# Start the Next.js application
CMD ["pnpm", "start"]
