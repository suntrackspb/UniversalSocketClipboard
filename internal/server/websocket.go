package server

import (
	"log"
	"net/http"
	"time"

	"github.com/denisuvarov/openwrt-clipboard/internal/protocol"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  protocol.ReadBufferSize,
	WriteBufferSize: protocol.WriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем подключения из локальной сети
		return true
	},
}

// WebSocketConn обертка над websocket.Conn
type WebSocketConn struct {
	*websocket.Conn
}

// HandleWebSocket обрабатывает WebSocket соединения
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Генерируем ID клиента (можно использовать UUID)
	clientID := generateClientID(r.RemoteAddr)

	client := &Client{
		ID:   clientID,
		Hub:  hub,
		Conn: &WebSocketConn{conn},
		Send: make(chan []byte, 256),
	}

	// Регистрируем клиента
	client.Hub.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}

// readPump читает сообщения от клиента
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(protocol.ClientTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(protocol.ClientTimeout))
		return nil
	})

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error from client %s: %v", c.ID, err)
			}
			break
		}

		// Парсим сообщение
		msg, err := protocol.FromJSON(messageData)
		if err != nil {
			log.Printf("Invalid message from client %s: %v", c.ID, err)
			continue
		}

		// Валидация
		if err := msg.Validate(); err != nil {
			log.Printf("Message validation failed from client %s: %v", c.ID, err)
			continue
		}

		// Проверяем размер содержимого
		if len(msg.Content) > protocol.MaxContentSize {
			log.Printf("Content too large from client %s: %d bytes", c.ID, len(msg.Content))
			errorMsg := protocol.NewErrorMessage(c.ID, "content too large")
			if errData, err := errorMsg.ToJSON(); err == nil {
				c.Send <- errData
			}
			continue
		}

		// Обрабатываем сообщение в зависимости от типа
		switch msg.Type {
		case protocol.TypeClipboardUpdate:
			log.Printf("Clipboard update from client %s (hash: %s, size: %d bytes)",
				c.ID, msg.Hash[:8], len(msg.Content))

			// Проверяем дедупликацию
			if msg.Hash != "" && c.LastHash == msg.Hash {
				log.Printf("Duplicate clipboard update from client %s, ignoring", c.ID)
				continue
			}

			// Обновляем хеш клиента
			c.LastHash = msg.Hash

			// Рассылаем обновление всем остальным клиентам
			c.Hub.Broadcast(msg, c.ID)

		case protocol.TypeClientHello:
			log.Printf("Client hello from %s", c.ID)
			ackMsg := protocol.NewMessage(protocol.TypeServerAck, "server", "connected")
			if ackData, err := ackMsg.ToJSON(); err == nil {
				c.Send <- ackData
			}

		case protocol.TypePing:
			pongMsg := protocol.NewMessage(protocol.TypePong, "server", "")
			if pongData, err := pongMsg.ToJSON(); err == nil {
				c.Send <- pongData
			}

		default:
			log.Printf("Unknown message type from client %s: %s", c.ID, msg.Type)
		}
	}
}

// writePump отправляет сообщения клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(protocol.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub закрыл канал
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error to client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error to client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// generateClientID генерирует ID клиента на основе адреса
func generateClientID(remoteAddr string) string {
	return remoteAddr + "-" + time.Now().Format("20060102150405")
}
