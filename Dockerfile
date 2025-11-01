# Root Dockerfile for Railway (Frontend Only)

# Multi-stage: build React (Vite) in Node, then serve with Nginx

ARG NODE_VERSION=20-alpine
ARG NGINX_VERSION=alpine

# Build stage
FROM node:${NODE_VERSION} AS builder
WORKDIR /build

# Install dependencies
COPY web/package*.json ./
RUN npm ci

# Build
COPY web/ ./
RUN npm run build

# Runtime stage
FROM nginx:${NGINX_VERSION}

# SPA nginx config with /health and static caching
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]

