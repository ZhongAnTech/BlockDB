package mongodb

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

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
	// 1, parse command
	// 2, send to internal queue
	// 3, consume queue and dispatch the command to every interested parties
	//    including chain logger and the real backend mongoDB server
	// 4, response to conn
	for {
		conn.SetReadDeadline(time.Now().Add(m.config.IdleConnectionTimeout))
		str, err := bufio.NewReader(conn).ReadString('\n')
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
		} else {
			fmt.Println(str)
		}
	}

	return nil
}
