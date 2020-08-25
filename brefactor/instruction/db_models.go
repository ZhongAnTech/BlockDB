package instruction

// current data
type MasterDataDoc struct {
	OpHash     string            `json:"op_hash"`
	Collection string            `json:"collection"` //操作的数据表
	Feature    CollectionFeature `json:"feature"`
	PublicKey  string            `json:"public_key"` //公钥
	Signature  string            `json:"signature"`  //签名
	Timestamp  int64             `json:"timestamp"`
}

// history data
type MasterHistoryDoc struct {
	OpHash     string            `json:"op_hash"`
	Version    int               `json:"version"`
	TxHash     string            `json:"tx_hash"`
	Collection string            `json:"collection"` //操作的数据表
	Feature    CollectionFeature `json:"feature"`
	PublicKey  string            `json:"public_key"` //公钥
	Signature  string            `json:"signature"`  //签名
	Timestamp  int64             `json:"timestamp"`
}

// operation
type MasterOpRecordDoc struct {
	OpHash     string            `json:"op_hash"`
	TxHash     string            `json:"tx_hash"`
	Collection string            `json:"collection"` //操作的数据表
	Feature    CollectionFeature `json:"feature"`
	PublicKey  string            `json:"public_key"` //公钥
	Signature  string            `json:"signature"`  //签名
	Timestamp  int64             `json:"timestamp"`
}

// info table
type MasterDocInfoDoc struct {
	Collection string `json:"collection"` //操作的数据表
	Version    int    `json:"version"`
	CreatedAt  int64  `json:"created_at"` // timestamp ms
	CreatedBy  string `json:"created_by"`
	ModifiedAt int64  `json:"modified_at"` // timestamp ms
	ModifiedBy string `json:"modified_by"`
}

// Audit table. merged to oprecord
type AuditModel struct {
	OpHash string `json:"_id"` //数据的hash
	//Collection string                 `json:"collection"` //操作的数据表
	Operation string                 `json:"operation"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`       //操作记录
	PublicKey string                 `json:"public_key"` //公钥
	Signature string                 `json:"signature"`  //签名
}

// OpDoc is the task queue filled by chain sync.
// update OpDoc once the OpDoc is executed.
type OpDoc struct {
	Order      int32  `json:"oder"`
	IsExecuted bool   `json:"is_executed"`
	TxHash     string `json:"tx_hash"`
	OpHash     string `json:"op_hash"`
	OpStr      string `json:"op_str"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"public_key"`
}

// oprecord table. One for each collection
type OpRecordDoc struct {
	DocId   string `json:"doc_id"`  // 文档Id
	OpHash  string `json:"op_hash"` //数据的hash
	Version int    `json:"version"`
	//Collection string                 `json:"collection"` //操作的数据表
	Operation string                 `json:"operation"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`       //操作记录
	PublicKey string                 `json:"public_key"` //公钥
	Signature string                 `json:"signature"`  //签名
}

// history table。
type HistoryDoc struct {
	DocId   string `json:"doc_id"` // 文档Id
	Version int    `json:"version"`
	//Collection string                 `json:"collection"` //操作的数据表
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"history"`    //历史版本
	PublicKey string                 `json:"public_key"` //公钥
	Signature string                 `json:"signature"`  //签名
}

// info table
type DocInfoDoc struct {
	DocId   string `json:"doc_id"` // 文档Id
	Version int    `json:"version"`
	//Collection   string `json:"collection"` //操作的数据表
	CreatedAt  int64  `json:"created_at"` // timestamp ms
	CreatedBy  string `json:"created_by"`
	ModifiedAt int64  `json:"modified_at"` // timestamp ms
	ModifiedBy string `json:"modified_by"`
}
