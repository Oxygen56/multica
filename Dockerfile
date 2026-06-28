# --- Build stage ---
FROM golang:1.26-alpine AS builder

ARG GOPROXY
ARG GOSUMDB
ENV GOPROXY=${GOPROXY:-https://proxy.golang.org,direct}
ENV GOSUMDB=${GOSUMDB:-sum.golang.org}

RUN apk add --no-cache git

WORKDIR /src

# Cache dependencies (Docker layer cache + BuildKit module cache)
COPY server/go.mod server/go.sum ./server/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd server && go mod download

# Copy server source
COPY server/ ./server/

# Build binaries sequentially. --mount=type=cache for /root/.cache/go-build
# persists compiled packages across builds and across commits — even when the
# COPY layer above is invalidated, Go reuses cached .a files for unchanged
# packages.  This is the primary optimization: on incremental changes compilation
# drops from a full rebuild to a few seconds.
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o bin/server ./cmd/server

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o bin/multica ./cmd/multica

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server && CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/migrate ./cmd/migrate

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server && CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/backfill_task_usage_hourly ./cmd/backfill_task_usage_hourly

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server && CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/backfill_codex_usage_cache ./cmd/backfill_codex_usage_cache

# --- Runtime stage ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /src/server/bin/server .
COPY --from=builder /src/server/bin/multica .
COPY --from=builder /src/server/bin/migrate .
COPY --from=builder /src/server/bin/backfill_task_usage_hourly .
COPY --from=builder /src/server/bin/backfill_codex_usage_cache .
COPY server/migrations/ ./migrations/
COPY docker/entrypoint.sh .
RUN sed -i 's/\r$//' entrypoint.sh && chmod +x entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:8080/health 2>/dev/null || exit 1

ENTRYPOINT ["./entrypoint.sh"]
