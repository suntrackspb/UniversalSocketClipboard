package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/denisuvarov/openwrt-clipboard/internal/client"
	"github.com/denisuvarov/openwrt-clipboard/internal/protocol"
)

const defaultServerURL = "ws://192.168.1.1:9090/ws"

var (
	serverURL = flag.String("server", "", "WebSocket server URL (overrides config file)")
	clientID  = flag.String("id", "", "Client ID (auto-generated if empty)")
	debug     = flag.Bool("debug", false, "Enable debug logging (connection errors, reconnects, etc.)")
	version   = "dev" // Будет заменено при сборке через -ldflags
)

func main() {
	flag.Parse()

	// Адрес сервера: флаг > конфиг-файл > значение по умолчанию
	if *serverURL == "" {
		if url, ok := client.LoadServerURL(); ok {
			*serverURL = url
			if *debug {
				log.Printf("Using server URL from config file")
			}
		} else {
			*serverURL = defaultServerURL
		}
	}

	log.Printf("OpenWRT Clipboard Client %s", version)

	// Генерируем Client ID если не указан
	if *clientID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		*clientID = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}

	log.Printf("Client ID: %s", *clientID)
	log.Printf("Server URL: %s", *serverURL)

	// Создаем WebSocket клиента
	wsClient := client.NewWSClient(*serverURL, *clientID, *debug)

	// Пытаемся подключиться (если не получится - handleDisconnect будет пытаться бесконечно)
	// Не используем log.Fatalf чтобы не завершать программу при первой неудаче
	_ = wsClient.Connect() // Игнорируем ошибку - реконнект будет в handleDisconnect

	// Создаем монитор буфера обмена
	clipMonitor := client.NewClipboardMonitor(*debug, func(content string) {
		// Проверяем размер
		if len(content) > protocol.MaxContentSize {
			if *debug {
				log.Printf("Clipboard content too large (%d bytes), not sending", len(content))
			}
			return
		}

		// Отправляем на сервер
		wsClient.SendClipboard(content)
	})

	// Запускаем монитор
	if err := clipMonitor.Start(); err != nil {
		log.Fatalf("Failed to start clipboard monitor: %v", err)
	}

	// Запускаем WebSocket клиента
	wsClient.Start()

	// Обрабатываем сообщения от сервера
	go func() {
		for msg := range wsClient.ReceiveChan() {
			switch msg.Type {
			case protocol.TypeClipboardUpdate:
				// Игнорируем свои собственные сообщения
				if msg.ClientID == *clientID {
					continue
				}

				if *debug {
					log.Printf("Received clipboard update from %s (hash: %s, size: %d bytes)",
						msg.ClientID, msg.Hash[:8], len(msg.Content))
				}

				// Обновляем локальный буфер обмена
				if err := clipMonitor.SetClipboard(msg.Content); err != nil && *debug {
					log.Printf("Failed to update clipboard: %v", err)
				}

			case protocol.TypeServerAck:
				if *debug {
					log.Printf("Server acknowledged connection")
				}

			case protocol.TypeError:
				if *debug {
					log.Printf("Server error: %s", msg.Error)
				}

			case protocol.TypePong:
				// Игнорируем pong сообщения

			default:
				if *debug {
					log.Printf("Unknown message type: %s", msg.Type)
				}
			}
		}
	}()

	if *debug {
		log.Printf("Client started successfully")
		log.Printf("Monitoring clipboard and syncing with server...")
	}

	// Ожидаем сигнала завершения
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	if *debug {
		log.Println("Shutting down client...")
	}
	clipMonitor.Stop()
	wsClient.Close()
	if *debug {
		log.Println("Client stopped")
	}
}

// getLogPath возвращает путь для лог-файла
func getLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "clipboard-client.log"
	}
	return filepath.Join(home, ".clipboard-client.log")
}
