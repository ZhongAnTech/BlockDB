package message

import "github.com/annchain/BlockDB/processors"

type ReservedMessage struct {
	header MessageHeader

	// TODO body not implemented
}

func NewReservedMessage(header *MessageHeader, b []byte) *ReservedMessage {

	return nil
}

func (m *ReservedMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
