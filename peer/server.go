package peer

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Peer struct {
	Id         uuid.UUID `json:"uuid"`
	Name       string    `json:"page"`
	ListenAddr string    `json:"listenAddr"`
	Version    string    `json:version`
}

type Server struct {
	Peer
	listener net.Listener
	mu       sync.Mutex
	PeerList map[uuid.UUID]*Peer
}

func MkServer(addr string, name string) (s *Server) {
	s = &Server{
		Peer: Peer{
			Id:         uuid.New(),
			ListenAddr: addr,
			Name:       name,
			Version:    "v1.0",
		},
		PeerList: make(map[uuid.UUID]*Peer),
	}

	log.WithFields(log.Fields{
		"id":   s.Id,
		"name": s.Name,
		"port": s.ListenAddr,
	}).Info("New Server")
	return
}

func (s *Server) StartPeer() {
	// Start server
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		panic(err)
	}
	s.listener = l

	go s.ListenLoop()
}

func (s *Server) ListenLoop() {
	log.WithFields(log.Fields{
		"port":     s.ListenAddr,
		"listener": s.listener.Addr().String(),
	}).Info("Listening for new Connections")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error in connection loop %s\n", err)
		}

		// Start goroutine to handle connection lifetime
		log.WithFields(log.Fields{
			"remoteAddr": conn.RemoteAddr(),
			"localAddr":  conn.LocalAddr(),
		}).Info("Received connection: ")

		go s.handshake(conn)
	}
}

func (s *Server) handshake(conn net.Conn) {

	// Server sends self
	// Peer accepts or reject
	// if OK : Add peer to peerlist and handle conn
	// if NOT close conn
	enc := json.NewEncoder(conn)
	enc.Encode(&s.Peer)

	dec := json.NewDecoder(conn)
	p := &Peer{}
	dec.Decode(&Peer{})
	if p.Version == s.Version {
		go s.handlePeer(conn)
	} else {

		log.WithFields(log.Fields{
			"peer_id":        p.Id,
			"peer_version":   p.Version,
			"server_version": s.Version,
		}).Info("Invalid Peer")
		conn.Close()
	}

}

func (s *Server) handlePeer(conn net.Conn) {

	buff := make([]byte, 1024)
	for {
		if _, err := conn.Read(buff); err != nil {

			log.Infoln("Connection closed")
			break
		}
		log.Infoln("%v", string(buff))
		conn.Write([]byte("Hi back!\n"))

	}
	conn.Close()
}
