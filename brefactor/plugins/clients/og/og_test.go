package og

import (
	"context"
	"fmt"
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
		MongoUrl:   "mongodb://nbstock.top:30003",
		LedgerUrl:  "http://nbstock.top:30010/new_archive",
		RetryTimes: 3,
	}

	storageExecutor, err := storage.Connect(context.Background(),"mongodb://nbstock.top:30003", "blockdb", "SCRAM-SHA-256", "rw", "comecome" )
	if err != nil {
		t.Error(err.Error())
	}

	p := OgClient{
		Config: config,
		StorageExecutor: storageExecutor,
	}
	p.Start()
	defer p.Stop()
	fmt.Println(blockMess)

	p.EnqueueSendToLedger(&blockMess)

	//data := gettestData()
	//p.EnqueueSendToLedger(&data)
	//time.Sleep(time.Second)

}
