# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Install build dependencies for CGO and cross-compilation
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set up arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the main application with proper cross-compilation support
# CGO is required for SQLite support
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -a -ldflags="-linkmode external -extldflags '-static'" \
    -o riskmatrix ./cmd/server

# Build the MITRE import tool
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -a -ldflags="-linkmode external -extldflags '-static'" \
    -o mitre-importer ./cmd/import-mitre

# Final stage - use platform-specific base image
FROM --platform=$TARGETPLATFORM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata wget && \
    # Create non-root user for security
    addgroup -g 1001 -S riskmatrix && \
    adduser -u 1001 -S riskmatrix -G riskmatrix

# Set working directory
WORKDIR /app

# Copy binaries from the builder stage
COPY --from=builder /app/riskmatrix .
COPY --from=builder /app/mitre-importer .

# Copy entrypoint script
COPY --from=builder /app/docker-entrypoint.sh .

# Copy web assets and configs
COPY --from=builder /app/web ./web
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/data/mitre.csv ./data/

# Create necessary directories and set permissions
RUN mkdir -p data logs && \
    chown -R riskmatrix:riskmatrix /app && \
    chmod +x ./riskmatrix ./mitre-importer ./docker-entrypoint.sh

# Switch to non-root user
USER riskmatrix

# Expose port
EXPOSE 8080

# Set environment variables
ENV DB_PATH=/app/data/riskmatrix.db
ENV CONFIG_PATH=/app/configs/config.json
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set entrypoint and default command
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["./riskmatrix"]