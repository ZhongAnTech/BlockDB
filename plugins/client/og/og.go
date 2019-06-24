package og

import (
	"encoding/json"
	"errors"
	"github.com/annchain/BlockDB/httplib"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

type OgProcessor struct {
	config   OgProcessorConfig
	dataChan chan string
	quit     chan bool
}
type OgProcessorConfig struct {
	IdleConnectionTimeout time.Duration
	LedgerUrl             string
	BufferSize            int
	RetryTimes            int
}

func (m *OgProcessor) Stop() {
	m.quit <- true
}

func (m *OgProcessor) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
	go m.ConsumeQueue()
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
		config:   config,
		dataChan: make(chan string, config.BufferSize),
	}
}

func (o *OgProcessor) EnqueueSendToLedger(data string) {
	o.dataChan <- data
	resData, err := o.sendToLedger(data)

	if err != nil {
		logrus.WithError(err).Warn("send data to og failed")
		return
	}
	logrus.WithField("res ", resData).Debug("got response")
}

func (o *OgProcessor) ConsumeQueue() {
outside:
	for {
		select {
		case data := <-o.dataChan:
			retry := 0
			for ; retry < o.config.RetryTimes; retry++ {
				resData, err := o.sendToLedger(data)
				if err != nil {
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				}
				logrus.WithField("response", resData).Debug("got response")
			}
			if retry == o.config.RetryTimes {
				logrus.WithField("data", data).Error("failed to send data to ledger. Abandon.")
			}
		case <-o.quit:
			break outside
		}
	}
	logrus.Info("OgProcessor stopped")
}

type TxReq struct {
	Data string `json:"data"`
}

type Response struct {
	//"data": "0x2f0d3ee49d9eb21a75249b348541574d11f6f70f36c50892b89db3e1dc4a591a",
	//"message": ""
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logrus.Debugf("%s took %s", name, elapsed)
}

func (o *OgProcessor) sendToLedger(data string) (resData interface{}, err error) {
	defer timeTrack(time.Now(), "sendToLedger")

	req := httplib.Post(o.config.LedgerUrl)
	req.SetTimeout(time.Second*10, time.Second*10)
	txReq := TxReq{
		Data: data,
	}
	_, err = req.JSONBody(&txReq)
	if err != nil {
		logrus.WithError(err).Error("error on encoding tx")
		return nil, err
	}
	d, _ := json.MarshalIndent(&txReq, "", "\t")
	logrus.WithField("data ", string(d)).Trace("send data to og")

	var res Response

	err = req.ToJSON(&res)
	if err != nil {
		logrus.WithError(err).Warn("send data failed")
		str, e := req.String()
		logrus.WithField("res ", str).WithError(e).Warn("got response")
		return nil, err
	}
	if res.Message != "" {
		err = errors.New(res.Message)
		logrus.WithError(err).Warn("got error from og")
		return nil, err
	}
	//logrus.Debug(res,res.Data)
	return res.Data, nil
}
