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
	Version    string    `json:"version"`
}

func (p *Peer) String() string {

	return fmt.Sprintf("ID[%s] Addr[%s]", p.Id.String()[:4], p.ListenAddr)
}

type Server struct {
	Peer
	listener net.Listener

	mu       sync.Mutex
	PeerList []*Peer
	connMap  map[uuid.UUID]*net.Conn
	tx       chan *Peer
}

func MkServer(addr string, name string, version string) (s *Server) {
	s = &Server{
		Peer: Peer{
			Id:         uuid.New(),
			ListenAddr: addr,
			Name:       name,
			Version:    version,
		},
		PeerList: []*Peer{},
		connMap:  make(map[uuid.UUID]*net.Conn),
		tx:       make(chan *Peer),
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
	go s.broadcastPeerList()

	//TODO : add flag
	go s.status()
}

func (s *Server) status() {
	ticker := time.NewTicker(time.Duration(1) * time.Second)

	for range ticker.C {
		s.mu.Lock()
		n := len(s.PeerList)
		s.mu.Unlock()
		log.WithFields(log.Fields{
			"Id":             s.Id.String()[:5],
			"PeerName":       s.Name,
			"ConnectedPeers": n,
		}).Info("Peer Status")
	}
}

func (s *Server) ListenLoop() {
	log.WithFields(log.Fields{
		"listener": s.listener.Addr().String(),
	}).Debug("Listening for new Connections")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error in connection loop %s\n", err)
		}

		// Start goroutine to handle connection lifetime
		log.WithFields(log.Fields{
			"remoteAddr": conn.RemoteAddr(),
			"localAddr":  conn.LocalAddr(),
		}).Debug("Received connection: ")

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) error {

	// Authorize Peer
	p, err := s.authorizePeer(conn)
	if err != nil {
		conn.Close()
		return err
	}

	if err := s.AddPeer(conn, p); err != nil {
		panic("Error in adding Peer")
	}

	// Broadcast Internal PeerList to the connected peer
	s.tx <- p
	return nil
}

func (s *Server) AddPeer(conn net.Conn, p *Peer) error {
	log.WithFields(log.Fields{
		"server": s.ListenAddr,
		"peer":   p.ListenAddr,
	}).Debug("Adding Peer.")
	s.mu.Lock()
	if _, exists := s.connMap[p.Id]; exists {
		return nil
	}
	s.PeerList = append(s.PeerList, p)
	s.connMap[p.Id] = &conn
	s.mu.Unlock()
	return nil

}

func (s *Server) authorizePeer(conn net.Conn) (*Peer, error) {
	// Block and wait for response
	p, err := ReceiveHandshake(conn)
	if err != nil {
		return nil, err
	}
	if isAuthorized(s, p) {
		if err := SendHandshake(conn, &s.Peer); err != nil {
			return nil, err
		}
		return p, nil
	}

	log.WithFields(log.Fields{
		"peer_id":        p.Id,
		"peer_version":   p.Version,
		"server_version": s.Version,
	}).Errorf("Invalid Peer")

	return nil, fmt.Errorf("invalid Peer")
}

func (s *Server) broadcastPeerList() {
	for newPeer := range s.tx {
		s.mu.Lock()
		// Send the Server peerlist
		listAddr := []string{}
		for _, p := range s.PeerList {
			if p.ListenAddr != newPeer.ListenAddr {
				listAddr = append(listAddr, p.ListenAddr)
			}

		}
		conn := s.connMap[newPeer.Id]
		SendPeerList(*conn, listAddr)
		s.mu.Unlock()
	}

}

func (s *Server) lookupAddr(addr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.PeerList {
		if addr == p.ListenAddr {
			return true
		}
	}
	return false

}

func (s *Server) Connect(addr string) error {
	if s.lookupAddr(addr) {
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second*1)
	if err != nil {
		return err
	}

	SendHandshake(conn, &s.Peer)
	p, err := ReceiveHandshake(conn)
	if err != nil {
		log.Errorln("%s : Server Closed connection", s.Name)
		conn.Close()
		return err
	}

	s.AddPeer(conn, p)

	peerList, err := ReceivePeerList(conn)
	if err != nil {
		// TODO: dunno what would happen here?
		// Maybe ask later in protocol
		log.Errorln("%s : Server Closed connection", s.Name)
		return err
	}

	log.WithFields(log.Fields{
		"receiver":          s.Name,
		"sender":            addr,
		"received_peerlist": peerList,
		"memory_peerlist":   s.PeerList,
	}).Debug("PeerList")

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, addr := range peerList {
		go s.Connect(addr)
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
		time.Sleep(10 * time.Second)
		log.Infoln("%v", string(buff))
		conn.Write([]byte("Hi back!\n"))

	}
	conn.Close()
}
