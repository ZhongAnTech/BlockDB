package og

import (
	"context"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/ZhongAnTech/BlockDB/brefactor/storage"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewOgProcessor(t *testing.T) {

	var blockMess = core_interface.BlockDBMessage{
		OpHash:    "0x3475623705236",
		Signature: "0x169153474351324",
		PublicKey: "0x769153474351324",
		Data:      `{"op":"insert","collection":"sample_collection","op_data":{"name":"fudan","address":{"city":"Shanghai","road":"xxx"},"logo":{"url":"http://a.png"},"teachers":["T1","T2","T3",]}}`,
	}

	logrus.SetLevel(logrus.TraceLevel)
	config := &OgClientConfig{
		MongoUrl:   "mongodb://localhost:27017",
		LedgerUrl:  "http://nbstock.top:30010/new_archive",
		RetryTimes: 3,

	}

	storageExecutor, err := storage.Connect(context.Background(),"mongodb://localhost:27017", "blockdb", "","nil","" )
	if err != nil {
		t.Error(err.Error())
	}

	p := OgClient{
		Config: config,
		StorageExecutor: storageExecutor,
		dataChan: make(chan *core_interface.BlockDBMessage,10),
		httpClient: createHTTPClient(),
	}
	p.EnqueueSendToLedger(&blockMess)
	p.Start()
	p.ConsumeQueue()





	//data := gettestData()
	//p.EnqueueSendToLedger(&data)
	//time.Sleep(time.Second)

}
