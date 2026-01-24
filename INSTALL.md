# Установка и автозапуск клиента

Инструкции по установке клиента как системного сервиса для автоматического запуска при включении/пробуждении системы.

## Linux (systemd)

### Автоматическая установка

```bash
sudo ./install-linux.sh
```

### Ручная установка

1. Скопировать бинарник:
```bash
sudo cp clipboard-client-linux /usr/local/bin/clipboard-client
sudo chmod +x /usr/local/bin/clipboard-client
```

2. Создать systemd unit файл `/etc/systemd/system/clipboard-client.service`:
```ini
[Unit]
Description=OpenWRT Clipboard Client
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
Environment="DISPLAY=:0"
Environment="XAUTHORITY=/home/YOUR_USERNAME/.Xauthority"
ExecStart=/usr/local/bin/clipboard-client -server ws://192.168.1.1:9090/ws
# Для отладки добавьте -debug:
# ExecStart=/usr/local/bin/clipboard-client -server ws://192.168.1.1:9090/ws -debug
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. Включить и запустить:
```bash
sudo systemctl daemon-reload
sudo systemctl enable clipboard-client
sudo systemctl start clipboard-client
```

### Управление сервисом

```bash
# Статус
sudo systemctl status clipboard-client

# Остановка
sudo systemctl stop clipboard-client

# Запуск
sudo systemctl start clipboard-client

# Перезапуск
sudo systemctl restart clipboard-client

# Просмотр логов
journalctl -u clipboard-client -f

# Отключить автозапуск
sudo systemctl disable clipboard-client
```

---

## macOS (LaunchAgent)

### Установка

1. Скопировать бинарник:
```bash
sudo cp clipboard-client-macos /usr/local/bin/clipboard-client
sudo chmod +x /usr/local/bin/clipboard-client
```

2. Создать LaunchAgent файл `~/Library/LaunchAgents/com.clipboard.client.plist`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clipboard.client</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/clipboard-client</string>
        <string>-server</string>
        <string>ws://192.168.1.1:9090/ws</string>
        <!-- Для отладки раскомментируйте следующую строку: -->
        <!-- <string>-debug</string> -->
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/dev/null</string>
    <key>StandardErrorPath</key>
    <string>/dev/null</string>
</dict>
</plist>
```

3. Загрузить агент:
```bash
launchctl load ~/Library/LaunchAgents/com.clipboard.client.plist
```

### Управление

```bash
# Запустить
launchctl start com.clipboard.client

# Остановить
launchctl stop com.clipboard.client

# Перезагрузить конфигурацию
launchctl unload ~/Library/LaunchAgents/com.clipboard.client.plist
launchctl load ~/Library/LaunchAgents/com.clipboard.client.plist

# Проверить статус
launchctl list | grep clipboard

# Удалить автозапуск
launchctl unload ~/Library/LaunchAgents/com.clipboard.client.plist
rm ~/Library/LaunchAgents/com.clipboard.client.plist
```

---

## Windows (Task Scheduler)

### Установка

1. Скопировать `clipboard-client-windows.exe` в `C:\Program Files\clipboard-client\`

2. Создать задачу в Планировщике заданий:

**Через PowerShell (от имени администратора):**
```powershell
$action = New-ScheduledTaskAction -Execute "C:\Program Files\clipboard-client\clipboard-client-windows.exe" -Argument "-server ws://192.168.1.1:9090/ws"
$trigger = New-ScheduledTaskTrigger -AtLogOn
$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName "ClipboardClient" -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Description "OpenWRT Clipboard Client"
```

**Или через GUI:**
1. Открыть Планировщик заданий (Task Scheduler)
2. Создать задачу (Create Task)
3. **General:**
   - Имя: `ClipboardClient`
   - Запуск: `Только когда пользователь вошел в систему` (Run only when user is logged on)
   - Отметить: `Запускать с наивысшими правами` (Run with highest privileges)
4. **Triggers:**
   - New → `При входе в систему` (At log on)
5. **Actions:**
   - New → Program: `C:\Program Files\clipboard-client\clipboard-client-windows.exe`
   - Arguments: `-server ws://192.168.1.1:9090/ws`
   - Для отладки: `-server ws://192.168.1.1:9090/ws -debug`
6. **Conditions:**
   - Снять галочку "Запускать задачу только при питании от электросети"
7. **Settings:**
   - Отметить "Запускать задачу как можно скорее после пропуска запланированного запуска"
   - Отметить "При сбое выполнения задачи перезапускать через: 1 минута"
   - Попытки перезапуска: 3

### Управление

```powershell
# Запустить задачу
Start-ScheduledTask -TaskName "ClipboardClient"

# Остановить задачу
Stop-ScheduledTask -TaskName "ClipboardClient"

# Проверить статус
Get-ScheduledTask -TaskName "ClipboardClient"

# Удалить задачу
Unregister-ScheduledTask -TaskName "ClipboardClient" -Confirm:$false
```

**Или через GUI:**
- Планировщик заданий → Библиотека планировщика заданий → ClipboardClient → Правая кнопка → Запустить/Остановить/Удалить

---

## Режим отладки

По умолчанию клиент работает тихо - не логирует ошибки подключения и реконнекты.

Для включения подробного логирования используйте флаг `-debug`:

```bash
clipboard-client -server ws://192.168.1.1:9090/ws -debug
```

**Без `-debug`:**
- Только важные сообщения (начало работы)
- Нет логов об ошибках подключения
- Нет логов о реконнектах
- Нет логов об ошибках отправки/получения

**С `-debug`:**
- Все логи включены
- Видны все попытки подключения
- Видны все ошибки и реконнекты
- Полезно для отладки проблем

---

## Настройка

### Изменение URL сервера

**Linux (systemd):**
```bash
sudo systemctl edit clipboard-client
```
Добавить:
```ini
[Service]
Environment="SERVER_URL=ws://192.168.1.1:9090/ws"
ExecStart=
ExecStart=/usr/local/bin/clipboard-client -server ${SERVER_URL}
```

**macOS:**
Отредактировать `~/Library/LaunchAgents/com.clipboard.client.plist` и изменить аргумент `-server`.

**Windows:**
Отредактировать задачу в Планировщике заданий и изменить аргументы.

---

## Проверка работы

После установки проверьте что клиент запущен:

**Linux:**
```bash
systemctl status clipboard-client
```

**macOS:**
```bash
launchctl list | grep clipboard
ps aux | grep clipboard-client
```

**Windows:**
```powershell
Get-Process | Where-Object {$_.ProcessName -like "*clipboard*"}
```

---

## Удаление

**Linux:**
```bash
sudo systemctl stop clipboard-client
sudo systemctl disable clipboard-client
sudo rm /etc/systemd/system/clipboard-client.service
sudo systemctl daemon-reload
sudo rm /usr/local/bin/clipboard-client
```

**macOS:**
```bash
launchctl unload ~/Library/LaunchAgents/com.clipboard.client.plist
rm ~/Library/LaunchAgents/com.clipboard.client.plist
sudo rm /usr/local/bin/clipboard-client
```

**Windows:**
```powershell
Unregister-ScheduledTask -TaskName "ClipboardClient" -Confirm:$false
Remove-Item "C:\Program Files\clipboard-client\" -Recurse -Force
```
