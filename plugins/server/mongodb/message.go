package mongodb

import (
	"github.com/annchain/BlockDB/processors"
	"net"
)

type MongoMessage interface {
	WriteTo(net.Conn) error
	ParseCommand() []*processors.LogEvent
}

type RequestMessage struct {
	host string
	op int
	payload []byte
}

const (
	opReply  = 1
	opUpdate = 2001
	opInsert = 2002
	reserved = 2003
	opQuery = 2004
	opGetMore = 2005
	opDelete = 2006
	opKillCursor = 2007
	opCommand = 2010
	opCommandReply = 2011
	opMsg = 2013
)

func (m *RequestMessage) Read() bool {
	if m.op == opQuery || m.op == opGetMore || m.op == opKillCursor {
		return true
	}
	if m.op == opUpdate || m.op == opInsert || m.op == opDelete {
		return false
	}
	// TODO more options need to be considered.
	return false
}

func (m *RequestMessage) Decode(bytes []byte) error {
	// TODO decode bytes to MongoMessage

	return nil
}

func (m *RequestMessage) ParseCommand() []*processors.LogEvent {
	// TODO parse mongo message to processor log events.

	return nil
}

func (m *RequestMessage) WriteTo(c net.Conn) error {
	// TODO write msg to connection

	return nil
}

type ResponseMessage struct {
//	TODO
}

func (m *ResponseMessage) ReadFromMongo(c net.Conn) error {
	// TODO read response from mongodb connection.

	return nil
}

func (m *ResponseMessage) WriteTo(c net.Conn) error {
	// TODO write msg to connection

	return nil
}

func (m *ResponseMessage) ParseCommand() []*processors.LogEvent {
	// TODO parse mongo response message to processor log events.

	return nil
}