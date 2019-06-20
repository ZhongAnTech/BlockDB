package message

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/annchain/BlockDB/common/bytes"
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

// codes below should be deleted.

type RequestMessage struct {
	host    string
	op      OpCode
	payload []byte
}

func (m *RequestMessage) ReadOnly() bool {
	if m.op == OpQuery || m.op == OpGetMore || m.op == OpKillCursors {
		return true
	}
	if m.op == OpUpdate || m.op == OpInsert || m.op == OpDelete {
		return false
	}
	// TODO more options need to be considered.
	return false
}

func (m *RequestMessage) Decode(b []byte) error {
	m.op = OpCode(bytes.GetInt32(b, 4))
	m.payload = b

	return nil
}

func (m *RequestMessage) ParseCommand() []*processors.LogEvent {
	// TODO parse mongo message to processor log events.

	return nil
}

func (m *RequestMessage) WriteTo(c net.Conn) error {
	// TODO write msg to connection

	_, err := c.Write(m.payload)
	return err
}

type ResponseMessage struct {
	payload []byte
}

func (m *ResponseMessage) ReadFromMongo(c net.Conn) error {
	// TODO read response from mongodb connection.

	reader := bufio.NewReader(c)

	header := make([]byte, HeaderLen)
	_, err := reader.Read(header)
	if err != nil {
		return err
	}

	msgSize := bytes.GetInt32(header, 0)
	if msgSize == HeaderLen {
		m.payload = header
		return nil
	}

	body := make([]byte, msgSize-HeaderLen)
	_, err = reader.Read(body)
	if err != nil {
		return err
	}

	m.payload = append(header, body...)
	return nil
}

func (m *ResponseMessage) WriteTo(c net.Conn) error {
	// TODO write msg to connection

	_, err := c.Write(m.payload)
	return err
}

func (m *ResponseMessage) ParseCommand() []*processors.LogEvent {
	// TODO parse mongo response message to processor log events.

	return nil
}
