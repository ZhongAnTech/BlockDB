package message

import "github.com/annchain/BlockDB/processors"

type GetMoreMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewGetMoreMessage(header *MessageHeader, b []byte) *GetMoreMessage {

	return nil
}

func (m *GetMoreMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
