package mongodb

import (
	"bufio"
	"net"

	"github.com/annchain/BlockDB/common/bytes"
	"github.com/annchain/BlockDB/processors"
)

type MongoMessage interface {
	WriteTo(net.Conn) error
	ParseCommand() []*processors.LogEvent
}

type RequestMessage struct {
	host    string
	op      int32
	payload []byte
}

const (
	opReply        int32 = 1
	opUpdate       int32 = 2001
	opInsert       int32 = 2002
	reserved       int32 = 2003
	opQuery        int32 = 2004
	opGetMore      int32 = 2005
	opDelete       int32 = 2006
	opKillCursor   int32 = 2007
	opCommand      int32 = 2010
	opCommandReply int32 = 2011
	opMsg          int32 = 2013
)

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
	m.op = bytes.GetInt32(b, 4)
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
