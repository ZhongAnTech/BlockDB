package instruction

type AuditModel struct {
	OpHash     string                 `json:"_id"`        //数据的hash
	Collection string                 `json:"collection"` //操作的数据表
	Operation  string                 `json:"operation"`
	Timestamp  string                 `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`       //操作记录
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
}

type Op struct {
	Order      int32  `json:"oder"`
	IsExecuted bool   `json:"is_executed"`
	TxHash     string `json:"tx_hash"`
	OpHash     string `json:"op_hash"`
	PublicKey  string `json:"public_key"`
	Signature  string `json:"signature"`
	OpStr      string `json:"op_str"`
}

type OperationRecords struct {
	OpHash     string                 `json:"op_hash"` //数据的hash
	Version    int                    `json:"version"`
	Collection string                 `json:"collection"` //操作的数据表
	Operation  string                 `json:"operation"`
	Timestamp  string                 `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`       //操作记录
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
}

type DocHistory struct {
	OpHash     string                 `json:"op_hash"` //数据的hash
	Version    int                    `json:"version"`
	Collection string                 `json:"collection"` //操作的数据表
	Timestamp  string                 `json:"timestamp"`
	Data       map[string]interface{} `json:"history"`    //历史版本
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
}

type DocInfo struct {
	OpHash       string `json:"op_hash"` //数据的hash
	Version      int    `json:"version"`
	Collection   string `json:"collection"` //操作的数据表
	CreateTime   string `json:"create_time"`
	CreateBy     string `json:"create_by"`
	LastModified string `json:"last_modified"`
}
