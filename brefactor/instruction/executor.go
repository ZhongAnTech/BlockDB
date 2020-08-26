package instruction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

var (
	DataType     = "data"     // for actual data storage
	HistoryType  = "history"  // data history versions
	OpRecordType = "oprecord" //
	DocInfoType  = "info"     // document info
	//AuditCollection = "audit"
)

var NamePattern = map[string]string{
	DataType:     "%s_" + DataType,
	HistoryType:  "%s_" + HistoryType,
	OpRecordType: "%s_" + OpRecordType,
	DocInfoType:  "%s_" + DocInfoType,
	//AuditCollection: "%s_" + AuditCollection,
}

var InitCollections = []string{DataType, HistoryType, OpRecordType, DocInfoType}

var (
	filter = bson.M{"is_executed": "false"}
	sort   = bson.M{"oder": 1}
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
	jumpTable       map[string]func(opContext GeneralCommand) error
}

func (t *InstructionExecutor) InitDefault() {
	t.quit = make(chan bool)
	t.jumpTable = map[string]func(opStr GeneralCommand) error{
		CreateCollection: t.createCollection,
		UpdateCollection: t.updateColl,
		Insert:           t.insertDoc,
		Update:           t.updateDoc,
		Delete:           t.deleteDoc,
		CreateIndex:      t.createIndex,
		DropIndex:        t.dropIndex,
	}
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
		// to json
		data, err := bson.Marshal(ins)
		if err != nil {
			logrus.WithField("value", ins).Warn("failed to marshal command")
			continue
		}

		// do deserialization and validations
		op := OpDoc{}
		err = bson.Unmarshal(data, &op)
		if err != nil {
			logrus.WithField("value", ins).Warn("failed to unmarshal op")
			continue
		}

		// TODO: signature validation (do not validate inside executor)
		// TODO: hash validation

		//opStrObject["op_hash"] = op.OpHash
		//opStrObject["signature"] = op.Signature

		err = t.Execute(GeneralCommand{
			TxHash:    op.TxHash,
			OpHash:    op.OpHash,
			OpStr:     op.OpStr,
			Signature: op.Signature,
			PublicKey: op.PublicKey,
		})
		if err != nil {
			logrus.WithError(err).WithField("op", op.OpStr).Error("failed to execute op")
			// TODO: retry or mark as failed, according to err type
			continue
		}

		ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)
		// update excute state
		exeFilter := bson.M{"op_hash": op.OpHash}
		update := bson.M{"is_executed": true}
		_, err = t.storageExecutor.Update(ctx, CommandCollection, exeFilter, update, "set")
		if err != nil {
			log.Println("failed to update execute state.")
			continue
		}
	}
	return true
}

func (t *InstructionExecutor) Execute(command GeneralCommand) (err error) {
	opStrObject := make(map[string]interface{})
	err = json.Unmarshal([]byte(command.OpStr), &opStrObject)
	if err != nil {
		return errors.New("failed to unmarshal opStr")
	}

	op := opStrObject["op"].(string)
	opFunction, ok := t.jumpTable[op]

	if !ok {
		return errors.New("unsupported op: " + op)
	}

	err = opFunction(command)
	return
}

//更新Coll
func (t *InstructionExecutor) UpdateCollectionFeatures(collection string, feature map[string]interface{}) (bool, *CollectionCommand) {
	flag := false
	var curColl *CollectionCommand
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
//func UpdateCollectionIndex(collection string,index map[string]string)(bool,*IndexCommand){
//	flag:=false
//	var curIndex *IndexCommand
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
