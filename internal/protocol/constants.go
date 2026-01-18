package protocol

import "time"

const (
	// MaxContentSize - максимальный размер содержимого (10 MB)
	MaxContentSize = 10 * 1024 * 1024

	// MaxClients - максимальное количество подключенных клиентов
	MaxClients = 20

	// ClientTimeout - таймаут для неактивных клиентов
	ClientTimeout = 5 * time.Minute

	// PingInterval - интервал отправки ping сообщений
	PingInterval = 30 * time.Second

	// PongTimeout - таймаут ожидания pong ответа
	PongTimeout = 10 * time.Second

	// MessageMaxAge - максимальный возраст сообщения
	MessageMaxAge = 1 * time.Minute

	// ReadBufferSize - размер буфера чтения WebSocket
	ReadBufferSize = 1024

	// WriteBufferSize - размер буфера записи WebSocket
	WriteBufferSize = 1024
)
