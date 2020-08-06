package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/W1llyu/ourjson"
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

	url                       = "http://47.100.122.212:30022" /* 远程RPC调试 */
	defaultValue              = "0"                           /* 缺省转账额 */
	defaultData               = "="                           /* 缺省交易数据 */
	defaultCryptoType         = "secp256k1"                   /* 缺省加密类型 */
	defaultTokenID            = 0                             /* 缺省安全令牌 */
	newTransactionRPCMethod   = "new_transaction"             /* 新建交易RPC方法 */
	queryNonceRPCMethod       = "query_nonce"                 /* 查询nonceRPC方法 */
	queryTransactionRPCMethod = "transaction"                 /* 查询交易是否上链 */

	countOfTXPerCoroutine = 10 /* 每个协程发送的交易数目 */
	countOfCoroutine      = 3  /* 协程数目 */
)

var (
	wg        sync.WaitGroup /* 协程等待组 */
	txHashes  chan string    /* 交易哈希 */
	heightMin uint64         /* 记录交易哈希区块的最小高度 */
	heightMax uint64         /* 记录交易哈希区块的最大高度 */
)

// NonceResponse 更新nonce的回复消息
type NonceResponse struct {
	Nonce uint64 `json:"data"`
	Err   string `json:"err"`
}

// TXResponse 交易回复消息
type TXResponse struct {
	Data string `json:"data"`
	Err  string `json:"err"`
}

// UpdateNonce 更新nonce，跟链同步
func UpdateNonce(account *og.SampleAccount) uint64 {
	getURL := url + "/" + queryNonceRPCMethod
	req, err := http.NewRequest("GET", getURL+"?address="+account.Address.String(), nil)
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

// TX 交易
func TX(from *og.SampleAccount, to *og.SampleAccount) []byte {
	postURL := url + "/" + newTransactionRPCMethod
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

// QueryTX 查询交易是否上链
func QueryTX(hash string) bool {
	getURL := url + "/" + queryTransactionRPCMethod
	req, err := http.NewRequest("GET", getURL+"?hash="+hash, nil)
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
	jsonObj, err := ourjson.ParseObject(string(respBody))
	if err != nil {
		fmt.Println(err)
	}
	data := jsonObj.GetJsonObject("data")
	errStr, err := jsonObj.GetString("err")
	if err != nil {
		fmt.Println(err)
	}
	if errStr != "" {
		return false /* 未上链，JSON报错 */
	}
	transaction := data.GetJsonObject("transaction")
	height, err := transaction.GetInt64("height")
	if err != nil {
		fmt.Println(err)
	}
	if uint64(height) < heightMin {
		heightMin = uint64(height)
	}
	if uint64(height) > heightMax {
		heightMax = uint64(height)
	}
	return true
}

// 每个协程的任务
func handle(from *og.SampleAccount, to *og.SampleAccount) {
	defer wg.Done()
	for i := 0; i < countOfTXPerCoroutine; i++ {
		tr := &TXResponse{}
		err := json.Unmarshal(TX(from, to), tr)
		if err != nil {
			print(err)
		}
		if tr.Err != "" {
			print(tr.Err)
			continue
		}
		if tr.Data != "" {
			hash := strings.TrimPrefix(tr.Data, "0x")
			txHashes <- hash
			fmt.Println("TX\t" + hash)
		}
	}
}

func main() {
	account0 := og.NewAccount(privKey0)
	account1 := og.NewAccount(privKey1)
	UpdateNonce(account0)
	txHashes = make(chan string, countOfTXPerCoroutine*countOfCoroutine)
	heightMin = ^uint64(0) /* 最大uint64类型数 */
	heightMax = uint64(0)  /* 最小uint64类型数 */
	for i := 0; i < countOfCoroutine; i++ {
		wg.Add(1)
		go handle(account0, account1)
	}
	wg.Wait()
	fmt.Println("Wait 1 min...")
	time.Sleep(time.Minute * 1)
	countOfTXSucceeded := 0
	len := len(txHashes)
	for i := 0; i < len; i++ {
		if QueryTX(<-txHashes) {
			countOfTXSucceeded++
		}
	}
	fmt.Println("Total number of TX:" + strconv.Itoa(countOfTXPerCoroutine*countOfCoroutine))
	fmt.Println("Succeeded number of TX:" + strconv.Itoa(countOfTXSucceeded))
}
