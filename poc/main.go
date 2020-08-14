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
	countOfTXPerCoroutine int              /* 每个协程发送的交易数目 */
	countOfCoroutine      int              /* 协程数目 */
	accountPairs          []og.AccountPair /* 账户对 */
	txResp                chan og.TXInfo   /* 交易返回 */
	wg                    sync.WaitGroup   /* 协程等待组 */
)

// 每个协程的任务
func handle(from *og.Account, to *og.Account) {
	defer wg.Done()
	for i := 0; i < countOfTXPerCoroutine; i++ {
		fmt.Println("?")
		txResp <- *og.NewTXInfo(og.TX(from, to))
	}
}

func main() {
	// 通过命令行参数获得每个协程发送的交易数目和协程数目
	countOfTXPerCoroutine, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	countOfCoroutine, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	// 构造账户对，每个协程对应对应1个账户对
	accountPairs = make([]og.AccountPair, countOfCoroutine)
	for i := 0; i < countOfCoroutine; i++ {
		accountPairs[i] = *og.NewAccountPair()
		accountPairs[i].From.InitNonce()
	}
	// 并发地、循环地发送交易
	txResp = make(chan og.TXInfo, countOfTXPerCoroutine*countOfCoroutine)
	wg.Add(countOfCoroutine)
	for i := 0; i < countOfCoroutine; i++ { /* 为了高效地发送交易，新建账户和交易是分离的 */
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
	len := len(txResp)
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
			fmt.Println("TX" + hash + "\t" + strconv.Itoa(sendTimestamp))
			ok, timestamp := queryer.QueryTX(hash) /* 查询上链时间戳 */
			if ok {
				countOfTXSucceeded++
				if timestamp >= sendTimestamp {
					delaySum += (timestamp - sendTimestamp)
					delayNum++
				}
			}
		}
	}
	fmt.Println("TPS:" + strconv.FormatFloat(queryer.GetTPS(), 'E', -1, 64))
	fmt.Println("Average delay:" + strconv.Itoa(delaySum/delayNum) + "ms")
	fmt.Println("Total number of TX:" + strconv.Itoa(countOfTXPerCoroutine*countOfCoroutine))
	fmt.Println("Succeeded number of TX:" + strconv.Itoa(countOfTXSucceeded))
}
