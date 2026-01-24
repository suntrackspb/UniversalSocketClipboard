# OpenWRT Clipboard

Централизованный буфер обмена для локальной сети с роутером OpenWRT в качестве сервера.

## Возможности

- Автоматическая синхронизация буфера обмена между устройствами
- Поддержка Windows, Linux, macOS
- WebSocket для real-time коммуникации
- Минимальное потребление ресурсов на роутере

## Установка

### Сервер (OpenWRT)

```bash
# Скачать из Releases или собрать самостоятельно
./build.sh
./deploy.sh
```

### Клиент

```bash
# Windows
clipboard-client-windows.exe -server ws://192.168.1.1:9090/ws

# Linux / macOS
clipboard-client -server ws://192.168.1.1:9090/ws
```

## Сборка

```bash
# Все платформы
make all

# Или быстрая сборка
./build.sh
```

## Технологии

- Go 1.21+
- WebSocket (gorilla/websocket)
- golang-design/clipboard

## Лицензия

MIT
