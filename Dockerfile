# Multi-stage build for WhatsApp Multi-Session Manager

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

# Set working directory
WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Go application
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

# Stage 3: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

# Create app user for security
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Create directories with proper permissions
RUN mkdir -p /app/data /app/logs /app/whatsapp/sessions /app/whatsapp/logs /app/config /app/database && \
    touch /app/database/sessions.db && \
    chown -R appuser:appgroup /app

# Copy entrypoint script
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# Copy built application from builder stage
COPY --from=builder /app/whatsapp-multi-session .
# Copy frontend build from frontend-builder stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

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

# Create volume mount points
VOLUME ["/app/data", "/app/sessions", "/app/logs"]

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]

# Run the application
CMD ["./whatsapp-multi-session"]