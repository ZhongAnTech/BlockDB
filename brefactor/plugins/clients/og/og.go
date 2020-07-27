package og

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type OgClientConfig struct {
	LedgerUrl  string
	RetryTimes int
}

type OgArchiveResponse struct {
	// TODO: you need to fix the response structure.
	Message string
	Data    interface{}
}

type OgClient struct {
	Config OgClientConfig

	dataChan   chan *core_interface.BlockDBMessage
	quit       chan bool
	httpClient *http.Client
}

func (m *OgClient) Name() string {
	return "OGClient"
}

func (m *OgClient) InitDefault() {
	m.dataChan = make(chan *core_interface.BlockDBMessage)
	m.quit = make(chan bool)
	m.httpClient = createHTTPClient()
}

func (m *OgClient) Stop() {
	m.quit <- true
}

func (m *OgClient) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
	go m.ConsumeQueue()
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 15, // TODO: use config number.
		},
		Timeout: time.Duration(10) * time.Second, // TODO: user config number
	}

	return client
}

func (o *OgClient) ConsumeQueue() {
outside:
	for {
		logrus.WithField("size", len(o.dataChan)).Debug("og queue size")
		select {
		case msg := <-o.dataChan:
			retry := 0
			var resData interface{}
			var err error
			for ; retry < o.Config.RetryTimes; retry++ {
				resData, err = o.sendToLedger(msg)
				// TODO: check the message returned by OG.
				if err != nil {
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else if resData == nil {
					err = fmt.Errorf("response is nil")
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else {
					logrus.WithField("response", resData).Debug("got response")
					// TODO: mark this message as "send ok" in your own task db.
					//err = o.originalDataProcessor.InsertOne(fmt.Sprintf("%v", resData), event.data)
					//if err != nil {
					//	logrus.WithField("response", resData).Error("write data err")
					//}
					break
				}
			}
			// TODO: mark this message as "failed" in your own task db.
			// future queries will come to see if the task succeeded or not
			if retry == o.Config.RetryTimes {
				err = fmt.Errorf("failed to send data to ledger. Abandon. %v", err)
				logrus.WithField("data", msg).Error("failed to send data to ledger. Abandon.")
			}
			//event.callbackChan <- err

		case <-o.quit:
			break outside
		}
	}
	logrus.Info("OgProcessor stopped")
}

func (o *OgClient) EnqueueSendToLedger(command *core_interface.BlockDBMessage) error {
	o.dataChan <- command
	return nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logrus.Debugf("%s took %s", name, elapsed)
}

func (o *OgClient) sendToLedger(message *core_interface.BlockDBMessage) (resData OgArchiveResponse, err error) {
	defer timeTrack(time.Now(), "sendToLedger")

	dataBytes, err := json.Marshal(message)
	if err != nil {
		logrus.WithError(err).Fatal("impl: you should provide a method to marshal json")
	}

	req, err := http.NewRequest("POST", o.Config.LedgerUrl, bytes.NewBuffer(dataBytes))
	logrus.WithField("data ", string(dataBytes)).Trace("send data to og")

	response, err := o.httpClient.Do(req)

	if err != nil {
		logrus.WithError(err).Warn("send data failed")
		return
	}
	// Close the connection to reuse it
	defer response.Body.Close()
	// Let's check if the work actually is done
	// We have seen inconsistencies even when we get 200 OK response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithError(err).Fatalf("Couldn't parse response body.")
		return
	}
	var respj OgArchiveResponse
	err = json.Unmarshal(body, &respj)
	if err != nil {
		logrus.WithField("response ", string(body)).WithError(err).Warnf(
			"got error from og, status %d ,%s ", response.StatusCode, response.Status)
		return respj, err
	}
	if respj.Message != "" {
		err = errors.New(respj.Message)
		logrus.WithError(err).Warn("got error from og")
		return
	}
	//check code
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("got response code %d ,response status %s", response.StatusCode, response.Status)
		return
	}
	return respj, nil
}
