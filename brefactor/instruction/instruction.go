package instruction

const (
	CreateCollection string = "create_collection"
	UpdateCollection string = "update_collection"
	Insert           string = "insert"
	Update           string = "update"
	Delete           string = "delete"
	CreateIndex      string = "hint_create_index"
	DropIndex        string = "hint_drop_index"

	CommandCollection string = "_op"
	MasterCollection  string = "_master"
)

//var Colls []*CollectionCommand

//var Indexes []*IndexCommand

type CollectionFeature struct {
	AllowUpdate        bool     `json:"allow_update"`
	AllowDelete        bool     `json:"allow_delete"`
	Cooperate          bool     `json:"cooperate"`
	AllowInsertMembers []string `json:"allow_insert_members"`
	AllowUpdateMembers []string `json:"allow_update_members"`
	AllowDeleteMembers []string `json:"allow_delete_members"`
}

type GeneralCommand struct {
	TxHash    string `json:"tx_hash"`
	OpHash    string `json:"op_hash"`
	OpStr     string `json:"op_str"`
	Signature string `json:"signature"`
	PublicKey string `json:"public_key"`
}

type CollectionCommand struct {
	//OpHash     string                 `json:"op_hash"`    //产生数据的hash作为主键
	Op         string            `json:"op"`
	Collection string            `json:"collection"` //要操作的数据表
	Feature    CollectionFeature `json:"feature"`
	PublicKey  string            `json:"public_key"` //公钥
}

type InsertCommand struct {
	//OpHash     string                 `json:"op_hash"`    //产生数据的hash作为主键
	Op         string                 `json:"op"`
	Collection string                 `json:"collection"` //要操作的数据表
	Data       map[string]interface{} `json:"data"`
	PublicKey  string                 `json:"public_key"` //公钥
}

type UpdateCommand struct {
	//OpHash     string                 `json:"op_hash"`
	Op         string                 `json:"op"`
	Collection string                 `json:"collection"` //要操作的数据表
	Query      map[string]string      `json:"query"`
	Set        map[string]interface{} `json:"set"`
	Unset      []string               `json:"unset"`
	PublicKey  string                 `json:"public_key"` //公钥
}

type DeleteCommand struct {
	//OpHash     string            `json:"op_hash"`
	Op         string            `json:"op"`
	Collection string            `json:"collection"` //要操作的数据表
	Query      map[string]string `json:"query"`
	PublicKey  string            `json:"public_key"` //公钥
}

type IndexCommand struct {
	//OpHash     string            `json:"op_hash"`
	Op         string            `json:"op"`
	Collection string            `json:"collection"` //要操作的数据表
	Index      map[string]string `json:"index"`
	PublicKey  string            `json:"public_key"` //公钥
}
