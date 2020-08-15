package ogws

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/annchain/BlockDB/processors"
	"github.com/latifrons/gorews"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type OGWSClient struct {
	url         *url.URL
	auditWriter AuditWriter
	client      *gorews.GorewsClient
	OGHeight    int
}

func (o *OGWSClient) SetHeight(height int) {
	o.OGHeight = height
}

func NewOGWSClient(ustr string, auditWriter AuditWriter) *OGWSClient {
	// connect to ws server
	u, err := url.Parse(ustr)
	if err != nil {
		logrus.WithField("url", viper.GetString("og.wsclient.url")).Fatal("cannot parse ogws client")
	}

	return &OGWSClient{
		url:         u,
		auditWriter: auditWriter,
	}
}

func (o *OGWSClient) Start() {

	logrus.WithField("url", o.url).Info("connecting to ws")

	o.client = gorews.NewGorewsClient()
	var headers http.Header
	err := o.client.Start(o.url.String(), headers, time.Second*5, time.Second*5, time.Second*5)
	if err != nil {
		logrus.WithError(err).Fatal("init ws client")
	}

	logrus.WithField("url", o.url).Info("connected to ws")

	o.client.Outgoing <- []byte("{\"event\":\"new_tx\"}")

	go func() {
		for {
			msg := <-o.client.Incoming
			_, err := o.handleMessage(msg)
			if err != nil {
				logrus.WithError(err).Warn("failed to handle message: " + string(msg))
			}
		}
	}()
}

func (o *OGWSClient) Stop() {
	o.client.Stop()
}

func (OGWSClient) Name() string {
	return "OGWSClient"
}

const (
	TxTypeSequencer = 1
	TxTypeArchive   = 4
)

func (o *OGWSClient) handleMessage(bytes []byte) (result OGMessageList, err error) {
	var ogmss OGMessageList
	err = json.Unmarshal(bytes, &ogmss)
	if err != nil {
		return
	}
	for _, ogms := range ogmss.Nodes {
		switch ogms.Type {
		case TxTypeSequencer:
			o.SetHeight(ogms.Height)
			logrus.Info("height is updated")
		case TxTypeArchive:
			// base64 decode
			dataBytes, err := base64.StdEncoding.DecodeString(ogms.DataBase64)
			if err != nil {
				logrus.WithError(err).Warn("failed to decode base64 string. Skip this event.")
				continue
			}
			if o.OGHeight < ogms.Height-2 {
				o.SetHeight(ogms.Height)
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
			o.auditWriter.GetOriginalDataProcessor().DeleteOne(ogms.Hash)
			logrus.WithField("event timestamp", auditEvent.Height).Debug("wrote audit event")

		default:
			continue
		}
	}
	return
}
