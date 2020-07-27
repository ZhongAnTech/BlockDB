package core

import (
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/clients/og"
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/listeners/web"
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

	// TODO: RPC server to receive http requests. (Wu Jianhang)
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
	// TODO: Command executor (Fang Ning)
	// CommandExecutor

	// TODO: External data storage facilities. (Dai Yunong)
	// StorageExecutor

	// TODO: Blockchain sender to send new tx consumed from queue. (Ding Qingyun)
	client := &og.OgClient{
		Config: og.OgClientConfig{
			LedgerUrl:  viper.GetString("blockchain.og.url"),
			RetryTimes: viper.GetInt("blockchain.og.retry_times"),
		},
	}
	client.InitDefault()
	n.components = append(n.components, client)

	// TODO: Sync manager to sync from lastHeight to maxHeight. (Wu Jianhang)
	// LedgerSyncer

	// TODO: Websocket server to receive new sequencer messages. (Ding Qingyun)
	// BlockchainListener

}
