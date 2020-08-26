package og

import (
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//获取最新高度
func Http() int {
	response, err := http.Get("http://nbstock.top:30022//v1/sequencer")
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	str := string(body)

	s1 := strings.Split(str, ",")
	s2 := strings.Split(s1[4], ":")
	height, err := strconv.Atoi(s2[1])
	if err != nil {
		fmt.Println("can't trans string to int")
	}
	fmt.Println(height)
	return height
}

//获取该高度的交易hash
func GetHashes(url string) []string {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	str := string(body)
	fmt.Println(str)
	if strings.Contains(str, "\"hashes\":null") {
		return nil
	}
	s1 := strings.Split(str, "[")
	s2 := strings.Split(s1[1], "]")
	s3 := strings.Split(s2[0], ",")
	fmt.Println(s3)
	return s3
}

//获取交易内容
func HttpHash(url string) string {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	fmt.Println(body)
	str := string(body)
	fmt.Println(str)
	return str
}

func hashByHeight(height int) {
	url1 := "http://nbstock.top:30022/transaction_hashes?height=" + strconv.Itoa(height)
	hashes := GetHashes(url1)

	if hashes == nil {
		fmt.Println("no tx with type = 4 in height: ", height)
	} else {
		var txDatas []Archive
		for _, v := range hashes {
			fmt.Println(v)
			url2 := "http://nbstock.top:30022/transaction?hash=" + v[1:len(v)-1]
			txData := HttpHash(url2)
			if strings.Contains(txData, "\"type\":4") == true {
				//验签，反序列化放到结构体，存入数据库

				txDatas = ToStruct(txData)
				fmt.Println("type=4---------", txData)

			}

		}
		//排序
		test(txDatas)
	}
}

func DownChain() {
	height := 0
	preHeight := 0
	for {
		if height == Http() {
			continue
		}
		time.Sleep(10 * time.Millisecond)
		height = Http()
		if height-1 > preHeight {
			for i := preHeight + 1; i < height; i++ {
				hashByHeight(i)
			}
		}
		preHeight = height
		hashByHeight(height)
	}

}
