package client

import (
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
)

// ClipboardMonitor отслеживает изменения буфера обмена
type ClipboardMonitor struct {
	lastHash     string
	onChange     func(content string)
	pollInterval time.Duration
	stopChan     chan struct{}
}

// NewClipboardMonitor создает новый монитор буфера обмена
func NewClipboardMonitor(onChange func(content string)) *ClipboardMonitor {
	return &ClipboardMonitor{
		onChange:     onChange,
		pollInterval: 500 * time.Millisecond,
		stopChan:     make(chan struct{}),
	}
}

// Start запускает мониторинг буфера обмена
func (m *ClipboardMonitor) Start() error {
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
	// Читаем текущее содержимое
	text, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Failed to read clipboard: %v", err)
		return
	}

	if len(text) == 0 {
		return
	}

	// Вычисляем хеш
	hash := computeHash(text)

	// Проверяем изменения
	if hash != m.lastHash {
		m.lastHash = hash
		log.Printf("Local clipboard changed (hash: %s, size: %d bytes)", hash[:min(8, len(hash))], len(text))

		// Вызываем коллбек
		if m.onChange != nil {
			m.onChange(text)
		}
	}
}

// updateLastHash обновляет последний хеш без вызова коллбека
func (m *ClipboardMonitor) updateLastHash() {
	text, err := clipboard.ReadAll()
	if err == nil && len(text) > 0 {
		m.lastHash = computeHash(text)
	}
}

// SetClipboard устанавливает содержимое буфера обмена
func (m *ClipboardMonitor) SetClipboard(text string) error {
	// Обновляем хеш перед установкой, чтобы избежать петли
	m.lastHash = computeHash(text)

	log.Printf("Clipboard updated from server (size: %d bytes)", len(text))

	err := clipboard.WriteAll(text)
	if err != nil {
		log.Printf("Failed to write clipboard: %v", err)
		return err
	}

	return nil
}

// Stop останавливает мониторинг
func (m *ClipboardMonitor) Stop() {
	close(m.stopChan)
	log.Printf("Clipboard monitor stopped")
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
