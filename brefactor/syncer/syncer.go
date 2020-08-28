package syncer

import (
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/clients/og"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ChainOperator is the pulling way to query the latest sequencer and txs.
type ChainOperator interface {
	QueryHeight() (int64, error)
	QueryTxHashByHeight(url string) ([]string, error)
	QueryTxByHash(url string) ([]byte, error)
	// TODO: Qingyun enrich interface and implement plugins/clients/og/OgChainOperator
}

// ChainEventReceiver is the pushing way to receive the latest sequencer
type ChainEventReceiver interface {
	Connect() int64
	EventChannel() chan int64 // Maybe you will pass a more complicate object in the channel
	// TODO: Qingyun enrich interface and implement plugins/clients/og/OgChainEventReceiver

}

type OgChainSyncerConfig struct {
	LatestHeightUrl string
	WebsocketUrl string
}

type OgChainSyncer struct {
	// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
	// TODO: receive websocket push (receive latest height) and sync (realtime)

	SyncerConfig OgChainSyncerConfig
	// table: op
	MaxSyncedHeight int64
	ChainOperator   ChainOperator
	InfoReceiver    ChainEventReceiver
	storageExecutor core_interface.StorageExecutor
	quit chan bool
}

func (o *OgChainSyncer) Start() {
	// load max height from ledger

	go o.loop()
}

func (o *OgChainSyncer) Stop() {
	panic("implement me")
}

func (o *OgChainSyncer) Name() string {
	return "OgChainSyncer"
}

func (o *OgChainSyncer) QueryHeight() (int64, error) {
	response, err := http.Get(o.SyncerConfig.LatestHeightUrl)
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
	//panic("implement me")
}

func (o *OgChainSyncer) QueryTxHashByHeight(url string) ([]string, error) {
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

func (o *OgChainSyncer) QueryTxByHash(url string) (string, error) {
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

func (o *OgChainSyncer) Connect() int64 {
	dialer := websocket.Dialer{}
	connect, _, err := dialer.Dial(o.SyncerConfig.WebsocketUrl,nil)
	if nil != err {
		log.Print(err)

	}
	defer connect.Close()
	var height int
	for {
		_, messageData, err := connect.ReadMessage()
		if nil != err {
			log.Print(err)
			break
		}
		str := string(messageData)
		s1 := strings.Split(str, ",")
		s2 := strings.Split(s1[5], ":")
		height, err := strconv.Atoi(s2[1])
		fmt.Println(height)

	}
	return int64(height)
}

func (o *OgChainSyncer) EventChannel() chan int64 {
	h := make(chan int64)
	h <- o.Connect()
	return h
}



func (o *OgChainSyncer) loop() {
	for {
		select {
		case <-o.quit:
			return
		case newHeight := <-o.InfoReceiver.EventChannel():
			// TODO: compare local max height and sync if behind.
			if newHeight > o.MaxSyncedHeight {
				// TODO: sync.
				for i := o.MaxSyncedHeight + 1; i <= newHeight; i++ {
					url1 := "http://nbstock.top:30022/transaction_hashes?height=" + strconv.Itoa(int(i))
					hashes, err := o.QueryTxHashByHeight(url1)
					if err != nil {
						fmt.Println("can't get txhash in newHeight-block")
					}
					if hashes == nil {
						fmt.Println("no tx with type = 4 in height: ", i)
					} else {
						var txDatas []og.Archive
						for _, v := range hashes {
							fmt.Println(v)
							url2 := "http://nbstock.top:30022/transaction?hash=" + v[1:len(v)-1]
							txData, err := o.QueryTxByHash(url2)
							if err != nil {
								fmt.Println("query tx by hash fail..")
							}
							if strings.Contains(txData, "\"type\":4") == true {
								//验签，反序列化放到结构体，存入数据库
								txDatas = og.ToStruct(txData)
								fmt.Println("type=4---------", txData)

							}

						}
						og.Test(txDatas)
					}
					o.MaxSyncedHeight = newHeight
				}
			}

		//case <- timeout (be very careful when you handle the timer reset to prevent blocking.)
		default:

			// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
			// TODO: receive websocket push (receive latest height) and sync (realtime)
			for {
				time.Sleep(10*time.Second)
				latestHeight,err := o.QueryHeight()
				url1 := "http://nbstock.top:30022/transaction_hashes?height=" + strconv.Itoa(int(latestHeight))
				hashes, err := o.QueryTxHashByHeight(url1)
				if err != nil {
					fmt.Println("can't get txhash in newHeight-block")
				}
				if hashes == nil {
					fmt.Println("no tx with type = 4 in height: ", latestHeight)
				} else {
					var txDatas []og.Archive
					for _, v := range hashes {
						fmt.Println(v)
						url2 := "http://nbstock.top:30022/transaction?hash=" + v[1:len(v)-1]
						txData, err := o.QueryTxByHash(url2)
						if err != nil {
							fmt.Println("query tx by hash fail..")
						}
						if strings.Contains(txData, "\"type\":4") == true {
							//验签，反序列化放到结构体，存入数据库
							txDatas = og.ToStruct(txData)
							fmt.Println("type=4---------", txData)

						}

					}
					og.Test(txDatas)
				}
			}
		}

	}
}
