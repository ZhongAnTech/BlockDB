package instruction

import (
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
	"time"
)

func OpRecord(op string, version int, hash string, coll string, timestamp string, data map[string]interface{}, pk string, sig string) error {
	//oprecord:=OpRecordDoc{hash,version,coll,op,timestamp,data,pk,sig}
	//TODO: Insert(HistoryDataBase,OpRecordType,oprecord)
	historydb := mongoutils.InitMgo(url, BlockDataBase, OpRecordType)
	record := bson.M{{"op_hash", hash}, {"version", version}, {"collection", coll}, {"operation", op},
		{"timestamp", timestamp}, {"data", data}, {"public_key", pk}, {"signature", sig}}
	_, err := historydb.Insert(record)
	if err != nil {
		log.Fatal("failed to insert data to OpRecord.")
		return err
	}
	_ = historydb.Close()
	return nil
}

func HistoryRecord(op string, hash string, version int, coll string, timestamp string, data map[string]interface{}, pk string, sig string) error {
	historydb := mongoutils.InitMgo(url, BlockDataBase, HistoryType)
	switch op {
	case Update:
		blockdb := mongoutils.InitMgo(url, BlockDataBase, coll)
		filter := bson.M{{"op_hash", hash}}
		response, err := blockdb.Select(filter, bson.M{}, 10, 0)
		if err != nil {
			return err
		}
		if len(response.Content) == 0 {
			return errors.New("not found ophash in history.")
		}
		c := &InsertCommand{}
		err = json.Unmarshal([]byte(response.Content[0]), &c)
		if err != nil {
			return err
		}
		data = c.Data
		_ = blockdb.Close()
	}
	record := bson.M{{"op_hash", hash}, {"version", version}, {"collection", coll}, {"timestamp", timestamp},
		{"data", data}, {"public_key", pk}, {"signature", sig}}
	_, err := historydb.Insert(record)
	if err != nil {
		log.Fatal("failed to insert data to history.")
		return err
	}
	_ = historydb.Close()
	return nil
}

func InsertInfo(hash string, coll string, pubkey string, timestamp string) (int, error) {
	infodb := mongoutils.InitMgo(url, BlockDataBase, DocInfoType)
	filter := bson.M{{"op_hash", hash}}
	response, err := infodb.Select(filter, bson.M{}, 10, 0)
	if err != nil {
		return -1, err
	}
	if len(response.Content) > 0 {
		return -1, errors.New("ophash hash existed.")
	}
	info := &DocInfoDoc{hash, 0, coll, timestamp, pubkey, timestamp}
	info_data := bson.M{{"op_hash", info.OpHash}, {"version", 0}, {"collection", info.Collection},
		{"create_time", info.CreateTime}, {"create_by", info.CreateBy}, {"last_modified", info.LastModified}}
	_, err = infodb.Insert(info_data)
	if err != nil {
		return -1, err
	}
	_ = infodb.Close()
	return 0, nil
}

func UpdateInfo(hash string) (int, error) {
	infodb := mongoutils.InitMgo(url, BlockDataBase, DocInfoType)
	filter := bson.M{{"op_hash", hash}}
	response, err := infodb.Select(filter, bson.M{}, 10, 0)
	if err != nil {
		return -1, err
	}
	if len(response.Content) == 0 {
		return -1, errors.New("ophash doesn't exist.")
	}
	c := make(map[string]interface{})
	err = json.Unmarshal([]byte(response.Content[0]), &c)
	if err != nil {
		return -1, err
	}
	version_map := c["version"].(map[string]interface{})
	version, err := strconv.Atoi(version_map["$numberInt"].(string))
	if err != nil {
		return -1, err
	}
	version = version + 1
	lastModified := strconv.FormatInt(time.Now().Unix(), 10)
	update := bson.M{{"version", version}, {"last_modified", lastModified}}
	_, err = infodb.Update(filter, update, "set")
	if err != nil {
		return -1, err
	}
	_ = infodb.Close()
	return version, nil
}

func GetOpRecordsById(hash string, coll string) []OpRecordDoc {
	var res []OpRecordDoc
	//TODO:SelectById(hash)
	return res
}

func GetHistoryRecord(hash string, coll string) []HistoryDoc {
	var res []HistoryDoc
	//TODO:SelectById(hash)
	return res
}
