package main

import (
	"github.com/aminediro/gochat/peer"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	peer1 := peer.MkServer(":3000", "Peer1", "v1.0")
	peer2 := peer.MkServer(":4000", "Peer2", "v1.0")
	peer1.StartPeer()
	peer2.StartPeer()

	peer2.Connect(peer1.ListenAddr)

	select {}
}
