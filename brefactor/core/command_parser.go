package core

import (
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
)

type DefaultJsonCommandParser struct {
}

func (d DefaultJsonCommandParser) FromJson(json string) (core_interface.BlockDBCommand, error) {
	logrus.WithField("json", json).Info("TODO: process json command")
	return core_interface.DefaultBlockDBCommand{}, nil
}
