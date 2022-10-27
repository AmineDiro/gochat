package main

import (
	"github.com/aminediro/gochat/peer"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	p := peer.MkServer(":4000", "Peer1")
	p.StartPeer()

	select {}
}
