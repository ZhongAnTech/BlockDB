package mongodb

import (
	"bufio"
	"fmt"
	"github.com/annchain/BlockDB/common/bytes"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

const headerLen = 16

type MongoProcessor struct {
	config MongoProcessorConfig

	readPool  *Pool
	writePool *Pool
}
type MongoProcessorConfig struct {
	IdleConnectionTimeout time.Duration
}

func (m *MongoProcessor) Stop() {
	logrus.Info("MongoProcessor stopped")
}

func (m *MongoProcessor) Start() {
	logrus.Info("MongoProcessor started")
	// start consuming queue
}

func NewMongoProcessor(config MongoProcessorConfig) *MongoProcessor {
	return &MongoProcessor{
		config:    config,
		readPool:  NewPool(10),
		writePool: NewPool(10),
	}
}

func (m *MongoProcessor) ProcessConnection(conn net.Conn) error {

	fmt.Println("start process connection")

	// 1, parse command
	// 2, dispatch the command to every interested parties
	//    including chain logger and the real backend mongoDB server
	// 3, response to conn
	for {
		conn.SetReadDeadline(time.Now().Add(m.config.IdleConnectionTimeout))

		var b [headerLen]byte
		cmdHeader := b[:]
		_, err := bufio.NewReader(conn).Read(cmdHeader)
		if err != nil {
			if err == io.EOF {
				logrus.Info("target closed")
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				logrus.Info("target timeout")
				conn.Close()
				return nil
			}
			return err
		}

		// query command
		msgSize := bytes.GetInt32(cmdHeader, 0)
		fmt.Println("msgsize: ", msgSize)

		// TODO handle full msg not only header msg
		err = m.messageHandler(cmdHeader, conn)
		if err != nil {
			// TODO handle err
			return err
		}

		break
	}
	return nil
}

func (m *MongoProcessor) messageHandler(bytes []byte, client net.Conn) error {

	var msg RequestMessage
	err := msg.Decode(bytes)
	if err != nil {
		// TODO handle err
		return err
	}

	var pool *Pool
	if msg.Read() {
		pool = m.readPool
	} else {
		pool = m.writePool
	}
	server := pool.Acquire()
	defer pool.FreeConn(server)

	err = msg.WriteTo(server)
	if err != nil {
		// TODO handle err
		return err
	}

	var msgResp ResponseMessage
	err = msgResp.ReadFromMongo(server)
	if err != nil {
		// TODO handle err
		return err
	}
	err = msgResp.WriteTo(client)
	if err != nil {
		// TODO handle err
		return err
	}

	err = m.handleBlockDBEvents(&msgResp)
	if err != nil {
		// TODO handle err
		return err
	}

	return nil
}

func (m *MongoProcessor) handleBlockDBEvents(msg MongoMessage) error {
	// TODO not implemented yet

	events := msg.ParseCommand()

	fmt.Println("block db events: ", events)

	return nil
}
