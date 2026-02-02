# Universal Socket Clipboard

---

## Русский

Централизованный буфер обмена для локальной сети с роутером OpenWRT в качестве сервера.

### Возможности

- Автоматическая синхронизация буфера обмена между устройствами
- Поддержка Windows, Linux, macOS
- WebSocket для real-time коммуникации
- Минимальное потребление ресурсов на роутере

### Установка

**Сервер (OpenWRT):**

```bash
# Скачать из Releases или собрать самостоятельно
./build.sh
./deploy.sh
```

**Клиент:**

```bash
# Windows
clipboard-client-windows.exe -server ws://192.168.1.1:9090/ws

# Linux / macOS
clipboard-client -server ws://192.168.1.1:9090/ws
```

Подробнее: [INSTALL_RU.md](INSTALL_RU.md) — установка и автозапуск клиента.

### Сборка

```bash
# Все платформы
make all

# Или быстрая сборка
./build.sh
```

### Технологии

- Go 1.21+
- WebSocket (gorilla/websocket)
- golang-design/clipboard

### Лицензия

MIT

---

## English

A centralized clipboard for your local network, with an OpenWRT router as the server.

### Features

- Automatic clipboard sync across devices
- Windows, Linux, macOS support
- WebSocket for real-time communication
- Low resource usage on the router

### Installation

**Server (OpenWRT):**

```bash
# Download from Releases or build locally
./build.sh
./deploy.sh
```

**Client:**

```bash
# Windows
clipboard-client-windows.exe -server ws://192.168.1.1:9090/ws

# Linux / macOS
clipboard-client -server ws://192.168.1.1:9090/ws
```

See [INSTALL_EN.md](INSTALL_EN.md) for client installation and autostart.

### Build

```bash
# All platforms
make all

# Or quick build
./build.sh
```

### Tech stack

- Go 1.21+
- WebSocket (gorilla/websocket)
- golang-design/clipboard

### License

MIT
