package ogws

import (
	"go.mongodb.org/mongo-driver/bson"
)

// AuditWriter is the OG
type AuditWriter interface {
	// receive OG event and write it to the backend storage
	WriteOGMessage(o *AuditEvent) error
}

type MongoDBAuditWriter struct {
	db *MongoDBDatabase
}

func NewMongoDBAuditWriter(db *MongoDBDatabase) *MongoDBAuditWriter {
	return &MongoDBAuditWriter{
		db: db,
	}
}

func (m *MongoDBAuditWriter) WriteOGMessage(o *AuditEvent) error {
	bytes, err := bson.Marshal(o)
	if err != nil {
		return err
	}
	err = m.db.Put(bytes)
	if err != nil {
		return err
	}
	return nil
}
