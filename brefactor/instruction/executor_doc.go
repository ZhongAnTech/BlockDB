package instruction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

func (t *InstructionExecutor) insertDoc(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)

	cmd := &InsertCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal insert gcmd.")
		return
	}

	version := 1
	ts := ts()

	// permission verification
	ok, err := t.PermissionVerify(Insert, cmd.Collection, cmd.PublicKey)
	if err != nil {
		logrus.WithError(err).Warn("error on insertDoc")
		return
	}
	if !ok {
		err = errors.New("user does not have permission to perform insertDoc")
		return
	}

	// TODO: insert doc data
	dataDoc := DataDoc{
		DocId:     gcmd.OpHash,
		Timestamp: ts,
		Data:      cmd.Data,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
	}
	dataDocM, err := toDoc(dataDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(cmd.Collection, DataType), dataDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create data document: " + cmd.Collection)
		// TODO: consider revert the changes or retry or something.
	}

	// TODO: insert doc info
	docInfoDoc := DocInfoDoc{
		DocId:      gcmd.OpHash,
		Version:    int64(version),
		CreatedAt:  ts,
		CreatedBy:  gcmd.PublicKey,
		ModifiedAt: ts,
		ModifiedBy: gcmd.PublicKey,
	}
	docInfoDocM, err := toDoc(docInfoDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(cmd.Collection, DocInfoType), docInfoDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create data DocInfoDoc document")
		// TODO: consider revert the changes or retry or something.
	}

	// TODO: insert doc history
	historyDoc := HistoryDoc{
		DocId:     gcmd.OpHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: ts,
		Version:   int64(version),
		Data:      cmd.Data,
	}
	err = t.InsertDocHistory(ctx, historyDoc, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: insert doc oprecord
	opRecordDoc := OpRecordDoc{
		DocId:     gcmd.OpHash,
		OpHash:    gcmd.OpHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: ts,
		Version:   int64(version),
		Operation: cmd.Op,
		Data:      cmd.Data,
	}
	err = t.InsertOpRecord(ctx, opRecordDoc, cmd.Collection)

	return
}

func (t *InstructionExecutor) updateDoc(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)

	cmd := &UpdateCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal update gcmd.")
		return err
	}

	// permission verification
	ok, err := t.PermissionVerify(Update, cmd.Collection, cmd.PublicKey)
	if err != nil {
		logrus.WithError(err).Warn("error on updateDoc")
		return
	}
	if !ok {
		err = errors.New("user does not have permission to perform updateDoc")
		return
	}

	actionTs := ts()
	id := cmd.Query["op_hash"]
	// get current version
	filter := bson.M{
		"doc_id": id,
	}
	oldVersion, err := t.GetCurrentVersion(ctx, filter, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: uodate doc info
	update := bson.M{
		"version":     oldVersion + 1,
		"modified_at": actionTs,
		"modified_by": cmd.PublicKey,
	}
	err = t.UpdateDocInfo(ctx, filter, update, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: update doc data
	//update set
	if len(cmd.Set) != 0 {
		set_update := bson.M{}
		for k, v := range cmd.Set {
			set_update["data."+k] = v
		}
		count, err := t.storageExecutor.Update(ctx, t.formatCollectionName(cmd.Collection, DataType),
			filter, set_update, "set")
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("unexpected update: results: %d", count)
		}
	}

	//update unset
	if len(cmd.Unset) != 0 {
		unset_update := bson.M{}
		for _, k := range cmd.Unset {
			unset_update["data."+k] = ""
		}
		count, err := t.storageExecutor.Update(ctx, t.formatCollectionName(cmd.Collection, DataType),
			filter, unset_update, "unset")
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("unexpected update: results: %d", count)
		}
	}

	dataDocCurrentMList, err := t.storageExecutor.Select(ctx,
		t.formatCollectionName(cmd.Collection, DataType),
		filter, nil, 1, 0)
	if err != nil {
		return
	}
	if len(dataDocCurrentMList.Content) == 0 {
		err = errors.New("data not found: " + cmd.Collection)
	}
	dataDocCurrentM := dataDocCurrentMList.Content[0]
	data, err := bson.Marshal(dataDocCurrentM)
	if err != nil {
		logrus.WithField("value", dataDocCurrentM).Warn("failed to marshal data")
		return
	}

	// do deserialization and validations
	dataDoc := DataDoc{}
	err = bson.Unmarshal(data, &dataDoc)
	if err != nil {
		logrus.WithField("value", dataDocCurrentM).Warn("failed to unmarshal data")
		return
	}

	// TODO: insert doc history
	historyDoc := HistoryDoc{
		DocId:     id,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,

		Version: oldVersion + 1,
		Data:    dataDoc.Data,
	}
	err = t.InsertDocHistory(ctx, historyDoc, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: insert doc oprecord
	opRecord := make(map[string]interface{})
	opRecord["query"] = cmd.Query
	opRecord["set"] = cmd.Set
	opRecord["unset"] = cmd.Unset

	opRecordDoc := OpRecordDoc{
		DocId:     id,
		OpHash:    gcmd.OpHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,
		Operation: cmd.Op,

		Version: oldVersion + 1,
		Data:    opRecord,
	}
	err = t.InsertOpRecord(ctx, opRecordDoc, cmd.Collection)

	return
}

