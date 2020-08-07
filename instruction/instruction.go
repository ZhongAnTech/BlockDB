package instruction

const (
	CreateCollection int = iota
	UpdateCollection
	Insert
	Update
	Delete
	CreateIndex
	DropIndex
	CollIndexDataBase 	string = "collindexdb"
	CollCollection		string = "coll"
	IndexCollection		string = "index"
	BlockDataBase 		string = "blockdb"
	BlockCollection		string = "block"
	HistoryDataBase		string = "historydb"
	HistoryCollection 	string = "history"
	OpRecordCollection	string = "oprecord"
	AuditDataBase		string = "auditdb"
	AuditCollection		string = "audit"
)

var Colls []*BlockDBCommandCollection
var Indexes []*BlockDBCommandIndex

type BlockDBCommandCollection struct{
	Hash		  string 				`json:"_hash"`	//产生数据的hash作为主键
	Collection    string      			`json:"collection"`	//要操作的数据表
	Feature		  map[string]interface{}`json:"feature"`
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
	Timestamp  	  string      			`json:"timestamp"`
}

type BlockDBCommandInsert struct{
	Hash		  string 				`json:"_hash"`	//产生数据的hash作为主键
	Collection    string      			`json:"collection"`	//要操作的数据表
	Data          map[string]interface{}`json:"data"`
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
	Timestamp  	  string      			`json:"timestamp"`
}

type BlockDBCommandUpdate struct{
	Collection    string      			`json:"collection"`	//要操作的数据表
	Query 		  map[string]string     `json:"query"`
	Set           map[string]string  	`json:"set"`
	Unset         []string 			    `json:"unset"`
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
	Timestamp  	  string      			`json:"timestamp"`
}

type BlockDBCommandDelete struct{
	Collection    string      			`json:"collection"`	//要操作的数据表
	Query 		  map[string]string     `json:"query"`
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
	Timestamp  	  string      			`json:"timestamp"`
}

type BlockDBCommandIndex struct{
	Collection    string      			`json:"collection"`	//要操作的数据表
	Index 		  map[string]string     `json:"index"`
	PublicKey     string      			`json:"public_key"`	//公钥
	Signature     string      			`json:"signature"`	//签名
	Timestamp  	  string      			`json:"timestamp"`
}