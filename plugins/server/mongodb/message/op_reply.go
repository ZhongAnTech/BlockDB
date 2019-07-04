package message

import (
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
)

type ReplyMessage struct {
	Header    *MessageHeader `json:"header"`
	Flags     replyFlags     `json:"flags"`
	CursorID  int64          `json:"cursor_id"`
	StartFrom int32          `json:"start_from"`
	Number    int32          `json:"number"`
	Documents []bson.M       `json:"documents"`
}

func NewReplyMessage(header *MessageHeader, b []byte) (*ReplyMessage, error) {

	//fmt.Println("new reply data: ", b)

	p := make([]byte, len(b))
	copy(p, b)

	pos := HeaderLen

	// read flags
	flags := newReplyFlags(p, pos)
	pos += 4
	// read cursor id
	cursorID := bytes.GetInt64(p, pos)
	pos += 8
	// read start_from
	startFrom := bytes.GetInt32(p, pos)
	pos += 4
	// read number returned
	number := bytes.GetInt32(p, pos)
	pos += 4

	// read documents
	var docs []bson.M
	bytesLeft := int(header.MessageSize) - pos
	for bytesLeft > 0 {
		doc, docSize, err := readDocument(b, pos)
		if err != nil {
			return nil, fmt.Errorf("read doc error: %v", err)
		}
		docs = append(docs, doc.Map())
		bytesLeft -= docSize
	}

	rm := &ReplyMessage{}
	rm.Header = header
	rm.Flags = flags
	rm.CursorID = cursorID
	rm.StartFrom = startFrom
	rm.Number = number
	rm.Documents = docs

	return rm, nil
}

func (rm *ReplyMessage) ExtractBasic() (user, db, collection, op, docId string) {
	// TODO

	return
}

type replyFlags struct {
	CursorNotFound   bool `json:"cursor_not_found"`
	QueryFailure     bool `json:"query_failure"`
	ShardConfigStale bool `json:"shard_config_stale"`
	AwaitCapable     bool `json:"await_capable"`
}

func newReplyFlags(b []byte, pos int) replyFlags {
	return replyFlags{
		CursorNotFound:   isFlagSetInt32(b, pos, 0),
		QueryFailure:     isFlagSetInt32(b, pos, 1),
		ShardConfigStale: isFlagSetInt32(b, pos, 2),
		AwaitCapable:     isFlagSetInt32(b, pos, 3),
	}
}

func (rf *replyFlags) MarshalJSON() ([]byte, error) {
	r := map[string]bool{}
	if rf.CursorNotFound {
		r["cursor_not_found"] = true
	}
	if rf.QueryFailure {
		r["query_failure"] = true
	}
	if rf.ShardConfigStale {
		r["shard_config_stale"] = true
	}
	if rf.AwaitCapable {
		r["await_capable"] = true
	}
	return json.Marshal(r)
}
