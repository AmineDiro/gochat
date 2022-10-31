package peer

import (
	"fmt"

	"github.com/aminediro/gochat/chat"
)

type PeerTerminal struct{}

func (t *PeerTerminal) Input(tx chan<- *chat.Message) error {

	var payload chat.Payload
	if _, err := fmt.Scanln(&payload); err != nil {
		return err
	}

	tx <- &chat.Message{Payload: payload}
	return nil
}
