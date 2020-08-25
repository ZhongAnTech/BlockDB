package instruction

import "encoding/json"

func (t *InstructionExecutor) createIndex(instruction OpContext) error {
	com := &IndexCommand{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal create_index command.")
		return err
	}
	//TODO: Verification of signature
	com.Timestamp = timestamp
	data := make(map[string]interface{})
	data["index"] = com.Index
	version, err := InsertInfo(com.OpHash, com.Collection, com.PublicKey, com.Timestamp)
	if err != nil {
		log.Println("failed to insert info.")
		return err
	}
	//Indexes=append(Indexes,com)
	blockdb := mongoutils.InitMgo(url, BlockDataBase, com.Collection)
	for k, v := range com.Index {
		_, err = blockdb.CreateIndex(k, "data."+v)
		if err != nil {
			log.Println("failed to create index for: data." + v)
			return err
		}
	}
	_ = blockdb.Close()

	err = OpRecord(CreateIndex, version, com.OpHash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	//HistoryRecord(com.OpHash,info.Version,com.Collection,timestamp,data,com.PublicKey,com.Signature)
	err = Audit(CreateIndex, com.OpHash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	return nil
}

func (t *InstructionExecutor) dropIndex(instruction OpContext) error {
	strconv.FormatInt(time.Now().Unix(), 10)

	com := &IndexCommand{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal drop_index command.")
		return err
	}
	//ok,index:=UpdateCollectionIndex(com.Collection,com.Index)
	com.Timestamp = ts()
	data := make(map[string]interface{})
	data["index"] = com.Index
	version, err := InsertInfo(com.OpHash, com.Collection, com.PublicKey, com.Timestamp)
	if err != nil {
		log.Println("failed to insert info.")
		return err
	}

	blockdb := mongoutils.InitMgo(url, BlockDataBase, com.Collection)
	for k := range com.Index {
		err = blockdb.DropIndex(k)
		if err != nil {
			log.Println("failed to drop index: " + k)
			return err
		}
	}
	_ = blockdb.Close()

	err = OpRecord(DropIndex, version, com.OpHash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	err = Audit(DropIndex, com.OpHash, com.Collection, timestamp, data, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	//data["index"]=index.Index
	//HistoryRecord("",com.Collection,timestamp,data,com.PublicKey,com.Signature)
	return nil
}
