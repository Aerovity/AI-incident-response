# Multi-stage build for AI Incident Response System
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o incident-ai .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/incident-ai .

# Copy scripts directory if it exists
COPY --from=builder /build/scripts ./scripts 2>/dev/null || true

# Expose service port
EXPOSE 8080

# Set environment variables
ENV OPENAI_API_KEY=""
ENV SERVICE_PORT=8080

# Run the application
CMD ["./incident-ai"]
