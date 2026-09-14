# =============================================================================
# Multi-stage Dockerfile для Network Scanner CLI
# =============================================================================
# Сборка CLI-версии в минимальном образе Alpine Linux
# =============================================================================

# ---------------------------------------------------------------------------
# Этап 1: Сборка
# ---------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git gcc musl-dev

# Установка рабочей директории
WORKDIR /app

# Копирование go.mod и go.sum для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка CLI с версией и временем сборки
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /usr/local/bin/network-scanner \
    ./cmd/network-scanner

# ---------------------------------------------------------------------------
# Этап 2: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.20

# Метаданные
LABEL org.opencontainers.image.title="Network Scanner CLI"
LABEL org.opencontainers.image.description="Network scanner CLI tool for discovering and analyzing network devices"
LABEL org.opencontainers.image.source="https://github.com/Ai-Gi/network-scanner"
LABEL org.opencontainers.image.version="2.3.0"

# Установка зависимостей для запуска
RUN apk add --no-cache \
    iproute2 \
    iputils \
    bash \
    ca-certificates

# Создание пользователя для безопасности
RUN addgroup -S scanner && adduser -S scanner -G scanner

# Копирование бинарника из builder
COPY --from=builder /usr/local/bin/network-scanner /usr/local/bin/network-scanner

# Создание директорий для данных
RUN mkdir -p /data/scans /data/config && \
    chown -R scanner:scanner /data

# Переменные окружения
ENV NETWORK_SCANNER_DATA_DIR=/data/scans
ENV NETWORK_SCANNER_CONFIG_DIR=/data/config
ENV NETWORK_SCANNER_LOG_LEVEL=info

# Переключение на непривилегированного пользователя
USER scanner

# Точка входа
ENTRYPOINT ["network-scanner"]

# Команда по умолчанию
CMD ["--help"]
