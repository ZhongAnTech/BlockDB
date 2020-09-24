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

		Version:    int64(version),
		Collection: cmd.Collection,
		Feature:    cmd.Feature,
	}
	err = t.InsertMasterHistory(ctx, masterHistoryDoc)
	if err != nil {
		return
	}

	// create master doc info
	masterDocInfoDoc := MasterDocInfoDoc{
		Collection: cmd.Collection,
		Version:    int64(version),
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
	err = t.InsertMasterOpRecord(ctx, masterOpRecordDoc)
	if err != nil {
		return
	}

	//TODO:set collection feature to cache

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
	ok, err := t.PermissionVerify(UpdateCollection, cmd.Collection, cmd.PublicKey)
	if err != nil {
		logrus.WithError(err).Warn("error on updateColl")
		return
	}
	if !ok {
		err = errors.New("user does not have permission to perform updateColl")
		return
	}

	// get current version
	filter := bson.M{
		"collection": cmd.Collection,
	}
	oldVersion, err := t.GetCurrentVersion(ctx, filter, MasterCollection)
	if err != nil {
		return
	}

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
	update = bson.M{
		"feature": cmd.Feature,
	}
	count, err = t.storageExecutor.Update(ctx, t.formatCollectionName(MasterCollection, DataType),
		filter, update, "set")
	if err != nil {
		return
	}
	if count != 1 {
		return fmt.Errorf("unexpected update: results: %d", count)
	}

	masterDataDocCurrentMList, err := t.storageExecutor.Select(ctx,
		t.formatCollectionName(MasterCollection, DataType),
		filter, nil, 1, 0)
	if err != nil {
		return
	}
	if len(masterDataDocCurrentMList.Content) == 0 {
		err = errors.New("master data not found: " + cmd.Collection)
	}
	masterDataDocCurrentM := masterDataDocCurrentMList.Content[0]
	data, err := bson.Marshal(masterDataDocCurrentM)
	if err != nil {
		logrus.WithField("value", masterDataDocCurrentM).Warn("failed to marshal master data")
		return
	}

	// do deserialization and validations
	masterDataDoc := MasterDataDoc{}
	err = bson.Unmarshal(data, &masterDataDoc)
	if err != nil {
		logrus.WithField("value", masterDataDocCurrentM).Warn("failed to unmarshal master data")
		return
	}

	// TODO: insert master_history
	masterHistoryDoc := MasterHistoryDoc{
		OpHash:    gcmd.OpHash,
		TxHash:    gcmd.TxHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,

		Version:    oldVersion + 1,
		Collection: cmd.Collection,
		Feature:    masterDataDoc.Feature,
	}
	err = t.InsertMasterHistory(ctx, masterHistoryDoc)
	if err != nil {
		return
	}

	// TODO: insert master_oprecord
	masterOpRecordDoc := MasterOpRecordDoc{
		OpHash:    gcmd.OpHash,
		TxHash:    gcmd.TxHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,

		Collection: cmd.Collection,
		Feature:    cmd.Feature,
	}
	err = t.InsertMasterOpRecord(ctx, masterOpRecordDoc)

	return
}

func (t *InstructionExecutor) InsertMasterOpRecord(ctx context.Context, masterOpRecordDoc MasterOpRecordDoc) (err error) {
	masterOpRecordDocM, err := toDoc(masterOpRecordDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, OpRecordType), masterOpRecordDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master opRecord document")
		// TODO: consider revert the changes or retry or something.
	}
	return
}

func (t *InstructionExecutor) InsertMasterHistory(ctx context.Context, masterHistoryDoc MasterHistoryDoc) (err error) {
	masterHistoryDocM, err := toDoc(masterHistoryDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(MasterCollection, HistoryType), masterHistoryDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create master history document")
		// TODO: consider revert the changes or retry or something.
	}
	return
}

func (t *InstructionExecutor) GetCurrentVersion(ctx context.Context, filter bson.M, coll string) (cur int64, err error) {
	docInfoDocCurrentMList, err := t.storageExecutor.Select(ctx,
		t.formatCollectionName(coll, DocInfoType),
		filter, nil, 1, 0)
	if err != nil {
		return -1, err
	}
	if len(docInfoDocCurrentMList.Content) == 0 {
		err = errors.New("data doc info not found: " + coll)
	}

	docInfoDocCurrentM := docInfoDocCurrentMList.Content[0]
	cur = docInfoDocCurrentM["version"].(int64)

	return cur, err
}
