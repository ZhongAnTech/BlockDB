package message

import (
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
)

type QueryMessage struct {
	Header *MessageHeader `json:"header"`
	Flags  queryFlags     `json:"flags"`
	Coll   string         `json:"collection"`
	Skip   int32          `json:"skip"`
	Limit  int32          `json:"limit"`
	Query  bson.M         `json:"query"`
	Fields bson.M         `json:"fields"`
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
	querySize := bytes.GetUInt32(p, 0)
	queryBytes := p[:querySize]

	var queryBson bson.D
	err := bson.Unmarshal(queryBytes, &queryBson)
	if err != nil {
		return nil, fmt.Errorf("read query document error, cannot unmarshal it to bson, err: %v", err)
	}
	p = p[querySize:]

	// read fields
	var fieldsBson bson.D
	if len(p) > 0 {
		fieldsSize := bytes.GetUInt32(p, 0)
		fieldsBytes := p[:fieldsSize]
		err = bson.Unmarshal(fieldsBytes, &fieldsBson)
		if err != nil {
			return nil, fmt.Errorf("read fields document error, cannot unmarshal it to bson, err: %v", err)
		}
	}

	qm := &QueryMessage{}
	qm.Header = header
	qm.Flags = flags
	qm.Coll = coll
	qm.Skip = skip
	qm.Limit = limit
	qm.Query = queryBson.Map()
	if fieldsBson != nil {
		qm.Fields = fieldsBson.Map()
	}

	return qm, nil
}

func (qm *QueryMessage) ExtractBasic() (user, db, collection, op, docId string) {
	// TODO

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

func (qf *queryFlags) MarshalJSON() ([]byte, error) {
	r := map[string]bool{}
	if qf.Reserved {
		r["reserved"] = true
	}
	if qf.TailableCursor {
		r["tailable_cursor"] = true
	}
	if qf.SlaveOk {
		r["slave_ok"] = true
	}
	if qf.OplogReplay {
		r["log_reply"] = true
	}
	if qf.NoCursorTimeout {
		r["no_cursor_timeout"] = true
	}
	if qf.AwaitData {
		r["await_data"] = true
	}
	if qf.Exhaust {
		r["exhaust"] = true
	}
	if qf.Partial {
		r["partial"] = true
	}
	return json.Marshal(r)
}
