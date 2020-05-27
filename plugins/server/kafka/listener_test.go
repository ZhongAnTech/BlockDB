package kafka

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/annchain/BlockDB/plugins/server/jsondata"
	"github.com/spf13/viper"
)

func TestListener(t *testing.T) {

	f, _ := os.Open(`D:\ws\gitpublic\annchain\BlockDB\config.toml`)
	defer f.Close()
	viper.SetConfigType("toml")

	viper.ReadConfig(f)
	viper.Get("listener.kafka.address")
	viper.Debug()

	config := KafkaProcessorConfig{
		Topic:   viper.GetString("listener.kafka.topic"),
		Address: viper.GetString("listener.kafka.address"),
		GroupId: viper.GetString("listener.kafka.group_id"),
	}
	fmt.Println(config)
	l := NewKafkaListener(config, &jsondata.JsonDataProcessor{}, &ledgerSender{})
	l.Start()

	for true {
		time.Sleep(time.Second)
	}
}

type ledgerSender struct {
}

func (l *ledgerSender) EnqueueSendToLedger(data interface{}) {
	fmt.Println(data)
}
