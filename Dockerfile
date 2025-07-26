# Multi-stage build for WhatsApp Multi-Session Manager
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o whatsapp-multi-session .

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

# Create app user for security
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Create directories with proper permissions
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

# Copy built application from builder stage
COPY --from=builder /app/whatsapp-multi-session .
COPY --from=builder /app/frontend/dist ./frontend/dist

# Set proper permissions
RUN chown appuser:appgroup whatsapp-multi-session && \
    chmod +x whatsapp-multi-session

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Run the application
CMD ["./whatsapp-multi-session"]