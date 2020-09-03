package instruction

// current data
type MasterDataDoc struct {
	OpHash     string            `bson:"op_hash"`
	Collection string            `bson:"collection"` //操作的数据表
	Feature    CollectionFeature `bson:"feature"`
	PublicKey  string            `bson:"public_key"` //公钥
	Signature  string            `bson:"signature"`  //签名
	Timestamp  int64             `bson:"timestamp"`
}

// history data
type MasterHistoryDoc struct {
	OpHash     string            `bson:"op_hash"`
	Version    int64               `bson:"version"`
	TxHash     string            `bson:"tx_hash"`
	Collection string            `bson:"collection"` //操作的数据表
	Feature    CollectionFeature `bson:"feature"`
	PublicKey  string            `bson:"public_key"` //公钥
	Signature  string            `bson:"signature"`  //签名
	Timestamp  int64             `bson:"timestamp"`
}

// operation
type MasterOpRecordDoc struct {
	OpHash     string            `bson:"op_hash"`
	TxHash     string            `bson:"tx_hash"`
	Collection string            `bson:"collection"` //操作的数据表
	Feature    CollectionFeature `bson:"feature"`
	PublicKey  string            `bson:"public_key"` //公钥
	Signature  string            `bson:"signature"`  //签名
	Timestamp  int64             `bson:"timestamp"`
}

// info table
type MasterDocInfoDoc struct {
	Collection string `bson:"collection"` //操作的数据表
	Version    int64    `bson:"version"`
	CreatedAt  int64  `bson:"created_at"` // timestamp ms
	CreatedBy  string `bson:"created_by"`
	ModifiedAt int64  `bson:"modified_at"` // timestamp ms
	ModifiedBy string `bson:"modified_by"`
}

// Audit table. merged to oprecord
type AuditModel struct {
	OpHash string `bson:"_id"` //数据的hash
	//Collection string                 `json:"collection"` //操作的数据表
	Operation string                 `bson:"operation"`
	Timestamp string                 `bson:"timestamp"`
	Data      map[string]interface{} `bson:"data"`       //操作记录
	PublicKey string                 `bson:"public_key"` //公钥
	Signature string                 `bson:"signature"`  //签名
}

// OpDoc is the task queue filled by chain sync.
// update OpDoc once the OpDoc is executed.
type OpDoc struct {
	Order      int32  `bson:"oder"`
	IsExecuted bool   `bson:"is_executed"`
	TxHash     string `bson:"tx_hash"`
	OpHash     string `bson:"op_hash"`
	OpStr      string `bson:"op_str"`
	Signature  string `bson:"signature"`
	PublicKey  string `bson:"public_key"`
}

// data table
type DataDoc struct {
	DocId   string `bson:"doc_id"` // 文档Id
	Timestamp int64                 `bson:"timestamp"`
	Data      map[string]interface{} `bson:"data"`
	PublicKey string                 `bson:"public_key"` //公钥
	Signature string                 `bson:"signature"`  //签名
}


// oprecord table. One for each collection
type OpRecordDoc struct {
	DocId   string `bson:"doc_id"`  // 文档Id
	OpHash  string `bson:"op_hash"` //数据的hash
	Version int64    `bson:"version"`
	//Collection string                 `json:"collection"` //操作的数据表
	Operation string                 `bson:"operation"`
	Timestamp int64                 `bson:"timestamp"`
	Data      map[string]interface{} `bson:"data"`       //操作记录
	PublicKey string                 `bson:"public_key"` //公钥
	Signature string                 `bson:"signature"`  //签名
}

// history table。
type HistoryDoc struct {
	DocId   string `bson:"doc_id"` // 文档Id
	Version int64    `bson:"version"`
	//Collection string                 `json:"collection"` //操作的数据表
	Timestamp int64                 `bson:"timestamp"`
	Data      map[string]interface{} `bson:"history"`    //历史版本
	PublicKey string                 `bson:"public_key"` //公钥
	Signature string                 `bson:"signature"`  //签名
}

// info table
type DocInfoDoc struct {
	DocId   string `bson:"doc_id"` // 文档Id
	Version int64    `bson:"version"`
	//Collection   string `json:"collection"` //操作的数据表
	CreatedAt  int64  `bson:"created_at"` // timestamp ms
	CreatedBy  string `bson:"created_by"`
	ModifiedAt int64  `bson:"modified_at"` // timestamp ms
	ModifiedBy string `bson:"modified_by"`
}
