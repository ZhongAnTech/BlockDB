package message

import "github.com/annchain/BlockDB/processors"

type QueryMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewQueryMessage(header *MessageHeader, b []byte) *QueryMessage {

	return nil
}

func (m *QueryMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
