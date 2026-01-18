package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// MessageType определяет типы сообщений в протоколе
type MessageType string

const (
	// TypeClipboardUpdate - обновление содержимого буфера обмена
	TypeClipboardUpdate MessageType = "clipboard_update"
	// TypeClientHello - приветствие от клиента при подключении
	TypeClientHello MessageType = "client_hello"
	// TypeServerAck - подтверждение от сервера
	TypeServerAck MessageType = "server_ack"
	// TypeError - сообщение об ошибке
	TypeError MessageType = "error"
	// TypePing - пинг для поддержания соединения
	TypePing MessageType = "ping"
	// TypePong - ответ на пинг
	TypePong MessageType = "pong"
)

// Message - основная структура сообщения
type Message struct {
	Type      MessageType `json:"type"`
	Content   string      `json:"content,omitempty"`
	ClientID  string      `json:"client_id"`
	Timestamp int64       `json:"timestamp"`
	Hash      string      `json:"hash,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// ClipboardData - данные буфера обмена
type ClipboardData struct {
	Text      string `json:"text"`
	Format    string `json:"format"` // "text", "image", etc.
	Size      int    `json:"size"`
	Timestamp int64  `json:"timestamp"`
}

// NewMessage создает новое сообщение
func NewMessage(msgType MessageType, clientID string, content string) *Message {
	msg := &Message{
		Type:      msgType,
		Content:   content,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}

	// Вычисляем хеш для дедупликации
	if content != "" {
		msg.Hash = ComputeHash(content)
	}

	return msg
}

// NewErrorMessage создает сообщение об ошибке
func NewErrorMessage(clientID string, errorText string) *Message {
	return &Message{
		Type:      TypeError,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
		Error:     errorText,
	}
}

// ToJSON сериализует сообщение в JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON десериализует сообщение из JSON
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Validate проверяет корректность сообщения
func (m *Message) Validate() error {
	if m.Type == "" {
		return ErrInvalidMessageType
	}
	if m.ClientID == "" {
		return ErrMissingClientID
	}
	if m.Timestamp == 0 {
		return ErrInvalidTimestamp
	}
	return nil
}

// ComputeHash вычисляет SHA256 хеш строки
func ComputeHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IsRecent проверяет, не устарело ли сообщение
func (m *Message) IsRecent(maxAge time.Duration) bool {
	msgTime := time.Unix(m.Timestamp, 0)
	return time.Since(msgTime) <= maxAge
}
