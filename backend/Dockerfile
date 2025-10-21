# Use official Go image as build environment
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./

# Copy source code first to generate go.sum
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Use lightweight alpine image as runtime environment
FROM alpine:latest



# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /home/appuser

# Copy binary file from build stage
COPY --from=builder /app/main .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Create storage directory with proper ownership
RUN mkdir -p storage/files && \
    chown -R appuser:appgroup storage && \
    chown -R appuser:appgroup .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run application
CMD ["./main"]