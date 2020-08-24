package core

import (
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
)

type DefaultCommandProcessor struct {
}

func (d DefaultCommandProcessor) Process(command core_interface.BlockDBCommand) (core_interface.CommandProcessResult, error) {
	logrus.WithField("cmd", command).Info("TODO: process this command")
	return core_interface.CommandProcessResult{
		Hash: "0x00",
		OK:   true,
	}, nil
}
