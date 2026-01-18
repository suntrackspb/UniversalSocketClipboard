package client

import (
	"log"
	"time"

	"golang.design/x/clipboard"
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
	// Инициализируем библиотеку clipboard
	err := clipboard.Init()
	if err != nil {
		return err
	}

	log.Printf("Clipboard monitor started")

	// Получаем текущее содержимое
	m.updateLastHash()

	// Запускаем мониторинг в фоне
	go m.monitorLoop()

	return nil
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
	content := clipboard.Read(clipboard.FmtText)
	if len(content) == 0 {
		return
	}

	text := string(content)

	// Вычисляем хеш
	hash := computeHash(text)

	// Проверяем изменения
	if hash != m.lastHash {
		m.lastHash = hash
		log.Printf("Local clipboard changed (hash: %s, size: %d bytes)", hash[:8], len(text))

		// Вызываем коллбек
		if m.onChange != nil {
			m.onChange(text)
		}
	}
}

// updateLastHash обновляет последний хеш без вызова коллбека
func (m *ClipboardMonitor) updateLastHash() {
	content := clipboard.Read(clipboard.FmtText)
	if len(content) > 0 {
		m.lastHash = computeHash(string(content))
	}
}

// SetClipboard устанавливает содержимое буфера обмена
func (m *ClipboardMonitor) SetClipboard(text string) error {
	// Обновляем хеш перед установкой, чтобы избежать петли
	m.lastHash = computeHash(text)

	clipboard.Write(clipboard.FmtText, []byte(text))
	log.Printf("Clipboard updated from server (size: %d bytes)", len(text))

	return nil
}

// Stop останавливает мониторинг
func (m *ClipboardMonitor) Stop() {
	close(m.stopChan)
	log.Printf("Clipboard monitor stopped")
}

// computeHash вычисляет хеш строки (простая реализация)
func computeHash(data string) string {
	// Используем ту же функцию что и в протоколе
	// Импортируем из protocol если нужно
	return hashString(data)
}

// hashString простая хеш-функция для избежания циклических зависимостей
func hashString(s string) string {
	// Простой FNV-1a hash
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	for i := 0; i < len(s); i++ {
		hash ^= uint64(s[i])
		hash *= prime64
	}

	// Преобразуем в строку
	return string(rune(hash))
}
