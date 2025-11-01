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

# Support PORT from Railway; default 8080
ENV PORT=8080

# Use a template so nginx listens on $PORT
COPY nginx/nginx.conf.template /etc/nginx/conf.d/default.conf.template

# Copy built assets
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["/bin/sh", "-c", "envsubst '$PORT' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'" ]
