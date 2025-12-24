#!/bin/bash

# Скрипт для зупинки Backend (Go/Docker) та Frontend (NPM)

# Кольори для виводу
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}====================================================${NC}"
echo -e "${YELLOW}=== Зупинка всіх компонентів додатку ===${NC}"
echo -e "${YELLOW}====================================================${NC}"

# 1. Зупинка Frontend (Vite)
echo -e "${YELLOW}=== 1. Зупинка Frontend (Vite) ===${NC}"
FRONTEND_PID=$(lsof -ti:5173 2>/dev/null || lsof -ti:3000 2>/dev/null || lsof -ti:5174 2>/dev/null)
if [ ! -z "$FRONTEND_PID" ]; then
    echo "Зупинка Frontend процесу (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null
    sleep 1
    # Перевірка, чи процес дійсно зупинився
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        echo -e "${RED}Процес не зупинився, використовую kill -9...${NC}"
        kill -9 $FRONTEND_PID 2>/dev/null
    fi
    echo -e "${GREEN}✓ Frontend зупинено${NC}"
else
    echo -e "${YELLOW}Frontend не запущено (процес не знайдено на портах 5173/3000/5174)${NC}"
fi

# 2. Зупинка Backend (Go)
echo -e "${YELLOW}=== 2. Зупинка Backend (Go) ===${NC}"
# Спробуємо знайти процес за портом 8080
BACKEND_PID=$(lsof -ti:8080 2>/dev/null)
if [ ! -z "$BACKEND_PID" ]; then
    echo "Зупинка Backend процесу (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null
    sleep 1
    # Перевірка, чи процес дійсно зупинився
    if kill -0 $BACKEND_PID 2>/dev/null; then
        echo -e "${RED}Процес не зупинився, використовую kill -9...${NC}"
        kill -9 $BACKEND_PID 2>/dev/null
    fi
    echo -e "${GREEN}✓ Backend зупинено${NC}"
else
    # Також спробуємо знайти за назвою процесу
    BACKEND_PID=$(pgrep -f "go run cmd/api/main.go" 2>/dev/null | head -1)
    if [ ! -z "$BACKEND_PID" ]; then
        echo "Зупинка Backend процесу (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null
        sleep 1
        if kill -0 $BACKEND_PID 2>/dev/null; then
            kill -9 $BACKEND_PID 2>/dev/null
        fi
        echo -e "${GREEN}✓ Backend зупинено${NC}"
    else
        echo -e "${YELLOW}Backend не запущено (процес не знайдено)${NC}"
    fi
fi

# 3. Зупинка Docker Compose
echo -e "${YELLOW}=== 3. Зупинка Docker Compose (PostgreSQL, Redis, MinIO) ===${NC}"
cd backend
if [ -f "docker-compose.yml" ]; then
    echo "Зупинка Docker контейнерів..."
    docker-compose down
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Docker контейнери зупинено${NC}"
    else
        echo -e "${RED}✗ Помилка при зупинці Docker контейнерів${NC}"
    fi
else
    echo -e "${YELLOW}docker-compose.yml не знайдено${NC}"
fi
cd .. # Повернення до кореня проєкту

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}=== Всі компоненти зупинено ===${NC}"
echo -e "${GREEN}====================================================${NC}"

