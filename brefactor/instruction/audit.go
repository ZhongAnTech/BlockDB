package instruction

import (
	"github.com/ZhongAnTech/BlockDB/brefactor/storage"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func Audit(op string, hash string, coll string, timestamp string, data map[string]interface{}, pk string, sig string) error {
	auditdb := storage.InitMongo(url, BlockDataBase, AuditCollection)
	audit := bson.D{{"op_hash", hash}, {"collection", coll}, {"operation", op}, {"timestamp", timestamp},
		{"data", data}, {"public_key", pk}, {"signature", sig}}
	_, err := auditdb.Insert(audit)
	if err != nil {
		log.Fatal("failed to insert data to history.")
		return err
	}
	_ = auditdb.Close()
	return nil
}

//{"op":"create_collection","name":"sample_collection","feature":{"allow_update":false, "allow_update_members": ["0x123456", "0x123456", "0x123456", "0x123456"]},"public_key": "0x769153474351324"}
