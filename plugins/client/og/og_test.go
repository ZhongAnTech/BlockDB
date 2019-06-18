package og

import "testing"

func TestNewOgProcessor(t *testing.T) {
	p := NewOgProcessor(OgProcessorConfig{LedgerUrl: "http://172.28.152.101:8000//new_archive"})
	p.sendToLedger("this is a message")
}
