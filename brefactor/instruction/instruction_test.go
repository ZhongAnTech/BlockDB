package instruction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/storage"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	//str:=`{"op":"insert", "collection":"sample_collection", "data":{"name":"fudan", "address":{"city":"Shanghai", "road":"xxx"}, "logo":{"url":"http://a.png"}, "teachers":["T1", "T2", "T3"]}, "public_key":"0x769153474351324", "signature":"0x169153474351324"}`
	//str:=`{"op": "update","name": "sample_collection","query": {"op_hash": "0x739483392203"},"set": {"name": "fudanNew","address.city": "Shanghai North East","logo": {}},"unset": ["teachers"],"public_key": "0x769153474351324","signature": "0x169153474351324","op_hash":"0x739483392203"}`
	str := `{"op": "hint_create_index","name": "sample_collection","index": {"idx_address_city": "address.city","idx_address_city2": "address.city2"},"public_key": "0x769153474351324","signature": "0x169153474351324"}`
	cmd := &IndexCommand{}
	err := json.Unmarshal([]byte(str), cmd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cmd)
	//str2,err:=json.Marshal(cmd)
	//if err != nil {
	//	fmt.Println("marshal err")
	//}
	//fmt.Println(string(str2))
}

func TestInstruction(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var executor InstructionExecutor
	executor.InitDefault()
	executor.Config = InstructionExecutorConfig{10, time.Second * 3, time.Second * 3, time.Second * 3}
	client, err := storage.Connect(ctx, "mongodb://localhost:27017", "blockdb", "", "", "")
	if err != nil {
		panic(err)
	}
	executor.storageExecutor = client

	//executor.storageExecutor=storage.Connect(ctx, "mongodb://127.0.0.1:27017", "block", "", "", "")
	err = executor.InitCollection(ctx, MasterCollection)
	if err != nil {
		panic(err)
	}
	executor.doBatchJob()

}
