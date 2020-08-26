package og

import (
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestNewOgProcessor(t *testing.T) {

	var blockMess = core_interface.BlockDBMessage{
		OpHash:    "0x3475623705236",
		Signature: "0x169153474351324",
		PublicKey: "0x769153474351324",
		Data:      `{"op":"insert","collection":"sample_collection","op_data":{"name":"fudan","address":{"city":"Shanghai","road":"xxx"},"logo":{"url":"http://a.png"},"teachers":["T1","T2","T3",]}}`,
	}

	logrus.SetLevel(logrus.TraceLevel)
	p := NewOgClient(OgClientConfig{LedgerUrl: "http://nbstock.top:30022/new_archive", RetryTimes: 3})
	p.Start()
	defer p.Stop()
	fmt.Println(blockMess)

	p.EnqueueSendToLedger(&blockMess)

	//data := gettestData()
	//p.EnqueueSendToLedger(&data)
	time.Sleep(time.Second)

}
