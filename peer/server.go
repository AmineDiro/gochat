package peer

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aminediro/gochat/chat"
	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Peer struct {
	Id         uuid.UUID `json:"uuid"`
	Name       string    `json:"name"`
	ListenAddr string    `json:"listenAddr"`
	Version    string    `json:"version"`
}

func (p *Peer) String() string {

	return fmt.Sprintf("Name[%s]", p.Name)
}

type Server struct {
	Peer
	listener net.Listener
	verbose  bool
	cx       chan net.Conn
	Tx       chan *chat.Message
	Rx       chan *chat.Message

	mu          sync.Mutex
	PeerList    []*Peer
	lenPeerList uint32
	connMap     map[uuid.UUID]*net.Conn
}

func MkServer(c ServerConfig) (s *Server) {
	s = &Server{
		Peer: Peer{
			Id:         uuid.New(),
			ListenAddr: c.ListenAddr,
			Name:       c.Name,
			Version:    c.Version,
		},
		verbose:  c.Verbose,
		PeerList: []*Peer{},
		connMap:  make(map[uuid.UUID]*net.Conn),
		cx:       make(chan net.Conn),
		Rx:       make(chan *chat.Message),
		Tx:       make(chan *chat.Message),
	}

	log.WithFields(log.Fields{
		"id":   s.Id,
		"name": s.Name,
		"port": s.ListenAddr,
	}).Debug("New Server")
	return
}

func (s *Server) StartPeer() {
	// Start server
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		panic(err)
	}
	s.listener = l

	go s.listenLoop()
	go s.commLoop()
	go s.serverProxy()
	if s.verbose {
		go s.printStatus()
	}
}

func (s *Server) printStatus() {
	ticker := time.NewTicker(time.Duration(1) * time.Second)

	for range ticker.C {
		log.WithFields(log.Fields{
			"Id":             s.Id.String()[:5],
			"PeerName":       s.Name,
			"Version":        s.Version,
			"Addr":           s.ListenAddr,
			"ConnectedPeers": s.lenPeerList,
			"NbConnections":  len(s.connMap),
		}).Info("Peer Status")
	}
}

func (s *Server) listenLoop() {
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

		s.cx <- conn

	}
}

func (s *Server) serverProxy() {
	for conn := range s.cx {
		if err := s.initProtocol(conn); err != nil {
			log.Info("Connection close.", err)
		}
	}

}
func (s *Server) commLoop() {
	for msg := range s.Rx {
		go s.broadcastMsg(msg)
	}
}

func (s *Server) broadcastMsg(msg *chat.Message) {
	for rid, conn := range s.connMap {
		err := chat.SendMessage(*conn, msg)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"msgID":      msg.MsgID.String(),
				"senderId":   msg.SenderID.String(),
				"receiverId": rid.String(),
			}).Errorln("Error while sending")
		}
	}
}

func (s *Server) initProtocol(conn net.Conn) error {
	// Authorize Peer
	p, err := s.authorizePeer(conn)
	if err != nil {
		conn.Close()
		return err
	}

	// If OK Add peer
	if err := s.addPeer(conn, p); err != nil {
		conn.Close()
		return err
	}

	// If OK Broadcast internal PeerList
	if err := s.broadcastPeerList(conn, p); err != nil {

		conn.Close()
		return err
	}

	go s.chatPeer(conn)
	return nil
}

func (s *Server) addPeer(conn net.Conn, p *Peer) error {
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
	atomic.AddUint32(&s.lenPeerList, 1)
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

func (s *Server) broadcastPeerList(conn net.Conn, newP *Peer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Send the Server peerlist
	listAddr := []string{}
	for _, p := range s.PeerList {
		if !reflect.DeepEqual(p.Id, newP.Id) {
			listAddr = append(listAddr, p.ListenAddr)
		}

	}
	// Get connection
	return SendPeerList(conn, listAddr)
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

func (s *Server) Connect(addrs ...string) error {
	for _, addr := range addrs {
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

		if err := s.addPeer(conn, p); err != nil {
			log.Errorln("Add Peer failed : ", err)
		}

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

		go s.Connect(peerList...)

		go s.chatPeer(conn)
	}

	return nil
}

func (s *Server) chatPeer(conn net.Conn) {

	for {
		msg, err := chat.ReceiveMessage(conn)
		if err != nil {
			//TODO : drop peer form peerlist
			break
		}
		s.Tx <- &msg
	}
	conn.Close()
}
