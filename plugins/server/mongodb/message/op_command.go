package message

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type CommandMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewCommandMessage(header *MessageHeader, b []byte) (*CommandMessage, error) {

	fmt.Println("new command data: ", b)
	return nil, nil
}

func (m *CommandMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
