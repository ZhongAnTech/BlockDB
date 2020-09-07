package storage

import (
	"context"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
)

type MongoClient struct {
	client      *mongo.Client
	database    *mongo.Database
	collections map[string]*mongo.Collection // for collection cache
}

func Connect(ctx context.Context, url string, databaseName string, authMechanism string, username string, password string) *MongoClient {
	clientOptions := options.Client().ApplyURI(url)
	if authMechanism != "" {
		clientOptions.Auth = &options.Credential{
			AuthMechanism: authMechanism,
			//AuthMechanismProperties: nil,
			//AuthSource:              "",
			Username: username,
			Password: password,
			//PasswordSet:             false,
		}
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	mClient := &MongoClient{}
	mClient.client = client
	mClient.database = client.Database(databaseName)
	mClient.collections = make(map[string]*mongo.Collection)
	return mClient
}

//插入一个文档
func (mc *MongoClient) Insert(ctx context.Context, collectionName string, val bson.M) (string, error) {
	collect := mc.ensureColl(collectionName)
	id, err := collect.InsertOne(ctx, val)
	if err != nil {
		logrus.WithError(err).Warn("failed to insert to ")
	}
	return id.InsertedID.(primitive.ObjectID).Hex(), err

}

//根据key value删除集合下所有符合条件的文档
func (mc *MongoClient) Delete(ctx context.Context, collectionName string, id string) (int64, error) {
	collect := mc.ensureColl(collectionName)
	filter := bson.M{"_id": id}
	count, err := collect.DeleteMany(ctx, filter, nil)
	if err != nil {
		logrus.WithError(err).Warn("failed to delete")
	}
	return count.DeletedCount, err
}

//根据fileter查询文档
func (mc *MongoClient) Select(ctx context.Context, collectionName string,
	filter bson.M, sort bson.M, limit int64, skip int64) (response core_interface.SelectResponse, err error) {

	collect := mc.ensureColl(collectionName)

	result, err := collect.Find(ctx, filter, options.Find().SetSort(sort).SetLimit(limit).SetSkip(skip))
	if err != nil {
		logrus.WithError(err).Warn("failed to select")
		return
	}

	for result.Next(ctx) {
		var ele bson.M
		err := result.Decode(&ele)
		if err != nil {
			logrus.WithError(err).Warn("failed to select")
			return
		}
		response.Content = append(response.Content, ele)
	}
	return
}

//根据主键查数据
func (mc *MongoClient) SelectById(ctx context.Context, collectionName string, id string) (response core_interface.SelectResponse, err error) {
	collect := mc.ensureColl(collectionName)

	filter := bson.M{"_id": id}
	result, err := collect.Find(ctx, filter)
	if err != nil {
		logrus.WithError(err).Warn("failed to select")
		return
	}
	for result.Next(ctx) {
		var ele bson.M
		err := result.Decode(&ele)
		if err != nil {
			logrus.WithError(err).Warn("failed to select")
			return
		}
		response.Content = append(response.Content, ele)
	}
	return
}

//TODO根据filter更新所有符合条件的文档
func (mc *MongoClient) Update(ctx context.Context, collectionName string, filter, update bson.M, operation string) (count int64, err error) {
	collect := mc.ensureColl(collectionName)

	var result *mongo.UpdateResult
	switch operation {
	case "set":
		update1 := bson.M{"$set": update}
		result, err = collect.UpdateMany(ctx, filter, update1)
		if err != nil {
			logrus.WithError(err).Warn("failed to update")
			return
		}
	case "unset":
		update1 := bson.M{"$unset": update}
		result, err = collect.UpdateMany(ctx, filter, update1)
		if err != nil {
			logrus.WithError(err).Warn("failed to update")
			return
		}
	}
	count = result.ModifiedCount
	return
}
func (mc *MongoClient) CreateCollection(ctx context.Context, collectionName string) (err error) {
	res := mc.database.RunCommand(ctx, bson.M{"create": collectionName})
	err = res.Err()
	if err != nil {
		logrus.WithError(err).Warn("failed to create collection")
		return
	}
	return
}

//创建单个索引
func (mc *MongoClient) CreateIndex(ctx context.Context, collectionName string, indexName, column string) (createdIndexName string, err error) {
	collect := mc.ensureColl(collectionName)

	Doc := bsonx.Doc{{column, bsonx.Int32(1)}}
	idx := mongo.IndexModel{
		Keys:    Doc,
		Options: options.Index().SetUnique(false).SetName(indexName),
	}
	createdIndexName, err = collect.Indexes().CreateOne(ctx, idx)
	if err != nil {
		logrus.WithError(err).Warn("failed to create index")
	}
	return
}

//index名字
func (mc *MongoClient) DropIndex(ctx context.Context, collectionName string, indexName string) (err error) {
	collect := mc.ensureColl(collectionName)

	_, err = collect.Indexes().DropOne(ctx, indexName)
	if err != nil {
		logrus.WithError(err).Warn("failed to drop index")
	}
	return

}

//返回数据库大小、索引大小、文档个数、索引个数
func (mc *MongoClient) CollectionInfo(ctx context.Context, collection string) (resp core_interface.CollectionInfoResponse, err error) {
	res := mc.database.RunCommand(ctx, bson.M{"collStats": collection})
	var document bson.M
	err = res.Decode(&document)
	if err != nil {
		logrus.WithError(err).Warn("failed to get collection info")
	}

	resp = core_interface.CollectionInfoResponse{
		StorageSize:    document["storageSize"].(int32),
		TotalIndexSize: document["totalIndexSize"].(int32),
		Count:          document["count"].(int32),
		NIndexes:       document["nindexes"].(int32),
	}
	return
}

//func (m *MongoCollection)CreateAccount()error
func (mc *MongoClient) Close(ctx context.Context) error {
	err := mc.database.Client().Disconnect(ctx)

	if err != nil {
		logrus.WithError(err).Warn("failed to close")
	}
	return err
}

func (mc *MongoClient) ensureColl(name string) *mongo.Collection {
	if v, ok := mc.collections[name]; ok {
		return v
	} else {
		v = mc.database.Collection(name)
		mc.collections[name] = v
		return v
	}
}
