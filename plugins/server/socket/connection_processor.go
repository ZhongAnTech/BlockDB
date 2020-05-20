package socket

import (
	"bufio"
	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/processors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type SocketConnectionProcessorConfig struct {
	IdleConnectionTimeout time.Duration
}

type SocketConnectionProcessor struct {
	config        SocketConnectionProcessorConfig
	dataProcessor processors.DataProcessor
	ledgerWriter  backends.LedgerWriter
}

func NewSocketProcessor(config SocketConnectionProcessorConfig, dataProcessor processors.DataProcessor, ledgerWriter backends.LedgerWriter) *SocketConnectionProcessor {
	return &SocketConnectionProcessor{
		config:        config,
		dataProcessor: dataProcessor,
		ledgerWriter:  ledgerWriter,
	}
}

func (s *SocketConnectionProcessor) ProcessConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(s.config.IdleConnectionTimeout))
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
		//fmt.Println(str)
		//fmt.Println(hex.Dump(bytes))
		events, err := s.dataProcessor.ParseCommand([]byte(str))
		if events == nil || err != nil {
			logrus.WithError(err).Warn("nil command")
			continue
		}
		for _, event := range events {
			event.Ip = conn.RemoteAddr().String()
			s.ledgerWriter.EnqueueSendToLedger(event)
		}
	}
}

func (SocketConnectionProcessor) Start() {
	logrus.Info("SocketConnectionProcessor started")
}

func (SocketConnectionProcessor) Stop() {
	logrus.Info("SocketConnectionProcessor stopped")
}
