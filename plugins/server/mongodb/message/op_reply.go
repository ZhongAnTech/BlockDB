package message

import (
	"github.com/annchain/BlockDB/processors"
)

type ReplyMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewReplyMessage(header *MessageHeader, b []byte) (*ReplyMessage, error) {

	//fmt.Println("new reply data: ", b)

	return nil, nil
}

func (m *ReplyMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
