package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/W1llyu/ourjson"
)

// AccountReq 新建账户请求字段
type AccountReq struct {
	Algorithm string `json:"algorithm"` /* 加密算法 */
}

func newAccountReq() *AccountReq {
	return &AccountReq{
		Algorithm: "secp256k1",
	}
}

func main() {
	url := "http://47.100.122.212:30022/new_account"
	post, err := json.Marshal(newAccountReq())
	if err != nil {
		fmt.Println(err)
	}
	postBuffer := bytes.NewBuffer(post)
	req, err := http.NewRequest("POST", url, postBuffer)
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
	data := jsonObj.GetJsonObject("data")
	kr, err := data.GetString("privkey")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(kr)
}
