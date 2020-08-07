package instruction


type OperationRecords struct{
	Hash		  string 				`json:"_hash"`	//数据的hash
	Collection    string      			`json:"collection"`	//操作的数据表
	Operation	  int					`json:"operation"`
	Timestamp	  string				`json:"timestamp"`
	Data		  map[string]interface{}`json:"data"`	//操作记录
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
}

type DocHistory struct {
	Hash		  string 				`json:"_hash"`	//数据的hash
	Collection    string      			`json:"collection"`	//操作的数据表
	Timestamp	  string				`json:"timestamp"`
	Data		  map[string]interface{}`json:"history"`	//历史版本
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
}


func OpRecord(op int,hash string,coll string,timestamp string,data map[string]interface{},pk string,sig string){
	oprecord:=OperationRecords{hash,coll,op,timestamp,data,pk,sig}
	//TODO: Insert(HistoryDataBase,OpRecordCollection,oprecord)
}

func HistoryRecord(hash string,coll string,timestamp string,data map[string]interface{},pk string,sig string){
	history:=DocHistory{hash,coll,timestamp,data,pk,sig}
	//TODO: Insert(HistoryDataBase,HistoryCollection,history)
}

