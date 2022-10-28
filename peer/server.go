package peer

import (
	"fmt"
	"net"
	"sync"
	"time"

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
	mu       sync.RWMutex
	PeerList map[uuid.UUID]*Peer
}

func MkServer(addr string, name string, version string) (s *Server) {
	s = &Server{
		Peer: Peer{
			Id:         uuid.New(),
			ListenAddr: addr,
			Name:       name,
			Version:    version,
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

	// go s.status()
}

func (s *Server) status() {
	ticker := time.NewTicker(time.Duration(2) * time.Second)

	for range ticker.C {
		s.mu.RLock()
		n := len(s.PeerList)
		s.mu.RUnlock()
		log.WithFields(log.Fields{
			"Id":             s.Id.String()[:5],
			"ConnectedPeers": n,
		}).Info("Peer Status")
	}
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

		go s.authorizePeer(conn)
	}
}

func (s *Server) authorizePeer(conn net.Conn) error {
	// Block and wait for response
	p, err := ReceiveHandshake(conn)
	if err != nil {
		return err
	}
	if p.Version == s.Version {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.PeerList[p.Id] = p
		if err := SendHandshake(conn, &s.Peer); err != nil {
			return err
		}
		go s.handlePeer(conn)
	} else {
		log.WithFields(log.Fields{
			"peer_id":        p.Id,
			"peer_version":   p.Version,
			"server_version": s.Version,
		}).Errorln("Invalid Peer")
		conn.Close()
	}
	return nil
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

func (s *Server) Connect(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second*1)
	if err != nil {
		return err
	}
	SendHandshake(conn, &s.Peer)

	//TODO :
	// OK from server
	// Not ok from server conn closed
	p, err := ReceiveHandshake(conn)
	if err != nil {
		conn.Close()
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PeerList[p.Id] = p

	return nil
}
