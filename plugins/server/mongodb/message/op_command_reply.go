package message

import "github.com/annchain/BlockDB/processors"

type CommandReplyMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewCommandReplyMessage(header *MessageHeader, b []byte) *CommandReplyMessage {

	return nil
}

func (m *CommandReplyMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
