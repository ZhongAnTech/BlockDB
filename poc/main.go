package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/W1llyu/ourjson"
	"github.com/annchain/OG/common/crypto"
)

var (
	url = "http://localhost:8000"
)

const (
	defaultValue         = 1                 /* 缺省转账额 */
	defaultData          = ""                /* 缺省数据 */
	defaultCryptoType    = "secp256k1"       /* 加密类型 */
	defaultTokenID       = int64(0)          /* 缺省安全令牌 */
	defaultSeed          = ""                /* 缺省随机种子 */
	accountRPCMethod     = "new_account"     /* 新建账户RPC方法 */
	transactionRPCMethod = "new_transaction" /* 交易RPC方法 */

	countOfAccounts = 8       /* 账户数目 */
	maxConcurrency  = 1000000 /* 最大并发数 */
)

// AccountReq 新建账户请求字段
type AccountReq struct {
	Algorithm string `json:"algorithm"` /* 加密算法 */
}

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

func newAccountReq() *AccountReq {
	return &AccountReq{
		Algorithm: defaultCryptoType,
	}
}

func newAccount() *Account {
	// 发送新建账户请求
	postURL := url + "/" + accountRPCMethod /* 新建账户URL */
	post, err := json.Marshal(newAccountReq())
	if err != nil {
		fmt.Println(err)
	}
	postBuffer := bytes.NewBuffer(post)
	req, err := http.NewRequest("POST", postURL, postBuffer)
	if err != nil {
		fmt.Println(err)
	}
	// 接收账户私钥和公钥
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
	data := jsonObj.GetJsonObject("data")
	kr, err := data.GetString("privkey") /* 私钥，十六进制字符串类型 */
	if err != nil {
		fmt.Println(err)
	}
	ku, err := data.GetString("pubkey") /* 公钥，十六进制字符串类型 */
	if err != nil {
		fmt.Println(err)
	}
	// 已知公钥算地址
	pubKey, err := hex.DecodeString(ku)
	if err != nil {
		fmt.Println(err)
	}
	hexPubKeyHash := hex.EncodeToString(crypto.Keccak256(pubKey))
	addr := "0x" + hexPubKeyHash[len(hexPubKeyHash)-40:]
	return &Account{
		PublicKey:  ku,
		PrivateKey: kr,
		Nonce:      0,
		Addr:       addr,
	}
}

func newTX(from Account, toAddr string) *TX {
	// 已知字符串类型私钥算ecdsa.PrivateKey类型私钥
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = crypto.S256()
	kr, err := hex.DecodeString(from.PrivateKey)
	if err != nil {
		fmt.Println(err)
	}
	priv.D = new(big.Int).SetBytes(kr)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(kr)
	// 已知私钥算签名
	hash := sha256.Sum256([]byte(defaultData))
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		fmt.Println(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return &TX{
		Nonce:         from.Nonce,
		FromAddr:      from.Addr,
		ToAddr:        toAddr,
		Value:         strconv.Itoa(defaultValue),
		Data:          defaultData,
		CryptoType:    defaultCryptoType,
		Signature:     string(signature),
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

// 我把交易设计成从偶秩账户到它的后缀，这个函数会循环返回0、2、4……、0……
func nextFromRank(r int) int {
	if r == countOfAccounts-2 {
		return 0
	}
	return r + 2
}

func main() {
	// 新建账户
	var accounts [countOfAccounts]Account
	for i := 0; i < countOfAccounts; i++ {
		accounts[i] = *newAccount()
	}
	fromRank := 0
	// 不同账户之间死循环发送交易请求
	for {
		// 相同账户之间并发发送交易请求
		for i := 0; i < maxConcurrency; i++ {
			go transaction(accounts[fromRank], accounts[fromRank+1].Addr)
		}
		fromRank = nextFromRank(fromRank)
	}
}
