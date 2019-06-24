package message

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type InsertMessage struct {
	header *MessageHeader

	// TODO body not implemented
}

func NewInsertMessage(header *MessageHeader, b []byte) (*InsertMessage, error) {

	fmt.Println("new insert data: ", b)

	return nil, nil
}

func (m *InsertMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
