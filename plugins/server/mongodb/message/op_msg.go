package message

import "github.com/annchain/BlockDB/processors"

type MsgMessage struct {
	header MessageHeader

	// TODO body not implemented
}

func NewMsgMessage(header *MessageHeader, b []byte) *MsgMessage {

	return nil
}

func (m *MsgMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
