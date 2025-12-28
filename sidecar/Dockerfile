FROM golang:1.24-bookworm AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY *.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o github-token-sidecar .

# Final stage - alpine for shell support in init containers
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/github-token-sidecar /usr/local/bin/github-token-sidecar

# Create non-root user
RUN adduser -D -u 1001 appuser
USER appuser

ENTRYPOINT ["/usr/local/bin/github-token-sidecar"]
