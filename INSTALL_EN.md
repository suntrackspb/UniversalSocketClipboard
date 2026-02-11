# Client installation and autostart

Instructions for installing the client as a system service so it starts automatically on boot or wake.

## Configuration file

You can omit the server URL from the command line and set it in a config file. If `-server` is not passed, the client looks for the config file and reads the URL from it; otherwise it uses the default `ws://192.168.1.1:9090/ws`.

**Config file location by OS:**

| OS      | Config file path |
|---------|-------------------|
| Linux   | `~/.config/clipboard-client/config` (or `$XDG_CONFIG_HOME/clipboard-client/config`) |
| macOS   | `~/Library/Application Support/clipboard-client/config` |
| Windows | `%APPDATA%\clipboard-client\config` (e.g. `C:\Users\<user>\AppData\Roaming\clipboard-client\config`) |

**File format** — a single line with the `server=` key (lines starting with `#` are ignored):

```
# WebSocket server URL
server=ws://192.168.1.1:9090/ws
```

The `-server` command-line flag overrides the config file.

## Linux (systemd)

### Automatic installation

```bash
sudo ./install-linux.sh
```

### Manual installation

1. Copy the binary:
```bash
sudo cp clipboard-client-linux /usr/local/bin/clipboard-client
sudo chmod +x /usr/local/bin/clipboard-client
```

2. Create a systemd unit file `/etc/systemd/system/clipboard-client.service`:
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
# For debugging add -debug:
# ExecStart=/usr/local/bin/clipboard-client -server ws://192.168.1.1:9090/ws -debug
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable clipboard-client
sudo systemctl start clipboard-client
```

### Service management

```bash
# Status
sudo systemctl status clipboard-client

# Stop
sudo systemctl stop clipboard-client

# Start
sudo systemctl start clipboard-client

# Restart
sudo systemctl restart clipboard-client

# View logs
journalctl -u clipboard-client -f

# Disable autostart
sudo systemctl disable clipboard-client
```

---

## macOS (LaunchAgent)

### Installation

1. Copy the binary:
```bash
sudo cp clipboard-client-macos /usr/local/bin/clipboard-client
sudo chmod +x /usr/local/bin/clipboard-client
```

2. Create LaunchAgent file `~/Library/LaunchAgents/com.clipboard.client.plist`:
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
        <!-- For debugging uncomment the next line: -->
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

3. Load the agent:
```bash
launchctl load ~/Library/LaunchAgents/com.clipboard.client.plist
```

### Management

```bash
# Start
launchctl start com.clipboard.client

# Stop
launchctl stop com.clipboard.client

# Reload configuration
launchctl unload ~/Library/LaunchAgents/com.clipboard.client.plist
launchctl load ~/Library/LaunchAgents/com.clipboard.client.plist

# Check status
launchctl list | grep clipboard

# Remove autostart
launchctl unload ~/Library/LaunchAgents/com.clipboard.client.plist
rm ~/Library/LaunchAgents/com.clipboard.client.plist
```

---

## Windows (Task Scheduler)

### Installation

1. Copy `clipboard-client-windows.exe` to `C:\Program Files\clipboard-client\`

2. Create a task in Task Scheduler:

**Via PowerShell (run as Administrator):**
```powershell
$action = New-ScheduledTaskAction -Execute "C:\Program Files\clipboard-client\clipboard-client-windows.exe" -Argument "-server ws://192.168.1.1:9090/ws"
$trigger = New-ScheduledTaskTrigger -AtLogOn
$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName "ClipboardClient" -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Description "OpenWRT Clipboard Client"
```

**Or via GUI:**
1. Open Task Scheduler
2. Create Task (not Create Basic Task)
3. **General:**
   - Name: `ClipboardClient`
   - Run: "Only when the user is logged on"
   - Check: "Run with highest privileges"
4. **Triggers:**
   - New → "At log on"
5. **Actions:**
   - New → Program: `C:\Program Files\clipboard-client\clipboard-client-windows.exe`
   - Arguments: `-server ws://192.168.1.1:9090/ws`
   - For debugging: `-server ws://192.168.1.1:9090/ws -debug`
6. **Conditions:**
   - Uncheck "Start the task only if the computer is on AC power"
7. **Settings:**
   - Check "Run task as soon as possible after a scheduled start is missed"
   - Check "If the task fails, restart every: 1 minute"
   - Restart attempts: 3

### Management

```powershell
# Start task
Start-ScheduledTask -TaskName "ClipboardClient"

# Stop task
Stop-ScheduledTask -TaskName "ClipboardClient"

# Check status
Get-ScheduledTask -TaskName "ClipboardClient"

# Delete task
Unregister-ScheduledTask -TaskName "ClipboardClient" -Confirm:$false
```

**Or via GUI:**
- Task Scheduler → Task Scheduler Library → ClipboardClient → Right-click → Run / End / Delete

---

## Debug mode

By default the client runs quietly and does not log connection errors or reconnects.

To enable verbose logging, use the `-debug` flag:

```bash
clipboard-client -server ws://192.168.1.1:9090/ws -debug
```

**Without `-debug`:**
- Only important messages (startup)
- No connection error logs
- No reconnect logs
- No send/receive error logs

**With `-debug`:**
- All logging enabled
- All connection attempts visible
- All errors and reconnects visible
- Useful for troubleshooting

---

## Configuration

### Changing the server URL

**Linux (systemd):**
```bash
sudo systemctl edit clipboard-client
```
Add:
```ini
[Service]
Environment="SERVER_URL=ws://192.168.1.1:9090/ws"
ExecStart=
ExecStart=/usr/local/bin/clipboard-client -server ${SERVER_URL}
```

**macOS:**
Edit `~/Library/LaunchAgents/com.clipboard.client.plist` and change the `-server` argument.

**Windows:**
Edit the task in Task Scheduler and change the arguments.

---

## Verifying it works

After installation, verify the client is running:

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

## Uninstall

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
