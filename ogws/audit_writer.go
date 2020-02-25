package ogws

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// AuditWriter is the OG
type AuditWriter interface {
	// receive OG event and write it to the backend storage
	WriteOGMessage(o *AuditEvent) error
}

type MongoDBAuditWriter struct {
	connectionString string
	coll             *mongo.Collection
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
	m.createUsersIndex()
	return m
}

func (m *MongoDBAuditWriter) createUsersIndex() {
	unique := true
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := m.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"hash": 1},
			Options: &options.IndexOptions{Unique: &unique}},
	})
	if err != nil {
		logrus.WithError(err).Warn("create index error")
	}
}
