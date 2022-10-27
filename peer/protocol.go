package peer

import "github.com/google/uuid"

type Handshake struct {
	id         uuid.UUID `json:"uuid"`
	name       string    `json:"page"`
	listenAddr string    `json:"listenAddr"`
}
