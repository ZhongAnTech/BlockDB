package instruction

type AuditModel struct{
	Hash		  string 				`json:"_hash"`	//数据的hash
	Collection    string      			`json:"collection"`	//操作的数据表
	Operation	  int					`json:"operation"`
	Timestamp	  string				`json:"timestamp"`
	Data		  map[string]interface{}`json:"data"`	//操作记录
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
}


func Audit(op int,hash string,coll string,timestamp string,data map[string]interface{},pk string,sig string){
	audit:=AuditModel{hash,coll,op,timestamp,data,pk,sig}
	//TODO: Insert(AuditDataBase,AuditCollection,audit)
}
