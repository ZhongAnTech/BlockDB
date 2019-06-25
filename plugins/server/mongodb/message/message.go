package message

import (
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"github.com/globalsign/mgo/bson"
	"strconv"
)

const (
	HeaderLen = 16
)

type Message struct {
	DBUser   string       `json:"db_user"`
	MongoMsg MongoMessage `json:"db_log"`
}

type MongoMessage interface {

	// ParseCommand parses mongo message to json string.
	// ParseCommand() string
}

// ParseCommand parses message to json string.
func (m *Message) ParseCommand() string {
	b, _ := json.Marshal(m)
	return string(b)
}

type MessageHeader struct {
	MessageSize uint32 `json:"size"`
	RequestID   uint32 `json:"req_id"`
	ResponseTo  uint32 `json:"resp_to"`
	OpCode      OpCode `json:"opcode"`
}

func DecodeHeader(b []byte) (*MessageHeader, error) {
	if len(b) < HeaderLen {
		return nil, fmt.Errorf("not enough length for header decoding, expect %d, get %d", HeaderLen, len(b))
	}

	m := &MessageHeader{
		MessageSize: bytes.GetUInt32(b, 0),
		RequestID:   bytes.GetUInt32(b, 4),
		ResponseTo:  bytes.GetUInt32(b, 8),
		OpCode:      OpCode(bytes.GetUInt32(b, 12)),
	}
	return m, nil
}

// readCString read collection full name from byte, starting at pos.
// Return the collection full name in string and the length of full name.
func readCString(b []byte, pos int) (string, int, error) {
	index := -1
	for i := pos; i < len(b); i++ {
		if b[i] == byte(0) {
			index = i
			break
		}
	}
	if index < 0 {
		return "", 0, fmt.Errorf("cannot read full collection name from bytes: %x", b)
	}

	cBytes := b[pos : index+1]
	s := ""
	for len(cBytes) > 0 {
		s = s + string(cBytes[0])
		cBytes = cBytes[1:]
	}

	return s, index - pos + 1, nil
}

// readDocument read a bson.Document from a byte array. The read start from "pos",
// returns a bson.Document and the size of the document in bytes. Or return error
// if the read meets any problems.
func readDocument(b []byte, pos int) (bson.D, int, error) {
	if len(b) < pos+4 {
		return nil, 0, fmt.Errorf("document too small for single size")
	}
	size := bytes.GetUInt32(b, pos)
	if len(b) < pos+int(size) {
		return nil, 0, fmt.Errorf("document too small for doc")
	}
	docB := b[pos : pos+int(size)]
	var doc bson.D
	err := bson.Unmarshal(docB, &doc)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot unmarshal it to bson, err: %v", err)
	}
	return doc, int(size), nil
}

// isFlagSetUint32 checks flag status. Return true when it is on, otherwise false.
func isFlagSetUInt32(b []byte, pos int, flagPos int) bool {
	// flag must in [0, 31]
	if flagPos > 31 || flagPos < 0 {
		return false
	}

	p := bytes.GetUInt32(b, pos)
	if p&FlagUIntSet[flagPos] > 0 {
		return true
	}
	return false
}

// isFlagInt32Set checks flag status. Return true when it is on, otherwise false.
func isFlagSetInt32(b []byte, pos int, flagPos int) bool {
	// flag must in [0, 31]
	if flagPos > 31 || flagPos < 0 {
		return false
	}

	p := bytes.GetInt32(b, pos)
	if p&FlagIntSet[flagPos] > 0 {
		return true
	}
	return false
}

func init() {

	// init all flag positions
	for i := 0; i < len(FlagUIntSet); i++ {
		ui32, _ := strconv.ParseUint(flagSetBinary[i], 2, 32)
		FlagUIntSet[i] = uint32(ui32)
		i32, _ := strconv.ParseInt(flagSetBinary[i], 2, 32)
		FlagIntSet[i] = int32(i32)
	}

}

var FlagUIntSet [32]uint32
var FlagIntSet [32]int32

var flagSetBinary [32]string = [32]string{
	"00000000" + "00000000" + "00000000" + "00000001",
	"00000000" + "00000000" + "00000000" + "00000010",
	"00000000" + "00000000" + "00000000" + "00000100",
	"00000000" + "00000000" + "00000000" + "00001000",
	"00000000" + "00000000" + "00000000" + "00010000",
	"00000000" + "00000000" + "00000000" + "00100000",
	"00000000" + "00000000" + "00000000" + "01000000",
	"00000000" + "00000000" + "00000000" + "10000000",

	"00000000" + "00000000" + "00000001" + "00000000",
	"00000000" + "00000000" + "00000010" + "00000000",
	"00000000" + "00000000" + "00000100" + "00000000",
	"00000000" + "00000000" + "00001000" + "00000000",
	"00000000" + "00000000" + "00010000" + "00000000",
	"00000000" + "00000000" + "00100000" + "00000000",
	"00000000" + "00000000" + "01000000" + "00000000",
	"00000000" + "00000000" + "10000000" + "00000000",

	"00000000" + "00000001" + "00000000" + "00000000",
	"00000000" + "00000010" + "00000000" + "00000000",
	"00000000" + "00000100" + "00000000" + "00000000",
	"00000000" + "00001000" + "00000000" + "00000000",
	"00000000" + "00010000" + "00000000" + "00000000",
	"00000000" + "00100000" + "00000000" + "00000000",
	"00000000" + "01000000" + "00000000" + "00000000",
	"00000000" + "10000000" + "00000000" + "00000000",

	"00000001" + "00000000" + "00000000" + "00000000",
	"00000010" + "00000000" + "00000000" + "00000000",
	"00000100" + "00000000" + "00000000" + "00000000",
	"00001000" + "00000000" + "00000000" + "00000000",
	"00010000" + "00000000" + "00000000" + "00000000",
	"00100000" + "00000000" + "00000000" + "00000000",
	"01000000" + "00000000" + "00000000" + "00000000",
	"10000000" + "00000000" + "00000000" + "00000000",
}
