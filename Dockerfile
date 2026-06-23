# ─── Build Stage ─────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy and download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -extldflags '-static'" \
    -o /readmeai ./cmd/main.go

# ─── Runtime Stage ───────────────────────────────────────────────────
FROM alpine:3.19 AS runtime

WORKDIR /app

# CA certs for HTTPS calls (GitHub API + OpenAI)
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /readmeai /app/readmeai

# Copy static web assets
COPY web/ /app/web/

# Non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/readmeai"]
