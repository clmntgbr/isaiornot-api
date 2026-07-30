FROM golang:1.25-alpine AS base

WORKDIR /app

RUN apk add --no-cache git ffmpeg

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM base AS development

RUN go install github.com/air-verse/air@v1.67.1
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
ENV PATH="/go/bin:${PATH}"

EXPOSE 3000
CMD ["air", "-c", ".air.toml"]

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

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o scheduler \
    ./cmd/scheduler

FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata ffmpeg

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

COPY --from=builder --chown=appuser:appuser /app/api .
COPY --from=builder --chown=appuser:appuser /app/dispatcher .
COPY --from=builder --chown=appuser:appuser /app/metadata .
COPY --from=builder --chown=appuser:appuser /app/heuristic .
COPY --from=builder --chown=appuser:appuser /app/aimodel .
COPY --from=builder --chown=appuser:appuser /app/scheduler .

USER appuser

EXPOSE 3000
CMD ["./api"]

FROM alpine:latest AS scheduler

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

COPY --from=builder --chown=appuser:appuser /app/scheduler .

USER appuser

CMD ["./scheduler"]
