package message

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type MsgMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewMsgMessage(header *MessageHeader, b []byte) (*MsgMessage, error) {

	fmt.Println("new msg data: ", b)
	return nil, nil
}

func (m *MsgMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
