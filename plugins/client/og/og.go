package og

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type OgProcessor struct {
	config     OgProcessorConfig
	dataChan   chan interface{}
	quit       chan bool
	httpClient *http.Client
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
		config:     config,
		dataChan:   make(chan interface{}, config.BufferSize),
		quit:       make(chan bool),
		httpClient: createHTTPClient(),
	}
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 15,
		},
		Timeout: time.Duration(10) * time.Second,
	}

	return client
}

func (o *OgProcessor) EnqueueSendToLedger(data interface{}) {
	o.dataChan <- data
	//resData, err := o.sendToLedger(data)

	//if err != nil {
	//	logrus.WithError(err).Warn("send data to og failed")
	//	return
	//}
	//logrus.WithField("res ", resData).Debug("got response")
}

func (o *OgProcessor) ConsumeQueue() {
outside:
	for {
		logrus.WithField("size", len(o.dataChan)).Debug("og queue size")
		select {
		case data := <-o.dataChan:
			retry := 0
			for ; retry < o.config.RetryTimes; retry++ {
				resData, err := o.sendToLedger(data)
				if err != nil {
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else {
					logrus.WithField("response", resData).Debug("got response")
					break
				}
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
	Data []byte `json:"data"`
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

func (o *OgProcessor) sendToLedger(data interface{}) (resData interface{}, err error) {
	defer timeTrack(time.Now(), "sendToLedger")

	dataBytes, err := json.Marshal(data)
	if err != nil {
		//you should provide a method to marshal json
		panic(err)
		return nil, err
	}
	txReq := TxReq{
		Data: dataBytes,
	}
	dataBytes, err = json.Marshal(txReq)
	if err != nil {
		//you should provide a method to marshal json
		panic(err)
		return nil, err
	}
	req, err := http.NewRequest("POST", o.config.LedgerUrl, bytes.NewBuffer(dataBytes))
	logrus.WithField("data ", string(dataBytes)).Trace("send data to og")

	response, err := o.httpClient.Do(req)

	if err != nil {
		logrus.WithError(err).Warn("send data failed")
		return nil, err
	}
	// Close the connection to reuse it
	defer response.Body.Close()
	// Let's check if the work actually is done
	// We have seen inconsistencies even when we get 200 OK response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithError(err).Fatalf("Couldn't parse response body.")
		return nil, err
	}
	var respj Response
	err = json.Unmarshal(body, &respj)
	if err != nil {
		logrus.WithField("response ",string(body)).WithError(err).Warn("got error from og")
		return respj, err
	}
	if respj.Message != "" {
		err = errors.New(respj.Message)
		logrus.WithError(err).Warn("got error from og")
		return nil, err
	}
	//check code
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("got response code %d ,response status %s", response.StatusCode, response.Status)
		return nil, err
	}
	return respj.Data, nil
}
