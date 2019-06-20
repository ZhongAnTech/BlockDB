package message

import "github.com/annchain/BlockDB/processors"

type UpdateMessage struct {
	header     *MessageHeader
	collection string
	selector   string
	update     string
}

func NewUpdateMessage(header *MessageHeader, b []byte) *UpdateMessage {

	m := &UpdateMessage{}
	m.header = header

	m.collection, _ = readCString(b, HeaderLen+4)

	return nil
}

func (m *UpdateMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
