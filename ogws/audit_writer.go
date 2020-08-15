package ogws

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// AuditWriter is the OG
type AuditWriter interface {
	// receive OG event and write it to the backend storage
	WriteOGMessage(o *AuditEvent) error
	GetCollection() *mongo.Collection
	GetOriginalDataProcessor() OriginalDataProcessor
	Query(f bson.M, limit, offset int64) ([]RawData, int64, error)
}

type MongoDBAuditWriter struct {
	connectionString      string
	coll                  *mongo.Collection
	originalDataProcessor OriginalDataProcessor
}

func (m *MongoDBAuditWriter) WriteOGMessage(o *AuditEvent) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	bytes, err := bson.Marshal(o)
	if err != nil {
		return err
	}
	_, err = m.coll.InsertOne(ctx, bytes)
	if err != nil {
		return err
	}
	return nil
}

func NewMongoDBAuditWriter(connectionString string, database string, collection string) *MongoDBAuditWriter {
	m := &MongoDBAuditWriter{connectionString: connectionString}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.connectionString))
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to audit mongo")
	}
	coll := client.Database(database).Collection(collection)
	m.coll = coll
	m.originalDataProcessor = &originalDataProcessor{
		coll: client.Database(database).Collection(collection + "_original_data_"),
	}
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logrus.WithError(err).Error("ping mongo err,will panic")
		panic(err)
	}
	m.createUsersIndex(m.coll)
	m.createUsersIndex(m.originalDataProcessor.GetCollection())
	return m
}

func (m *MongoDBAuditWriter) createUsersIndex(coll *mongo.Collection) {
	unique := true
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"hash": 1},
			Options: &options.IndexOptions{Unique: &unique}},
	})
	if err != nil {
		logrus.WithError(err).Warn("create index error")
	}
}

func (m *MongoDBAuditWriter) GetCollection() *mongo.Collection {
	return m.coll
}

func (m *MongoDBAuditWriter) GetOriginalDataProcessor() OriginalDataProcessor {
	return m.originalDataProcessor
}
