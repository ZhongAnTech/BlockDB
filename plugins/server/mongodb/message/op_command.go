package message

import "github.com/annchain/BlockDB/processors"

type CommandMessage struct {
	header MessageHeader

	// TODO body not implemented
}

func NewCommandMessage(header *MessageHeader, b []byte) *CommandMessage {

	return nil
}

func (m *CommandMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
