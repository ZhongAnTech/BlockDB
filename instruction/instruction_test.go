package instruction

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestUnmarshal(t *testing.T){
	//str:=`{"op":"insert", "collection":"sample_collection", "data":{"name":"fudan", "address":{"city":"Shanghai", "road":"xxx"}, "logo":{"url":"http://a.png"}, "teachers":["T1", "T2", "T3"]}, "public_key":"0x769153474351324", "signature":"0x169153474351324"}`
	//str:=`{"op": "update","name": "sample_collection","query": {"op_hash": "0x739483392203"},"set": {"name": "fudanNew","address.city": "Shanghai North East","logo": {}},"unset": ["teachers"],"public_key": "0x769153474351324","signature": "0x169153474351324","op_hash":"0x739483392203"}`
	str:=`{"op": "hint_create_index","name": "sample_collection","index": {"idx_address_city": "address.city","idx_address_city2": "address.city2"},"public_key": "0x769153474351324","signature": "0x169153474351324"}`
	cmd:=&BlockDBCommandIndex{}
	err:=json.Unmarshal([]byte(str),cmd)
	if err != nil{
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
		op:= c["op"].(string)
		err=Execute(op,str)
		if err != nil {
			panic(err)
		}
		i++
		fmt.Println(i)

	}
	fmt.Println("文件读取结束...")
}



