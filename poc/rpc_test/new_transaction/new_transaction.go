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

	"github.com/annchain/OG/common/crypto"
)

// 2个账户的公钥、私钥
const (
	privKey0 = "0x01a03846356c336844979e7972916b9b045071a57e532457576dd65e89952d9154"
	pubKey0  = "0x010414018666329d97625809f0b8daf7ea052e7dac810dbfcf67ac31d9543bd2937123a984a681b4102c2fce907afc52df4edffb602134fc52e58f8c7b8e04d25dc8"
	privKey1 = "0x01e32e537bd309f0c97ccec33c1d016c3b7e3561760ad574364fdf7ef07fc876ac"
	pubKey1  = "0x010442bd7bd58f46c2c3b5d69d20227c30a8dd937306122722b170c935cbb0319e170cd4af39f3f4e51abbe5910a24925153567cda2961768e99951cfbecd4e489af"

	url                  = "http://47.100.122.212:30022" /* 远程RPC调试 */
	defaultValue         = 0                             /* 缺省转账额 */
	defaultData          = ""                            /* 缺省数据 */
	defaultCryptoType    = "secp256k1"                   /* 加密类型 */
	defaultTokenID       = int64(0)                      /* 缺省安全令牌 */
	transactionRPCMethod = "new_transaction"             /* 交易RPC方法 */
)

// Account 账户字段
type Account struct {
	PublicKey  string /* 公钥，十六进制字符串类型 */
	PrivateKey string /* 私钥，十六进制字符串类型 */
	Nonce      int64  /* 账户发出交易次数 */
	Addr       string /* 地址，十六进制字符串类型 */
}

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

func getAddr(ku string) string {
	pubKey, err := hex.DecodeString(strings.TrimPrefix(ku, "0x"))
	if err != nil {
		fmt.Println(err)
	}
	hexPubKeyHash := hex.EncodeToString(crypto.Keccak256(pubKey))
	return "0x" + hexPubKeyHash[len(hexPubKeyHash)-40:]
}

func newTX(from Account, toAddr string) *TX {
	// TODO 签名
	return &TX{
		Nonce:         from.Nonce,
		FromAddr:      from.Addr,
		ToAddr:        toAddr,
		Value:         strconv.Itoa(defaultValue),
		Data:          defaultData,
		CryptoType:    defaultCryptoType,
		Signature:     "",
		FromPublicKey: from.PublicKey,
		TokenID:       defaultTokenID,
	}
}

func transaction(from Account, toAddr string) string {
	postURL := url + "/" + transactionRPCMethod /* 新交易URL */
	post, err := json.Marshal(newTX(from, toAddr))
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
	from.Nonce++
	return string(respBody)
}

func main() {
	account0 := new(Account)
	account0.PublicKey = pubKey0
	account0.PrivateKey = privKey0
	account0.Nonce = 0
	account0.Addr = getAddr(pubKey0)
	account1 := new(Account)
	account1.PublicKey = pubKey1
	account1.PrivateKey = privKey1
	account1.Nonce = 0
	account1.Addr = getAddr(pubKey1)
	fmt.Println(transaction(*account0, account0.Addr))
}
