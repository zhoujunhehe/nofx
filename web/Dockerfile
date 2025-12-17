# ═══════════════════════════════════════════════════════════════
# NOFX Frontend Dockerfile for Railway
# Multi-stage build: Node.js build → Caddy runtime
# ═══════════════════════════════════════════════════════════════

ARG NODE_VERSION=20-alpine

# ──────────────────────────────────────────────────────────────
# Build Stage (Node.js)
# ──────────────────────────────────────────────────────────────
FROM node:${NODE_VERSION} AS builder

ENV NPM_CONFIG_UPDATE_NOTIFIER=false
ENV NPM_CONFIG_FUND=false

WORKDIR /build

# Install dependencies first (better caching)
COPY package*.json ./
RUN npm ci

# Copy source and build
COPY . ./
RUN npm run build

# ──────────────────────────────────────────────────────────────
# Runtime Stage (Caddy)
# ──────────────────────────────────────────────────────────────
FROM caddy:alpine

WORKDIR /app

# Copy Caddyfile and format it
COPY Caddyfile ./
RUN caddy fmt Caddyfile --overwrite

# Copy built static files
COPY --from=builder /build/dist ./dist

# Railway uses PORT environment variable
ENV PORT=3000

EXPOSE ${PORT}

CMD ["caddy", "run", "--config", "Caddyfile", "--adapter", "caddyfile"]
