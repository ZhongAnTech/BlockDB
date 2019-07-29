package ogws

import "github.com/annchain/BlockDB/processors"

type OGMessageList struct {
	Nodes []OGMessage `json:"nodes"`
}

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

type AuditEvent struct {
	Type         int                  `json:"type"`
	Hash         string               `json:"hash"`
	ParentsHash  []string             `json:"parents_hash"`
	AccountNonce int                  `json:"account_nonce"`
	Height       int                  `json:"height"`
	PublicKey    string               `json:"public_key"`
	Signature    string               `json:"signature"`
	MineNonce    int                  `json:"mine_nonce"`
	Weight       int                  `json:"weight"`
	Version      int                  `json:"version"`
	Data         *processors.LogEvent `json:"data"`
}
