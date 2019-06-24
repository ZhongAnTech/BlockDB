package message

import (
	"github.com/annchain/BlockDB/processors"
)

type CommandReplyMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewCommandReplyMessage(header *MessageHeader, b []byte) (*CommandReplyMessage, error) {

	//fmt.Println("new command reply data: ", b)
	return nil, nil
}

func (m *CommandReplyMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
