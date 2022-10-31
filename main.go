package main

import (
	"flag"
	"fmt"

	"github.com/aminediro/gochat/chat"
	"github.com/aminediro/gochat/peer"
	log "github.com/sirupsen/logrus"
)

type Peers []string

func (i *Peers) String() string {
	return "List of Peers"
}

func (ps *Peers) Set(value string) error {
	*ps = append(*ps, value)
	return nil
}

func main() {

	var ps Peers
	addr := flag.String("port", "", "Address of the Peer")
	name := flag.String("name", "", "Peer name")
	version := flag.String("version", "1.0", "Protocol version")
	verbose := flag.Bool("v", false, "Protocol version")
	flag.Var(&ps, "peers", "Peer list to connect to")

	flag.Parse()

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		// FullTimestamp: true,
	})

	config := peer.ServerConfig{
		Name:       *name,
		ListenAddr: *addr,
		Version:    *version,
		Verbose:    *verbose,
	}
	// Start Peer Server
	s := peer.MkServer(config)
	s.StartPeer()
	s.Connect(ps...)

	//Start terminal
	rx := make(chan *chat.Message)
	tx := make(chan *chat.Message)
	t := peer.MkTerminal(tx, rx)
	t.Start()

	go func() {
		for msg := range tx {
			fmt.Printf("\n<<< Received %s\n", msg.Payload)

		}

	}()

	select {}
}
