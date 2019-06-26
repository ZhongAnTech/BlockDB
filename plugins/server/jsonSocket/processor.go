package jsonSocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/processors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type JsonSocketProcessorConfig struct {
	IdleConnectionTimeout time.Duration
}

type JsonSocketProcessor struct {
	config       JsonSocketProcessorConfig
	ledgerWriter backends.LedgerWriter
}

func NewJsonSocketProcessor(config JsonSocketProcessorConfig, ledgerWriter backends.LedgerWriter) *JsonSocketProcessor {
	return &JsonSocketProcessor{
		config:       config,
		ledgerWriter: ledgerWriter,
	}
}

func (m *JsonSocketProcessor) Start() {
	logrus.Info("JsonSocketProcessor started")
}

func (m *JsonSocketProcessor) Stop() {
	logrus.Info("JsonSocketProcessor stopped")
}

func (m *JsonSocketProcessor) ProcessConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(m.config.IdleConnectionTimeout))
		str, err := reader.ReadString(byte(0))
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
		str = str[:len(str)-1]
		// query command
		fmt.Println(str)
		//fmt.Println(hex.Dump(bytes))
		event := m.ParseCommand([]byte(str))
		if event == nil {
			logrus.WithError(err).Warn("nil command")
			continue
		}
		event.Ip = conn.RemoteAddr().String()

		m.ledgerWriter.EnqueueSendToLedger(event)
	}
}

func (m *JsonSocketProcessor) ParseCommand(bytes []byte) *processors.LogEvent {
	var c interface{}
	if err := json.Unmarshal(bytes, &c); err != nil {
		logrus.WithError(err).Warn("bad format")
		return nil
	}
	event := processors.LogEvent{
		Timestamp: time.Now().Unix(),
		Data:      c,
		Type:      "json",
	}
	return &event

}
