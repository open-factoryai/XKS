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

# Build binary
RUN make build-local

# Runtime stage - use Microsoft's official Azure CLI image
FROM mcr.microsoft.com/azure-cli:2.57.0

# Create non-root user
RUN adduser -D -s /bin/bash xks

# Copy binary from builder
COPY --from=builder /app/dist/xks /usr/local/bin/xks

# Set permissions
RUN chmod +x /usr/local/bin/xks

# Switch to non-root user
USER xks

# Set working directory
WORKDIR /home/xks

# Default command
ENTRYPOINT ["xks"]
CMD ["--help"]