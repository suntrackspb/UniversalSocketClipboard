package client

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	goclipboard "github.com/atotto/clipboard"
)

// ClipboardMonitor отслеживает изменения буфера обмена
type ClipboardMonitor struct {
	lastHash     string
	onChange     func(content string)
	pollInterval time.Duration
	stopChan     chan struct{}
	debug        bool
}

// NewClipboardMonitor создает новый монитор буфера обмена
func NewClipboardMonitor(debug bool, onChange func(content string)) *ClipboardMonitor {
	return &ClipboardMonitor{
		onChange:     onChange,
		pollInterval: 500 * time.Millisecond,
		stopChan:     make(chan struct{}),
		debug:        debug,
	}
}

// Start запускает мониторинг буфера обмена
func (m *ClipboardMonitor) Start() error {
	if m.debug {
		log.Printf("Clipboard monitor started, polling every %v", m.pollInterval)
	}

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
	content, err := goclipboard.ReadAll()
	if err != nil {
		if m.debug {
			log.Printf("Failed to read clipboard: %v", err)
		}
		return
	}
	if len(content) == 0 {
		return
	}

	// Игнорируем пути к файлам
	if isFilePath(content) {
		return
	}

	// Вычисляем хеш
	hash := computeHash(content)

	// Проверяем изменения
	if hash != m.lastHash {
		m.lastHash = hash
		if m.debug {
			log.Printf("Local clipboard changed (hash: %s, size: %d bytes)", hash[:min(8, len(hash))], len(content))
		}

		// Вызываем коллбек
		if m.onChange != nil {
			m.onChange(content)
		}
	}
}

// updateLastHash обновляет последний хеш без вызова коллбека
func (m *ClipboardMonitor) updateLastHash() {
	text, err := goclipboard.ReadAll()
	if err != nil {
		return
	}
	if len(text) == 0 {
		return
	}

	// Игнорируем пути к файлам
	if isFilePath(text) {
		return
	}

	m.lastHash = computeHash(text)
}

// SetClipboard устанавливает содержимое буфера обмена
func (m *ClipboardMonitor) SetClipboard(content string) error {
	// Обновляем хеш перед установкой, чтобы избежать петли
	m.lastHash = computeHash(content)

	if m.debug {
		log.Printf("Clipboard updated from server (size: %d bytes)", len(content))
	}
	err := goclipboard.WriteAll(content)
	if err != nil {
		if m.debug {
			log.Printf("Failed to write clipboard: %v", err)
		}
		return err
	}

	return nil
}

// Stop останавливает мониторинг
func (m *ClipboardMonitor) Stop() {
	close(m.stopChan)
	if m.debug {
		log.Printf("Clipboard monitor stopped")
	}
}

// isFilePath проверяет является ли строка путем к файлу
func isFilePath(text string) bool {
	// Убираем пробелы по краям
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return false
	}

	// Проверяем паттерны путей
	// Unix/Mac пути начинаются с /
	if strings.HasPrefix(text, "/") {
		// Проверяем что это реальный файл
		if info, err := os.Stat(text); err == nil && !info.IsDir() {
			return true
		}
	}

	// Windows пути начинаются с C:\ или подобное
	if len(text) >= 3 && text[1] == ':' && (text[2] == '\\' || text[2] == '/') {
		if info, err := os.Stat(text); err == nil && !info.IsDir() {
			return true
		}
	}

	// file:// URI
	if strings.HasPrefix(text, "file://") {
		return true
	}

	return false
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
