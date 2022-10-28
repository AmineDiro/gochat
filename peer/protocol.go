package peer

import (
	"encoding/json"
	"net"

	"github.com/google/uuid"
)

type Handshake struct {
	id         uuid.UUID `json:"uuid"`
	name       string    `json:"page"`
	listenAddr string    `json:"listenAddr"`
}

func SendHandshake(conn net.Conn, peer *Peer) error {
	enc := json.NewEncoder(conn)
	return enc.Encode(peer)
}

func ReceiveHandshake(conn net.Conn) (*Peer, error) {
	dec := json.NewDecoder(conn)
	p := &Peer{}
	err := dec.Decode(p)
	return p, err
}
