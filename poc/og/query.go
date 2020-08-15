package og

import (
	"BlockDB/poc/common"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/W1llyu/ourjson"
)

// Queryer 查询器
type Queryer struct {
	timestampOfHeight map[int]int /* 键：区块高度，值：时间戳 */
	HeightMin         int         /* 交易最小高度 */
	HeightMax         int         /* 交易最大高度 */
	heightInited      bool        /* 交易最小、最大高度是否被初始化 */
}

// NewQueryer 构造查询器
func NewQueryer() *Queryer {
	return &Queryer{
		timestampOfHeight: make(map[int]int),
		heightInited:      false,
	}
}

// QuerySequencerTimestamp 查询指定高度区块时间戳
func (q *Queryer) querySequencerTimestamp(height int) int {
	timestamp, ok := q.timestampOfHeight[height]
	if ok {
		return timestamp
	}
	getURL := common.URL + "/" + common.QuerySequencerRPCMethod
	req, err := http.NewRequest("GET", getURL+"?seq_id="+strconv.Itoa(height), nil)
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
	data := jsonObj.GetJsonObject("data")
	timestamp, err = data.GetInt("Timestamp")
	if err != nil {
		fmt.Println(err)
	}
	if timestamp == 0 /* Annchain.OG存在区块时间戳是零的故障，暂时用上个区块的时间戳替代 */ {
		timestamp = q.querySequencerTimestamp(height - 1)
	}
	q.timestampOfHeight[height] = timestamp
	return timestamp
}

// QueryTX 查询交易相关信息，返回是否上链、区块时间戳
func (q *Queryer) QueryTX(hash string) (bool, int) {
	getURL := common.URL + "/" + common.QueryTransactionRPCMethod
	req, err := http.NewRequest("GET", getURL+"?hash="+hash, nil)
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
		return false, 0 /* 未上链，JSON报错 */
	}
	data := jsonObj.GetJsonObject("data")
	transaction := data.GetJsonObject("transaction")
	height, err := transaction.GetInt("height")
	if err != nil {
		fmt.Println(err)
	}
	if !q.heightInited {
		q.HeightMin = height
		q.HeightMax = height
		q.heightInited = true
	} else {
		if height < q.HeightMin {
			q.HeightMin = height
		}
		if height > q.HeightMax {
			q.HeightMax = height
		}
	}
	return true, q.querySequencerTimestamp(height)
}

// QueryCountOfTX 查询指定高度区块交易数目
func queryCountOfTX(height int) int {
	getURL := common.URL + "/" + common.QueryTransactionsRPCMethod
	req, err := http.NewRequest("GET", getURL+"?height="+strconv.Itoa(height), nil)
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
	data := jsonObj.GetJsonObject("data")
	txHashes := data.GetJsonArray("hashes")
	if err != nil {
		fmt.Println(err)
	}
	return len(txHashes.Values())
}

// GetTPS 获得每秒处理交易数（TPS）,如果交易总数不足，不能计算TPS，那么返回0
func (q *Queryer) GetTPS() float64 {
	count := 0
	if q.HeightMax-q.HeightMin > 1 /* 交易至少完全填充了1个区块，可以计算TPS */ {
		for i := q.HeightMin + 1; i < q.HeightMax; i++ {
			count += queryCountOfTX(i)
		}
		return float64(count*1000) / float64(q.querySequencerTimestamp(q.HeightMax-1)-q.querySequencerTimestamp(q.HeightMin))
	}
	return 0
}
