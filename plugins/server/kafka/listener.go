package kafka

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/processors"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaProcessorConfig struct {
	Topic   string
	Address string
	GroupId  string
}

type KafkaListener struct {
	config        KafkaProcessorConfig
	ledgerWriter  backends.LedgerWriter
	dataProcessor processors.DataProcessor

	wg      sync.WaitGroup
	stopped bool
}

func (k *KafkaListener) Name() string {
	return "KafkaListener"
}

func NewKafkaListener(config KafkaProcessorConfig, dataProcessor processors.DataProcessor, ledgerWriter backends.LedgerWriter) *KafkaListener {
	return &KafkaListener{
		config:        config,
		ledgerWriter:  ledgerWriter,
		dataProcessor: dataProcessor,
	}
}

func (k *KafkaListener) Start() {
	//ps, _ := kafka.LookupPartitions(context.Background(), "tcp", k.config.Address, k.config.Topic)

	// currently we will listen to all partitions
	//for _, p := range ps {
	//	k.wg.Add(1)
	//	go k.doListen(p)
	//}
	k.wg.Add(1)
	go k.doListen()
	logrus.Info("KafkaListener started")
}

func (k *KafkaListener) Stop() {
	k.stopped = true
	k.wg.Wait()
	logrus.Info("KafkaListener stopped")
}

func (k *KafkaListener) doListen() {
	brokers := strings.Split(k.config.Address, ";")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     k.config.Topic,
		MinBytes:  1,    // 1B
		MaxBytes:  10e6, // 10MB,
		GroupID:k.config.GroupId,
	})
	defer func() {
		_ = r.Close()
		k.wg.Done()
	}()

	//deadlineContext, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*3))
	//err := r.SetOffsetAt(deadlineContext, time.Now())
	//if err != nil {
	//	logrus.WithError(err).Error("cannot set offset to partition")
	//	return
	//}
	logrus.WithField("brokers",brokers).WithField("groupid",k.config.GroupId).WithField("topic", k.config.Topic).Info("kafka  consumer started")

	for !k.stopped {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logrus.WithError(err).Error("read msg error")
			time.Sleep(time.Second * 1)
			continue
		}
		s := string(m.Value)
		logrus.WithFields(logrus.Fields{
			"partition": m.Partition,
			"offset":    m.Offset,
			"msg":       s,
		}).Info("message")

		events := k.dataProcessor.ParseCommand(m.Value)
		for _, event := range events {
			k.ledgerWriter.EnqueueSendToLedger(event)
		}
	}

}
