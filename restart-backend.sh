#!/bin/bash

# Скрипт для перезапуску Backend (Go)

# Кольори для виводу
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=====================================${NC}"
echo -e "${YELLOW}=== Перезапуск Backend (Go) ===${NC}"
echo -e "${YELLOW}=====================================${NC}"

cd backend || { echo -e "${RED}✗ Помилка: папка 'backend' не знайдена.${NC}"; exit 1; }

# 1. Зупинка поточного процесу Backend
echo "Пошук запущеного процесу Backend..."

# Знаходимо процес по порту 8080
BACKEND_PID=$(lsof -t -i:8080 -sTCP:LISTEN 2>/dev/null)

if [ -z "$BACKEND_PID" ]; then
    # Якщо не знайдено по порту, шукаємо по команді
    BACKEND_PID=$(pgrep -f "go run cmd/api/main.go" 2>/dev/null)
fi

if [ -n "$BACKEND_PID" ]; then
    echo "Знайдено процес Backend на PID: $BACKEND_PID. Зупиняємо..."
    kill "$BACKEND_PID" 2>/dev/null
    sleep 2
    
    # Перевірка, чи процес зупинився
    if ps -p "$BACKEND_PID" > /dev/null 2>&1; then
        echo "Процес не зупинився, використовуємо kill -9..."
        kill -9 "$BACKEND_PID" 2>/dev/null
        sleep 1
    fi
    echo -e "${GREEN}✓ Backend зупинено.${NC}"
else
    echo -e "${YELLOW}⚠ Процес Backend не знайдено (можливо, вже зупинений).${NC}"
fi

# 2. Запуск Backend знову
echo ""
echo "Запуск Backend..."
nohup go run cmd/api/main.go > go_backend.log 2>&1 &
NEW_BACKEND_PID=$!
echo "Backend запущено на PID: $NEW_BACKEND_PID. Логи: backend/go_backend.log"

# 3. Перевірка Health Check з retry логікою
echo "Очікування запуску Go-сервера..."
MAX_ATTEMPTS=10
ATTEMPT=1
HEALTH_CHECK_PASSED=false

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    # Перевірка, чи процес ще запущений
    if ! ps -p $NEW_BACKEND_PID > /dev/null 2>&1; then
        echo -e "${RED}✗ Помилка: Процес Backend завершився. Перевірте backend/go_backend.log${NC}"
        exit 1
    fi
    
    # Спроба health check
    if curl -s http://localhost:8080/health 2>/dev/null | grep -q "ok"; then
        echo -e "${GREEN}✓ Backend успішно перезапущено на http://localhost:8080 (спроба $ATTEMPT/$MAX_ATTEMPTS)${NC}"
        HEALTH_CHECK_PASSED=true
        break
    fi
    
    if [ $ATTEMPT -lt $MAX_ATTEMPTS ]; then
        echo "Спроба $ATTEMPT/$MAX_ATTEMPTS: Backend ще не готовий, очікування 2 секунди..."
        sleep 2
    fi
    
    ATTEMPT=$((ATTEMPT + 1))
done

if [ "$HEALTH_CHECK_PASSED" = false ]; then
    echo -e "${RED}✗ Помилка: Backend не відповідає після $MAX_ATTEMPTS спроб. Перевірте backend/go_backend.log${NC}"
    exit 1
fi

cd .. # Повернення до кореня проєкту

