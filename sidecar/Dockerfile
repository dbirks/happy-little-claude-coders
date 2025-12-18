FROM golang:1.24-bookworm AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o github-token-sidecar .

# Final stage - minimal image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /build/github-token-sidecar /usr/local/bin/github-token-sidecar

# Run as non-root user (distroless nonroot uses UID 65532)
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/github-token-sidecar"]
