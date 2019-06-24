package og

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.TraceLevel)

}

func TestNewOgProcessor(t *testing.T) {
	p := NewOgProcessor(OgProcessorConfig{LedgerUrl: "http://172.28.152.101:8000//new_archive"})
	p.EnqueueSendToLedger("this is a message")
}
