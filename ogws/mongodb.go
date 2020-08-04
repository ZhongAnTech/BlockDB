package ogws

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

/*
 * MongoDB operator
 */
type MongoDBDatabase struct {
	config MongoDBConfig
	lock   sync.RWMutex
	client *mongo.Client
	coll   *mongo.Collection
}
type MongoDBConfig struct {
	Uri        string
	Database   string
	Collection string
	UserName   string
	Password   string
	AuthMethod string
}

func NewMongoDBDatabase(config MongoDBConfig) *MongoDBDatabase {
	database := &MongoDBDatabase{
		config: config,
	}
	database.ConnectMongoDB()
	return database
}

func (db *MongoDBDatabase) ConnectMongoDB() {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	// user Connection database

	// Set client options
	clientOptions := options.Client().ApplyURI(db.config.Uri).SetAuth(options.Credential{
		//AuthMechanism: db.config.AuthMethod,
		//AuthMechanismProperties: nil,
		AuthSource: db.config.Database,
		Username:   db.config.UserName,
		Password:   db.config.Password,
		//PasswordSet:             false,
	})

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		logrus.Fatal(err)
	}

	db.coll = client.Database(db.config.Database).Collection(db.config.Collection)

	// Check the connection
	err = client.Ping(ctx, nil)

	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Connected to user MongoDB!")

}

func (db *MongoDBDatabase) Stats() {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	// user Connection database

	// Set client options
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("%s", db.config.Uri)).SetAuth(options.Credential{
		//AuthMechanism:           "",
		//AuthMechanismProperties: nil,
		AuthSource: db.config.Database,
		Username:   db.config.UserName,
		Password:   db.config.Password,
		//PasswordSet:             false,
	})

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		logrus.Fatal(err)
	}
	mdb := client.Database(db.config.Database)

	result := mdb.RunCommand(context.Background(), bson.M{"collStats": db.config.Collection})

	var document bson.M
	err = result.Decode(&document)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Collection size: %v Bytes\n", document["size"])
	fmt.Printf("Average object size: %v Bytes\n", document["avgObjSize"])
	fmt.Printf("Storage size: %v Bytes\n", document["storageSize"])
	fmt.Printf("Total index size: %v Bytes\n", document["totalIndexSize"])
}

func (db *MongoDBDatabase) Put(value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := db.coll.InsertOne(ctx, value)
	return err
}

func (db *MongoDBDatabase) Has(key []byte) (bool, error) {
	v, err := db.Get(key)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

func (db *MongoDBDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.M{
		"_id": string(key),
	}
	v := db.coll.FindOne(ctx, filter)
	if v.Err() != nil {
		return nil, v.Err()
	}
	doc := bson.M{}

	err := v.Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}
	return []byte(doc["value"].(string)), nil
}

func (db *MongoDBDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.M{
		"_id": string(key),
	}
	_, err := db.coll.DeleteOne(ctx, filter)
	return err
}

func (db *MongoDBDatabase) Close() {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_ = db.client.Disconnect(ctx)
}
