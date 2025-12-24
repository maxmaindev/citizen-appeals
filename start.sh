#!/bin/bash

# Скрипт для автоматичного запуску Backend (Go/Docker) та Frontend (NPM)

# Кольори для виводу
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}=== 1. Запуск Backend: Docker Compose (PostgreSQL) ===${NC}"
echo -e "${GREEN}====================================================${NC}"

# 1. Запуск Docker Compose у фоновому режимі
cd backend
echo "Запуск контейнерів PostgreSQL, Redis та MinIO..."
docker-compose up -d

# Перевірка, чи запущена БД
# Важливо: даємо БД час для запуску (хоча б 5 секунд)
echo "Очікування запуску бази даних (5 секунд)..."
sleep 5

# Перевірка Health Check (опціонально, але корисно)
if docker exec citizen_appeals_db pg_isready -U postgres; then
    echo -e "${GREEN}✓ PostgreSQL запущено та готово.${NC}"
else
    echo -e "${RED}✗ Помилка: PostgreSQL не запустився. Перевірте логи Docker.${NC}"
    exit 1
fi

echo -e "${GREEN}==============================${NC}"
echo -e "${GREEN}=== 2. Запуск Backend (Go) ===${NC}"
echo -e "${GREEN}==============================${NC}"

# 2. Запуск Go-сервера у фоновому режимі
# Використовуємо & для запуску у фоні та `nohup` для ігнорування SIGHUP
# Логи Go-сервера будуть записані у go_backend.log
nohup go run cmd/api/main.go > go_backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend запущено на PID: $BACKEND_PID. Логи: backend/go_backend.log"
cd .. # Повернення до кореня проєкту

# 3. Перевірка Health Check бекенду
echo "Очікування запуску Go-сервера (3 секунди)..."
sleep 3
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo -e "${GREEN}✓ Backend Go запущено на http://localhost:8080${NC}"
else
    echo -e "${RED}✗ Помилка: Backend не відповідає. Перевірте backend/go_backend.log${NC}"
    exit 1
fi

echo -e "${GREEN}==================================${NC}"
echo -e "${GREEN}=== 3. Запуск Frontend (Vite) ===${NC}"
echo -e "${GREEN}==================================${NC}"

# 4. Запуск Frontend
cd frontend
echo "Запуск Frontend (npm run dev)..."
npm run dev

# Примітка: Ця команда (npm run dev) залишатиметься активною
# і блокуватиме термінал, поки ви не зупините її (Ctrl+C).