package og

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type OgChainOperator struct {
	GetHeightUrl string
	Height int64
	TxHash []string
	Tx string
}

func (oc *OgChainOperator) QueryHeight() (int64, error) {
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
	return int64(height), err
}
func (oc *OgChainOperator) QueryTxHashByHeight(url string) ([]string, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	str := string(body)
	fmt.Println(str)
	if strings.Contains(str, "\"hashes\":null") {
		return nil,err
	}
	s1 := strings.Split(str, "[")
	s2 := strings.Split(s1[1], "]")
	s3 := strings.Split(s2[0], ",")
	fmt.Println(s3)
	return s3,err
}
func (oc *OgChainOperator) QueryTxByHash(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	fmt.Println(body)
	str := string(body)
	fmt.Println(str)
	return str,err
}
