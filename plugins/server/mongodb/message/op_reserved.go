package message

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type ReservedMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewReservedMessage(header *MessageHeader, b []byte) (*ReservedMessage, error) {

	fmt.Println("new reserved data: ", b)

	return nil, nil
}

func (m *ReservedMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
