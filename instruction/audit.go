package instruction

import (
	"github.com/annchain/BlockDB/plugins/server/mongodb/mongoutils"
	"log"
	"go.mongodb.org/mongo-driver/bson"
)

type AuditModel struct{
	OpHash		  string 				`json:"_id"`	//数据的hash
	Collection    string      			`json:"collection"`	//操作的数据表
	Operation	  string				`json:"operation"`
	Timestamp	  string				`json:"timestamp"`
	Data		  map[string]interface{}`json:"data"`	//操作记录
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
}


func Audit(op string,hash string,coll string,timestamp string,data map[string]interface{},pk string,sig string)error{
	auditdb:= mongoutils.InitMgo(url,BlockDataBase,AuditCollection)
	audit:=bson.D{{"op_hash",hash},{"collection",coll},{"operation",op},{"timestamp",timestamp},
		{"data",data},{"public_key",pk},{"signature",sig}}
	_,err:=auditdb.Insert(audit)
	if err != nil {
		log.Fatal("failed to insert data to history.")
		return err
	}
	_=auditdb.Close()
	return nil
}

//{"op":"create_collection","name":"sample_collection","feature":{"allow_update":false, "allow_update_members": ["0x123456", "0x123456", "0x123456", "0x123456"]},"public_key": "0x769153474351324"}