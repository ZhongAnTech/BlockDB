package syncer

import (
	"context"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/clients/og"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
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
	Connect()
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

type Archive struct {
	Height       int      `json:"height"`
	Type         int      `json:"type"`
	TxHash       string   `json:"tx_hash"`
	OpHash       string   `json:"op_hash"`
	PublicKey    string   `json:"public_key"`
	Signature    string   `json:"signature"`
	Parents      []string `json:"parents"`
	AccountNonce int      `json:"account_nonce"`
	MindNonce    int      `json:"mind_nonce"`
	Weight       int      `json:"weight"`
	Data         string   `json:"data"`
}

type Op struct {
	Order      int    `json:"order"`
	Height     int    `json:"height"`
	IsExecuted bool   `json:"is_executed"`
	TxHash     string `json:"tx_hash"`
	OpHash     string `json:"op_hash"`
	PublicKey  string `json:"public_key"`
	Signature  string `json:"signature"`
	OpStr      string `json:"op_str"`
}

func (o *OgChainSyncer) Start() {
	// load max height from ledger

	go o.loop()
}

func (o *OgChainSyncer) Stop() {
	o.quit <- true
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



type LastHeight struct {
	height int64
}


func (o *OgChainSyncer) loop() {
	for {
		select {
		case <-o.quit:
			return
		case newHeight := <-o.InfoReceiver.EventChannel():
			// TODO: compare local max height and sync if behind.

			//newHeight来自ws推送
			//假如重新启动，就从数据库里面查之前的高度
			if o.MaxSyncedHeight == 0 {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
				content,err := o.storageExecutor.Select(ctx,"lastHeight",bson.M{},nil,1,0)
				if err != nil {
					fmt.Println("can't get lateHeight from db")
				}
				for _,v := range content.Content {
					a := LastHeight{}
					bsonBytes, _ := bson.Marshal(v)
					bson.Unmarshal(bsonBytes, &a)
					o.MaxSyncedHeight = a.height
				}
			}
			if newHeight > o.MaxSyncedHeight {
				// TODO: sync.
				for i := o.MaxSyncedHeight + 1; i <= newHeight; i++ {
					url1 := o.SyncerConfig.LatestHeightUrl + "/transaction_hashes?height=" + strconv.Itoa(int(i))
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
							url2 := o.SyncerConfig.LatestHeightUrl + "/transaction?hash=" + v[1:len(v)-1]
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
						//遍历结构体集合，插入“op” Collection中
						for i, v := range txDatas {
							var op = Op{
								Order:      i,
								Height:     v.Height,
								IsExecuted: false,
								TxHash:     v.TxHash,
								OpHash:     v.OpHash,
								PublicKey:  v.PublicKey,
								Signature:  v.Signature,
								OpStr:      v.Data,
							}

							fmt.Println("op: ", op)

							ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
							//update := bson.D{{"$set", data}}
							id, err := o.storageExecutor.Insert(ctx,"op",bson.M{
								"is_executed" : op.IsExecuted,
								"tx_hash" : op.TxHash,
								"op_hash" : op.OpHash,
								"public_key" : op.PublicKey,
								"signature" : op.Signature,
								"op_str" : op.OpStr,
							})
							fmt.Println(id, err)

							filter := bson.M{
								"tx_hash" : op.TxHash,
								"op_hash" : op.OpHash,
								"status" : 0,
							}

							update := bson.M{
								"tx_hash" : op.TxHash,
								"op_hash" : op.OpHash,
								"status" : 1,
							}
							o.storageExecutor.Update(ctx,"isOnChain",filter,update,"set")
						}
					}
					ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
					//将此时的高度替换旧的存入数据库中
					o.storageExecutor.Update(ctx,"lastHeight",bson.M{"lastHeigh":o.MaxSyncedHeight},bson.M{"lastHeight":newHeight},"unset")
					o.MaxSyncedHeight = newHeight
				}
			}

		//case <- timeout (be very careful when you handle the timer reset to prevent blocking.)
		case <- time.After(time.Second*60):
			latestHeight,err := o.QueryHeight()
			if err != nil {
				fmt.Println("fail to query new height")
			}
			for i := o.MaxSyncedHeight + 1; i <= latestHeight; i++ {
				url1 := o.SyncerConfig.LatestHeightUrl + "/transaction_hashes?height=" + strconv.Itoa(int(i))
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
						url2 := o.SyncerConfig.LatestHeightUrl + "/transaction?hash=" + v[1:len(v)-1]
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
					//遍历结构体集合，插入“op” Collection中
					for i, v := range txDatas {
						var op = Op{
							Order:      i,
							Height:     v.Height,
							IsExecuted: false,
							TxHash:     v.TxHash,
							OpHash:     v.OpHash,
							PublicKey:  v.PublicKey,
							Signature:  v.Signature,
							OpStr:      v.Data,
						}

						fmt.Println("op: ", op)

						ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
						//update := bson.D{{"$set", data}}
						id, err := o.storageExecutor.Insert(ctx,"op",bson.M{
							"is_executed" : op.IsExecuted,
							"tx_hash" : op.TxHash,
							"op_hash" : op.OpHash,
							"public_key" : op.PublicKey,
							"signature" : op.Signature,
							"op_str" : op.OpStr,
						})
						fmt.Println(id, err)

						filter := bson.M{
							"tx_hash" : op.TxHash,
							"op_hash" : op.OpHash,
							"status" : 0,
						}

						update := bson.M{
							"tx_hash" : op.TxHash,
							"op_hash" : op.OpHash,
							"status" : 1,
						}
						o.storageExecutor.Update(ctx,"isOnChain",filter,update,"set")
					}
				}
				ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
				//将此时的高度替换旧的存入数据库中
				o.storageExecutor.Update(ctx,"lastHeight",bson.M{"lastHeigh":o.MaxSyncedHeight},bson.M{"lastHeight":latestHeight},"unset")
				o.MaxSyncedHeight = latestHeight
			}
		default:

			// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
			// TODO: receive websocket push (receive latest height) and sync (realtime)

		}

	}
}
