package main

import (
	"time"

	"github.com/aminediro/gochat/peer"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func startPeers() {

}

func main() {
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		// FullTimestamp: true,
	})

	peer1 := peer.MkServer(":3100", "Peer1", "v1.0")
	peer2 := peer.MkServer(":3200", "Peer2", "v1.0")
	peer3 := peer.MkServer(":3300", "Peer3", "v1.0")
	peer4 := peer.MkServer(":3400", "Peer4", "v1.0")

	peer1.StartPeer()
	peer2.StartPeer()
	peer3.StartPeer()
	peer4.StartPeer()

	peer2.Connect(peer1.ListenAddr)
	time.Sleep(time.Millisecond * 100)
	peer3.Connect(peer1.ListenAddr)
	time.Sleep(time.Millisecond * 100)
	peer4.Connect(peer2.ListenAddr)
	select {}
}
