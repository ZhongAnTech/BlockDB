package message

import (
	"encoding/json"
	"fmt"

	"github.com/annchain/BlockDB/common/bytes"
	"github.com/annchain/BlockDB/processors"
	"github.com/globalsign/mgo/bson"
)

type QueryMessage struct {
	header *MessageHeader

	flags      queryFlags
	collection string
	skip       int32
	limit      int32
	query      string
	fields     string
}

func NewQueryMessage(header *MessageHeader, b []byte) *QueryMessage {

	// TODO handle errors. Be aware of fatal messages from client.
	p := make([]byte, len(b))
	copy(p, b)

	fmt.Println(p)

	p = p[HeaderLen:]

	// read flags
	flags := newQueryFlags(p, 0)
	p = p[4:]

	// read collection full name
	coll, collLen, _ := readCString(p, 0)
	p = p[collLen:]

	skip := bytes.GetInt32(p, 0)
	limit := bytes.GetInt32(p, 4)
	p = p[8:]

	// read query document
	docSize := bytes.GetInt32(p, 0)
	docBytes := p[:docSize]

	var docBson bson.D
	bson.Unmarshal(docBytes, &docBson)
	doc, _ := json.Marshal(docBson.Map())

	// read fields
	// TODO fields needed.

	qm := &QueryMessage{}
	qm.header = header
	qm.flags = flags
	qm.collection = coll
	qm.skip = skip
	qm.limit = limit
	qm.query = string(doc)

	return qm
}

func (m *QueryMessage) ParseCommand() []*processors.LogEvent {

	return nil
}

type queryFlags struct {
	Reserved        bool
	TailableCursor  bool
	SlaveOk         bool
	OplogReplay     bool
	NoCursorTimeout bool
	AwaitData       bool
	Exhaust         bool
	Partial         bool
}

func newQueryFlags(b []byte, pos int) queryFlags {
	q := queryFlags{
		Reserved:        isFlagSet(b, pos, 0),
		TailableCursor:  isFlagSet(b, pos, 1),
		SlaveOk:         isFlagSet(b, pos, 2),
		OplogReplay:     isFlagSet(b, pos, 3),
		NoCursorTimeout: isFlagSet(b, pos, 4),
		AwaitData:       isFlagSet(b, pos, 5),
		Exhaust:         isFlagSet(b, pos, 6),
		Partial:         isFlagSet(b, pos, 7),
	}
	return q
}
