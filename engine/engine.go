package engine

import (
	"github.com/annchain/BlockDB/listener"
	"github.com/annchain/BlockDB/plugins/server/mongodb"
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
	// MongoDB incoming
	if viper.GetBool("listener.mongodb.enabled") {
		// Incoming connection handler
		p := mongodb.NewMongoProcessor(mongodb.MongoProcessorConfig{
			IdleConnectionTimeout: time.Second * time.Duration(viper.GetInt("listener.mongodb.idle_connection_seconds")),
		})
		l := listener.NewGeneralTCPListener(p, viper.GetInt("listener.mongodb.incoming_port"),
			viper.GetInt("listener.mongodb.incoming_max_connection"))
		n.components = append(n.components, l)
	}

}
