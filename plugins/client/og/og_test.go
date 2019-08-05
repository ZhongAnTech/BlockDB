package og

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.TraceLevel)

}

type testData struct {
	A int
	B string
	c float64
	D Complex
	F testObjectData
}

type Complex complex128

func (c Complex) MarshalJSON() ([]byte, error) {
	r := real(c)
	i := imag(c)
	var s string
	if i > 0 {
		s = fmt.Sprintf("%f + i%f", r, i)
	} else {
		s = fmt.Sprintf("%f - i%f", real(c), -i)
	}
	return json.Marshal(&s)
}

type testObjectData struct {
	H []byte
	K uint32
	L string
}

func TestNewOgProcessor(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	p := NewOgProcessor(OgProcessorConfig{LedgerUrl: "http://172.28.152.101:8040//new_archive",RetryTimes:3,BufferSize:15,})
	p.Start()
	defer p.Stop()
	p.EnqueueSendToLedger("this is a message")
	data := gettestData()
	p.EnqueueSendToLedger(data)
	time.Sleep(time.Second)
}

func gettestData() *testData {
	data := testData{
		A: 45566,
		B: "what is this ? a message ?, test message",
		c: 56.78,
		D: complex(34.566, 78.9023),
		F: testObjectData{
			H: []byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			K: 67,
			L: "this this a string of test message",
		},
	}
	return &data
}

func TestBatch(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	data := gettestData()
	p := NewOgProcessor(OgProcessorConfig{LedgerUrl: "http://172.28.152.101:8000//new_archive", BufferSize: 100, RetryTimes: 3})
	p.Start()
	defer p.Stop()
	for {
		select {
		case <-time.After(20 * time.Microsecond):
			go p.EnqueueSendToLedger(data)
		}

	}
}
