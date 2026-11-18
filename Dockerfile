# Stage 1: Build static pure-Go binary (CGO_ENABLED=0)
FROM golang:1.27-alpine AS builder

WORKDIR /src

# Download dependencies first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source and embedded OpenAPI docs
COPY main.go ./
COPY docs ./docs
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

# Stage 2: Minimal runtime image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    mkdir -p /data && \
    chown -R appuser:appgroup /data

WORKDIR /app

COPY --from=builder /out/server /app/server

ENV PORT=8080 \
    GIN_MODE=release \
    DB_PATH=/data/todo.db

USER appuser

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/app/server"]
