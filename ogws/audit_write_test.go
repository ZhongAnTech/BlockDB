package ogws

import (
	"encoding/json"
	"testing"
	"time"

	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
)

type InnerData struct {
	Type   string `json:"type"`
	Person Person `json:"person"`
	Loc    string `json:"loc"`
}

type Person struct {
	Age  int
	Name string
}

func TestWriter(t *testing.T) {
	var o = &AuditEvent{}
	o.AccountNonce = 23
	o.Hash = "hash"
	o.Height = 34
	o.MineNonce = 21
	o.Data = &AuditEventDetail{
		Ip:         "1.3.4.5",
		Identity:   "23455",
		PrimaryKey: "haah",
		Timestamp:  time.Now().Format(time.RFC3339),
		Data: InnerData{Person: Person{
			Age:  0,
			Name: "",
		},
			Type: "haha",
			Loc:  "shanghai",
		},
	}
	bytes, err := bson.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	var e = &AuditEvent{}
	err = bson.Unmarshal(bytes, e)
	out, err := json.MarshalIndent(e, "", "\t")
	t.Log(string(out))

	bytes, err = bson2.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	e = &AuditEvent{}
	err = bson2.Unmarshal(bytes, e)
	out, err = json.MarshalIndent(e, "", "\t")
	t.Log(string(out))
}
