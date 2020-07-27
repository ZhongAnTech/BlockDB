package core

import (
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/web"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BlockDB struct {
	components []Component
}

func (n *BlockDB) Start() {
	for _, component := range n.components {
		logrus.Infof("Starting %s", component.Name())
		component.Start()
		logrus.Infof("Started: %s", component.Name())

	}
	logrus.Info("BlockDB engine started")
}

func (n *BlockDB) Stop() {
	//status.Stopped = true
	for i := len(n.components) - 1; i >= 0; i-- {
		component := n.components[i]
		logrus.Infof("Stopping %s", component.Name())
		component.Stop()
		logrus.Infof("Stopped: %s", component.Name())
	}
	logrus.Info("BlockDB engine stopped gracefully")
}

func (n *BlockDB) Name() string {
	panic("implement me")
}

func (n *BlockDB) InitDefault() {
	n.components = []Component{}
}

func (n *BlockDB) Setup() {
	// init components.
	if viper.GetBool("listener.http.enabled") {
		p := &web.HttpListener{
			JsonCommandParser:       &DefaultJsonCommandParser{}, // parse json command
			BlockDBCommandProcessor: &DefaultCommandProcessor{},  // send command to ledger
			Config: web.HttpListenerConfig{
				Port:             viper.GetInt("listener.http.port"),
				MaxContentLength: viper.GetInt64("listener.http.max_content_length"),
			},
		}

		p.Setup()
		n.components = append(n.components, p)
	}

	// Dependency check on External data storage facilities.

	// Blockchain sender to send new tx consumed from queue.

	// Websocket server to receive new sequencer messages.

	// RPC server to receive http requests.

}
