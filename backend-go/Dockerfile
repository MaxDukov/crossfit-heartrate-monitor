# syntax=docker/dockerfile:1
# ── Этап 1: сборка ────────────────────────────────────────────
# CGO обязателен: openant-go использует google/gousb (libusb).
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev pkgconfig libusb-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /out/cf-server ./cmd/server

# ── Этап 2: рантайм ───────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata libusb && \
    adduser -D -u 1000 cf

COPY --from=builder /out/cf-server /usr/local/bin/cf-server

USER cf
ENV CF_DB_PATH=/data/cf_monitor.db \
    CF_PORT=8000

VOLUME /data
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -qO- http://127.0.0.1:8000/api/health || exit 1

CMD ["cf-server"]
