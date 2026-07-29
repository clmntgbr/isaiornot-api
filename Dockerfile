# ============================================
# Base stage - Common dependencies
# ============================================
FROM golang:1.25-alpine AS base

WORKDIR /app

# Install git and ffmpeg for Go modules / video frame extraction
RUN apk add --no-cache git ffmpeg

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .


# ============================================
# Development stage - With Air hot reload
# ============================================
FROM base AS development

RUN go install github.com/air-verse/air@v1.67.1

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
ENV PATH="/go/bin:${PATH}"

# Expose port
EXPOSE 3000

# Run with Air
CMD ["air", "-c", ".air.toml"]


# ============================================
# Builder stage - Build all binaries
# ============================================
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o dispatcher \
    ./cmd/dispatcher

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o metadata \
    ./cmd/metadata

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o heuristic \
    ./cmd/heuristic

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o aimodel \
    ./cmd/aimodel

# ============================================
# Production stage - Minimal runtime
# ============================================
FROM alpine:latest AS production

# Install ca-certificates, ffmpeg for HTTPS and video frame extraction
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copy binaries from builder
COPY --from=builder --chown=appuser:appuser /app/api .
COPY --from=builder --chown=appuser:appuser /app/dispatcher .
COPY --from=builder --chown=appuser:appuser /app/metadata .
COPY --from=builder --chown=appuser:appuser /app/heuristic .
COPY --from=builder --chown=appuser:appuser /app/aimodel .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Run the server application by default
CMD ["./api"]
