package protocol

import "errors"

var (
	// ErrInvalidMessageType - неверный тип сообщения
	ErrInvalidMessageType = errors.New("invalid message type")

	// ErrMissingClientID - отсутствует ID клиента
	ErrMissingClientID = errors.New("missing client ID")

	// ErrInvalidTimestamp - неверный timestamp
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrContentTooLarge - содержимое слишком большое
	ErrContentTooLarge = errors.New("content size exceeds limit")

	// ErrMessageExpired - сообщение устарело
	ErrMessageExpired = errors.New("message has expired")
)
