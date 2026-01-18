package client

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	goclipboard "github.com/atotto/clipboard"
	"golang.design/x/clipboard"
)

// ClipboardMonitor отслеживает изменения буфера обмена
type ClipboardMonitor struct {
	lastHash     string
	lastFilePath string // Кэш последнего пути к файлу
	onChange     func(content string)
	pollInterval time.Duration
	stopChan     chan struct{}
	useAdvanced  bool // Использовать ли golang.design/clipboard
}

// NewClipboardMonitor создает новый монитор буфера обмена
func NewClipboardMonitor(onChange func(content string)) *ClipboardMonitor {
	return &ClipboardMonitor{
		onChange:     onChange,
		pollInterval: 500 * time.Millisecond,
		stopChan:     make(chan struct{}),
		useAdvanced:  false, // По умолчанию используем atotto (без разрешений)
	}
}

// Start запускает мониторинг буфера обмена
func (m *ClipboardMonitor) Start() error {
	// Пробуем инициализировать golang.design/clipboard
	err := clipboard.Init()
	if err != nil {
		log.Printf("⚠️  Advanced clipboard (golang.design) not available: %v", err)
		log.Printf("📋 Using basic clipboard (atotto) - text only, no permissions needed")
		m.useAdvanced = false
	} else {
		log.Printf("✨ Advanced clipboard initialized - images support available!")
		m.useAdvanced = true
	}

	log.Printf("Clipboard monitor started, polling every %v", m.pollInterval)

	// Получаем текущее содержимое
	m.updateLastHash()

	// Запускаем мониторинг в фоне
	go m.monitorLoop()

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// monitorLoop основной цикл мониторинга
func (m *ClipboardMonitor) monitorLoop() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkClipboard()
		case <-m.stopChan:
			return
		}
	}
}

// checkClipboard проверяет изменения в буфере обмена
func (m *ClipboardMonitor) checkClipboard() {
	var content string
	var err error

	if m.useAdvanced {
		// Используем golang.design/clipboard
		// Приоритет: Изображение > Файлы > Текст

		// 1. Проверяем изображение
		imgData := clipboard.Read(clipboard.FmtImage)
		if len(imgData) > 0 {
			content = encodeImage(imgData)
			log.Printf("📷 Image detected in clipboard (%d bytes raw)", len(imgData))
		} else {
			// 2. Проверяем файлы (через текстовый формат - пути к файлам)
			textData := clipboard.Read(clipboard.FmtText)
			if len(textData) > 0 {
				text := string(textData)

				// Проверяем это путь к файлу?
				if isFilePath(text) {
					// Пытаемся найти полный путь
					fullPath := findFullPath(text)
					if fullPath != "" {
						// Проверяем изменился ли файл (по пути)
						if fullPath == m.lastFilePath {
							// Тот же файл, пропускаем (избегаем повторного чтения)
							return
						}

						m.lastFilePath = fullPath

						// Это файл! Читаем содержимое и кодируем
						fileContent, err := readFileContent(fullPath)
						if err == nil {
							content = encodeFile(fullPath, fileContent)
							log.Printf("📁 File detected: %s (%d bytes)", fullPath, len(fileContent))
						} else {
							// Не удалось прочитать файл, передаем путь
							content = "FILE_PATH:" + fullPath
							log.Printf("📁 File path detected: %s (content not readable)", fullPath)
						}
					} else {
						// Не нашли файл, передаем как текст
						content = text
						m.lastFilePath = "" // Сбрасываем кэш
					}
				} else {
					m.lastFilePath = "" // Сбрасываем кэш если это не файл
					// Обычный текст
					content = text
				}
			} else {
				return
			}
		}
	} else {
		// Используем atotto/clipboard (только текст)
		content, err = goclipboard.ReadAll()
		if err != nil {
			log.Printf("Failed to read clipboard: %v", err)
			return
		}
		if len(content) == 0 {
			return
		}
	}

	// Вычисляем хеш
	hash := computeHash(content)

	// Проверяем изменения
	if hash != m.lastHash {
		m.lastHash = hash
		log.Printf("Local clipboard changed (hash: %s, size: %d bytes)", hash[:min(8, len(hash))], len(content))

		// Вызываем коллбек
		if m.onChange != nil {
			m.onChange(content)
		}
	} else if m.lastFilePath != "" {
		// Если это файл и хеш не изменился - сбрасываем кэш пути (файл больше не в буфере)
		m.lastFilePath = ""
	}
}