func (t *InstructionExecutor) deleteDoc(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)

	cmd := &DeleteCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal delete gcmd.")
		return err
	}

	// permission verification
	ok, err := t.PermissionVerify(Delete, cmd.Collection, cmd.PublicKey)
	if err != nil {
		logrus.WithError(err).Warn("error on deleteDoc")
		return
	}
	if !ok {
		err = errors.New("user does not have permission to perform deleteDoc")
		return
	}

	actionTs := ts()
	op_hash := cmd.Query["op_hash"]
	// get current version
	filter := bson.M{
		"doc_id": op_hash,
	}
	oldVersion, err := t.GetCurrentVersion(ctx, filter, cmd.Collection)
	if err != nil {
		return
	}

	id, err := t.GetDocId(ctx, filter, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: update doc info
	update := bson.M{
		"version":     oldVersion + 1,
		"modified_at": actionTs,
		"modified_by": cmd.PublicKey,
	}
	err = t.UpdateDocInfo(ctx, filter, update, cmd.Collection)
	if err != nil {
		return
	}

	// TODO: delete doc data
	count, err := t.storageExecutor.Delete(ctx, t.formatCollectionName(cmd.Collection, DataType), id)
	if err != nil {
		return
	}
	if count != 1 {
		return fmt.Errorf("unexpected delete: results: %d", count)
	}

	// TODO: insert doc history
	historyDoc := HistoryDoc{
		DocId:     op_hash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,

		Version: oldVersion + 1,
		Data:    nil,
	}
	err = t.InsertDocHistory(ctx, historyDoc, cmd.Collection)
	if err != nil {
		return
	}

	opRecord := make(map[string]interface{})
	opRecord["query"] = cmd.Query

	// TODO: insert doc oprecord
	opRecordDoc := OpRecordDoc{
		DocId:     op_hash,
		OpHash:    gcmd.OpHash,
		PublicKey: gcmd.PublicKey,
		Signature: gcmd.Signature,
		Timestamp: actionTs,
		Operation: cmd.Op,

		Version: oldVersion + 1,
		Data:    opRecord,
	}
	err = t.InsertOpRecord(ctx, opRecordDoc, cmd.Collection)

	return
}

func (t *InstructionExecutor) InsertOpRecord(ctx context.Context, opRecordDoc OpRecordDoc, coll string) (err error) {
	opRecordDocM, err := toDoc(opRecordDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(coll, OpRecordType), opRecordDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create data opRecord document: " + coll)
		// TODO: consider revert the changes or retry or something.
	}
	return
}

func (t *InstructionExecutor) InsertDocHistory(ctx context.Context, historyDoc HistoryDoc, coll string) (err error) {
	historyDocM, err := toDoc(historyDoc)
	if err != nil {
		return
	}
	_, err = t.storageExecutor.Insert(ctx, t.formatCollectionName(coll, HistoryType), historyDocM)
	if err != nil {
		logrus.WithError(err).Error("failed to create data history document: " + coll)
		// TODO: consider revert the changes or retry or something.
	}
	return
}

func (t *InstructionExecutor) UpdateDocInfo(ctx context.Context, filter bson.M, update bson.M, coll string) (err error) {
	count, err := t.storageExecutor.Update(ctx, t.formatCollectionName(coll, DocInfoType),
		filter, update, "set")
	if err != nil {
		return
	}
	if count != 1 {
		return fmt.Errorf("unexpected update: results: %d", count)
	}
	return
}

func (t *InstructionExecutor) GetDocId(ctx context.Context, filter bson.M, coll string) (id string, err error) {
	docDataDocCurrentMList, err := t.storageExecutor.Select(ctx,
		t.formatCollectionName(coll, DataType),
		filter, nil, 1, 0)
	if err != nil {
		return "", err
	}
	if len(docDataDocCurrentMList.Content) == 0 {
		err = errors.New("data doc not found: " + coll)
	}

	docInfoDocCurrentM := docDataDocCurrentMList.Content[0]
	id = docInfoDocCurrentM["_id"].(primitive.ObjectID).Hex()

	return id, err
}
