package main

import (
	"BlockDB/poc/common"
	"BlockDB/poc/og"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	countOfTXsPerCoroutine int                    /* 每个协程发送的交易数目 */
	countOfCoroutines      int                    /* 协程数目 */
	accountPairs           []og.AccountPair       /* 账户对 */
	txResp                 chan og.TXInfo         /* 交易返回 */
	wg                     sync.WaitGroup         /* 协程等待组 */
	testResults            map[string]interface{} /* 测试结果 */
)

// 每个协程的任务
func handle(from *og.Account, to *og.Account) {
	defer wg.Done()
	for i := 0; i < countOfTXsPerCoroutine; i++ {
		txResp <- *og.NewTXInfo(og.TX(from, to))
	}
}

func main() {
	// 通过命令行参数获得每个协程发送的交易数目和协程数目
	var err error /* 为了正确地传值给countOfTXsPerCoroutine和countOfCoroutines */
	countOfTXsPerCoroutine, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	countOfCoroutines, err = strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	// 构造账户对，每个协程对应对应1个账户对
	accountPairs = make([]og.AccountPair, countOfCoroutines)
	for i := 0; i < countOfCoroutines; i++ {
		accountPairs[i] = *og.NewAccountPair()
		accountPairs[i].From.InitNonce()
	}
	// 并发地、循环地发送交易
	txResp = make(chan og.TXInfo, countOfTXsPerCoroutine*countOfCoroutines)
	wg.Add(countOfCoroutines)
	for i := 0; i < countOfCoroutines; i++ { /* 为了高效地发送交易，新建账户和交易是分离的 */
		go handle(accountPairs[i].From, accountPairs[i].To)
	}
	wg.Wait()
	// 等待1分钟，交易上链
	fmt.Println("Wait 1 min...")
	time.Sleep(time.Minute * 1)
	// 处理交易返回，计算上链成功率和上链延迟平均数
	countOfTXSucceeded := 0    /* 上链成功数 */
	delaySum := 0              /* 上链延迟总和，单位：毫秒，用来计算上链延迟平均数 */
	delayNum := 0              /* 被测到存在上链延迟的交易数目，用来计算上链延迟平均数 */
	queryer := og.NewQueryer() /* 查询器 */
	testResults = make(map[string]interface{})
	len := len(txResp)
	fmt.Println("Hash\tSent Timestamp\tSequencer Timestamp")
	for i := 0; i < len; i++ {
		txr := <-txResp
		tr := &og.TXResponse{}
		err := json.Unmarshal(txr.RespBody, tr)
		if err != nil {
			fmt.Println(err)
		}
		if tr.Err != "" {
			fmt.Println(tr.Err)
			continue
		}
		if tr.Data != "" {
			hash := strings.TrimPrefix(tr.Data, "0x")         /* 交易哈希 */
			sendTimestamp := common.Nano2Milli(txr.Timestamp) /* 交易发送时间戳 */
			ok, timestamp := queryer.QueryTX(hash)            /* 查询上链时间戳 */
			if ok {
				countOfTXSucceeded++
				if timestamp >= sendTimestamp {
					delaySum += (timestamp - sendTimestamp)
					delayNum++
					fmt.Println(hash + "\t" + strconv.Itoa(sendTimestamp) + "\t" + strconv.Itoa(timestamp))
				}
			}
		}
	}
	testResults["TPS"] = queryer.GetTPS()
	testResults["Success Rate of TX/%"] = float64(countOfTXSucceeded*100) / float64(countOfTXsPerCoroutine*countOfCoroutines)
	testResults["Average Delay/ms"] = float64(delaySum) / float64(delayNum)
	fmt.Println(testResults)
}
