package message

import "github.com/annchain/BlockDB/processors"

type ReplyMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewReplyMessage(header *MessageHeader, b []byte) *ReplyMessage {

	return nil
}

func (m *ReplyMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
