package og

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/annchain/BlockDB/ogws"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type OgProcessor struct {
	config                OgProcessorConfig
	dataChan              chan *msgEvent
	quit                  chan bool
	quitFetch             chan bool
	httpClient            *http.Client
	originalDataProcessor ogws.OriginalDataProcessor
}
type OgProcessorConfig struct {
	IdleConnectionTimeout time.Duration
	LedgerUrl             string
	BufferSize            int
	RetryTimes            int
}

type msgEvent struct {
	callbackChan chan error
	data         interface{}
}

func (m *OgProcessor) Stop() {
	m.quit <- true
	m.quitFetch<-true
}

func (m *OgProcessor) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
	go m.ConsumeQueue()
}

func (m *OgProcessor) Name() string {
	return "OgProcessor"
}

func NewOgProcessor(config OgProcessorConfig, originalDataProcessor ogws.OriginalDataProcessor) *OgProcessor {
	_, err := url.Parse(config.LedgerUrl)
	if err != nil {
		panic(err)
	}
	return &OgProcessor{
		config:                config,
		dataChan:              make(chan *msgEvent, config.BufferSize),
		quit:                  make(chan bool),
		httpClient:            createHTTPClient(),
		originalDataProcessor: originalDataProcessor,
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

func (o *OgProcessor) EnqueueSendToLedger(data interface{}) error {
	me := &msgEvent{
		callbackChan: make(chan error),
		data:         data,
	}

	o.dataChan <- me

	// waiting for callback
	select {
	case err := <-me.callbackChan:
		if err != nil {
			return err
		}
		return nil
	}
}

func (o *OgProcessor) ConsumeQueue() {
outside:
	for {
		logrus.WithField("size", len(o.dataChan)).Debug("og queue size")
		select {
		case event := <-o.dataChan:
			retry := 0
			var resData interface{}
			var err error
			for ; retry < o.config.RetryTimes; retry++ {
				resData, err = o.sendToLedger(event.data)
				if err != nil {
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else if resData == nil {
					err = fmt.Errorf("response is nil")
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else {
					logrus.WithField("response", resData).Debug("got response")
					err = o.originalDataProcessor.InsertOne(fmt.Sprintf("%v", resData), event.data)
					if err != nil {
						logrus.WithField("response", resData).Error("write data err")
					}
					break
				}
			}
			if retry == o.config.RetryTimes {
				err = fmt.Errorf("failed to send data to ledger. Abandon. %v", err)
				logrus.WithField("data", event.data).Error("failed to send data to ledger. Abandon.")
			}
			event.callbackChan <- err

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
		logrus.WithField("response ", string(body)).WithError(err).Warnf(
			"got error from og , status %d ,%s ", response.StatusCode, response.Status)
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
