package message

import (
	"encoding/hex"
	"fmt"
	"github.com/annchain/BlockDB/processors"
)

type QueryMessage struct {
	header *MessageHeader

	flag       string
	collection string
	skip       int32
	limit      int32
	query      string
	fields     string
}

func NewQueryMessage(header *MessageHeader, b []byte) *QueryMessage {

	// TODO not implemented yet.

	fmt.Println(hex.Dump(b))
	return nil
}

func (m *QueryMessage) ParseCommand() []*processors.LogEvent {

	return nil
}
