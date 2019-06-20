package message

import "github.com/annchain/BlockDB/processors"

type DeleteMessage struct {
	header MessageHeader

	// TODO body not implemented
}

func NewDeleteMessage(header *MessageHeader, b []byte) *DeleteMessage {

	return nil
}

func (m *DeleteMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
