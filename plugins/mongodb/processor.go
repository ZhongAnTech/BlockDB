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
}

func NewMongoProcessor() *MongoProcessor {
	return &MongoProcessor{}
}

func (m *MongoProcessor) ProcessConnection(conn net.Conn) {
	return
}
