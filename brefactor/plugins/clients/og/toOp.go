package og

import (
	"encoding/json"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/plugins/serve/mongo"
	"go.mongodb.org/mongo-driver/bson"

	"sort"
	"strings"
)

type allData struct {
	Type        int         `json:"type"`
	Transaction string      `json:"transaction"`
	Sequencer   string      `json:"sequencer"`
	archive     interface{} `json:"archive"`
}

type Archive struct {
	Height       int      `json:"hieght"`
	Type         int      `json:"type"`
	TxHash       string   `json:"tx_hash"`
	OpHash       string   `json:"op_hash"`
	PublicKey    string   `json:"public_key"`
	Signature    string   `json:"signature"`
	Parents      []string `json:"parents"`
	AccountNonce int      `json:"account_nonce"`
	MindNonce    int      `json:"mind_nonce"`
	Weight       int      `json:"weight"`
	Data         string   `json:"data"`
}

type Op struct {
	Order      int    `json:"order"`
	Height     int    `json:"height"`
	IsExecuted bool   `json:"is_executed"`
	TxHash     string `json:"tx_hash"`
	OpHash     string `json:"op_hash"`
	PublicKey  string `json:"public_key"`
	Signature  string `json:"signature"`
	OpStr      string `json:"op_str"`
}

var order int32 = 0

type Archives []Archive
type ByHash struct {
	Archives
}

func (s ByHash) Len() int {
	panic("implement me")
}

func (s ByHash) Swap(i, j int) {
	panic("implement me")
}

func (s ByHash) Less(i, j int) bool {
	return s.Archives[i].TxHash < s.Archives[j].TxHash
}

func ToStruct(str string) Archives {
	s1 := strings.Split(str, "},")
	s2 := strings.Split(s1[0], "\"data\": {")
	fmt.Println(s2[1])

	s3 := strings.Split(s2[1], "\"archive\":")
	fmt.Println(s3[1])

	var archiveMsg Archive
	//反序列化
	err := json.Unmarshal([]byte(s3[1]), &archiveMsg)
	if err != nil {
		fmt.Printf("unmarshal err = %v\n", err)
	}
	fmt.Printf("反序列化后 Data = %v\n", archiveMsg)

	//pubKeyBytes,err := hex.DecodeString(archiveMsg.PublicKey)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//pubKey,err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//opHash, err := hex.DecodeString(archiveMsg.OpHash)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//signatureBytes, err := hex.DecodeString(archiveMsg.Signature)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//data := Normalize(string(archiveMsg.Data))
	//hash := sha256.Sum256([]byte(data))
	//
	//if !bytes.Equal(opHash, hash[:]) {
	//	fmt.Println("invalid op_hash")
	//}
	//
	//isSuccess, err := pubKey.Verify(hash[:], signatureBytes)
	//if err != nil || !isSuccess {
	//	fmt.Println("invalid signature")
	//}

	var archiveMsgs []Archive
	archiveMsgs = append(archiveMsgs, archiveMsg)
	fmt.Println("archiveMsgs: ", archiveMsgs)
	return archiveMsgs
}

func test(archiveMsgs []Archive) {
	sort.Sort(ByHash{archiveMsgs})
	for i, v := range archiveMsgs {
		var op = Op{
			Order:      i,
			Height:     v.Height,
			IsExecuted: false,
			TxHash:     v.TxHash,
			OpHash:     v.OpHash,
			PublicKey:  v.PublicKey,
			Signature:  v.Signature,
			OpStr:      v.Data,
		}

		fmt.Println("op: ", op)
		mgo := mongo.InitMgo("mongodb://localhost:27017", "test", "op")
		mgo2 := mongo.InitMgo("mongodb://localhost:27017", "test", "isOnChain")

		//update := bson.D{{"$set", data}}
		id, err := mgo.Insert(bson.D{
			{"is_executed", op.IsExecuted},
			{"tx_hash", op.TxHash},
			{"op_hash", op.OpHash},
			{"public_key", op.PublicKey},
			{"signature", op.Signature},
			{"op_str", op.OpStr},
		})
		fmt.Println(id, err)

		update := bson.D{
			{"tx_hash", op.TxHash},
			{"op_hash", op.OpHash},
			{"status", 0},
		}

		update2 := bson.D{
			{"tx_hash", op.TxHash},
			{"op_hash", op.OpHash},
			{"status", 1},
		}
		mgo2.Update(update, update2, "unset")
	}

}
