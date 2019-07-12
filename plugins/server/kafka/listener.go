package Kafka

import (
	"context"
	"fmt"
	"github.com/annchain/BlockDB/backends"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type KafkaProcessorConfig struct {
	Topic   string
	Address string
}

type KafkaListener struct {
	config       KafkaProcessorConfig
	ledgerWriter backends.LedgerWriter
	wg           sync.WaitGroup
	stopped      bool
}

func (m *KafkaListener) Name() string {
	return "KafkaListener"
}

func NewKafkaListener(config KafkaProcessorConfig, ledgerWriter backends.LedgerWriter) *KafkaListener {
	return &KafkaListener{
		config:       config,
		ledgerWriter: ledgerWriter,
	}
}

func (m *KafkaListener) Start() {
	ps, _ := kafka.LookupPartitions(context.Background(), "tcp", m.config.Address, m.config.Topic)

	// currently we will listen to all partitions
	for _, p := range ps {
		m.wg.Add(1)
		go m.doListen(p)
	}
	logrus.Info("KafkaListener started")
}

func (m *KafkaListener) Stop() {
	m.stopped = true
	m.wg.Wait()
	logrus.Info("KafkaListener stopped")
}

func (m *KafkaListener) doListen(partition kafka.Partition) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   strings.Split(m.config.Address, ";"),
		Topic:     m.config.Topic,
		Partition: partition.ID,
		MinBytes:  1,    // 1B
		MaxBytes:  10e6, // 10MB
	})
	defer func() {
		_ = r.Close()
		m.wg.Done()
	}()

	deadlineContext, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*3))
	err := r.SetOffsetAt(deadlineContext, time.Now())
	if err != nil {
		return
	}
	logrus.WithField("partition", partition.ID).WithField("topic", m.config.Topic).Info("kafka partition consumer started")

	for !m.stopped {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logrus.WithError(err).WithField("partition", partition.ID).Error("partition error")
			time.Sleep(time.Second * 1)
			continue
		}
		fmt.Printf("[%v],topic:[%v],partition:[%v],offset:[%v],key:[%s]\n", m.Time, m.Topic, m.Partition, m.Offset, string(m.Key))
		fmt.Println(string(m.Value))
	}

}
