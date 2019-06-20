package message

import (
	"bufio"
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

type MessageHeader struct {
	MessageSize int32
	RequestID   int32
	ResponseTo  int32
	OpCode      OpCode
}

func (m *Message) ParseCommand() []*processors.LogEvent {
	// TODO
	return nil
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
