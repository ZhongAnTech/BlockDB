package storage

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func TestMgo(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	mgo := Connect(ctx, "mongodb://paichepai.win:27017", "test", "", "", "")
	hex1, err := mgo.Insert(ctx, "coll", bson.D{{"a", 1}, {"b", "abc"}})
	if err != nil {
		t.Error("fail to insert: ", err)
	}

	hex2, err := mgo.Insert(ctx, "coll", bson.D{{"a", 2}, {"b", "efg"}})
	if err != nil {
		t.Error("fail to insert: ", err)
	}

	_, err = mgo.Update(ctx, "coll", bson.D{{"a", 1}, {"b", "abc"}}, bson.D{{"a", 3}, {"b", "klm"}}, "set")
	if err != nil {
		t.Error("fail to update: ", err)
	}

	response, err := mgo.Select(ctx, "coll", bson.D{{"a", bson.D{{"$ne", nil}}}}, bson.D{{"a", -1}}, 0, 0)
	if err != nil {
		t.Error("fail to select: ", err)
	}
	fmt.Println(response)

	ciResp, err := mgo.CollectionInfo(ctx, "coll")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ciResp)

	_, err = mgo.Delete(ctx, "coll", hex1)
	if err != nil {
		t.Error("fail to delete: ", err)
	}

	_, err = mgo.Delete(ctx, "coll", hex2)
	if err != nil {
		t.Error("fail to delete: ", err)
	}

	hex3, err := mgo.Insert(ctx, "coll1", bson.D{{"a", 1}, {"b", "abc"}})
	if err != nil {
		t.Error("fail to insert: ", err)
	}

	_, err = mgo.Delete(ctx, "coll1", hex3)
	if err != nil {
		t.Error("fail to delete: ", err)
	}
}
