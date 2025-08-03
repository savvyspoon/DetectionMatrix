# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o riskmatrix ./cmd/server

# Final stage
FROM alpine:latest

# Install required packages
RUN apk --no-cache add ca-certificates sqlite

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/riskmatrix .

# Copy web assets and configs
COPY --from=builder /app/web ./web
COPY --from=builder /app/configs ./configs

# Create data directory
RUN mkdir -p data

# Expose port
EXPOSE 8080

# Set environment variables
ENV DB_PATH=/app/data/riskmatrix.db
ENV CONFIG_PATH=/app/configs/config.json

# Run the application
CMD ["./riskmatrix", "-db", "${DB_PATH}", "-config", "${CONFIG_PATH}"]