package client

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config filename inside the config directory.
const configFileName = "config"

// ConfigDir returns the directory for the client config file per OS:
//   - Linux:   $XDG_CONFIG_HOME/clipboard-client  (default ~/.config/clipboard-client)
//   - macOS:   ~/Library/Application Support/clipboard-client
//   - Windows: %APPDATA%\clipboard-client
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		// Prefer APPDATA (roaming); fallback to user profile.
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "clipboard-client"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "clipboard-client"), nil
	default:
		// Linux and other Unix: XDG_CONFIG_HOME, default ~/.config
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(home, ".config")
		}
		return filepath.Join(configHome, "clipboard-client"), nil
	}
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// LoadServerURL reads the config file and returns the server URL if the "server" key is set.
// Format: one line "server=ws://...", lines starting with # are ignored.
// Returns ("", false) if the file does not exist or "server" is not set.
func LoadServerURL() (string, bool) {
	path, err := ConfigPath()
	if err != nil {
		return "", false
	}

	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "server=") {
			url := strings.TrimSpace(strings.TrimPrefix(line, "server="))
			if url != "" {
				return url, true
			}
		}
	}
	return "", false
}
