# =============================================================================
# Multi-stage Dockerfile для Network Scanner CLI
# =============================================================================
# Сборка CLI-версии в минимальном образе Alpine Linux
# =============================================================================

# ---------------------------------------------------------------------------
# Этап 1: Сборка
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

# Аргументы сборки (передаются через --build-arg)
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Установка зависимостей для сборки (git нужен для go modules, ca-certificates для TLS)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Копируем go.mod/go.sum и скачиваем зависимости (кэширование слоя)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем статический бинарник
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o /bin/network-scanner \
    ./cmd/network-scanner

# ---------------------------------------------------------------------------
# Этап 2: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.20

# metadata
LABEL org.opencontainers.image.title="Network Scanner CLI" \
      org.opencontainers.image.description="Кроссплатформенная утилита для сканирования локальных сетей" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}"

# Минимальный набор для работы сети и TLS
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 scanner && \
    adduser -u 1000 -G scanner -h /home/scanner -D scanner

WORKDIR /home/scanner

# Копируем бинарник из builder
COPY --from=builder /bin/network-scanner /usr/local/bin/network-scanner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER scanner

# Точка входа — CLI сканер
ENTRYPOINT ["network-scanner"]
CMD ["--help"]
