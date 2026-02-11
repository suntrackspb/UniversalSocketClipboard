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

**Сервер:**

- OpenWRT (роутер):

  ```bash
  # Скачать из Releases или собрать самостоятельно
  ./build.sh
  ./deploy.sh
  ```

- Linux / Windows (сервер, NAS, хост):

  Соберите серверные бинарники:

  ```bash
  make server-linux    # Linux x64
  make server-windows  # Windows x64
  ```

  Затем скопируйте и запустите бинарник на нужной машине:

  - Linux:

    ```bash
    scp bin/clipboard-server-linux user@your-server:/usr/local/bin/clipboard-server
    ssh user@your-server 'chmod +x /usr/local/bin/clipboard-server && clipboard-server -addr :9090'
    ```

  - Windows:

    ```powershell
    # Скопируйте bin/clipboard-server-windows.exe и запустите:
    clipboard-server-windows.exe -addr :9090
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

Universal Socket Clipboard is a centralized clipboard for your local network powered by a simple WebSocket server that can run on an OpenWRT router, a regular Linux/Windows server, or a NAS.

### Features

- Automatic clipboard sync across devices
- Windows, Linux, macOS support
- WebSocket for real-time communication
- Low resource usage on the router

### Installation

**Server:**

- OpenWRT (router):

  ```bash
  # Download from Releases or build locally
  ./build.sh
  ./deploy.sh
  ```

- Linux / Windows (server, NAS, host):

  Build server binaries:

  ```bash
  make server-linux    # Linux x64
  make server-windows  # Windows x64
  ```

  Then copy and run the binary on your machine:

  - Linux:

    ```bash
    scp bin/clipboard-server-linux user@your-server:/usr/local/bin/clipboard-server
    ssh user@your-server 'chmod +x /usr/local/bin/clipboard-server && clipboard-server -addr :9090'
    ```

  - Windows:

    ```powershell
    # Copy bin/clipboard-server-windows.exe and run:
    clipboard-server-windows.exe -addr :9090
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
