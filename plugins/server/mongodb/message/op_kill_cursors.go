package message

import "github.com/annchain/BlockDB/processors"

type KillCursorsMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewKillCursorsMessage(header *MessageHeader, b []byte) *KillCursorsMessage {

	return nil
}

func (m *KillCursorsMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
