package message

import (
	"encoding/json"
	"fmt"

	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
)

type QueryMessage struct {
	Header *MessageHeader

	Flags      queryFlags `json:"flags"`
	Collection string     `json:"collection"`
	Skip       int32      `json:"skip"`
	Limit      int32      `json:"limit"`
	Query      string     `json:"query"`
	Fields     string     `json:"fields"`
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
	qm.Header = header
	qm.Flags = flags
	qm.Collection = coll
	qm.Skip = skip
	qm.Limit = limit
	qm.Query = string(doc)

	return qm
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
