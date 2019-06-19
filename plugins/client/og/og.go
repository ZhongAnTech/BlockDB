package og

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/annchain/BlockDB/httplib"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

type OgProcessor struct {
	config OgProcessorConfig
}
type OgProcessorConfig struct {
	IdleConnectionTimeout time.Duration
	LedgerUrl             string
}

func (m *OgProcessor) Stop() {
	logrus.Info("OgProcessor stopped")
}

func (m *OgProcessor) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
}

func (m *OgProcessor) Name() string {
	return "OgProcessor"
}

func NewOgProcessor(config OgProcessorConfig) *OgProcessor {
	_, err := url.Parse(config.LedgerUrl)
	if err != nil {
		panic(err)
	}
	return &OgProcessor{
		config: config,
	}
}

func (o *OgProcessor) SendToLedger(data string) {
	 resData,err := o.sendToLedger(data)
	if err!=nil {
		logrus.WithError(err).Warn("send data to og failed")
		return
	}
	logrus.WithField("res ",resData).Debug("got response")
}

type TxReq struct {
	Data string `json:"data"`
}

type Response struct {

	//"data": "0x2f0d3ee49d9eb21a75249b348541574d11f6f70f36c50892b89db3e1dc4a591a",
	//"message": ""
	Data  interface{} `json:"data"`
	Message string  `json:"message"`

}

func (o *OgProcessor) sendToLedger(data string) (resData interface{},err error ){
	req := httplib.Post(o.config.LedgerUrl)
	req.SetTimeout(time.Second*10, time.Second*10)
	txReq := TxReq{
		Data: data,
	}
	_, err = req.JSONBody(&txReq)
	if err != nil {
		panic(fmt.Errorf("encode tx errror %v", err))
	}
	d, _ := json.MarshalIndent(&txReq, "", "\t")
	logrus.WithField("data ", string(d)).Debug("send data to og")

	var res Response

	err = req.ToJSON(&res)
	if err != nil {
		logrus.WithError(err).Warn("send data failed")
		str,e := req.String()
		logrus.WithField("res ",str).WithError(e).Warn("got response")
		return  nil,err
	}
	if res.Message !="" {
		err = errors.New(res.Message)
		logrus.WithError(err).Warn("got error from og")
		return   nil,err
	}
	//logrus.Debug(res,res.Data)
	return res.Data,nil
}

