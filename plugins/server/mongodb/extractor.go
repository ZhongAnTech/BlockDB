package mongodb

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
	"time"

	//"github.com/sirupsen/logrus"

	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/multiplexer"
	"github.com/annchain/BlockDB/plugins/server/mongodb/message"
	"github.com/annchain/BlockDB/processors"
)

type ExtractorFactory struct {
	ledgerWriter backends.LedgerWriter
	config       *ExtractorConfig
}

type ExtractorConfig struct {
	IgnoreMetaQuery bool
}

func NewExtractorFactory(writer backends.LedgerWriter, config *ExtractorConfig) *ExtractorFactory {
	return &ExtractorFactory{
		ledgerWriter: writer,
		config:       config,
	}
}

func (e ExtractorFactory) GetInstance(context multiplexer.DialogContext) multiplexer.Observer {
	return &ExtractorObserver{
		req:  NewExtractor(context, e.ledgerWriter, e.config),
		resp: NewExtractor(context, e.ledgerWriter, e.config),
	}
}

type ExtractorObserver struct {
	req  ExtractorInterface
	resp ExtractorInterface
}

func (e *ExtractorObserver) GetIncomingWriter() io.Writer {
	return e.req
}

func (e *ExtractorObserver) GetOutgoingWriter() io.Writer {
	return e.resp
}

type ExtractorInterface interface {
	Write(p []byte) (n int, err error)

	// init is called when message struct is detected.
	init(header *message.MessageHeader)

	// reset() is called when extractor finishes one iteration of extraction.
	// All the variables are set to origin situations to wait next message
	// extraction.
	reset() error
}

type Extractor struct {
	context multiplexer.DialogContext
	header  *message.MessageHeader
	buf     []byte

	extract func(h *message.MessageHeader, b []byte) (*message.Message, error)

	writer backends.LedgerWriter
	config *ExtractorConfig

	mu sync.RWMutex
}

func NewExtractor(context multiplexer.DialogContext, writer backends.LedgerWriter, config *ExtractorConfig) *Extractor {
	r := &Extractor{}

	r.buf = make([]byte, 0)
	r.extract = extractMessage
	r.context = context
	r.writer = writer
	r.config = config

	return r
}

func (e *Extractor) Write(p []byte) (int, error) {

	//fmt.Println("new Write byte: ", hex.Dump(p))

	b := make([]byte, len(p))
	copy(b, p)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.buf = append(e.buf, b...)

	// init header
	if e.header == nil && len(e.buf) > message.HeaderLen {
		header, err := message.DecodeHeader(e.buf)
		if err != nil {
			return len(b), err
		}
		e.init(header)
	}
	// case that buf size no larger than header length or buf not matches msg size.
	if e.header == nil || uint32(len(e.buf)) < e.header.MessageSize {
		return len(b), nil
	}

	msg, err := e.extract(e.header, e.buf[:e.header.MessageSize])
	if err != nil {
		return len(b), err
	}
	// set user to context
	if msg.DBUser != "" {
		e.context.User = msg.DBUser
	} else {
		msg.DBUser = e.context.User
	}

	logEvent := &processors.LogEvent{
		Type:       "mongo",
		Ip:         e.context.Source.RemoteAddr().String(),
		Data:       msg,
		PrimaryKey: msg.DocID,
		Timestamp:  int64(time.Now().Unix()),
		Identity:   msg.DBUser,
	}

	data, _ := json.Marshal(logEvent)

	write := true
	if e.config.IgnoreMetaQuery {
		s := string(p)
		blacklist := []string{"buildinfo", "getlasterror", "architecture", "dbStats", "saslStart", "saslContinue", "listCollections", "collStats"}
		if msg.DB == "admin" {
			write = false
		} else {
			// check blacklist
			for _, word := range blacklist {
				if strings.Contains(s, word) {
					fmt.Println("blacklist event: ", string(data))
					write = false
					break
				}
			}
		}
	}
	if write {
		t, _ := json.Marshal(msg)
		logrus.WithField("ev", string(t)).Warn("log")
		e.writer.EnqueueSendToLedger(logEvent)
	}
	e.reset()

	return len(b), nil
}

func (e *Extractor) init(header *message.MessageHeader) {
	e.header = header
}

func (e *Extractor) reset() error {
	if e.header == nil {
		e.buf = make([]byte, 0)
		return nil
	}
	e.buf = e.buf[e.header.MessageSize:]
	e.header = nil
	return nil
}

func extractMessage(header *message.MessageHeader, b []byte) (*message.Message, error) {
	if len(b) != int(header.MessageSize) {
		return nil, fmt.Errorf("msg bytes length not equal to size in header. "+
			"Bytes length: %d, size in header: %d", len(b), header.MessageSize)
	}

	var err error
	var mm message.MongoMessage
	switch header.OpCode {
	case message.OpReply:
		mm, err = message.NewReplyMessage(header, b)
		break
	case message.OpUpdate:
		err = fmt.Errorf("Extraction for OpUpdate not implemented")
		//mm, err = message.NewUpdateMessage(header, b)
		break
	case message.OpInsert:
		err = fmt.Errorf("Extraction for OpInsert not implemented")
		//mm, err = message.NewInsertMessage(header, b)
		break
	case message.Reserved:
		err = fmt.Errorf("Extraction for Reserved not implemented")
		//mm, err = message.NewReservedMessage(header, b)
		break
	case message.OpQuery:
		mm, err = message.NewQueryMessage(header, b)
		break
	case message.OpGetMore:
		err = fmt.Errorf("Extraction for OpGetMore not implemented")
		//mm, err = message.NewGetMoreMessage(header, b)
		break
	case message.OpDelete:
		err = fmt.Errorf("Extraction for OpDelete not implemented")
		//mm, err = message.NewDeleteMessage(header, b)
		break
	case message.OpKillCursors:
		err = fmt.Errorf("Extraction for OpKillCursors not implemented")
		//mm, err = message.NewKillCursorsMessage(header, b)
		break
	case message.OpCommand:
		err = fmt.Errorf("Extraction for OpCommand not implemented")
		//mm, err = message.NewCommandMessage(header, b)
		break
	case message.OpCommandReply:
		err = fmt.Errorf("Extraction for OpCommandReply not implemented")
		//mm, err = message.NewCommandReplyMessage(header, b)
		break
	case message.OpMsg:
		mm, err = message.NewMsgMessage(header, b)
		break
	default:
		return nil, fmt.Errorf("unknown opcode: %d", header.OpCode)
	}
	if err != nil {
		return nil, fmt.Errorf("init mongo message error: %v", err)
	}

	user, db, collection, op, docId := mm.ExtractBasic()

	m := &message.Message{}
	m.DBUser = user
	m.DB = db
	m.Collection = collection
	m.Op = op
	m.DocID = docId
	m.MongoMsg = mm

	return m, nil
}