// updateLastHash обновляет последний хеш без вызова коллбека
func (m *ClipboardMonitor) updateLastHash() {
	var content string

	if m.useAdvanced {
		// Проверяем изображение
		imgData := clipboard.Read(clipboard.FmtImage)
		if len(imgData) > 0 {
			content = encodeImage(imgData)
		} else {
			// Проверяем текст (может быть путь к файлу)
			textData := clipboard.Read(clipboard.FmtText)
			if len(textData) > 0 {
				text := string(textData)
				if isFilePath(text) {
					fileContent, err := readFileContent(text)
					if err == nil {
						content = encodeFile(text, fileContent)
					} else {
						content = "FILE_PATH:" + text
					}
				} else {
					content = text
				}
			}
		}
	} else {
		text, err := goclipboard.ReadAll()
		if err != nil {
			return
		}
		content = text
	}

	if len(content) > 0 {
		m.lastHash = computeHash(content)
	}
}

// SetClipboard устанавливает содержимое буфера обмена
func (m *ClipboardMonitor) SetClipboard(content string) error {
	// Обновляем хеш перед установкой, чтобы избежать петли
	m.lastHash = computeHash(content)

	// Проверяем тип контента
	if strings.HasPrefix(content, "IMAGE_BASE64:") {
		// Это изображение
		imgData, err := decodeImage(content)
		if err != nil {
			return err
		}
		log.Printf("📷 Setting image from server (%d bytes raw)", len(imgData))
		if m.useAdvanced {
			clipboard.Write(clipboard.FmtImage, imgData)
		} else {
			log.Printf("⚠️  Image received but advanced clipboard not available")
			return fmt.Errorf("image not supported with basic clipboard")
		}
	} else if strings.HasPrefix(content, "FILE_BASE64:") {
		// Это файл
		filePath, fileContent, err := decodeFile(content)
		if err != nil {
			return err
		}
		// Сохраняем файл во временную директорию
		savedPath, err := saveReceivedFile(filePath, fileContent)
		if err != nil {
			log.Printf("⚠️  Failed to save file: %v", err)
			return err
		}
		log.Printf("📁 File saved to temp: %s (%d bytes)", savedPath, len(fileContent))

		// Копируем файл в буфер обмена
		// На macOS используем pbcopy для правильного формата файлов
		if err := copyFileToClipboard(savedPath); err != nil {
			log.Printf("⚠️  Failed to copy file to clipboard: %v, trying text format", err)
			// Fallback: используем текстовый формат
			if m.useAdvanced {
				clipboard.Write(clipboard.FmtText, []byte(savedPath))
			} else {
				goclipboard.WriteAll(savedPath)
			}
		}

		// Обновляем кэш чтобы не читать файл снова
		m.lastFilePath = savedPath
		m.lastHash = computeHash("FILE_PATH:" + savedPath)
	} else if strings.HasPrefix(content, "FILE_PATH:") {
		// Это только путь к файлу (файл не был передан)
		filePath := strings.TrimPrefix(content, "FILE_PATH:")
		// Копируем файл в буфер обмена
		if err := copyFileToClipboard(filePath); err != nil {
			log.Printf("⚠️  Failed to copy file to clipboard: %v, trying text format", err)
			// Fallback: используем текстовый формат
			if m.useAdvanced {
				clipboard.Write(clipboard.FmtText, []byte(filePath))
			} else {
				goclipboard.WriteAll(filePath)
			}
		}
		// Обновляем кэш чтобы не читать файл снова
		m.lastFilePath = filePath
		m.lastHash = computeHash("FILE_PATH:" + filePath)
	} else {
		// Это текст
		log.Printf("Clipboard updated from server (size: %d bytes)", len(content))
		if m.useAdvanced {
			clipboard.Write(clipboard.FmtText, []byte(content))
		} else {
			err := goclipboard.WriteAll(content)
			if err != nil {
				log.Printf("Failed to write clipboard: %v", err)
				return err
			}
		}
	}

	return nil
}

