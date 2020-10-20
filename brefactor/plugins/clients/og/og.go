package og

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"time"
)

const LOADING = 0
const LOAD_SUCC = 1
const LOAD_FAIL = 2

type OgClientConfig struct {
	MongoUrl   string
	LedgerUrl  string
	RetryTimes int
}

type OgArchiveResponse struct {
	// TODO: you need to fix the response structure.
	Message string
	Data    interface{}
}

type TxReq struct {
	Data []byte `json:"data"`
}

type OgClient struct {
	Config          *OgClientConfig
	StorageExecutor core_interface.StorageExecutor

	dataChan   chan *core_interface.BlockDBMessage
	quit       chan bool
	httpClient *http.Client
}

func (m *OgClient) Name() string {
	return "OGClient"
}

func (m *OgClient) InitDefault() {
	m.dataChan = make(chan *core_interface.BlockDBMessage, 30)
	m.quit = make(chan bool)
	m.httpClient = createHTTPClient()
}

func (m *OgClient) Stop() {
	m.quit <- true
}

//func Connection(url string) string {
//	resp, err := http.Get(url)
//	if err != nil {
//		fmt.Printf("http.Get()函数执行错误,错误为:%v\n", err)
//	}
//	defer resp.Body.Close()
//
//	body, err := ioutil.ReadAll(resp.Body)
//
//	if err != nil {
//		fmt.Printf("ioutil.ReadAll()函数执行出错,错误为:%v\n", err)
//	}
//	fmt.Println("connection succ..",string(body))
//	return string(string(body))
//}

func (m *OgClient) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
	go m.ConsumeQueue()
	go m.Reload()

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

type isOnChain struct {
	TxHash string `json:"tx_hash"`
	OpHash string `json:"op_hash"`
	//0:正在上链， 1:已经上链，2:上链失败
	Status int `json:"status"`
}

func (o *OgClient) Reload() {
	//取出上链失败的重新上链
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	selectResponse, err := o.StorageExecutor.Select(ctx, "dataToOG", nil, nil, 10, 0)
	if err != nil {
		fmt.Println("ERR: ", err)
	}
	if selectResponse.Content != nil {
		for _, result := range selectResponse.Content {
			a := core_interface.BlockDBMessage{}
			bsonBytes, _ := bson.Marshal(result)
			bson.Unmarshal(bsonBytes, &a)
			o.dataChan <- &a
		}
	}
}

func (o *OgClient) ConsumeQueue() {
	// TODO: adapt mongo to
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	//mgo3,_:= o.StorageExecutor.Insert(ctx,"op",bson.M{})
	o.StorageExecutor.CreateCollection(ctx, "dataToOG")
	o.StorageExecutor.CreateCollection(ctx, "isOnChain")
outside:
	for {
		logrus.WithField("size", len(o.dataChan)).Debug("og queue size")
		select {
		case msg := <-o.dataChan:
			//need to save msg in mongodb
			fmt.Println("start to consume queue")
			fmt.Println("receive msg from EnqueueSendToLedger: ",msg)
			id, err := o.StorageExecutor.Insert(ctx, "dataToOG", bson.M{
				//{"tx_hash",msg.TxHash},
				"public_key": msg.PublicKey,
				"signature":  msg.Signature,
				"op_hash":    msg.OpHash,
				"op_str":     msg.Data,
			})
			fmt.Println("######", id)

			retry := 0
			var resData OgArchiveResponse
			for ; retry < o.Config.RetryTimes; retry++ {

				resData, err = o.sendToLedger(msg)
				//需要重新上链而非break
				if resData.Data == nil {
					fmt.Println("return msg: ", resData.Message)

				} else {
					//txhash-ophash 存入isOnchain集合中
					txHash := resData.Data.(string)
					fmt.Println(".....", txHash)
					isOn := &isOnChain{
						TxHash: txHash,
						OpHash: msg.OpHash,
						Status: LOADING,
					}

					o.StorageExecutor.Insert(ctx, "isOnChain", bson.M{
						"tx_hash": txHash,
						"op_hash": isOn.OpHash,
						"status":  isOn.Status,
					})
				}

				// TODO: check the message returned by OG.
				if err != nil {
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else if resData.Data == nil {
					err = fmt.Errorf("response is nil")
					logrus.WithField("retry", retry).WithError(err).Warnf("failed to send to ledger")
				} else {
					logrus.WithField("response", resData).Debug("got response")
					// TODO: mark this message as "send ok" in your own task db.
					o.StorageExecutor.Delete(ctx, "dataToOG", id)

				}
			}

			// TODO: mark this message as "failed" in your own task db.
			// future queries will come to see if the task succeeded or not
			if retry == o.Config.RetryTimes {
				err = fmt.Errorf("failed to send data to ledger. Abandon. %v", err)
				logrus.WithField("data", msg).Error("failed to send data to ledger. Abandon.")

				//上链失败更新到isOnChain
				io := isOnChain{
					TxHash: resData.Data.(string),
					OpHash: msg.OpHash,
					Status: LOAD_FAIL,
				}

				o.StorageExecutor.Update(ctx, "isOnChain", bson.M{"tx_hash": io.TxHash, "op_hash": io.OpHash, "status": 0}, bson.M{"tx_hash": io.TxHash, "op_hash": io.OpHash, "status": 2}, "set")

			}
			//event.callbackChan <- err

		case <-o.quit:
			break outside
		}

	}
	logrus.Info("OgProcessor stopped")
}

// TODO: Wu jianhang call this method
func (o *OgClient) EnqueueSendToLedger(command *core_interface.BlockDBMessage) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	o.StorageExecutor.CreateCollection(ctx, "_op")
	fmt.Println("COMMAND:", command)
	command.Data = base64.StdEncoding.EncodeToString([]byte(command.Data))
	o.dataChan <- command
	fmt.Println(command)
	fmt.Println(len(o.dataChan))
	return nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logrus.Debugf("%s took %s", name, elapsed)
}

func (o *OgClient) sendToLedger(message *core_interface.BlockDBMessage) (resData OgArchiveResponse, err error) {
	defer timeTrack(time.Now(), "sendToLedger")

	dataBytes, err := json.Marshal(message)
	fmt.Println(dataBytes)
	if err != nil {
		logrus.WithError(err).Fatal("impl: you should provide a method to marshal json")
	}

	txReq := TxReq{
		Data: dataBytes,
	}
	dataBytes, err = json.Marshal(txReq)
	fmt.Println("dataBytes:",dataBytes)
	req, err := http.NewRequest("POST", o.Config.LedgerUrl, bytes.NewBuffer(dataBytes))
	logrus.WithField("data ", string(dataBytes)).Trace("send data to og")
	//返回*response，关于连接的信息
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
	fmt.Println("*********", respj)
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
