package instruction

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/storage"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
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

func TestInstruction(t *testing.T){
	file, err := os.Open("./instructions.txt")
	if err != nil {
		fmt.Println("文件打开失败 = ", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var executor InstructionExecutor
	executor.InitDefault()
	executor.Config=InstructionExecutorConfig{10,time.Second*3,time.Second*3,time.Second*3}
	executor.storageExecutor=storage.Connect(ctx, "mongodb://paichepai.win:27017", "blockdb", "", "", "")
	//executor.storageExecutor=storage.Connect(ctx, "mongodb://127.0.0.1:27017", "block", "", "", "")
	err =executor.InitCollection(ctx,MasterCollection)
	if err != nil{
		fmt.Println(err)
	}

	i:=0
	for {
		str, err := reader.ReadString('\n') //读到一个换行就结束
		if err == io.EOF {                  //io.EOF 表示文件的末尾
			break
		}
		if(len(str)==0){
			break
		}
		//fmt.Print(str)
		c := make(map[string]interface{})
		err = json.Unmarshal([]byte(str), &c)
		if err != nil {
			panic(err)
		}

		err = executor.Execute(GeneralCommand{
			TxHash:    c["tx_hash"].(string),
			OpHash:    c["op_hash"].(string),
			OpStr:     c["op_str"].(string),
			Signature: c["signature"].(string),
			PublicKey: c["public_key"].(string),
		})
		if err != nil {
			logrus.WithError(err).WithField("op", c["op_str"].(string)).Error("failed to execute op")
			// TODO: retry or mark as failed, according to err type
			continue
		}

		i++
		fmt.Println("i="+strconv.Itoa(i))
	}
	fmt.Println("文件读取结束...")
}

func TestInstructionFromDB(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var executor InstructionExecutor
	executor.InitDefault()
	executor.Config=InstructionExecutorConfig{10,time.Second*3,time.Second*3,time.Second*3}
	executor.storageExecutor=storage.Connect(ctx, "mongodb://paichepai.win:27017", "blockdb", "", "", "")
	//executor.storageExecutor=storage.Connect(ctx, "mongodb://127.0.0.1:27017", "block", "", "", "")
	err:=executor.InitCollection(ctx,MasterCollection)
	if err != nil{
		fmt.Println(err)
	}
	executor.doBatchJob()

}
