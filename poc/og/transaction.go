package og

import (
	"BlockDB/poc/common"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	og_common "github.com/annchain/OG/common"
	"github.com/annchain/OG/common/crypto"
	"github.com/annchain/OG/common/math"
	"github.com/annchain/OG/rpc"
	"github.com/annchain/OG/types"
	"github.com/annchain/OG/types/tx_types"
)

// TXResponse 交易回复消息
type TXResponse struct {
	Data string `json:"data"` /* 哈希 */
	Err  string `json:"err"`  /* 错误 */
}

// txRequest 交易请求
func txRequest(from *Account, to *Account) *rpc.NewTxRequest {
	nonce, err := from.ConsumeNonce()
	if err != nil {
		fmt.Println(err)
	}
	signer := crypto.NewSigner(from.PrivateKey.Type)
	tx := tx_types.Tx{
		TxBase: types.TxBase{
			Type:         0,
			Hash:         og_common.Hash{},
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
		Value:      common.DefaultValue,
		Data:       common.DefaultData,
		CryptoType: common.DefaultCryptoType,
		Signature:  hex.EncodeToString(sig.Bytes),
		Pubkey:     hex.EncodeToString(from.PublicKey.Bytes),
		TokenId:    common.DefaultTokenID,
	}
}

// TX 交易
func TX(from *Account, to *Account) []byte {
	postURL := common.URL + "/" + common.NewTransactionRPCMethod
	post, err := json.Marshal(txRequest(from, to))
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
	fmt.Println("A new transaction sent.")
	return respBody
}

// TXInfo 发送交易后，Annchain.OG回复的交易信息
type TXInfo struct {
	RespBody  []byte /* 回复内容 */
	Timestamp int64  /* 被构造时间戳 */
}

// NewTXInfo 构造回复
func NewTXInfo(rb []byte) *TXInfo {
	return &TXInfo{
		RespBody:  rb,
		Timestamp: time.Now().UnixNano(),
	}
}
