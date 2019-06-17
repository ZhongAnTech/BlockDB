package og

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/httplib"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/url"
	"time"
)

type OgProcessor struct {
	config OgProcessorConfig
}
type OgProcessorConfig struct {
	IdleConnectionTimeout time.Duration
	LedgerUrl             string
}

func (m *OgProcessor) Stop() {
	logrus.Info("OgProcessor stopped")
}

func (m *OgProcessor) Start() {
	logrus.Info("OgProcessor started")
	// start consuming queue
}

func (m *OgProcessor)Name()string{
	return "OgProcessor"
}

func NewOgProcessor(config OgProcessorConfig) *OgProcessor {
	_, err := url.Parse(config.LedgerUrl)
	if err != nil {
		panic(err)
	}
	return &OgProcessor{
		config: config,
	}
}

func (o *OgProcessor) SendToLedger(data []byte) {
	go o.sendToLedger(data)
}

type TxReq struct {
	Data []byte `json:"data"`
}

func (o *OgProcessor) sendToLedger(data []byte) {
	req := httplib.Post(o.config.LedgerUrl)
	req.SetTimeout(time.Second*10, time.Second*10)
	txReq := TxReq{
		Data: data,
	}
	_, err := req.JSONBody(&txReq)
	if err != nil {
		panic(fmt.Errorf("encode tx errror %v", err))
	}
	d, _ := json.MarshalIndent(&txReq, "", "\t")
	fmt.Println(string(d))

	str, err := req.String()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(str)
}

func (m *OgProcessor) ProcessConnection(conn net.Conn) error {
	// 1, parse command
	// 2, dispatch the command to every interested parties
	//    including chain logger and the real backend mongoDB server
	// 3, response to conn
	for {
		conn.SetReadDeadline(time.Now().Add(m.config.IdleConnectionTimeout))
		bytes, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				logrus.Info("target closed")
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				logrus.Info("target timeout")
				conn.Close()
				return nil
			}
			return err
		}
		// query command
		fmt.Println(hex.Dump(bytes))
		m.SendToLedger(bytes)
	}
	return nil
}
