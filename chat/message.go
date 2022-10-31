package chat

import "github.com/google/uuid"

type Header struct {
	MsgID      uuid.UUID `json:"id"`
	SenderID   string    `json:"senderId"`
	SenderName string    `json:"name"`
	Timestamp  int64     `json:"timestamp"`
}

type Payload string

type Message struct {
	Header  Header
	Payload Payload
}

type Terminal interface {
	Input() error
	Output() error
}
