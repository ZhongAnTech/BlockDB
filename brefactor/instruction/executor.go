package instruction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
	"time"
)

var (
	CollCollection     = "coll"
	HistoryCollection  = "history"
	OpRecordCollection = "oprecord"
	InfoCollection     = "info"
	AuditCollection    = "audit"
)

var NamePattern = map[string]string{
	CollCollection:     "%s_" + CollCollection,
	HistoryCollection:  "%s_" + HistoryCollection,
	OpRecordCollection: "%s_" + OpRecordCollection,
	InfoCollection:     "%s_" + InfoCollection,
	AuditCollection:    "%s_" + AuditCollection,
}

var InitCollections = []string{CollCollection, HistoryCollection, OpRecordCollection, InfoCollection, AuditCollection}

var (
	filter = bson.D{{"is_executed", "false"}}
	sort   = bson.D{{"oder", 1}}
)

type InstructionExecutorConfig struct {
	BatchSize     int64
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ErrorInterval time.Duration
}

type InstructionExecutor struct {
	Config          InstructionExecutorConfig
	storageExecutor core_interface.StorageExecutor
	quit            chan bool
}

func (t *InstructionExecutor) InitDefault() {
	t.quit = make(chan bool)
}

func (t *InstructionExecutor) Start() {
	go t.runCommand()
}

func (t *InstructionExecutor) Stop() {
	t.quit <- true
}

func (t *InstructionExecutor) Name() string {
	return "InstructionExecutor"
}

func (t *InstructionExecutor) formatCollectionName(collName string, collType string) string {
	return fmt.Sprintf(NamePattern[collType], collName)
}

// InitCollection setup all necessary collections to support business
func (t *InstructionExecutor) InitCollection(ctx context.Context, collName string) (err error) {
	for _, collType := range InitCollections {
		collectionFullName := t.formatCollectionName(collName, collType)
		err = t.storageExecutor.CreateCollection(ctx, collectionFullName)
		if err != nil {
			logrus.WithField("name", collectionFullName).Warn("failed to create collection")
			return err
		} else {
			logrus.WithField("name", collectionFullName).Info("collection created")
		}
		// create index for op_hash
		_, err = t.storageExecutor.CreateIndex(ctx, collectionFullName, "idx_op_hash", "op_hash")

		if err != nil {
			logrus.WithField("name", collectionFullName).Warn("failed to create index")
			return err
		} else {
			logrus.WithField("name", collectionFullName).Info("index created")
		}
	}
	return
}

// runCommand continuously fetches command from database.
func (t *InstructionExecutor) runCommand() {

	for {
		select {
		case <-t.quit:
			// TODO: do clean work.
			return
		default:
			didSome := t.doBatchJob()
			if didSome {
				continue
			}
			time.Sleep(t.Config.ErrorInterval)
		}
	}

}

func (t *InstructionExecutor) doBatchJob() (didSome bool) {

	ctx, _ := context.WithTimeout(context.Background(), t.Config.ReadTimeout)

	resp, err := t.storageExecutor.Select(ctx, CommandCollection, filter, sort, t.Config.BatchSize, 0)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch instructions")
		return false
	}
	if len(resp.Content) == 0 {
		logrus.Debug("no further instructions to be processed")
		return false
	}
	for _, ins := range resp.Content {
		// do deserialization and validations
		com := &Op{}
		err := json.Unmarshal([]byte(ins), com)
		if err != nil {
			log.Println("failed to unmarshal instruction.")
			continue
		}
		tem := make(map[string]interface{})
		err = json.Unmarshal([]byte(com.OpStr), &tem)
		if err != nil {
			log.Println("failed to unmarshal opStr.")
			continue
		}
		tem["op_hash"] = com.OpHash
		tem["signature"] = com.Signature
		op := tem["op"].(string)
		res, err := json.Marshal(tem)
		if err != nil {
			log.Println("failed to marshal opStr.")
			continue
		}
		err = Execute(op, string(res))
		if err != nil {
			log.Println(err)
			continue
		}

		//update excute state
		exe_filter := bson.D{{"op_hash", com.OpHash}}
		update := bson.D{{"is_executed", true}}
		_, err = instructiondb.Update(exe_filter, update, "set")
		if err != nil {
			log.Println("failed to update execute state.")
			continue
		}

	}
	return true
}

func Execute(op string, instruction string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	switch op {
	case CreateCollection:
		err := createColl(instruction, timestamp)
		if err != nil {
			return err
		}

	case UpdateCollection:
		err := updateColl(instruction, timestamp)
		if err != nil {
			return err
		}

	case Insert:
		err := insertDoc(instruction, timestamp)
		if err != nil {
			return err
		}

	case Update:
		err := updateDoc(instruction, timestamp)
		if err != nil {
			return err
		}

	case Delete:
		err := deleteDoc(instruction, timestamp)
		if err != nil {
			return err
		}

	case CreateIndex:
		err := createIndex(instruction, timestamp)
		if err != nil {
			return err
		}

	case DropIndex:
		err := dropIndex(instruction, timestamp)
		if err != nil {
			return err
		}

	}
	return nil
}

