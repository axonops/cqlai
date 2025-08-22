# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -ldflags="-w -s" -o cqlai cmd/cqlai/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 cqlai && \
    adduser -D -u 1000 -G cqlai cqlai

# Set working directory
WORKDIR /home/cqlai

# Copy binary from builder
COPY --from=builder /app/cqlai /usr/local/bin/cqlai

# Copy default config (optional)
COPY --from=builder /app/cqlai.json.example /home/cqlai/cqlai.json.example

# Set ownership
RUN chown -R cqlai:cqlai /home/cqlai

# Switch to non-root user
USER cqlai

# Set entrypoint
ENTRYPOINT ["cqlai"]

# Default command (can be overridden)
CMD ["--help"]