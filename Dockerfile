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

# Установка зависимостей для сборки
RUN apk add --no-cache git gcc musl-dev libpcap-dev

# Установка рабочей директории
WORKDIR /src

# Копируем go.mod/go.sum и скачиваем зависимости (кэширование слоя)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем бинарник.
# CGO_ENABLED=1 — нужен для gopacket/pcap (libpcap); SQLite-драйвер
# modernc.org/sqlite pure-Go и CGO не требует.
COPY . ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o /bin/network-scanner \
    ./cmd/network-scanner

# ---------------------------------------------------------------------------
# Этап 2: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.20

# ARG не наследуются из стадии builder — объявляем заново для меток
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Метаданные
LABEL org.opencontainers.image.title="Network Scanner CLI"
LABEL org.opencontainers.image.description="Кроссплатформенная утилита для сканирования локальных сетей"
LABEL org.opencontainers.image.source="https://github.com/RekadzeAV/network-scanner"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_TIME}"

# Минимальный набор для работы сети и TLS
RUN apk add --no-cache \
    iproute2 \
    iputils \
    bash \
    ca-certificates \
    libpcap \
    tzdata && \
    addgroup -g 1000 scanner && \
    adduser -u 1000 -G scanner -h /home/scanner -D scanner

WORKDIR /home/scanner

# Копируем бинарник из builder
COPY --from=builder /bin/network-scanner /usr/local/bin/network-scanner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Директории для данных и конфигурации
RUN mkdir -p /home/scanner/scans /home/scanner/config && \
    chown -R scanner:scanner /home/scanner

# Переменные окружения по умолчанию
ENV NETWORK_SCANNER_DATA_DIR=/home/scanner/scans
ENV NETWORK_SCANNER_CONFIG_DIR=/home/scanner/config
ENV NETWORK_SCANNER_LOG_LEVEL=info

USER scanner

# Точка входа — CLI сканер
ENTRYPOINT ["network-scanner"]
CMD ["--help"]
