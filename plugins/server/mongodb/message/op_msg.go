package message

import (
	"fmt"

	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
)

type MsgMessage struct {
	Header   *MessageHeader `json:"header"`
	Flags    msgFlags       `json:"flags"`
	Sections []section      `json:"sections"`
	CheckSum uint32         `json:"check_sum"`
}

func NewMsgMessage(header *MessageHeader, b []byte) (*MsgMessage, error) {

	// fmt.Println(fmt.Sprintf("new msg data: %x", b))

	p := make([]byte, len(b))
	copy(p, b)

	pos := HeaderLen

	// read flags
	flags := newMsgFlags(p, pos)
	pos += 4

	// read sections
	sectionsBytes := int(header.MessageSize) - pos
	if flags.CheckSumPresent {
		// reduce the length of checkSum
		sectionsBytes -= 4
	}
	var secs []section
	for sectionsBytes > 0 {
		sec, offset, err := newSection(p, pos)
		if err != nil {
			return nil, err
		}
		secs = append(secs, sec)
		pos += offset
		sectionsBytes -= offset
	}

	// read check sum
	checkSum := uint32(0)
	if flags.CheckSumPresent {
		checkSum = bytes.GetUInt32(p, pos)
	}

	mm := &MsgMessage{}
	mm.Header = header
	mm.Flags = flags
	mm.Sections = secs
	mm.CheckSum = checkSum

	return mm, nil
}

func (mm *MsgMessage) ExtractBasic() (user, db, collection, op, docId string) {
	if len(mm.Sections) <= 0 {
		return
	}
	secBody, ok := mm.Sections[0].(*sectionBody)
	if !ok {
		return
	}
	doc := secBody.Document.Map()
	// user
	if v, ok := doc["saslSupportedMechs"]; ok {
		user = v.(string)
	}
	// db
	if v, ok := doc["$db"]; ok {
		db = v.(string)
	}
	// op and collection
	if v, ok := doc["update"]; ok {
		op = "update"
		collection = v.(string)
	} else if v, ok := doc["insert"]; ok {
		op = "insert"
		collection = v.(string)
	} else if v, ok := doc["query"]; ok {
		op = "query"
		collection = v.(string)
	} else if v, ok := doc["delete"]; ok {
		op = "delete"
		collection = v.(string)
	}
	// doc id

	//if len(mm.Sections) <= 1 {
	//	return
	//}
	//secSeq, ok := mm.Sections[1].(*sectionDocumentSequence)
	//if !ok {
	//	return
	//}
	//
	return
}

type msgFlags struct {
	CheckSumPresent bool `json:"check_sum"`
	MoreToCome      bool `json:"more_to_come"`
	ExhaustAllowed  bool `json:"exhaust_allowed"`
}

func newMsgFlags(b []byte, pos int) msgFlags {

	flag := msgFlags{
		CheckSumPresent: isFlagSetUInt32(b, pos, 0),
		MoreToCome:      isFlagSetUInt32(b, pos, 1),
		ExhaustAllowed:  isFlagSetUInt32(b, pos, 16),
	}
	return flag
}

type section interface {
	kind() sectionType
}

func newSection(b []byte, pos int) (section, int, error) {
	if len(b) < pos+1 {
		return nil, 0, fmt.Errorf("document too small for section type")
	}

	sType := sectionType(b[pos])
	pos++

	switch sType {
	case singleDocument:
		doc, size, err := readDocument(b, pos)
		if err != nil {
			return nil, 0, err
		}
		s := &sectionBody{
			PayloadType: singleDocument,
			Document:    doc,
		}
		return s, 1 + int(size), nil

	case documentSequence:
		// read doc sequence size
		if len(b) < 4 {
			return nil, 0, fmt.Errorf("document too small for docSeq size")
		}
		size := bytes.GetUInt32(b, pos)
		if len(b) < pos+int(size) {
			return nil, 0, fmt.Errorf("document too small for docSeq")
		}
		pos += 4

		// read identifier
		identifier, idSize, err := readCString(b, pos)
		if err != nil {
			return nil, 0, fmt.Errorf("read cstring error: %v", err)
		}
		pos += idSize

		// read documents
		var docs []bson.D
		bytesLeft := int(size) - 4 - idSize
		for bytesLeft > 0 {
			doc, docSize, err := readDocument(b, pos)
			if err != nil {
				return nil, 0, fmt.Errorf("read doc error: %v", err)
			}
			docs = append(docs, doc)
			bytesLeft -= docSize
		}

		s := &sectionDocumentSequence{
			PayloadType: documentSequence,
			Size:        size,
			Identifier:  identifier,
			Documents:   docs,
		}
		return s, 1 + int(size), nil

	default:
		return nil, 0, fmt.Errorf("unknown section type: %v", sType)
	}

}

type sectionBody struct {
	PayloadType sectionType `json:"type"`
	Document    bson.D      `json:"document"`
}

func (s *sectionBody) kind() sectionType {
	return s.PayloadType
}

type sectionDocumentSequence struct {
	PayloadType sectionType `json:"type"`
	Size        uint32      `json:"size"`
	Identifier  string      `json:"identifier"`
	Documents   []bson.D    `json:"documents"`
}

func (s *sectionDocumentSequence) kind() sectionType {
	return s.PayloadType
}

type sectionType byte

const (
	singleDocument sectionType = iota
	documentSequence
)

//func (m *MsgMessage) ParseCommand() []*processors.LogEvent {
//
//	return nil
//}
