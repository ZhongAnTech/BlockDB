package mongodb

import (
	"github.com/sirupsen/logrus"
	"net"
)

type MongoProcessor struct {
}

func (m *MongoProcessor) Stop() {
	logrus.Info("MongoProcessor stopped")
}

func (m *MongoProcessor) Start() {
	logrus.Info("MongoProcessor started")
	// start consuming queue
}

func NewMongoProcessor() *MongoProcessor {
	return &MongoProcessor{}
}

func (m *MongoProcessor) ProcessConnection(conn net.Conn) {
	// 1, parse command
	// 2, send to internal queue
	// 3, consume queue and dispatch the command to every interested parties
	//    including chain logger and the real backend mongoDB server
	// 4, response to conn

	return
}
