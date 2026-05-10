# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev linux-headers argon2-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 make build-linux

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates argon2-dev

WORKDIR /root

# Copy binary from builder
COPY --from=builder /app/build/homechaind-linux-amd64 /usr/local/bin/homechaind

# Expose ports
EXPOSE 26656 26657 1317 8545 8546

# Set environment
ENV HOME=/root/.homechain

# Create data directory
RUN mkdir -p /root/.homechain/data

# Run the binary
CMD ["homechaind", "start"]
