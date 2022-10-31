package peer

import (
	"fmt"
	"time"

	"github.com/aminediro/gochat/chat"
)

type PeerTerminal struct {
	Tx chan<- *chat.Message
	Rx <-chan *chat.Message
}

func (t *PeerTerminal) Input() error {
	var payload chat.Payload

	fmt.Printf(">>> ")
	if _, err := fmt.Scanln(&payload); err != nil {
		return err
	}

	time.Now().Unix()
	t.Tx <- &chat.Message{
		Header:  chat.Header{Timestamp: time.Now().Unix()},
		Payload: payload}
	return nil
}

func (t *PeerTerminal) Output() error {
	for msg := range t.Rx {
		if _, err := fmt.Printf("[%s] %s\n", msg.Header.SenderName, msg.Payload); err != nil {
			return err
		}
	}
	return nil
}

func (t *PeerTerminal) Start() {

	go func() {
		for {
			if err := t.Input(); err != nil {
				panic("input error")
			}
		}
	}()

	go t.Output()
}

func MkTerminal(tx chan<- *chat.Message, rx <-chan *chat.Message) *PeerTerminal {
	return &PeerTerminal{
		Tx: tx,
		Rx: rx,
	}

}
