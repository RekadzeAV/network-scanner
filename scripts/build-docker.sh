#!/bin/bash
# =============================================================================
# Скрипт для сборки Docker-образа Network Scanner CLI
# =============================================================================
# Использование:
#   ./scripts/build-docker.sh              # Сборка с дефолтными параметрами
#   ./scripts/build-docker.sh 2.3.0        # Сборка с указанной версией
#   ./scripts/build-docker.sh latest sha  # Сборка с версией и commit hash
# =============================================================================

set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Параметры
VERSION="${1:-dev}"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT="${2:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

IMAGE_NAME="network-scanner"
IMAGE_TAG="${VERSION}"

echo -e "${GREEN}=================================================================${NC}"
echo -e "${GREEN}  Network Scanner CLI - Docker Build${NC}"
echo -e "${GREEN}=================================================================${NC}"
echo ""
echo -e "  Версия:      ${YELLOW}${VERSION}${NC}"
echo -e "  Время сборки: ${YELLOW}${BUILD_TIME}${NC}"
echo -e "  Commit:      ${YELLOW}${GIT_COMMIT}${NC}"
echo -e "  Образ:       ${YELLOW}${IMAGE_NAME}:${IMAGE_TAG}${NC}"
echo ""

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не установлен. Установите Docker и попробуйте снова.${NC}"
    exit 1
fi

# Проверка запущен ли Docker
if ! docker info &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не запущен. Запустите Docker и попробуйте снова.${NC}"
    exit 1
fi

echo -e "${GREEN}>>> Сборка Docker-образа...${NC}"
docker build \
    --build-arg VERSION="${VERSION}" \
    --build-arg BUILD_TIME="${BUILD_TIME}" \
    --build-arg GIT_COMMIT="${GIT_COMMIT}" \
    -t "${IMAGE_NAME}:${IMAGE_TAG}" \
    -f Dockerfile \
    .

echo ""
echo -e "${GREEN}>>> Проверка образа...${NC}"
docker image inspect "${IMAGE_NAME}:${IMAGE_TAG}" &> /dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Образ успешно собран!${NC}"
    
    echo ""
    echo -e "${GREEN}>>> Информация об образе:${NC}"
    docker images "${IMAGE_NAME}" --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}\t{{.CreatedAt}}"
else
    echo -e "${RED}❌ Ошибка проверки образа${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=================================================================${NC}"
echo -e "${GREEN}  Готово!${NC}"
echo -e "${GREEN}=================================================================${NC}"
echo ""
echo -e "  Запуск CLI:"
echo -e "    docker run --rm ${IMAGE_NAME}:${IMAGE_TAG} --help"
echo ""
echo -e "  Запуск сканирования:"
echo -e "    docker run --rm --network host --cap-add NET_ADMIN --cap-add NET_RAW ${IMAGE_NAME}:${IMAGE_TAG} scan --cidr 192.168.1.0/24"
echo ""
