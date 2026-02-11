# Установка сервера / Server installation

Ручная установка и автозапуск сервера Universal Socket Clipboard на OpenWRT, Linux и Windows.

Manual installation and autostart of the Universal Socket Clipboard server on OpenWRT, Linux, and Windows.

---

## Русский

### Обзор

Сервер доступен в трёх вариантах бинарников:

- **OpenWRT (ARM64)** — `clipboard-server-openwrt` — для роутеров
- **Linux (x64)** — `clipboard-server-linux` — для серверов, NAS, обычного Linux
- **Windows (x64)** — `clipboard-server-windows.exe` — для Windows-серверов и рабочих станций

Параметр запуска: `-addr` — адрес и порт HTTP (по умолчанию `:9090`). Эндпоинты: `/ws` (WebSocket), `/health` (JSON), `/` (веб-страница статуса).

---

### OpenWRT (роутер)

1. Соберите или скачайте бинарник `clipboard-server-openwrt`.
2. Загрузите на роутер и запустите:

```bash
# Сборка
make server-openwrt

# Загрузка (замените IP и пользователя при необходимости)
scp bin/clipboard-server-openwrt root@192.168.1.1:/tmp/clipboard-server
ssh root@192.168.1.1 "chmod +x /tmp/clipboard-server"

# Запуск вручную (в фоне)
ssh root@192.168.1.1 "nohup /tmp/clipboard-server -addr :9090 > /tmp/clipboard-server.log 2>&1 &"
```

Либо используйте скрипт развёртывания: `./deploy.sh` (см. README).

**Полезные команды на роутере:**

```bash
# Логи
tail -f /tmp/clipboard-server.log

# Остановка
killall clipboard-server

# Проверка
curl http://192.168.1.1:9090/health
```

---

### Linux (x64) — systemd

Подходит для постоянной работы сервера на Linux (VPS, NAS, домашний сервер).

#### 1. Копирование бинарника

```bash
sudo cp clipboard-server-linux /usr/local/bin/clipboard-server
sudo chmod +x /usr/local/bin/clipboard-server
```

#### 2. Создание systemd unit

Создайте файл `/etc/systemd/system/clipboard-server.service`:

```ini
[Unit]
Description=Universal Socket Clipboard Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/clipboard-server -addr :9090
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

При необходимости замените `User=root` на отдельного пользователя (например `User=clipboard`); тогда убедитесь, что порт 9090 не занят и при необходимости измените порт в `-addr`.

#### 3. Включение и запуск

```bash
sudo systemctl daemon-reload
sudo systemctl enable clipboard-server
sudo systemctl start clipboard-server
```

#### Управление сервисом

```bash
# Статус
sudo systemctl status clipboard-server

# Остановка
sudo systemctl stop clipboard-server

# Запуск
sudo systemctl start clipboard-server

# Перезапуск
sudo systemctl restart clipboard-server

# Логи
journalctl -u clipboard-server -f

# Отключить автозапуск
sudo systemctl disable clipboard-server
```

#### Смена порта

Отредактируйте unit или создайте override:

```bash
sudo systemctl edit clipboard-server
```

Добавьте/измените в секции `[Service]`:

```ini
[Service]
ExecStart=
ExecStart=/usr/local/bin/clipboard-server -addr :8080
```

Затем:

```bash
sudo systemctl daemon-reload
sudo systemctl restart clipboard-server
```

---

### Windows (x64)

1. Скопируйте `clipboard-server-windows.exe` на целевую машину (например в `C:\Program Files\clipboard-server\`).
2. Запуск вручную (PowerShell или cmd):

```powershell
cd "C:\Program Files\clipboard-server"
.\clipboard-server-windows.exe -addr :9090
```

Для постоянной работы можно создать задачу в Планировщике заданий (аналогично клиенту в INSTALL_RU.md / INSTALL_EN.md): триггер «При запуске» или «При входе в систему», действие — запуск `clipboard-server-windows.exe -addr :9090`.

---

### Эндпоинты сервера

- **`/ws`** — WebSocket для клиентов буфера обмена.
- **`/health`** — JSON с полями `status`, `clients`, `version` (для мониторинга).
- **`/`** — веб-страница со статусом и количеством клиентов.

---

## English

### Overview

The server is provided as three binaries:

- **OpenWRT (ARM64)** — `clipboard-server-openwrt` — for routers
- **Linux (x64)** — `clipboard-server-linux` — for servers, NAS, desktop Linux
- **Windows (x64)** — `clipboard-server-windows.exe` — for Windows servers and workstations

Launch option: `-addr` — HTTP address and port (default `:9090`). Endpoints: `/ws` (WebSocket), `/health` (JSON), `/` (status page).

---

### OpenWRT (router)

1. Build or download `clipboard-server-openwrt`.
2. Upload and run on the router:

```bash
# Build
make server-openwrt

# Upload (change IP and user if needed)
scp bin/clipboard-server-openwrt root@192.168.1.1:/tmp/clipboard-server
ssh root@192.168.1.1 "chmod +x /tmp/clipboard-server"

# Run manually (background)
ssh root@192.168.1.1 "nohup /tmp/clipboard-server -addr :9090 > /tmp/clipboard-server.log 2>&1 &"
```

Or use the deploy script: `./deploy.sh` (see README).

**Useful commands on the router:**

```bash
# Logs
tail -f /tmp/clipboard-server.log

# Stop
killall clipboard-server

# Check
curl http://192.168.1.1:9090/health
```

---

### Linux (x64) — systemd

Suitable for running the server permanently on Linux (VPS, NAS, home server).

#### 1. Copy the binary

```bash
sudo cp clipboard-server-linux /usr/local/bin/clipboard-server
sudo chmod +x /usr/local/bin/clipboard-server
```

#### 2. Create systemd unit

Create `/etc/systemd/system/clipboard-server.service`:

```ini
[Unit]
Description=Universal Socket Clipboard Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/clipboard-server -addr :9090
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

You can change `User=root` to a dedicated user (e.g. `User=clipboard`) if desired; ensure port 9090 is free or adjust `-addr` accordingly.

#### 3. Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable clipboard-server
sudo systemctl start clipboard-server
```

#### Service management

```bash
# Status
sudo systemctl status clipboard-server

# Stop
sudo systemctl stop clipboard-server

# Start
sudo systemctl start clipboard-server

# Restart
sudo systemctl restart clipboard-server

# Logs
journalctl -u clipboard-server -f

# Disable autostart
sudo systemctl disable clipboard-server
```

#### Changing the port

Edit the unit or create an override:

```bash
sudo systemctl edit clipboard-server
```

Add or change in the `[Service]` section:

```ini
[Service]
ExecStart=
ExecStart=/usr/local/bin/clipboard-server -addr :8080
```

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl restart clipboard-server
```

---

### Windows (x64)

1. Copy `clipboard-server-windows.exe` to the target machine (e.g. `C:\Program Files\clipboard-server\`).
2. Run manually (PowerShell or cmd):

```powershell
cd "C:\Program Files\clipboard-server"
.\clipboard-server-windows.exe -addr :9090
```

For running at startup, create a scheduled task (similar to the client in INSTALL_RU.md / INSTALL_EN.md): trigger “At startup” or “At log on”, action — run `clipboard-server-windows.exe -addr :9090`.

---

### Server endpoints

- **`/ws`** — WebSocket for clipboard clients.
- **`/health`** — JSON with `status`, `clients`, `version` (for monitoring).
- **`/`** — Web status page with client count.
