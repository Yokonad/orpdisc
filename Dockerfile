# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o monitor ./cmd/monitor

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 monitor

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/monitor .

# Create data directory
RUN mkdir -p /data && chown -R monitor:monitor /data

# Switch to non-root user
USER monitor

# Expose health check port
EXPOSE 8080

# Set environment
ENV DB_PATH=/data/data.db

# Run the binary
CMD ["./monitor"]
