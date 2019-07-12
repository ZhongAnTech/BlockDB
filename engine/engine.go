package engine

import (
	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/listener"
	"github.com/annchain/BlockDB/multiplexer"
	"github.com/annchain/BlockDB/plugins/client/og"
	"github.com/annchain/BlockDB/plugins/server/jsondata"
	"github.com/annchain/BlockDB/plugins/server/kafka"
	"github.com/annchain/BlockDB/plugins/server/log4j2"
	"github.com/annchain/BlockDB/plugins/server/mongodb"
	"github.com/annchain/BlockDB/plugins/server/socket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type Engine struct {
	components []Component
}

func NewEngine() *Engine {
	engine := new(Engine)
	engine.components = []Component{}
	engine.registerComponents()
	return engine
}

func (n *Engine) Start() {
	for _, component := range n.components {
		logrus.Infof("Starting %s", component.Name())
		component.Start()
		logrus.Infof("Started: %s", component.Name())

	}
	logrus.Info("BlockDB Engine Started")
}

func (n *Engine) Stop() {
	//status.Stopped = true
	for i := len(n.components) - 1; i >= 0; i-- {
		comp := n.components[i]
		logrus.Infof("Stopping %s", comp.Name())
		comp.Stop()
		logrus.Infof("Stopped: %s", comp.Name())
	}
	logrus.Info("BlockDB Engine Stopped")
}

func (n *Engine) registerComponents() {

	var defaultLedgerWriter backends.LedgerWriter

	if viper.GetBool("og.enabled") {
		url := viper.GetString("og.url")
		p := og.NewOgProcessor(og.OgProcessorConfig{LedgerUrl: url,
			IdleConnectionTimeout: time.Second * time.Duration(viper.GetInt("og.idle_connection_seconds")),
			BufferSize:            viper.GetInt("og.buffer_size"),
			RetryTimes:            viper.GetInt("og.retry_times"),
		})
		defaultLedgerWriter = p
		n.components = append(n.components, p)
	}

	// MongoDB incoming
	if viper.GetBool("listener.mongodb.enabled") {
		url := viper.GetString("backend.mongodb.url")
		if url != "" {
			builder := multiplexer.NewDefaultTCPConnectionBuilder(url)
			observerFactory := mongodb.NewExtractorFactory(defaultLedgerWriter, &mongodb.ExtractorConfig{
				IgnoreMetaQuery: viper.GetBool("listener.mongodb.ignore_meta_query"),
			})
			mp := multiplexer.NewMultiplexer(builder, observerFactory)
			l := listener.NewGeneralTCPListener(mp, viper.GetInt("listener.mongodb.incoming_port"),
				viper.GetInt("listener.mongodb.incoming_max_connection"))

			n.components = append(n.components, l)
		}
	}

	if viper.GetBool("listener.log4j2Socket.enabled") {
		// Incoming connection handler
		p := log4j2.NewLog4j2SocketProcessor(
			log4j2.Log4j2SocketProcessorConfig{
				IdleConnectionTimeout: time.Second * time.Duration(viper.GetInt("listener.log4j2Socket.idle_connection_seconds")),
			},
			defaultLedgerWriter,
		)
		l := listener.NewGeneralTCPListener(p, viper.GetInt("listener.log4j2Socket.incoming_port"),
			viper.GetInt("listener.log4j2Socket.incoming_max_connection"))
		n.components = append(n.components, l)
	}

	if viper.GetBool("listener.jsonSocket.enabled") {
		// Incoming connection handler
		p := socket.NewSocketProcessor(
			socket.SocketConnectionProcessorConfig{
				IdleConnectionTimeout: time.Second * time.Duration(viper.GetInt("listener.jsonSocket.idle_connection_seconds")),
			},
			jsondata.NewJsonDataProcessor(jsondata.JsonDataProcessorConfig{}),
			defaultLedgerWriter,
		)
		l := listener.NewGeneralTCPListener(p, viper.GetInt("listener.jsonSocket.incoming_port"),
			viper.GetInt("listener.jsonSocket.incoming_max_connection"))
		n.components = append(n.components, l)
	}

	if viper.GetBool("listener.kafka.enabled") {
		// Incoming connection handler
		p := kafka.NewKafkaListener(kafka.KafkaProcessorConfig{
			Topic:   viper.GetString("listener.kafka.topic"),
			Address: viper.GetString("listener.kafka.address"),
		},
			jsondata.NewJsonDataProcessor(jsondata.JsonDataProcessorConfig{}),
			defaultLedgerWriter,
		)
		n.components = append(n.components, p)
	}

}
