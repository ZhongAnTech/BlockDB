package instruction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// InitCollection setup all necessary collections to support business
func (t *InstructionExecutor) InitCollection(ctx context.Context, collName string) (err error) {
	// TODO: insert into master collection

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

func (t *InstructionExecutor) createCollection(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)

	cmd := &CollectionCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal create_collection gcmd.")
		return err
	}
	//缓存
	//Colls = append(Colls, cmd)

	// TODO: permission check on creating collection
	// YOU may need an additional table to record master permissions

	// create master data record
	version := 1
	ts := ts()
	masterDoc := MasterDataDoc{
		OpHash:     gcmd.OpHash,
		Signature:  gcmd.Signature,
		PublicKey:  gcmd.PublicKey,
		Timestamp:  ts,
		Collection: cmd.Collection,
		Feature:    cmd.Feature,
	}
	masterDocM, err := toDoc(masterDoc)
	if err != nil {
		return
	}

	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, DataType), masterDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master data document")
		// TODO: consider revert the changes or retry or something.
	}

	// insert master history
	masterHistoryDoc := MasterHistoryDoc{
		OpHash:    gcmd.OpHash,
		TxHash:    gcmd.TxHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: ts,

		Version:    version,
		Collection: cmd.Collection,
		Feature:    cmd.Feature,
	}
	masterHistoryDocM, err := toDoc(masterHistoryDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, HistoryType), masterHistoryDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master history document")
		// TODO: consider revert the changes or retry or something.
	}

	// insert master oprecord
	masterOpRecordDoc := MasterOpRecordDoc{
		OpHash:    gcmd.OpHash,
		TxHash:    gcmd.TxHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: ts,

		Collection: cmd.Collection,
		Feature:    cmd.Feature,
	}
	masterOpRecordDocM, err := toDoc(masterOpRecordDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, OpRecordType), masterOpRecordDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master opRecord document")
		// TODO: consider revert the changes or retry or something.
	}

	// create master doc info
	masterDocInfoDoc := MasterDocInfoDoc{
		Collection: cmd.Collection,
		Version:    version,
		CreatedAt:  ts,
		CreatedBy:  gcmd.PublicKey,
		ModifiedAt: ts,
		ModifiedBy: gcmd.PublicKey,
	}
	masterDocInfoDocM, err := toDoc(masterDocInfoDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, DocInfoType), masterDocInfoDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master DocInfoDoc document")
		// TODO: consider revert the changes or retry or something.
	}

	// create collection and its supporting collections
	err = t.InitCollection(ctx, cmd.Collection)
	if err != nil {
		logrus.WithError(err).Error("failed to init collection")
		return err
	}
	return
}

func (t *InstructionExecutor) updateColl(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)

	cmd := &CollectionCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		logrus.WithError(err).Warn("failed to unmarshal update_collection gcmd.")
		return
	}

	actionTs := ts()

	// permission verification
	if !t.PermissionVerify(UpdateCollection, cmd.Collection, cmd.PublicKey) {
		err = errors.New("user does not have permission to perform updateColl")
		logrus.WithError(err).Warn("error on updateColl")
		return
	}

	// get current version
	filter := bson.M{
		"collection": cmd.Collection,
	}
	masterDocInfoDocCurrentMList, err := t.storageExecutor.Select(ctx,
		t.formatCollectionName(MasterCollection, DocInfoType),
		filter, nil, 1, 0)
	if err != nil {
		return
	}

	if len(masterDocInfoDocCurrentMList.Content) == 0 {
		err = errors.New("master doc info now found: " + cmd.Collection)
	}

	masterDocInfoDocCurrentM := masterDocInfoDocCurrentMList.Content[0]
	id = masterDocInfoDocCurrentM["_id"]
	oldVersion := masterDocInfoDocCurrentM["version"].(int)

	// TODO: update master_docinfo
	update := bson.M{
		"version":     oldVersion + 1,
		"modified_at": actionTs,
		"modified_by": cmd.PublicKey,
	}
	count, err := t.storageExecutor.Update(ctx, t.formatCollectionName(MasterCollection, DocInfoType),
		filter, update, "set")
	if err != nil {
		return
	}
	if count != 1 {
		return fmt.Errorf("unexpected update: results: %d", count)
	}

	// TODO: update master_data

	// TODO: insert master_history

	// TODO: insert master_oprecord

	//ok, curColl := UpdateCollectionFeatures(com.Collection, com.Feature)
	//if ok {
	//	com.Timestamp = timestamp
	//	version, err := UpdateInfo(curColl.OpHash)
	//	if err != nil {
	//		log.Println("failed to update info.")
	//		return err
	//	}
	//	//update
	//	filter := bson.M{{"op_hash", curColl.OpHash}}
	//	colldb := mongoutils.InitMgo(url, BlockDataBase, CollCollection)
	//	update := bson.M{{"feature", com.Feature}}
	//	_, err = colldb.Update(filter, update, "set")
	//	if err != nil {
	//		log.Fatal("failed to insert data to colldb.")
	//		return err
	//	}
	//	_ = colldb.Close()
	//
	//	err = OpRecord(UpdateCollection, version, curColl.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
	//	if err != nil {
	//		return err
	//	}
	//	err = HistoryRecord(UpdateCollection, curColl.OpHash, version, com.Collection, timestamp, curColl.Feature, com.PublicKey, com.Signature)
	//	if err != nil {
	//		return err
	//	}
	//	err = Audit(UpdateCollection, curColl.OpHash, com.Collection, timestamp, com.Feature, com.PublicKey, com.Signature)
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	log.Println("collection " + com.Collection + " doesn't exist.")
	//}
	return
}
