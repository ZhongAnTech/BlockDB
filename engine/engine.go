package engine

import "github.com/sirupsen/logrus"

type Engine struct {
	Components []Component
}

func NewEngine() *Engine {
	engine := new(Engine)
	engine.Components = []Component{}

}

func (n *Engine) Start() {
	for _, component := range n.Components {
		logrus.Infof("Starting %s", component.Name())
		component.Start()
		logrus.Infof("Started: %s", component.Name())

	}
	logrus.Info("Engine Started")
}

func (n *Engine) Stop() {
	//status.Stopped = true
	for i := len(n.Components) - 1; i >= 0; i-- {
		comp := n.Components[i]
		logrus.Infof("Stopping %s", comp.Name())
		comp.Stop()
		logrus.Infof("Stopped: %s", comp.Name())
	}
	logrus.Info("Node Stopped")
}
