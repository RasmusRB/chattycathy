# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate cmd/migrate/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and curl for healthcheck
RUN apk --no-cache add ca-certificates curl

# Copy binaries from builder
COPY --from=builder /server .
COPY --from=builder /migrate .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command
CMD ["./server"]
