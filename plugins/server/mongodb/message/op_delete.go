package message

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type DeleteMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewDeleteMessage(header *MessageHeader, b []byte) (*DeleteMessage, error) {

	fmt.Println("new delete data: ", b)
	return nil, nil
}

func (m *DeleteMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
