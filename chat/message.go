package chat

import (
	"encoding/json"
	"net"

	"github.com/google/uuid"
)

type Message struct {
	MsgID      uuid.UUID `json:"id"`
	SenderID   uuid.UUID `json:"senderId"`
	SenderName string    `json:"name"`
	Timestamp  int64     `json:"timestamp"`
	Payload    string    `json:"payload"`
}

func SendMessage(conn net.Conn, msg *Message) error {
	enc := json.NewEncoder(conn)
	return enc.Encode(msg)
}

func ReceiveMessage(conn net.Conn) (Message, error) {
	var msg Message
	dec := json.NewDecoder(conn)
	err := dec.Decode(&msg)
	return msg, err
}
