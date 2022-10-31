package chat

type Header string
type Payload string

type Message struct {
	Header  Header
	Payload Payload
}

type Terminal interface {
	Input() (*Message, error)
	Output(*Message) error
}
