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
	url               = "http://47.100.122.212:30022" /* 远程RPC调试 */
	defaultValue      = "0"                           /* 缺省转账额 */
	defaultData       = "="                           /* 缺省交易数据 */
	defaultCryptoType = "secp256k1"                   /* 缺省加密类型 */
	defaultTokenID    = 0                             /* 缺省安全令牌 */

	newAccountRPCMethod        = "new_account"        /* 新建账户RPC方法 */
	newTransactionRPCMethod    = "new_transaction"    /* 新建交易RPC方法 */
	queryNonceRPCMethod        = "query_nonce"        /* 查询nonceRPC方法 */
	queryTransactionRPCMethod  = "transaction"        /* 查询单笔交易RPC方法 */
	queryTransactionsRPCMethod = "transaction_hashes" /* 查询多笔交易RPC方法 */
	querySequencerRPCMethod    = "sequencer"          /* 查询区块信息RPC方法 */

	countOfTXPerCoroutine = 10  /* 每个协程发送的交易数目 */
	countOfCoroutine      = 100 /* 协程数目 */
)

// AccountReq 新建账户请求
type AccountReq struct {
	Algorithm string `json:"algorithm"` /* 加密算法 */
}

// AccountPair 账户对
type AccountPair struct {
	From *og.SampleAccount /* 发送账户 */
	To   *og.SampleAccount /* 接收账户 */
}

// NonceResponse 更新nonce的回复消息
type NonceResponse struct {
	Nonce uint64 `json:"data"` /* nonce */
	Err   string `json:"err"`  /* 错误 */
}

// TXResponse 交易回复消息
type TXResponse struct {
	Data string `json:"data"` /* 哈希 */
	Err  string `json:"err"`  /* 错误 */
}

// TXInfo 交易信息
type TXInfo struct {
	Hash     string /* 哈希 */
	InitTime int    /* 新建时间时间戳 */
}

var (
	accountPairs      []AccountPair  /* 账户对 */
	wg                sync.WaitGroup /* 协程等待组 */
	tx                chan TXInfo    /* 交易哈希 */
	timestampOfHeight map[int]int    /* 键：区块高度，值：时间戳 */
	heightMin         int            /* 交易最小高度 */
	heightMax         int            /* 交易最大高度 */
	initFlag          bool           /* 交易最小、最大高度是否被初始化 */
)

// NewAccountReq 新建账户请求
func NewAccountReq() *AccountReq {
	return &AccountReq{
		Algorithm: "secp256k1",
	}
}

// NewAccountPair 新建账户对
func NewAccountPair(fromPrivKey string, toPrivKey string) *AccountPair {
	return &AccountPair{
		From: og.NewAccount(fromPrivKey),
		To:   og.NewAccount(toPrivKey),
	}
}

// NewTXInfo 新建交易信息
func NewTXInfo(hash string, timestamp int) *TXInfo {
	return &TXInfo{
		Hash:     hash,
		InitTime: timestamp,
	}
}

// NewAccount 新建账户，返回私钥
func NewAccount() string {
	postURL := url + "/" + newAccountRPCMethod
	post, err := json.Marshal(NewAccountReq())
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
	jsonObj, err := ourjson.ParseObject(string(respBody[:]))
	if err != nil {
		fmt.Println(err)
	}
	errStr, err := jsonObj.GetString("err")
	if err != nil {
		fmt.Println(err)
	}
	if errStr != "" {
		fmt.Println(errStr)
		return ""
	}
	data := jsonObj.GetJsonObject("data")
	kr, err := data.GetString("privkey")
	if err != nil {
		fmt.Println(err)
	}
	return kr
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

// QuerySequencerTimestamp 查询指定高度区块时间戳
func QuerySequencerTimestamp(height int) int {
	timestamp, ok := timestampOfHeight[height]
	if ok {
		return timestamp
	}
	getURL := url + "/" + querySequencerRPCMethod
	req, err := http.NewRequest("GET", getURL+"?seq_id="+strconv.Itoa(height), nil)
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
	timestamp, err = data.GetInt("Timestamp")
	if err != nil {
		fmt.Println(err)
	}
	timestampOfHeight[height] = timestamp
	if !initFlag {
		heightMin = height
		heightMax = height
		initFlag = true
	} else {
		if height < heightMin {
			heightMin = height
		}
		if height > heightMax {
			heightMax = height
		}
	}
	return timestamp
}

// QueryTX 查询交易相关信息，返回是否上链、区块时间戳
func QueryTX(hash string) (bool, int) {
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
	errStr, err := jsonObj.GetString("err")
	if err != nil {
		fmt.Println(err)
	}
	if errStr != "" {
		return false, 0 /* 未上链，JSON报错 */
	}
	data := jsonObj.GetJsonObject("data")
	transaction := data.GetJsonObject("transaction")
	height, err := transaction.GetInt("height")
	if err != nil {
		fmt.Println(err)
	}
	return true, QuerySequencerTimestamp(height)
}

// QueryCountOfTX 查询指定高度区块交易数目
func QueryCountOfTX(height int) int {
	getURL := url + "/" + queryTransactionsRPCMethod
	req, err := http.NewRequest("GET", getURL+"?height="+strconv.Itoa(height), nil)
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
	txHashes := data.GetJsonArray("hashes")
	if err != nil {
		fmt.Println(err)
	}
	return len(txHashes.Values())
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
			tx <- *NewTXInfo(hash, int(time.Now().UnixNano()/1e6))
			fmt.Println("TX\t" + hash)
		}
	}
}

// 每秒交易数、上链延迟、上链成功率、查询延迟，并发查询性能
func main() {
	accountPairs = make([]AccountPair, countOfCoroutine)
	for i := 0; i < countOfCoroutine; i++ {
		accountPairs[i] = *NewAccountPair(NewAccount(), NewAccount())
		UpdateNonce(accountPairs[i].From)
	}
	tx = make(chan TXInfo, countOfTXPerCoroutine*countOfCoroutine)
	timestampOfHeight = make(map[int]int)
	initFlag = false
	for i := 0; i < countOfCoroutine; i++ {
		wg.Add(1)
		go handle(accountPairs[i].From, accountPairs[i].To)
	}
	wg.Wait()
	fmt.Println("Wait 1 min...")
	time.Sleep(time.Minute * 1)

	countOfTXSucceeded := 0 /* 上链成功数 */
	delaySum := 0           /* 上链延迟总和，单位：毫秒 */
	delayNum := 0
	len := len(tx)
	for i := 0; i < len; i++ {
		txInfo := <-tx
		ok, timestamp := QueryTX(txInfo.Hash)
		if ok {
			countOfTXSucceeded++
			if timestamp >= txInfo.InitTime {
				delaySum += (timestamp - txInfo.InitTime)
				delayNum++
			}
		}
	}
	count := 0
	if heightMax-heightMin > 1 /* 交易至少完全填充了1个区块，可以计算TPS */ {
		for i := heightMin + 1; i < heightMax; i++ {
			count += QueryCountOfTX(i)
		}
		tps := float64(count*1000) / float64(QuerySequencerTimestamp(heightMax-1)-QuerySequencerTimestamp(heightMin))
		fmt.Println(tps)
	} else {
		fmt.Println("TPS: -")
	}
	fmt.Println("Average delay:" + strconv.Itoa(delaySum/delayNum) + "ms")
	fmt.Println("Total number of TX:" + strconv.Itoa(countOfTXPerCoroutine*countOfCoroutine))
	fmt.Println("Succeeded number of TX:" + strconv.Itoa(countOfTXSucceeded))
}
