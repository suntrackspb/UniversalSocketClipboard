package client

import (
	"log"
	"net/url"
	"time"

	"github.com/denisuvarov/openwrt-clipboard/internal/protocol"
	"github.com/gorilla/websocket"
)

// WSClient представляет WebSocket клиента
type WSClient struct {
	serverURL    string
	conn         *websocket.Conn
	clientID     string
	sendChan     chan *protocol.Message
	receiveChan  chan *protocol.Message
	reconnecting bool
}

// NewWSClient создает нового WebSocket клиента
func NewWSClient(serverURL, clientID string) *WSClient {
	return &WSClient{
		serverURL:   serverURL,
		clientID:    clientID,
		sendChan:    make(chan *protocol.Message, 10),
		receiveChan: make(chan *protocol.Message, 10),
	}
}

// Connect подключается к серверу
func (c *WSClient) Connect() error {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return err
	}

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.conn = conn
	log.Printf("Connected to server")

	// Отправляем приветствие
	helloMsg := protocol.NewMessage(protocol.TypeClientHello, c.clientID, "")
	if err := c.sendMessage(helloMsg); err != nil {
		log.Printf("Failed to send hello: %v", err)
	}

	return nil
}

// Start запускает клиента
func (c *WSClient) Start() {
	go c.readPump()
	go c.writePump()
}

// readPump читает сообщения от сервера
func (c *WSClient) readPump() {
	defer func() {
		if c.conn != nil {
			c.conn.Close()
		}
		c.handleDisconnect()
	}()

	for {
		if c.conn == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			c.handleDisconnect()
			return
		}

		msg, err := protocol.FromJSON(messageData)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Отправляем сообщение в канал получения
		select {
		case c.receiveChan <- msg:
		default:
			log.Printf("Receive channel full, dropping message")
		}
	}
}

// writePump отправляет сообщения на сервер
func (c *WSClient) writePump() {
	ticker := time.NewTicker(protocol.PingInterval)
	defer func() {
		ticker.Stop()
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		select {
		case msg := <-c.sendChan:
			if c.conn == nil {
				log.Printf("Not connected, cannot send message")
				continue
			}

			if err := c.sendMessage(msg); err != nil {
				log.Printf("Send error: %v", err)
				c.handleDisconnect()
				return
			}

		case <-ticker.C:
			if c.conn == nil {
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				c.handleDisconnect()
				return
			}
		}
	}
}

// sendMessage отправляет сообщение на сервер
func (c *WSClient) sendMessage(msg *protocol.Message) error {
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// SendClipboard отправляет обновление буфера обмена
func (c *WSClient) SendClipboard(content string) {
	msg := protocol.NewMessage(protocol.TypeClipboardUpdate, c.clientID, content)

	select {
	case c.sendChan <- msg:
		log.Printf("Sending clipboard update (hash: %s, size: %d bytes)", msg.Hash[:8], len(content))
	default:
		log.Printf("Send channel full, dropping clipboard update")
	}
}

// ReceiveChan возвращает канал для получения сообщений
func (c *WSClient) ReceiveChan() <-chan *protocol.Message {
	return c.receiveChan
}

// handleDisconnect обрабатывает разрыв соединения
func (c *WSClient) handleDisconnect() {
	if c.reconnecting {
		return
	}

	c.reconnecting = true
	c.conn = nil

	log.Printf("Disconnected from server, attempting to reconnect...")

	// Пытаемся переподключиться
	go func() {
		for {
			time.Sleep(5 * time.Second)

			log.Printf("Reconnecting...")
			if err := c.Connect(); err != nil {
				log.Printf("Reconnect failed: %v", err)
				continue
			}

			c.reconnecting = false
			log.Printf("Reconnected successfully")
			return
		}
	}()
}

// Close закрывает соединение
func (c *WSClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
