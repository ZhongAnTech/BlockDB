package message

import "github.com/annchain/BlockDB/processors"

type UpdateMessage struct {
	header     *MessageHeader
	collection string
	flags      string
	selector   string
	update     string
}

func NewUpdateMessage(header *MessageHeader, b []byte) *UpdateMessage {

	b = b[HeaderLen+4:]
	coll, collLen, _ := readCString(b, 0)
	b = b[collLen:]

	m := &UpdateMessage{}
	m.header = header
	m.collection = coll

	// TODO extract selector and update parts

	return nil
}

func (m *UpdateMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
