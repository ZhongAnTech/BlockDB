package message

import "github.com/annchain/BlockDB/processors"

type UpdateMessage struct {
	header     MessageHeader
	collection string
	selector   string
	update     string
}

func NewUpdateMessage(header *MessageHeader, b []byte) *UpdateMessage {

	return nil
}

func (m *UpdateMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
