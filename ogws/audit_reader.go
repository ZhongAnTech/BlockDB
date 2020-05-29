package ogws

import (
	"context"
	"encoding/json"
	"time"

	bson2 "github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RawData struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuditEvent
}

func (l *MongoDBAuditWriter) Query(filter bson.M, limit, skip int64) (resp []RawData, count int64, err error) {
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
		var o RawData
		var event AuditEvent
		val := cur.Current.Lookup("_id")
		var id primitive.ObjectID
		err := val.Unmarshal(&id)
		if err != nil {
			logrus.WithError(err).WithField("val", val).Error("decode id  failed")
			continue
		}
		o.Id = id
		err = bson2.Unmarshal(cur.Current, &event)
		if err != nil {
			logrus.WithError(err).Error("decode failed")
			continue
		}
		o.AuditEvent = event
		resp = append(resp, o)
	}
	err = cur.Err()
	if err != nil {
		logrus.WithError(err).Error("read  err")
		return
	}
	return
}
