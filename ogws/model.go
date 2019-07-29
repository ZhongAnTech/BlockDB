package ogws

import (
	"github.com/annchain/BlockDB/processors"
	"time"
)

type OGMessageList struct {
	Nodes []OGMessage `json:"nodes"`
}

// received from OG
type OGMessage struct {
	Type         int      `json:"type"`
	Hash         string   `json:"hash"`
	ParentsHash  []string `json:"parents_hash"`
	AccountNonce int      `json:"account_nonce"`
	Height       int      `json:"height"`
	PublicKey    string   `json:"public_key"`
	Signature    string   `json:"signature"`
	MineNonce    int      `json:"mine_nonce"`
	Weight       int      `json:"weight"`
	Version      int      `json:"version"`
	DataBase64   string   `json:"data"`
}

// push to MongoDB
type AuditEvent struct {
	Type         int               `json:"type"`
	Hash         string            `json:"hash"`
	ParentsHash  []string          `json:"parents_hash"`
	AccountNonce int               `json:"account_nonce"`
	Height       int               `json:"height"`
	PublicKey    string            `json:"public_key"`
	Signature    string            `json:"signature"`
	MineNonce    int               `json:"mine_nonce"`
	Weight       int               `json:"weight"`
	Version      int               `json:"version"`
	Data         *AuditEventDetail `json:"data"`
}

type AuditEventDetail struct {
	Identity   string      `json:"identity"`
	Type       string      `json:"type"`
	Ip         string      `json:"ip"`
	PrimaryKey string      `json:"primary_key"`
	Timestamp  string      `json:"timestamp"`
	Data       interface{} `json:"data"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
}

func FromLogEvent(l *processors.LogEvent) (a AuditEventDetail) {
	strt := time.Unix(0, l.Timestamp*int64(1000000))

	a = AuditEventDetail{
		Type:       l.Type,
		Data:       l.Data,
		PrimaryKey: l.PrimaryKey,
		Ip:         l.Ip,
		Identity:   l.Identity,
		After:      l.After,
		Before:     l.Before,
		Timestamp:  strt.Format("2006-01-02 15:04:05"),
	}
	return
}
