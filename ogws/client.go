package ogws

import (
	"encoding/base64"
	"encoding/json"
	"github.com/annchain/BlockDB/processors"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"time"
)

type OGWSClient struct {
	url         *url.URL
	quit        chan bool
	auditWriter AuditWriter
}

func NewOGWSClient(ustr string, auditWriter AuditWriter) *OGWSClient {
	// connect to ws server
	u, err := url.Parse(ustr)
	if err != nil {
		logrus.WithField("url", viper.GetString("og.wsclient.url")).Fatal("cannot parse ogws client")
	}

	return &OGWSClient{
		url:         u,
		quit:        make(chan bool),
		auditWriter: auditWriter,
	}
}

func (o *OGWSClient) Start() {

	logrus.WithField("url", o.url).Info("connecting to ws")

	c, _, err := websocket.DefaultDialer.Dial(o.url.String(), nil)
	if err != nil {
		logrus.WithError(err).Fatal("dial ws")
	}
	logrus.WithField("url", o.url).Info("connected to ws")

	err = c.WriteMessage(websocket.TextMessage, []byte("{\"event\":\"new_tx\"}"))
	if err != nil {
		logrus.WithError(err).Fatal("init ws")
	}

	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			o.handleMessage(message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			//err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			//if err != nil {
			//	log.Println("write:", err)
			//	return
			//}
		case <-o.quit:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (o *OGWSClient) Stop() {
	o.quit <- true
}

func (OGWSClient) Name() string {
	return "OGWSClient"
}

const TxTypeArchive = 4

func (o *OGWSClient) handleMessage(bytes []byte) (result OGMessageList, err error) {
	var ogmss OGMessageList
	err = json.Unmarshal(bytes, &ogmss)
	if err != nil {
		return
	}
	for _, ogms := range ogmss.Nodes {
		if ogms.Type != TxTypeArchive {
			continue
		}
		// base64 decode
		dataBytes, err := base64.StdEncoding.DecodeString(ogms.DataBase64)
		if err != nil {
			logrus.WithError(err).Warn("failed to decode base64 string. Skip this event.")
			continue
		}

		var logEvent processors.LogEvent
		err = json.Unmarshal(dataBytes, &logEvent)
		if err != nil {
			logrus.WithError(err).Warn("failed to decode logEvent. Skip this event.")
			continue
		}

		auditEventDetail := FromLogEvent(&logEvent)

		auditEvent := &AuditEvent{
			Signature:    ogms.Signature,
			Type:         ogms.Type,
			PublicKey:    ogms.PublicKey,
			AccountNonce: ogms.AccountNonce,
			Hash:         ogms.Hash,
			Height:       ogms.Height,
			MineNonce:    ogms.MineNonce,
			ParentsHash:  ogms.ParentsHash,
			Version:      ogms.Version,
			Weight:       ogms.Weight,
			Data:         &auditEventDetail,
		}
		err = o.auditWriter.WriteOGMessage(auditEvent)
		if err != nil {
			logrus.WithError(err).Warn("failed to write ledger.")
			continue
		}
	}
	return
}
