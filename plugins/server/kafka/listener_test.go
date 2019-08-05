package kafka

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"testing"
	"time"
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
	}
	fmt.Println(config)

	l := NewKafkaListener(config, nil)
	l.Start()

	for true {
		time.Sleep(time.Second)
	}
}
