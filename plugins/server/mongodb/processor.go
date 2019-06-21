package mongodb

import (
	"bufio"
	"fmt"
	"github.com/annchain/BlockDB/plugins/server/mongodb/message"

	"github.com/annchain/BlockDB/common/bytes"

	//"github.com/annchain/BlockDB/processors"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

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
	//TODO move mongo url into config
	url := "172.28.152.101:27017"

	return &MongoProcessor{
		config:    config,
		readPool:  NewPool(url, 10),
		writePool: NewPool(url, 10),
	}
}

func (m *MongoProcessor) ProcessConnection(conn net.Conn) error {
	defer conn.Close()

	fmt.Println("start process connection")

	// http://docs.mongodb.org/manual/faq/diagnostics/#faq-keepalive
	if conn, ok := conn.(*net.TCPConn); ok {
		conn.SetKeepAlivePeriod(2 * time.Minute)
		conn.SetKeepAlive(true)
	}

	reader := bufio.NewReader(conn)

	backend := m.writePool.Acquire()
	defer m.writePool.Release(backend)

	for {
		conn.SetReadDeadline(time.Now().Add(m.config.IdleConnectionTimeout))

		cmdHeader := make([]byte, message.HeaderLen)
		_, err := reader.Read(cmdHeader)
		if err != nil {
			if err == io.EOF {
				fmt.Println("target closed")
				logrus.Info("target closed")
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				fmt.Println("target timeout")
				logrus.Info("target timeout")
				conn.Close()
				return nil
			}
			return err
		}

		// query command
		msgSize := bytes.GetInt32(cmdHeader, 0)
		fmt.Println("msgsize: ", msgSize)

		cmdBody := make([]byte, msgSize-message.HeaderLen)
		_, err = reader.Read(cmdBody)
		if err != nil {
			fmt.Println("read body error: ", err)
			return err
		}
		fmt.Println(fmt.Sprintf("msg header: %x", cmdHeader))
		fmt.Println(fmt.Sprintf("msg body: %x", cmdBody))

		cmdFull := append(cmdHeader, cmdBody...)
		err = m.messageHandler(cmdFull, conn, backend)
		if err != nil {
			// TODO handle err
			return err
		}

	}
	return nil
}

func (m *MongoProcessor) messageHandler(bytes []byte, client, backend net.Conn) error {
	//
	//var msg message.RequestMessage
	//fmt.Println("Request--->")
	//fmt.Println(hex.Dump(bytes))
	//err := msg.Decode(bytes)
	//if err != nil {
	//	// TODO handle err
	//	return err
	//}
	//
	////var pool *Pool
	////if msg.ReadOnly() {
	////	pool = m.readPool
	////} else {
	////	pool = m.writePool
	////}
	////backend := pool.Acquire()
	////defer pool.Release(backend)
	//
	//err = msg.WriteTo(backend)
	//if err != nil {
	//	// TODO handle err
	//	return err
	//}
	//
	//var msgResp message.ResponseMessage
	//err = msgResp.ReadFromMongo(backend)
	//if err != nil {
	//	// TODO handle err
	//	return err
	//}
	//fmt.Println("<---Response")
	////fmt.Println(hex.Dump(msgResp.payload))
	//err = msgResp.WriteTo(client)
	//if err != nil {
	//	// TODO handle err
	//	return err
	//}
	//
	////err = m.handleBlockDBEvents(&msgResp)
	////if err != nil {
	////	// TODO handle err
	////	return err
	////}

	return nil
}

func (m *MongoProcessor) handleBlockDBEvents(msg message.Message) error {
	// TODO not implemented yet

	events := msg.ParseCommand()

	fmt.Println("block db events: ", events)

	return nil
}