// saveReceivedFile сохраняет полученный файл во временную директорию
func saveReceivedFile(originalPath string, content []byte) (string, error) {
	// Извлекаем имя файла
	fileName := originalPath
	if strings.Contains(fileName, "/") {
		parts := strings.Split(fileName, "/")
		fileName = parts[len(parts)-1]
	}

	// Всегда сохраняем во временную директорию
	// Это позволяет вставлять файл куда нужно через буфер обмена
	tmpDir := os.TempDir()

	// Создаем уникальное имя чтобы не конфликтовать с другими файлами
	timestamp := time.Now().Unix()
	savePath := fmt.Sprintf("%s/clipboard_%d_%s", tmpDir, timestamp, fileName)

	err := os.WriteFile(savePath, content, 0644)
	if err != nil {
		return "", err
	}

	// На macOS/Linux файл в буфере обмена - это путь к файлу
	// Когда пользователь вставляет его, система копирует файл в новое место
	// Поэтому сохраняем во временную директорию - файл будет доступен для вставки
	return savePath, nil
}

// Stop останавливает мониторинг
func (m *ClipboardMonitor) Stop() {
	close(m.stopChan)
	log.Printf("Clipboard monitor stopped")
}

// encodeImage кодирует изображение в base64 с префиксом
func encodeImage(imgData []byte) string {
	// Префикс для идентификации что это изображение
	prefix := "IMAGE_BASE64:"
	encoded := base64.StdEncoding.EncodeToString(imgData)
	return prefix + encoded
}

// decodeImage декодирует base64 изображение
func decodeImage(encoded string) ([]byte, error) {
	const prefix = "IMAGE_BASE64:"
	if len(encoded) < len(prefix) || encoded[:len(prefix)] != prefix {
		return nil, fmt.Errorf("not an image")
	}
	return base64.StdEncoding.DecodeString(encoded[len(prefix):])
}

// isFilePath проверяет является ли строка путем к файлу
func isFilePath(text string) bool {
	// Проверяем паттерны путей
	if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "file://") {
		// Проверяем что это реальный файл
		path := text
		if strings.HasPrefix(path, "file://") {
			path = strings.TrimPrefix(path, "file://")
		}
		info, err := os.Stat(path)
		return err == nil && !info.IsDir()
	}

	// На macOS при копировании файла может быть только имя файла
	// Проверяем что это похоже на имя файла (есть расширение)
	if strings.Contains(text, ".") && !strings.Contains(text, " ") {
		// Может быть имя файла, попробуем найти в стандартных местах
		return tryFindFile(text)
	}

	return false
}

// tryFindFile пытается найти файл по имени в стандартных директориях
func tryFindFile(fileName string) bool {
	return findFullPath(fileName) != ""
}

// findFullPath ищет полный путь к файлу по имени
func findFullPath(fileName string) string {
	// Если уже полный путь
	if strings.HasPrefix(fileName, "/") {
		if info, err := os.Stat(fileName); err == nil && !info.IsDir() {
			return fileName
		}
		return ""
	}

	// Ищем в стандартных директориях
	home := os.Getenv("HOME")
	searchDirs := []string{
		home + "/Desktop",
		home + "/Downloads",
		home + "/Documents",
		home + "/Pictures",
	}

	for _, dir := range searchDirs {
		path := dir + "/" + fileName
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

// readFileContent читает содержимое файла
func readFileContent(filePath string) ([]byte, error) {
	// Очищаем путь от file:// префикса
	path := filePath
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
	}

	// Ограничиваем размер файла (например, 10MB)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 10*1024*1024 {
		return nil, fmt.Errorf("file too large: %d bytes", info.Size())
	}

	return os.ReadFile(path)
}

