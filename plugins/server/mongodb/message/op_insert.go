package message

import "github.com/annchain/BlockDB/processors"

type InsertMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewInsertMessage(header *MessageHeader, b []byte) *InsertMessage {

	return nil
}

func (m *InsertMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
