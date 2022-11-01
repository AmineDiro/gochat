package peer

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/aminediro/gochat/chat"
	"github.com/google/uuid"
)

type Terminal interface {
	Input() error
	Output() error
}
type PeerTerminal struct {
	peer *Peer
	Tx   chan<- *chat.Message
	Rx   <-chan *chat.Message
}

func MkTerminal(p *Peer, tx chan<- *chat.Message, rx <-chan *chat.Message) *PeerTerminal {
	return &PeerTerminal{
		peer: p,
		Tx:   tx,
		Rx:   rx,
	}

}
func (t *PeerTerminal) Input() error {
	var payload string

	in := bufio.NewReader(os.Stdin)

	payload, _ = in.ReadString('\n')
	msgID, _ := uuid.NewUUID()
	t.Tx <- &chat.Message{
		MsgID:      msgID,
		SenderID:   t.peer.Id,
		SenderName: t.peer.Name,
		Timestamp:  time.Now().Unix(),
		Payload:    payload}
	fmt.Printf("[%s] ", t.peer.Name)
	return nil
}

func (t *PeerTerminal) Output() error {
	for msg := range t.Rx {
		if _, err := fmt.Printf("\n[%s] [%s] %s", t.peer.Name, msg.SenderName, msg.Payload); err != nil {
			return err
		}
		fmt.Printf("[%s] ", t.peer.Name)
	}
	return nil
}

func (t *PeerTerminal) Start() {

	fmt.Printf("[%s] ", t.peer.Name)
	go func() {
		for {
			if err := t.Input(); err != nil {
				fmt.Println("inputerror", err)
			}
		}
	}()

	go t.Output()
}
