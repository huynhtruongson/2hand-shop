# =============================================================================
# Stage 1: Builder
# Build context must be the workspace root (2hand-shop/) so that the shared
# internal/pkg module is accessible during the build.
#
# Usage:
#   docker build -f internal/services/commerce/Dockerfile -t commerce-service .
#
# Run the API server (default):
#   docker run commerce-service
#
# Run migrations:
#   docker run commerce-service migration
# =============================================================================
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# ---------------------------------------------------------------------------
# Copy shared package first (internal/pkg) — cached unless pkg changes
# ---------------------------------------------------------------------------
COPY ./internal/pkg/ ./internal/pkg/

# ---------------------------------------------------------------------------
# Copy commerce service source
# ---------------------------------------------------------------------------
COPY ./internal/services/commerce/ ./internal/services/commerce/

# ---------------------------------------------------------------------------
# Download dependencies for the commerce service
# ---------------------------------------------------------------------------
WORKDIR /workspace/internal/services/commerce

RUN go mod download

# ---------------------------------------------------------------------------
# Build both binaries
# CGO_ENABLED=0  → fully static binary (no libc dependency)
# -trimpath      → remove local paths from binary for reproducibility
# -ldflags       → strip debug info to reduce binary size
# ---------------------------------------------------------------------------
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /workspace/bin/app \
    ./cmd/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /workspace/bin/migration \
    ./cmd/migration

# =============================================================================
# Stage 2: Runtime image
# Uses a minimal distroless image — no shell, no package manager.
# Both binaries are included so the same image can run either command.
# =============================================================================
FROM alpine:latest

# Copy both compiled binaries
COPY --from=builder ./workspace/bin/app       ./app
COPY --from=builder ./workspace/bin/migration ./migration

# Expose the default service port
EXPOSE 8083

# Default command runs the API server.
# Override with "migration" to run database migrations:
#   docker run commerce-service /migration
ENTRYPOINT ["/app"]
