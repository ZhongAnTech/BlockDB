package ogws

import (
	"context"
	"encoding/json"
	"time"

	bson2 "github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UpdateTimeFormat = "2006-01-02T15:04:05.999Z07:00"

type OriginalDataProcessor interface {
	DeleteOne(hash string) error
	DeleteMany(hashes []string) error
	InsertOne(hash string, data interface{}) error
	UpdateHash(Id primitive.ObjectID, hash string) error
	GetExpired(duration time.Duration, limit, offset int64) ([]OriginalData, int64, error)
	Query(f bson.M, limit, offset int64) ([]OriginalData, int64, error)
	GetCollection() *mongo.Collection
}

type originalDataProcessor struct {
	coll *mongo.Collection
}

type OriginalRawData struct {
	Data       interface{} `json:"data" bson:"data"`
	Hash       string      `json:"hash" bson:"hash"`
	UpdateTime string      `json:"update_time" bson:"update_time"`
}

type OriginalData struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OriginalRawData
}

func (o *originalDataProcessor) DeleteOne(hash string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	f := bson.M{"hash": hash}
	_, err := o.coll.DeleteOne(ctx, f)
	if err != nil {
		return err
	}
	return nil
}

func (o *originalDataProcessor) DeleteMany(hashes []string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	f := bson.M{"hash": bson.M{"$in": hashes}}
	_, err := o.coll.DeleteMany(ctx, f)
	if err != nil {
		return err
	}
	return nil
}

func (o *originalDataProcessor) InsertOne(hash string, data interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	od := &OriginalRawData{
		Data:       data,
		Hash:       hash,
		UpdateTime: time.Now().Format(UpdateTimeFormat),
	}
	bytes, err := bson.Marshal(od)
	if err != nil {
		return err
	}
	_, err = o.coll.InsertOne(ctx, bytes)
	if err != nil {
		return err
	}
	return nil
}

func (o *originalDataProcessor) UpdateHash(Id primitive.ObjectID, hash string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	f := bson.M{"_id": Id}
	_, err := o.coll.UpdateOne(ctx, f, bson.M{"hash": hash})
	if err != nil {
		return err
	}
	return nil
}

func (o *originalDataProcessor) GetExpired(duration time.Duration, limit, offset int64) (resp []OriginalData, count int64, err error) {
	timeFilter := time.Now().Add(-duration).Format(UpdateTimeFormat)
	filter := bson.M{"update_time": bson.M{"$lt": timeFilter}}
	return o.Query(filter, limit, offset)
}

func (l *originalDataProcessor) Query(filter bson.M, limit, skip int64) (resp []OriginalData, count int64, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 8*time.Second)
	if logrus.GetLevel() > logrus.DebugLevel {
		logData, _ := json.Marshal(&filter)
		logrus.WithField("filter", string(logData)).Trace("query filter")
	}
	count, err = l.coll.CountDocuments(ctx, filter)
	if err != nil {
		return
	}
	cur, err := l.coll.Find(ctx, filter, &options.FindOptions{Limit: &limit, Skip: &skip, Sort: bson.M{"_id": -1}})
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var o OriginalData
		var rawData OriginalRawData
		val := cur.Current.Lookup("_id")
		var id primitive.ObjectID
		err := val.Unmarshal(&id)
		if err != nil {
			logrus.WithError(err).WithField("val", val).Error("decode id  failed")
			continue
		}
		o.Id = id
		err = bson2.Unmarshal(cur.Current, &rawData)
		if err != nil {
			logrus.WithError(err).Error("decode failed")
			continue
		}
		o.OriginalRawData = rawData
		resp = append(resp, o)
	}
	err = cur.Err()
	if err != nil {
		logrus.WithError(err).Error("read  err")
		return
	}
	return
}

func (m *originalDataProcessor) GetCollection() *mongo.Collection {
	return m.coll
}