// encodeFile кодирует файл в base64 с метаданными
func encodeFile(filePath string, fileContent []byte) string {
	// Префикс с путем и содержимым
	prefix := "FILE_BASE64:"
	encoded := base64.StdEncoding.EncodeToString(fileContent)
	return prefix + filePath + "|" + encoded
}

// decodeFile декодирует base64 файл
func decodeFile(encoded string) (string, []byte, error) {
	const prefix = "FILE_BASE64:"
	if len(encoded) < len(prefix) || encoded[:len(prefix)] != prefix {
		return "", nil, fmt.Errorf("not a file")
	}

	data := encoded[len(prefix):]
	parts := strings.SplitN(data, "|", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid file format")
	}

	filePath := parts[0]
	fileContent, err := base64.StdEncoding.DecodeString(parts[1])
	return filePath, fileContent, err
}

// copyFileToClipboard копирует файл в буфер обмена используя правильный формат
// Поддерживает macOS, Linux (X11/Wayland) и Windows
func copyFileToClipboard(filePath string) error {
	// Проверяем что файл существует
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file does not exist: %v", err)
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: используем osascript для правильного формата файлов
		script := fmt.Sprintf(`set the clipboard to (POSIX file "%s")`, filePath)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("osascript failed: %v", err)
		}
		log.Printf("📁 File copied to clipboard via osascript: %s", filePath)
		return nil

	case "linux":
		// Linux: пробуем разные способы в зависимости от окружения
		// Сначала пробуем wl-copy (Wayland)
		if cmd := exec.Command("wl-copy"); cmd.Run() == nil {
			// Wayland доступен, используем wl-copy с text/uri-list
			fileURL := fmt.Sprintf("file://%s\r\n", filePath)
			cmd := exec.Command("wl-copy", "--type", "text/uri-list")
			cmd.Stdin = strings.NewReader(fileURL)
			if err := cmd.Run(); err != nil {
				// Fallback на обычный текст
				return fmt.Errorf("wl-copy failed: %v", err)
			}
			log.Printf("📁 File copied to clipboard via wl-copy: %s", filePath)
			return nil
		}

		// X11: используем xclip с text/uri-list
		fileURL := fmt.Sprintf("file://%s\r\n", filePath)
		cmd := exec.Command("xclip", "-i", "-selection", "clipboard", "-t", "text/uri-list")
		cmd.Stdin = strings.NewReader(fileURL)
		if err := cmd.Run(); err != nil {
			// Fallback на обычный текст через xclip
			cmd := exec.Command("xclip", "-i", "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(filePath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("xclip failed: %v", err)
			}
			log.Printf("📁 File path copied to clipboard via xclip (text): %s", filePath)
			return nil
		}
		log.Printf("📁 File copied to clipboard via xclip (uri-list): %s", filePath)
		return nil

	case "windows":
		// Windows: используем PowerShell для копирования файла в буфер обмена
		// PowerShell может копировать файл как объект через Add-Type и Clipboard
		// Но проще использовать команду для копирования пути и надеяться что приложение распознает
		// Для полной поддержки нужен WinAPI с CF_HDROP, но это сложнее
		psScript := fmt.Sprintf(`[System.Windows.Forms.Clipboard]::SetText('%s')`, filePath)
		cmd := exec.Command("powershell", "-Command", psScript)
		if err := cmd.Run(); err != nil {
			// Fallback: пробуем через cmd
			cmd := exec.Command("cmd", "/c", "echo", filePath, "|", "clip")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Windows clipboard failed: %v", err)
			}
		}
		log.Printf("📁 File path copied to clipboard via PowerShell: %s", filePath)
		return nil

	default:
		return fmt.Errorf("file clipboard not implemented for %s", runtime.GOOS)
	}
}

// computeHash вычисляет хеш строки
func computeHash(data string) string {
	// Простой FNV-1a hash
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	for i := 0; i < len(data); i++ {
		hash ^= uint64(data[i])
		hash *= prime64
	}

	// Преобразуем в hex строку
	return fmt.Sprintf("%016x", hash)
}
