package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/annchain/BlockDB/poc/rpc_test/og"
	"github.com/annchain/OG/common"
	"github.com/annchain/OG/common/crypto"
	"github.com/annchain/OG/common/math"
	"github.com/annchain/OG/rpc"
	"github.com/annchain/OG/types"
	"github.com/annchain/OG/types/tx_types"
	"io/ioutil"
	"net/http"
)

// 2个账户的公钥、私钥
const (
	privKey0 = "0x01a03846356c336844979e7972916b9b045071a57e532457576dd65e89952d9154"
	pubKey0  = "0x010414018666329d97625809f0b8daf7ea052e7dac810dbfcf67ac31d9543bd2937123a984a681b4102c2fce907afc52df4edffb602134fc52e58f8c7b8e04d25dc8"
	privKey1 = "0x01e32e537bd309f0c97ccec33c1d016c3b7e3561760ad574364fdf7ef07fc876ac"
	pubKey1  = "0x010442bd7bd58f46c2c3b5d69d20227c30a8dd937306122722b170c935cbb0319e170cd4af39f3f4e51abbe5910a24925153567cda2961768e99951cfbecd4e489af"

	url                  = "http://localhost:8000" /* 远程RPC调试 */
	defaultValue         = 0                       /* 缺省转账额 */
	defaultData          = ""                      /* 缺省数据 */
	defaultCryptoType    = "secp256k1"             /* 加密类型 */
	defaultTokenID       = int64(0)                /* 缺省安全令牌 */
	transactionRPCMethod = "new_transaction"       /* 交易RPC方法 */
	nonceMethod          = "query_nonce"
)

//// Account 账户字段
//type Account struct {
//	PublicKey  string /* 公钥，十六进制字符串类型 */
//	PrivateKey string /* 私钥，十六进制字符串类型 */
//	Nonce      int64  /* 账户发出交易次数 */
//	Addr       string /* 地址，十六进制字符串类型 */
//}

// TX 交易字段
type TX struct {
	Nonce         int64  `json:"nonce"`       /* 支出账户发出交易次数 */
	FromAddr      string `json:"from"`        /* 支出账户地址 */
	ToAddr        string `json:"to"`          /* 收入账户地址 */
	Value         string `json:"value"`       /* 交易额 */
	Data          string `json:"data"`        /* 交易规则 */
	CryptoType    string `json:"crypto_type"` /* 加密类型 */
	Signature     string `json:"signature"`   /* 签名 */
	FromPublicKey string `json:"pubkey"`      /* 支出账户公钥，十六进制字符串类型 */
	TokenID       int64  `json:"token_id"`    /* 安全令牌 */
}

func newTX(from *og.SampleAccount, to *og.SampleAccount) *rpc.NewTxRequest {
	signer := crypto.NewSigner(from.PrivateKey.Type)

	// TODO 签名，query nonce
	// consume nonce
	nonce, err := from.ConsumeNonce()
	if err != nil {
		panic(err)
	}

	// This is an OG Tx to show what signature target is.
	// No need to send all stuff to OG.
	tx := tx_types.Tx{
		TxBase: types.TxBase{
			Type:         0,
			Hash:         common.Hash{},
			ParentsHash:  nil,
			AccountNonce: nonce,
			Height:       0,
			PublicKey:    nil,
			Signature:    nil,
			MineNonce:    0,
			Weight:       0,
			Version:      0,
		},
		From:    &from.Address,
		To:      to.Address,
		Value:   math.NewBigInt(0),
		TokenId: 0,
		Data:    nil,
	}

	fmt.Println(hex.EncodeToString(tx.SignatureTargets()))
	fmt.Println(hex.EncodeToString(from.PublicKey.Bytes))
	sig := signer.Sign(from.PrivateKey, tx.SignatureTargets())

	// This is an OG rpc Tx to show what should be sent to OG.
	return &rpc.NewTxRequest{
		Nonce:      nonce,
		From:       from.Address.String(),
		To:         to.Address.String(),
		Value:      "0",
		Data:       "=",
		CryptoType: defaultCryptoType,
		Signature:  hex.EncodeToString(sig.Bytes),
		Pubkey:     hex.EncodeToString(from.PublicKey.Bytes),
		TokenId:    0,
	}
}

func transaction(from *og.SampleAccount, to *og.SampleAccount) string {
	postURL := url + "/" + transactionRPCMethod /* 新交易URL */
	post, err := json.Marshal(newTX(from, to))
	if err != nil {
		fmt.Println(err)
	}
	postBuffer := bytes.NewBuffer(post)
	fmt.Println(postBuffer.String())
	req, err := http.NewRequest("POST", postURL, postBuffer)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return string(respBody)
}

type NonceResponse struct {
	Nonce uint64 `json:"data"`
	Err   string `json:"err"`
}

func updateNonce(account *og.SampleAccount) uint64 {
	postURL := url + "/" + nonceMethod /* 新交易URL */

	req, err := http.NewRequest("GET", postURL+"?address="+account.Address.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(respBody))
	nr := &NonceResponse{}
	err = json.Unmarshal(respBody, nr)
	if err != nil {
		print(err)
	}
	if nr.Err != "" {
		print(nr.Err)
	}
	account.SetNonce(nr.Nonce)
	return nr.Nonce
}

func main() {
	account0 := og.NewAccount(privKey0)
	account1 := og.NewAccount(privKey1)

	updateNonce(account0)
	// update nonce first.
	fmt.Println("result " + transaction(account0, account1))
}
