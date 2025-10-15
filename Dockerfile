# Builder stage: Debian-based for CGO (SQLite/LZ4)
FROM golang:1.23-bookworm AS builder

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      build-essential sqlite3 libsqlite3-dev liblz4-dev tzdata ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o rca-backend .

# Final stage: slim Debian runtime
FROM debian:bookworm-slim

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates tzdata sqlite3 liblz4-1 wget && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -m -s /bin/bash appuser

WORKDIR /app

# Copy the binary and config from builder stage
COPY --from=builder /app/rca-backend .
COPY --from=builder /app/config.yaml .

# Create data directory
RUN mkdir -p /app/data && chown -R appuser:appuser /app
USER appuser

# Expose port (Render will set PORT)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application (PORT override handled in app)
CMD ["./rca-backend", "--config", "config.yaml"]