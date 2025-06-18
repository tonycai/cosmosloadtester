# Multi-stage build for cosmosloadtester-cli
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the CLI application
RUN make cli

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    jq \
    bc

# Create a non-root user
RUN addgroup -g 1001 cosmosload && \
    adduser -D -u 1001 -G cosmosload cosmosload

# Set working directory
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/bin/cosmosloadtester-cli /usr/local/bin/cosmosloadtester-cli

# Copy examples and documentation
COPY --from=builder /app/examples/ /app/examples/
COPY --from=builder /app/CLI_README.md /app/README.md

# Copy and set up entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Create config directory and set permissions
RUN mkdir -p /home/cosmosload/.cosmosloadtester && \
    mkdir -p /app/results && \
    chown -R cosmosload:cosmosload /home/cosmosload/.cosmosloadtester && \
    chown -R cosmosload:cosmosload /app

# Switch to non-root user
USER cosmosload

# Set environment variables
ENV HOME=/home/cosmosload
ENV PATH="/usr/local/bin:${PATH}"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD cosmosloadtester-cli --version || exit 1

# Set entrypoint
ENTRYPOINT ["docker-entrypoint.sh"]

# Default command
CMD []

# Labels for metadata
LABEL maintainer="cosmosloadtester-team"
LABEL version="1.0.0"
LABEL description="Terminal-based Cosmos Load Testing Tool"
LABEL org.opencontainers.image.source="https://github.com/orijtech/cosmosloadtester"
LABEL org.opencontainers.image.title="cosmosloadtester-cli"
LABEL org.opencontainers.image.description="A powerful terminal-based load testing tool for Cosmos blockchain applications" 