func createColl(instruction string, timestamp string) error {
	com := &BlockDBCommandCollection{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Fatal("failed to unmarshal create_collection command.")
		return err
	}
	//TODO: Verification of signature
	com.Timestamp = timestamp
	//缓存
	Colls = append(Colls, com)
	//doc info
	version, err := InsertInfo(com.OpHash, com.Collection, com.PublicKey, com.Timestamp)
	if err != nil {
		log.Fatal("failed to insert info.")
		return err
	}
	//create collection
	blockdb := mongoutils.InitMgo(url, BlockDataBase, "")
	err = blockdb.CreateCollection(com.Collection)
	if err != nil {
		log.Fatal("failed to connect to block collection.")
		return err
	}
	_ = blockdb.Close()

	//create index for op_hash
	blockdb = mongoutils.InitMgo(url, BlockDataBase, com.Collection)
	_, err = blockdb.CreateIndex("op_hash", "op_hash")
	if err != nil {
		log.Fatal("failed to create index for " + com.Collection + " on op_hash.")
	}
	_ = blockdb.Close()

	//insert
	colldb := mongoutils.InitMgo(url, BlockDataBase, CollCollection)
	data := bson.D{{"op_hash", com.OpHash}, {"collection", com.Collection}, {"feature", com.Feature},
		{"public_key", com.PublicKey}, {"signature", com.Signature}, {"timestamp", com.Timestamp}}
	_, err = colldb.Insert(data)
	if err != nil {
		log.Fatal("failed to insert data to colldb.")
		return err
	}
	_ = colldb.Close()

	//op record
	err = OpRecord(CreateCollection, version, com.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	//history
	err = HistoryRecord(CreateCollection, com.OpHash, version, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	//audit
	err = Audit(CreateCollection, com.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
	if err != nil {
		return err
	}
	return nil
}

func updateColl(instruction string, timestamp string) error {
	com := &BlockDBCommandCollection{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal update_collection command.")
		return err
	}
	//TODO: Verification of signature
	//权限验证
	if Check(UpdateCollection, com.Collection, com.PublicKey) {
		ok, curColl := UpdateCollectionFeatures(com.Collection, com.Feature)
		if ok {
			com.Timestamp = timestamp
			version, err := UpdateInfo(curColl.OpHash)
			if err != nil {
				log.Println("failed to update info.")
				return err
			}
			//update
			filter := bson.D{{"op_hash", curColl.OpHash}}
			colldb := mongoutils.InitMgo(url, BlockDataBase, CollCollection)
			update := bson.D{{"feature", com.Feature}}
			_, err = colldb.Update(filter, update, "set")
			if err != nil {
				log.Fatal("failed to insert data to colldb.")
				return err
			}
			_ = colldb.Close()

			err = OpRecord(UpdateCollection, version, curColl.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
			if err != nil {
				return err
			}
			err = HistoryRecord(UpdateCollection, curColl.OpHash, version, com.Collection, timestamp, curColl.Feature, com.PublicKey, com.Signature)
			if err != nil {
				return err
			}
			err = Audit(UpdateCollection, curColl.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
			if err != nil {
				return err
			}
		} else {
			log.Println("collection " + com.Collection + " doesn't exist.")
		}
	} else {
		log.Println("update_collection permission denied")
	}
	return nil
}

func insertDoc(instruction string, timestamp string) error {
	com := &BlockDBCommandInsert{}
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
		data := bson.D{{"op_hash", com.OpHash}, {"collection", com.Collection}, {"data", com.Data},
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

func updateDoc(instruction string, timestamp string) error {
	com := &BlockDBCommandUpdate{}
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
		filter := bson.D{{"op_hash", hash}}
		if len(com.Set) != 0 {
			set_update := bson.D{}
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
			unset_update := bson.D{}
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

func deleteDoc(instruction string, timestamp string) error {
	com := &BlockDBCommandDelete{}
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

func createIndex(instruction string, timestamp string) error {
	com := &BlockDBCommandIndex{}
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

func dropIndex(instruction string, timestamp string) error {
	com := &BlockDBCommandIndex{}
	err := json.Unmarshal([]byte(instruction), com)
	if err != nil {
		log.Println("failed to unmarshal drop_index command.")
		return err
	}
	//TODO: Verification of signature
	//ok,index:=UpdateCollectionIndex(com.Collection,com.Index)
	com.Timestamp = timestamp
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

//权限验证
func Check(op string, collection string, publickey string) bool {
	flag := false
outside:
	for _, coll := range Colls {
		if coll.Collection == collection {
			if coll.PublicKey == publickey {
				switch op {
				case Insert, UpdateCollection:
					flag = true
				case Update:
					if coll.Feature["allow_update"].(bool) == true {
						flag = true
					}
				case Delete:
					if coll.Feature["allow_delete"].(bool) == true {
						flag = true
					}
				}
			} else if coll.Feature["cooperate"].(bool) == true {
				switch op {
				case Insert:
					allows := coll.Feature["allow_insert_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag = true
						}
					}
				case Update:
					allows := coll.Feature["allow_update_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag = true
						}
					}
				case Delete:
					allows := coll.Feature["allow_delete_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag = true
						}
					}
				}
			}
			break outside
		}
	}
	return flag
}

//更新Coll
func UpdateCollectionFeatures(collection string, feature map[string]interface{}) (bool, *BlockDBCommandCollection) {
	flag := false
	var curColl *BlockDBCommandCollection
	for _, curColl = range Colls {
		if curColl.Collection == collection {
			curColl.Feature = feature
			flag = true
			//for k:=range feature{
			//	curColl.Feature[k]=feature[k]
			//	flag=true
			//}
			break
		}
	}
	return flag, curColl
}

//更新Indexes
//func UpdateCollectionIndex(collection string,index map[string]string)(bool,*BlockDBCommandIndex){
//	flag:=false
//	var curIndex *BlockDBCommandIndex
//	for _,curIndex=range Indexes{
//		if curIndex.Collection == collection{
//			for k:=range index{
//				delete(curIndex.Index,k)
//				flag=true
//			}
//			break
//		}
//	}
//	return flag,curIndex
//}
