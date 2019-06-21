package message

import (
	"fmt"
	"time"

	"github.com/annchain/BlockDB/processors"
)

const (
	HeaderLen = 16
)

type Message struct {
	Sender    string
	DBUser    string
	TimeStamp time.Time
	MongoMsg  MongoMessage
}

type MongoMessage interface {
	//WriteTo(net.Conn) error
	ParseCommand() []*processors.LogEvent
}

func (m *Message) ParseCommand() []*processors.LogEvent {
	// TODO
	return nil
}

type MessageHeader struct {
	MessageSize int32
	RequestID   int32
	ResponseTo  int32
	OpCode      OpCode
}

// readCString read collection full name from byte, starting at pos.
// Return the collection full name in string and the length of full name.
func readCString(b []byte, pos int) (string, int, error) {
	index := -1
	for i := pos; i < len(b)-pos; i++ {
		if b[i] == byte(0) {
			index = i
			break
		}
	}
	if index < 0 {
		return "", 0, fmt.Errorf("cannot read full collection name from bytes: %x", b)
	}

	cBytes := b[pos : index+1]
	s := ""
	for len(cBytes) > 0 {
		s = s + string(cBytes[0])
		cBytes = cBytes[1:]
	}
	fmt.Println("collection full name: ", s)

	return s, index - pos + 1, nil
}

// isFlagSet checks flag status. Return true when it is on, otherwise false.
func isFlagSet(b []byte, pos int, flagPos int) bool {
	// flag must in [0, 31]
	if flagPos > 31 || flagPos < 0 {
		return false
	}

	offset := flagPos / 8
	left := uint(flagPos - offset*8)
	p := b[pos+offset]
	p = p << (8 - (left + 1))
	p = p >> (left)
	if p == 0 {
		return false
	}
	return true
}
