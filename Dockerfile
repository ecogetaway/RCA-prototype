# Use Go 1.23 as base image
FROM golang:1.23-alpine AS builder

# Install necessary build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rca-backend .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary and config from builder stage
COPY --from=builder /app/rca-backend .
COPY --from=builder /app/config.yaml .

# Create data directory
RUN mkdir -p /app/data

# Change ownership to non-root user
RUN chown -R appuser:appuser /app
USER appuser

# Expose port (Render will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the application with config file
CMD ["./rca-backend", "--config", "config.yaml"]