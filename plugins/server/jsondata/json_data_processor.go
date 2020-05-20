package jsondata

import (
	"encoding/json"
	"github.com/annchain/BlockDB/processors"
	"github.com/sirupsen/logrus"
	"time"
)

type JsonDataProcessorConfig struct {
}

type JsonDataProcessor struct {
	config JsonDataProcessorConfig
}

func NewJsonDataProcessor(config JsonDataProcessorConfig) *JsonDataProcessor {
	return &JsonDataProcessor{
		config: config,
	}
}

func (m *JsonDataProcessor) Start() {
	logrus.Info("JsonDataProcessor started")
}

func (m *JsonDataProcessor) Stop() {
	logrus.Info("JsonDataProcessor stopped")
}

func (m *JsonDataProcessor) ParseCommand(bytes []byte) (events []*processors.LogEvent, err error) {
	var c processors.LogEvent
	if err := json.Unmarshal(bytes, &c); err != nil {
		logrus.WithError(err).Warn("bad format")
		return nil, err
	}

	if c.Type == "" {
		c.Type = "json"
	}
	if c.Timestamp == 0 {
		c.Timestamp = time.Now().UnixNano() / 1e6
	}
	return []*processors.LogEvent{&c}, nil

}
