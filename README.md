# Universal Socket Clipboard

---

## Русский

Universal Socket Clipboard — это централизованный буфер обмена для локальной сети с простым WebSocket‑сервером, который можно запустить на роутере OpenWRT, обычном Linux/Windows сервере или NAS.

### Возможности

- Автоматическая синхронизация буфера обмена между устройствами
- Поддержка Windows, Linux, macOS
- WebSocket для real-time коммуникации
- Минимальное потребление ресурсов на роутере

### Установка

**Сервер:** Подробная установка и systemd-юнит — [INSTALL_SERVER.md](INSTALL_SERVER.md).

- OpenWRT (роутер): `./build.sh` и `./deploy.sh`, либо ручная загрузка (см. INSTALL_SERVER.md).
- Linux / Windows: соберите `make server-linux` или `make server-windows`, скопируйте бинарник и запустите `-addr :9090`; на Linux можно настроить systemd (пример юнита в INSTALL_SERVER.md).

**Клиент:**

```bash
# С указанием адреса в параметрах
clipboard-client -server ws://192.168.1.1:9090/ws

# Или без параметра — адрес берётся из конфиг-файла или значения по умолчанию
clipboard-client
```

Адрес сервера можно задать в конфиг-файле (см. [INSTALL_RU.md](INSTALL_RU.md)#конфигурационный-файл).

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

Universal Socket Clipboard is a centralized clipboard for your local network powered by a simple WebSocket server that can run on an OpenWRT router, a regular Linux/Windows server, or a NAS.

### Features

- Automatic clipboard sync across devices
- Windows, Linux, macOS support
- WebSocket for real-time communication
- Low resource usage on the router

### Installation

**Server:** Full installation and systemd unit — [INSTALL_SERVER.md](INSTALL_SERVER.md).

- OpenWRT: `./build.sh` and `./deploy.sh`, or manual setup (see INSTALL_SERVER.md).
- Linux / Windows: `make server-linux` or `make server-windows`, then copy and run with `-addr :9090`; on Linux you can use the systemd unit (example in INSTALL_SERVER.md).

**Client:**

```bash
# With server URL on the command line
clipboard-client -server ws://192.168.1.1:9090/ws

# Or without — URL is read from config file or default
clipboard-client
```

You can put the server URL in a config file (see [INSTALL_EN.md](INSTALL_EN.md)#configuration-file).

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
