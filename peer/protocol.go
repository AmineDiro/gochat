package peer

import (
	"encoding/json"
	"net"
)

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

func SendPeerList(conn net.Conn, l []string) error {
	enc := json.NewEncoder(conn)
	return enc.Encode(l)
}

func ReceivePeerList(conn net.Conn) ([]string, error) {
	dec := json.NewDecoder(conn)
	l := []string{}
	err := dec.Decode(&l)
	return l, err
}

func isAuthorized(s *Server, p *Peer) bool {
	return p.Version == s.Version
}
