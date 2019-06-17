package mongodb

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/annchain/BlockDB/processors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

const headerLen = 16

type MongoProcessor struct {
	config MongoProcessorConfig
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
		config: config,
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
		bytes := b[:]
		_, err := bufio.NewReader(conn).Read(bytes)
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
		fmt.Println(hex.Dump(bytes))
		events := m.ParseCommand(bytes)
		fmt.Println(events)

	}
	return nil
}

func (m *MongoProcessor) ParseCommand(bytes []byte) []processors.LogEvent {
	return nil
}
