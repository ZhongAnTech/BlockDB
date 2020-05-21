package engine

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/listener"
	"github.com/annchain/BlockDB/multiplexer"
	"github.com/annchain/BlockDB/ogws"
	"github.com/annchain/BlockDB/plugins/client/og"
	"github.com/annchain/BlockDB/plugins/server/jsondata"
	"github.com/annchain/BlockDB/plugins/server/kafka"
	"github.com/annchain/BlockDB/plugins/server/log4j2"
	"github.com/annchain/BlockDB/plugins/server/mongodb"
	"github.com/annchain/BlockDB/plugins/server/socket"
	"github.com/annchain/BlockDB/plugins/server/web"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	if viper.GetBool("debug.enabled") {
		port := viper.GetInt("debug.port")
		go logrus.Fatal(http.ListenAndServe("localhost:"+fmt.Sprintf("%d", port), nil))
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
			GroupId: viper.GetString("listener.kafka.group_id"),
		},
			jsondata.NewJsonDataProcessor(jsondata.JsonDataProcessorConfig{}),
			defaultLedgerWriter,
		)
		n.components = append(n.components, p)
	}
	auditWriter := ogws.NewMongoDBAuditWriter(
		viper.GetString("audit.mongodb.connection_string"),
		viper.GetString("audit.mongodb.database"),
		viper.GetString("audit.mongodb.collection"),
	)
	if viper.GetBool("og.wsclient.enabled") {
		w := ogws.NewOGWSClient(viper.GetString("og.wsclient.url"), auditWriter)
		n.components = append(n.components, w)
	}
	if viper.GetBool("listener.http.enabled") {
		p := web.NewHttpListener(web.HttpListenerConfig{
			Port:             viper.GetInt("listener.http.port"),
			EnableAudit:      viper.GetBool("listener.http.enable_audit"),
			EnableHealth:     viper.GetBool("listener.http.enable_health"),
			MaxContentLength: viper.GetInt64("listener.http.max_content_length"),
		},
			jsondata.NewJsonDataProcessor(jsondata.JsonDataProcessorConfig{}),
			defaultLedgerWriter,
			auditWriter,
		)
		n.components = append(n.components, p)
	}
}
