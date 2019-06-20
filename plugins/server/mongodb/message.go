package mongodb

import (
	"bufio"
	"net"

	"github.com/annchain/BlockDB/common/bytes"
	"github.com/annchain/BlockDB/processors"
)


const (
	headerLen = 16
)

type OpCode int32

const (
	opReply        = OpCode(1)
	opUpdate       = OpCode(2001)
	opInsert       = OpCode(2002)
	reserved       = OpCode(2003)
	opQuery        = OpCode(2004)
	opGetMore      = OpCode(2005)
	opDelete       = OpCode(2006)
	opKillCursor   = OpCode(2007)
	opCommand      = OpCode(2010)
	opCommandReply = OpCode(2011)
	opMsg          = OpCode(2013)
)

type MongoMessage interface {
	WriteTo(net.Conn) error
	ParseCommand() []*processors.LogEvent
}

type RequestMessage struct {
	host    string
	op      OpCode
	payload []byte
}

func (m *RequestMessage) ReadOnly() bool {
	if m.op == opQuery || m.op == opGetMore || m.op == opKillCursor {
		return true
	}
	if m.op == opUpdate || m.op == opInsert || m.op == opDelete {
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

	header := make([]byte, headerLen)
	_, err := reader.Read(header)
	if err != nil {
		return err
	}

	msgSize := bytes.GetInt32(header, 0)
	if msgSize == headerLen {
		m.payload = header
		return nil
	}

	body := make([]byte, msgSize-headerLen)
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

type MessageHeader struct {
	MessageSize int32
	RequestID int32
	ResponseTo int32
	OpCode OpCode
}
