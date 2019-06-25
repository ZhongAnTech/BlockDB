package message

import (
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
)

type QueryMessage struct {
	Header *MessageHeader

	Flags  queryFlags `json:"flags"`
	coll   string     `json:"collection"`
	Skip   int32      `json:"skip"`
	Limit  int32      `json:"limit"`
	Query  bson.D     `json:"query"`
	Fields string     `json:"fields"`
}

func NewQueryMessage(header *MessageHeader, b []byte) (*QueryMessage, error) {

	//fmt.Println("new query data: ", b)

	p := make([]byte, len(b))
	copy(p, b)

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
	docSize := bytes.GetUInt32(p, 0)
	docBytes := p[:docSize]

	var docBson bson.D
	err := bson.Unmarshal(docBytes, &docBson)
	if err != nil {
		return nil, fmt.Errorf("read query document error, cannot unmarshal it to bson, err: %v", err)
	}

	// read fields
	// TODO fields needed.

	qm := &QueryMessage{}
	qm.Header = header
	qm.Flags = flags
	qm.coll = coll
	qm.Skip = skip
	qm.Limit = limit
	qm.Query = docBson

	return qm, nil
}

func (qm *QueryMessage) ExtractBasic() (user, db, collection, op, docId string) {
	return
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
		Reserved:        isFlagSetInt32(b, pos, 0),
		TailableCursor:  isFlagSetInt32(b, pos, 1),
		SlaveOk:         isFlagSetInt32(b, pos, 2),
		OplogReplay:     isFlagSetInt32(b, pos, 3),
		NoCursorTimeout: isFlagSetInt32(b, pos, 4),
		AwaitData:       isFlagSetInt32(b, pos, 5),
		Exhaust:         isFlagSetInt32(b, pos, 6),
		Partial:         isFlagSetInt32(b, pos, 7),
	}
	return q
}
