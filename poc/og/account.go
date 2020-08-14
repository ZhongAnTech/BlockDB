package og

import (
	"BlockDB/poc/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/W1llyu/ourjson"
	og_common "github.com/annchain/OG/common"
	"github.com/annchain/OG/common/crypto"
	"go.uber.org/atomic"
)

// accountReq 开户请求
type accountReq struct {
	Algorithm string `json:"algorithm"` /* 加密算法 */
}

// newAccountReq 构造开户请求
func newAccountReq() *accountReq {
	return &accountReq{
		Algorithm: common.DefaultCryptoType,
	}
}

// Account 账户
type Account struct {
	PrivateKey  crypto.PrivateKey /* 私钥 */
	PublicKey   crypto.PublicKey  /* 公钥 */
	Address     og_common.Address /* 地址 */
	nonce       atomic.Uint64     /* nonce，每次发送交易后递增 */
	nonceInited bool              /* nonce是否已经被初始化 */
	mutex       sync.RWMutex      /* 读写锁 */
}

// NewAccount 构造账户
func NewAccount() *Account {
	postURL := common.URL + "/" + common.NewAccountRPCMethod
	post, err := json.Marshal(newAccountReq())
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
	jsonObj, err := ourjson.ParseObject(string(respBody))
	if err != nil {
		fmt.Println(err)
	}
	errStr, err := jsonObj.GetString("err")
	if err != nil {
		fmt.Println(err)
	}
	if errStr != "" {
		fmt.Println(errStr)
	}
	data := jsonObj.GetJsonObject("data")
	kr, err := data.GetString("privkey")
	if err != nil {
		fmt.Println(err)
	}
	priv, err := crypto.PrivateKeyFromString(kr)
	if err != nil {
		fmt.Println(err)
	}
	signer := crypto.NewSigner(priv.Type)
	pub := signer.PubKey(priv)
	return &Account{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    signer.Address(pub),
	}
}

// AccountPair 账户对
type AccountPair struct {
	From *Account /* 发送交易账户 */
	To   *Account /* 接收交易账户 */
}

// NewAccountPair 构造账户对
func NewAccountPair() *AccountPair {
	return &AccountPair{
		From: NewAccount(),
		To:   NewAccount(),
	}
}

// nonceResponse 更新nonce的回复消息
type nonceResponse struct {
	Nonce uint64 `json:"nonce"` /* nonce */
	Err   string `json:"err"`   /* 错误 */
}

// InitNonce 初始化账户的nonce，跟链同步
func (account *Account) InitNonce() {
	getURL := common.URL + "/" + common.QueryNonceRPCMethod
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
	nr := &nonceResponse{}
	err = json.Unmarshal(respBody, nr)
	if err != nil {
		fmt.Println(err)
	}
	if nr.Err != "" {
		fmt.Println(err)
	}
	account.mutex.Lock()
	defer account.mutex.Unlock()
	account.nonce.Store(nr.Nonce)
	account.nonceInited = true
}

// ConsumeNonce 消费nonce，返回新nonce和错误
func (account *Account) ConsumeNonce() (uint64, error) {
	account.mutex.Lock()
	defer account.mutex.Unlock()
	if !account.nonceInited {
		return 0, fmt.Errorf("nonce is not initialized")
	}
	account.nonce.Inc() /* nonce++，包装并返回 */
	return account.nonce.Load(), nil
}
