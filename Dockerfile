# Stage 1: Build frontend
FROM node:22-slim AS frontend-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /build

# Copy package files first for better caching
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy frontend source
COPY frontend/ ./

# Build arguments declared after dependency install for better layer caching
ARG NUXT_PUBLIC_API_URL
ARG NUXT_PUBLIC_TOKEN_URL
ARG NUXT_PUBLIC_AUTH0_DOMAIN=login.bcc.no
ARG NUXT_PUBLIC_AUTH0_CLIENT_ID
ARG NUXT_PUBLIC_AUTH0_AUDIENCE
ARG NUXT_PUBLIC_RUDDERSTACK_WRITE_KEY
ARG NUXT_PUBLIC_RUDDERSTACK_DATA_PLANE_URL
ARG NUXT_PUBLIC_VAPID_PUBLIC_KEY
ARG NUXT_PUBLIC_IS_STAGING
ARG NUXT_PUBLIC_FIREBASE_DATABASE
ARG NUXT_PUBLIC_FIREBASE_API_KEY
ARG NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN
ARG NUXT_PUBLIC_FIREBASE_PROJECT_ID
ARG APP_VERSION

ENV APP_VERSION=${APP_VERSION}
RUN pnpm run build

# Stage 2: Build Go backend
FROM golang:1.25 AS backend-builder

WORKDIR /build

# Copy dependency files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o server \
    ./cmd/server

# Stage 3: Runtime
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy backend binary
COPY --from=backend-builder /build/server /server

# Copy frontend static files
COPY --from=frontend-builder /build/.output/public /static

# Run as non-root user
USER 65534

# Expose API port
EXPOSE 8080

# Set default static files path
ENV STATIC_FILES_PATH=/static

# Run the server
ENTRYPOINT ["/server"]
