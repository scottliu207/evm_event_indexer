# ============ Build Stage ============
FROM golang:1.25.4-alpine AS builder

# Install git (needed for some go modules) and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" strips debug info for smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/indexer ./cmd/indexer

# ============ Runtime Stage ============
FROM alpine:3.19

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/indexer .

# Copy config file
COPY --from=builder /app/config/config.yaml ./config/

# Expose API port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/indexer"]