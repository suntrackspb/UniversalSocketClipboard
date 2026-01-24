#!/bin/bash

# Скрипт для загрузки и запуска сервера на роутере OpenWRT

# Конфигурация (можно переопределить через переменные окружения)
ROUTER_IP="${ROUTER_IP:-192.168.1.1}"
ROUTER_USER="${ROUTER_USER:-root}"
ROUTER_PORT="${ROUTER_PORT:-9090}"
SERVER_BIN="bin/clipboard-server-openwrt"

echo "🚀 OpenWRT Clipboard - Развертывание на роутере"
echo "==============================================="
echo ""
echo "Роутер: $ROUTER_USER@$ROUTER_IP"
echo "Порт: $ROUTER_PORT"
echo ""

# Проверяем наличие бинарника
if [ ! -f "$SERVER_BIN" ]; then
    echo "❌ Файл $SERVER_BIN не найден!"
    echo "Сначала соберите сервер: make server-openwrt"
    exit 1
fi

echo "✓ Найден бинарник: $SERVER_BIN ($(du -h $SERVER_BIN | cut -f1))"
echo ""

# Загрузка на роутер
echo "📤 Загрузка сервера на роутер..."
scp "$SERVER_BIN" "$ROUTER_USER@$ROUTER_IP:/tmp/clipboard-server" || {
    echo "❌ Ошибка загрузки файла на роутер"
    exit 1
}
echo "✓ Сервер загружен в /tmp/clipboard-server"
echo ""

# Установка прав и запуск
echo "🔧 Настройка и запуск сервера..."
ssh "$ROUTER_USER@$ROUTER_IP" << EOF
    # Делаем исполняемым
    chmod +x /tmp/clipboard-server
    
    # Останавливаем старый процесс если запущен
    killall clipboard-server 2>/dev/null
    
    # Запускаем в фоне
    nohup /tmp/clipboard-server -addr :$ROUTER_PORT > /tmp/clipboard-server.log 2>&1 &
    
    # Ждем немного
    sleep 2
    
    # Проверяем что запустился
    if pgrep clipboard-server > /dev/null; then
        echo "✅ Сервер успешно запущен!"
        echo ""
        echo "📊 Информация:"
        ps | grep clipboard-server | grep -v grep
    else
        echo "❌ Ошибка запуска сервера"
        echo "Лог ошибок:"
        cat /tmp/clipboard-server.log
        exit 1
    fi
EOF

echo ""
echo "===================================="
echo "✅ Развертывание завершено!"
echo ""
echo "🌐 Сервер доступен по адресу:"
echo "   http://$ROUTER_IP:$ROUTER_PORT"
echo ""
echo "📝 Полезные команды:"
echo "   Просмотр логов:     ssh $ROUTER_USER@$ROUTER_IP 'tail -f /tmp/clipboard-server.log'"
echo "   Остановка сервера:  ssh $ROUTER_USER@$ROUTER_IP 'killall clipboard-server'"
echo "   Проверка статуса:   curl http://$ROUTER_IP:$ROUTER_PORT/health"
echo ""
