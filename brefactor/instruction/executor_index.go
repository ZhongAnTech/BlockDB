package instruction

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"log"
	"context"
)

func (t *InstructionExecutor) createIndex(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)
	cmd := &IndexCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal create_index gcmd.")
		return
	}

	// permission verification
	if !t.PermissionVerify(CreateIndex, cmd.Collection, cmd.PublicKey) {
		err = errors.New("user does not have permission to perform createIndex")
		logrus.WithError(err).Warn("error on createIndex")
		return
	}

	// TODO: create index
	for k,v:=range cmd.Index{
		_, err = t.storageExecutor.CreateIndex(ctx,t.formatCollectionName(cmd.Collection, DataType),k,"data."+v)
		if err != nil {
			logrus.WithError(err).Error("failed to create index on document: "+cmd.Collection)
			// TODO: consider revert the changes or retry or something.
		}
	}


	return
}

func (t *InstructionExecutor) dropIndex(gcmd GeneralCommand) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.WriteTimeout)
	cmd := &IndexCommand{}
	err = json.Unmarshal([]byte(gcmd.OpStr), cmd)
	if err != nil {
		log.Fatal("failed to unmarshal drop_index gcmd.")
		return
	}

	// permission verification
	if !t.PermissionVerify(CreateIndex, cmd.Collection, cmd.PublicKey) {
		err = errors.New("user does not have permission to perform dropIndex")
		logrus.WithError(err).Warn("error on dropIndex")
		return
	}


	//TODO: drop index
	for k:=range cmd.Index{
		err = t.storageExecutor.DropIndex(ctx,t.formatCollectionName(cmd.Collection, DataType),k)
		if err != nil {
			logrus.WithError(err).Error("failed to drop index on document: "+cmd.Collection)
			// TODO: consider revert the changes or retry or something.
		}
	}


	return
}
