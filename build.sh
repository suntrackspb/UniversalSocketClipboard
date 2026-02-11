#!/bin/bash

# Скрипт быстрой сборки для всех платформ

echo "🚀 OpenWRT Clipboard - Быстрая сборка"
echo "===================================="
echo ""

# Проверяем наличие Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go 1.21+ и попробуйте снова."
    exit 1
fi

echo "✓ Go версия: $(go version)"
echo ""

# Устанавливаем зависимости
echo "📦 Установка зависимостей..."
go mod download
go mod tidy
echo "✓ Зависимости установлены"
echo ""

# Создаем директорию для бинарников
mkdir -p bin

# Сервер для OpenWRT (ARM64)
echo "🔨 Компиляция сервера для OpenWRT (ARM64)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o bin/clipboard-server-openwrt \
    ./cmd/server
echo "✓ Сервер OpenWRT: bin/clipboard-server-openwrt ($(du -h bin/clipboard-server-openwrt | cut -f1))"
echo ""

# Сервер для Linux (x64)
echo "🔨 Компиляция сервера для Linux (x64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o bin/clipboard-server-linux \
    ./cmd/server
echo "✓ Сервер Linux: bin/clipboard-server-linux ($(du -h bin/clipboard-server-linux | cut -f1))"
echo ""

# Сервер для Windows (x64)
echo "🔨 Компиляция сервера для Windows (x64)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o bin/clipboard-server-windows.exe \
    ./cmd/server
echo "✓ Сервер Windows: bin/clipboard-server-windows.exe ($(du -h bin/clipboard-server-windows.exe | cut -f1))"
echo ""

# Клиент для Windows
echo "🔨 Компиляция клиента для Windows (x64)..."
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w -H=windowsgui" \
    -trimpath \
    -o bin/clipboard-client-windows.exe \
    ./cmd/client
echo "✓ Windows: bin/clipboard-client-windows.exe ($(du -h bin/clipboard-client-windows.exe | cut -f1))"
echo ""

# Клиент для Linux
echo "🔨 Компиляция клиента для Linux (x64)..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o bin/clipboard-client-linux \
    ./cmd/client
echo "✓ Linux: bin/clipboard-client-linux ($(du -h bin/clipboard-client-linux | cut -f1))"
echo ""

# Клиент для macOS
echo "🔨 Компиляция клиента для macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o bin/clipboard-client-macos \
    ./cmd/client
echo "✓ macOS: bin/clipboard-client-macos ($(du -h bin/clipboard-client-macos | cut -f1))"
echo ""

echo "===================================="
echo "✅ Все бинарники успешно собраны!"
echo ""
echo "📂 Результаты сборки:"
ls -lh bin/
echo ""
echo "📝 Следующие шаги:"
echo "  1. Загрузить сервер на роутер: make deploy-server ROUTER_IP=192.168.1.1"
echo "  2. Запустить клиент на устройствах"
echo ""
