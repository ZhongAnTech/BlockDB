package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/annchain/BlockDB/poc/rpc_test/og"
	"github.com/annchain/OG/common"
	"github.com/annchain/OG/common/crypto"
	"github.com/annchain/OG/common/math"
	"github.com/annchain/OG/rpc"
	"github.com/annchain/OG/types"
	"github.com/annchain/OG/types/tx_types"
)

const (
	// 2个账户的私钥
	privKey0 = "0x01a03846356c336844979e7972916b9b045071a57e532457576dd65e89952d9154"
	privKey1 = "0x01e32e537bd309f0c97ccec33c1d016c3b7e3561760ad574364fdf7ef07fc876ac"

	url                  = "http://47.100.122.212:30022" /* 远程RPC调试 */
	defaultValue         = "0"                           /* 缺省转账额 */
	defaultData          = "="                           /* 缺省交易数据 */
	defaultCryptoType    = "secp256k1"                   /* 缺省加密类型 */
	defaultTokenID       = 0                             /* 缺省安全令牌 */
	transactionRPCMethod = "new_transaction"             /* 新建交易RPC方法 */
	nonceRPCMethod       = "query_nonce"                 /* 查询nonceRPC方法 */

	countOfTXPerCoroutine = 10 /* 每个协程发送的交易数目 */
	countOfCoroutine      = 3  /* 协程数目 */
)

var (
	wg sync.WaitGroup
)

// NonceResponse 更新nonce的回复消息
type NonceResponse struct {
	Nonce uint64 `json:"data"`
	Err   string `json:"err"`
}

// TransactionResponse 交易回复消息
type TransactionResponse struct {
	Data string `json:"data"`
	Err  string `json:"err"`
}

// UpdateNonce 更新nonce，跟链同步
func UpdateNonce(account *og.SampleAccount) uint64 {
	postURL := url + "/" + nonceRPCMethod
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

// NewTX 新建交易
func NewTX(from *og.SampleAccount, to *og.SampleAccount) *rpc.NewTxRequest {
	// Nonce
	nonce, err := from.ConsumeNonce() /* 消费nonce */
	if err != nil {
		fmt.Println(err)
	}
	// Signature
	signer := crypto.NewSigner(from.PrivateKey.Type)
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
	sig := signer.Sign(from.PrivateKey, tx.SignatureTargets()) /* 从私钥生成签名 */
	return &rpc.NewTxRequest{
		Nonce:      nonce,
		From:       from.Address.String(),
		To:         to.Address.String(),
		Value:      defaultValue,
		Data:       defaultData,
		CryptoType: defaultCryptoType,
		Signature:  hex.EncodeToString(sig.Bytes),
		Pubkey:     hex.EncodeToString(from.PublicKey.Bytes),
		TokenId:    defaultTokenID,
	}
}

// Transaction 交易
func Transaction(from *og.SampleAccount, to *og.SampleAccount) []byte {
	postURL := url + "/" + transactionRPCMethod
	post, err := json.Marshal(NewTX(from, to))
	if err != nil {
		fmt.Println(err)
	}
	postBuffer := bytes.NewBuffer(post)
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
	return respBody
}

// 每个协程的任务
func handle(from *og.SampleAccount, to *og.SampleAccount) {
	defer wg.Done()
	for i := 0; i < countOfTXPerCoroutine; i++ {
		tr := &TransactionResponse{}
		err := json.Unmarshal(Transaction(from, to), tr)
		if err != nil {
			print(err)
		}
		if tr.Err != "" {
			print(tr.Err)
			continue
		}
		if tr.Data != "" {
			fmt.Println(strings.TrimPrefix(tr.Data, "0x"))
		}
	}
}

func main() {
	account0 := og.NewAccount(privKey0)
	account1 := og.NewAccount(privKey1)
	UpdateNonce(account0)
	for i := 0; i < countOfCoroutine; i++ {
		wg.Add(1)
		go handle(account0, account1)
	}
	wg.Wait()
}
