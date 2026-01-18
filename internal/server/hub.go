package server

import (
	"log"
	"sync"

	"github.com/denisuvarov/openwrt-clipboard/internal/protocol"
)

// Client представляет подключенного клиента
type Client struct {
	ID       string
	Hub      *Hub
	Conn     *WebSocketConn
	Send     chan []byte
	LastHash string // Хеш последнего отправленного сообщения
}

// Hub управляет всеми подключенными клиентами
type Hub struct {
	// Зарегистрированные клиенты
	clients map[*Client]bool

	// Broadcast канал для всех клиентов
	broadcast chan *BroadcastMessage

	// Регистрация новых клиентов
	register chan *Client

	// Отмена регистрации клиентов
	unregister chan *Client

	// Мьютекс для безопасной работы с клиентами
	mu sync.RWMutex

	// Последнее состояние буфера обмена
	lastClipboard *protocol.Message
}

// BroadcastMessage содержит сообщение и исключения
type BroadcastMessage struct {
	Message   *protocol.Message
	ExcludeID string // ID клиента, которого нужно исключить из broadcast
}

// NewHub создает новый Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client, 10),
		unregister: make(chan *Client, 10),
		clients:    make(map[*Client]bool),
	}
}

// Run запускает основной цикл Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()

			// Проверяем лимит клиентов
			if len(h.clients) >= protocol.MaxClients {
				log.Printf("Max clients reached (%d), rejecting client %s", protocol.MaxClients, client.ID)
				close(client.Send)
				h.mu.Unlock()
				continue
			}

			h.clients[client] = true
			log.Printf("Client registered: %s (total: %d)", client.ID, len(h.clients))

			// Отправляем текущее состояние буфера новому клиенту
			if h.lastClipboard != nil {
				msg, err := h.lastClipboard.ToJSON()
				if err == nil {
					select {
					case client.Send <- msg:
					default:
						log.Printf("Failed to send initial clipboard to client %s", client.ID)
					}
				}
			}

			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("Client unregistered: %s (total: %d)", client.ID, len(h.clients))
			}
			h.mu.Unlock()

		case broadcastMsg := <-h.broadcast:
			h.mu.Lock()

			// Обновляем последнее состояние буфера
			if broadcastMsg.Message.Type == protocol.TypeClipboardUpdate {
				h.lastClipboard = broadcastMsg.Message
			}

			// Сериализуем сообщение один раз
			message, err := broadcastMsg.Message.ToJSON()
			if err != nil {
				log.Printf("Error serializing message: %v", err)
				h.mu.Unlock()
				continue
			}

			// Отправляем всем клиентам кроме отправителя
			for client := range h.clients {
				// Пропускаем клиента-отправителя
				if client.ID == broadcastMsg.ExcludeID {
					continue
				}

				// Проверяем дедупликацию
				if broadcastMsg.Message.Hash != "" && client.LastHash == broadcastMsg.Message.Hash {
					continue
				}

				select {
				case client.Send <- message:
					// Обновляем последний хеш клиента
					if broadcastMsg.Message.Hash != "" {
						client.LastHash = broadcastMsg.Message.Hash
					}
				default:
					// Канал переполнен, отключаем клиента
					log.Printf("Client %s send buffer full, disconnecting", client.ID)
					close(client.Send)
					delete(h.clients, client)
				}
			}

			h.mu.Unlock()
		}
	}
}

// ClientCount возвращает количество подключенных клиентов
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast отправляет сообщение всем клиентам
func (h *Hub) Broadcast(msg *protocol.Message, excludeClientID string) {
	h.broadcast <- &BroadcastMessage{
		Message:   msg,
		ExcludeID: excludeClientID,
	}
}
