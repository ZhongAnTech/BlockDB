package instruction

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
)

func (t *InstructionExecutor) insertDoc(instruction OpContext) error {
	com := &InsertCommand{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal insert command.")
		return err
	}
	//TODO: Verification of signature
	if Check(Insert, com.Collection, com.PublicKey) {
		com.Timestamp = timestamp
		version, err := InsertInfo(com.OpHash, com.Collection, com.PublicKey, com.Timestamp)
		if err != nil {
			log.Println("failed to insert info.")
			return err
		}
		//inset data
		data := bson.M{{"op_hash", com.OpHash}, {"collection", com.Collection}, {"data", com.Data},
			{"public_key", com.PublicKey}, {"signature", com.Signature}, {"timestamp", com.Timestamp}}
		blockdb := mongoutils.InitMgo(url, BlockDataBase, com.Collection)
		_, err = blockdb.Insert(data)
		if err != nil {
			log.Println("failed to insert data, ophash: " + com.OpHash)
			return err
		}
		_ = blockdb.Close()

		err = OpRecord(Insert, version, com.OpHash, com.Collection, timestamp, com.Data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = HistoryRecord(Insert, com.OpHash, version, com.Collection, timestamp, com.Data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = Audit(Insert, com.OpHash, com.Collection, timestamp, com.Data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
	} else {
		log.Println("insert permission denied")
	}
	return nil
}

func (t *InstructionExecutor) updateDoc(instruction OpContext) error {
	com := &UpdateCommand{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal update command.")
		return err
	}
	//fmt.Println(com)
	//TODO: Verification of signature
	if Check(Update, com.Collection, com.PublicKey) {
		com.Timestamp = timestamp
		hash := com.Query["op_hash"]
		version, err := UpdateInfo(hash)
		if err != nil {
			log.Println("failed to update info.")
			return err
		}
		//fmt.Println("finish update info")
		data := make(map[string]interface{})
		data["query"] = com.Query
		data["set"] = com.Set
		data["unset"] = com.Unset
		blockdb := mongoutils.InitMgo(url, BlockDataBase, com.Collection)
		filter := bson.M{{"op_hash", hash}}
		if len(com.Set) != 0 {
			set_update := bson.M{}
			for k, v := range com.Set {
				set_update = append(set_update, bson.E{"data." + k, v})
			}
			_, err = blockdb.Update(filter, set_update, "set")
			if err != nil {
				log.Println("failed to update data.")
				return err
			}

		}
		if len(com.Unset) != 0 {
			unset_update := bson.M{}
			for _, k := range com.Unset {
				unset_update = append(unset_update, bson.E{"data." + k, ""})
			}
			_, err = blockdb.Update(filter, unset_update, "unset")
			if err != nil {
				log.Println("failed to update data.")
				return err
			}
		}
		_ = blockdb.Close()

		err = OpRecord(Update, version, hash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = HistoryRecord(Update, hash, version, com.Collection, timestamp, data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = Audit(Update, hash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
	} else {
		log.Println("update permission denied")
	}
	return nil
}

func (t *InstructionExecutor) deleteDoc(instruction OpContext) error {
	com := &DeleteCommand{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal delete command.")
		return err
	}
	//TODO: Verification of signature
	//权限验证
	if Check(Delete, com.Collection, com.PublicKey) {
		com.Timestamp = timestamp
		hash := com.Query["op_hash"]
		data := make(map[string]interface{})
		data["query"] = com.Query
		blockdb := mongoutils.InitMgo(url, BlockDataBase, com.Collection)
		_, err = blockdb.Delete(hash)
		if err != nil {
			log.Println("failed to delete data.")
			return err
		}
		_ = blockdb.Close()

		version, err := UpdateInfo(hash)
		if err != nil {
			log.Println("failed to update info.")
			return err
		}
		err = OpRecord(Delete, version, hash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = HistoryRecord(Delete, hash, version, com.Collection, timestamp, nil, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
		err = Audit(Delete, hash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
		if err != nil {
			return err
		}
	} else {
		log.Println("delete permission denied")
	}
	return nil
}
