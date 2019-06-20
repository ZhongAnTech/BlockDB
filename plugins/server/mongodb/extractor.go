package mongodb

import (
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"sync"
)

type Extractor interface {
	Write(p []byte) (n int, err error)

	// init is called when message struct is detected.
	init(info interface{})

	// reset() is called when extractor finishes one iteration of extraction.
	// All the variables are set to origin situations to wait next message
	// extraction.
	reset() error
}

type RequestExtractor struct {
	header *MessageHeader
	buf    []byte

	extract func(b []byte) (MongoMessage, error)

	mu sync.RWMutex
}

func NewRequestExtractor() *RequestExtractor {
	r := &RequestExtractor{}

	r.buf = make([]byte, 0)
	r.extract = extractMongoMessage

	return r
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
func (e *RequestExtractor) Write(p []byte) (int, error) {
	b := make([]byte, len(p))
	copy(b, p)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.buf = append(e.buf, b...)

	// init header
	if e.header == nil && len(e.buf) > headerLen {
		header, err := decodeHeader(e.buf)
		if err != nil {
			return len(b), err
		}
		e.init(header)
	}
	// case that buf size no larger than header length or buf not matches msg size.
	if e.header == nil || int32(len(e.buf)) < e.header.MessageSize {
		return len(b), nil
	}

	msg, err := e.extract(e.buf[:e.header.MessageSize])
	if err != nil {
		return len(b), err
	}
	// TODO write msg to blockDB.
	fmt.Println(msg)

	e.reset()
	return len(b), nil
}

func (e *RequestExtractor) init(header *MessageHeader) {
	e.header = header
}

func (e *RequestExtractor) reset() error {
	if e.header == nil {
		e.buf = make([]byte, 0)
		return nil
	}
	e.buf = e.buf[e.header.MessageSize:]
	e.header = nil
	return nil
}

type ResponseExtractor struct {
}

func (e *ResponseExtractor) extract(b []byte) MongoMessage {
	// TODO

	return nil
}

func decodeHeader(b []byte) (*MessageHeader, error) {
	if len(b) < headerLen {
		return nil, fmt.Errorf("not enough length for header decoding, expect %d, get %d", headerLen, len(b))
	}

	m := &MessageHeader{
		MessageSize: bytes.GetInt32(b, 0),
		RequestID:   bytes.GetInt32(b, 4),
		ResponseTo:  bytes.GetInt32(b, 8),
		OpCode:      OpCode(bytes.GetInt32(b, 12)),
	}
	return m, nil
}

func extractMongoMessage(b []byte) (MongoMessage, error) {

	return nil, nil
}
