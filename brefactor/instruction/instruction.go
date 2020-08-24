package instruction

const (
	CreateCollection  string = "create_collection"
	UpdateCollection  string = "update_collection"
	Insert            string = "insert"
	Update            string = "update"
	Delete            string = "delete"
	CreateIndex       string = "hint_create_index"
	DropIndex         string = "hint_drop_index"
	CommandDataBase   string = "test"
	CommandCollection string = "op"
	BlockDataBase     string = "block"
)

var Colls []*BlockDBCommandCollection

//var Indexes []*BlockDBCommandIndex

type BlockDBCommandCollection struct {
	OpHash     string                 `json:"op_hash"`    //产生数据的hash作为主键
	Collection string                 `json:"collection"` //要操作的数据表
	Feature    map[string]interface{} `json:"feature"`
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
	Timestamp  string                 `json:"timestamp"`
}

type BlockDBCommandInsert struct {
	OpHash     string                 `json:"op_hash"`    //产生数据的hash作为主键
	Collection string                 `json:"collection"` //要操作的数据表
	Data       map[string]interface{} `json:"data"`
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
	Timestamp  string                 `json:"timestamp"`
}

type BlockDBCommandUpdate struct {
	OpHash     string                 `json:"op_hash"`
	Collection string                 `json:"collection"` //要操作的数据表
	Query      map[string]string      `json:"query"`
	Set        map[string]interface{} `json:"set"`
	Unset      []string               `json:"unset"`
	PublicKey  string                 `json:"public_key"` //公钥
	Signature  string                 `json:"signature"`  //签名
	Timestamp  string                 `json:"timestamp"`
}

type BlockDBCommandDelete struct {
	OpHash     string            `json:"op_hash"`
	Collection string            `json:"collection"` //要操作的数据表
	Query      map[string]string `json:"query"`
	PublicKey  string            `json:"public_key"` //公钥
	Signature  string            `json:"signature"`  //签名
	Timestamp  string            `json:"timestamp"`
}

type BlockDBCommandIndex struct {
	OpHash     string            `json:"op_hash"`
	Collection string            `json:"collection"` //要操作的数据表
	Index      map[string]string `json:"index"`
	PublicKey  string            `json:"public_key"` //公钥
	Signature  string            `json:"signature"`  //签名
	Timestamp  string            `json:"timestamp"`
}
