# Use official Go image as build environment
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Use lightweight alpine image as runtime environment
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy binary file from build stage
COPY --from=builder /app/main .

# Create storage directory
RUN mkdir -p storage/files && \
    chown -R appuser:appgroup storage

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run application
CMD ["./main"]