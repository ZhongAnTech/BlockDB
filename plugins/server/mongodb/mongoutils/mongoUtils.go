package mongoutils

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
)

type storageUtil interface {
	//m：插入的json对应对bson形式；返回插入成功后的主键id
	Insert(val bson.D) (string, error)
	//在collect中删除主键id为hash
	Delete(hash string)(int64,error)
	/**
	filter:筛选条件 为空：则表示全取
	sort:排序条件 为空：则表示不排序
	limit:查找出来的数据量;为0： 则表示全部
	skip:跳过skip条文档 为0：则表示逐条取
	skip+limit：跳过skip个文档后，取limit个文档
	*/
	Select(filter bson.D,sort bson.D,limit int64,skip int64)(Response,error)
	//在collect中查找主键id为hash的文档
	SelectById(hash string)(Response,error)
	//将filter更新为update
	Update(filter, update bson.D,operation string)(int64,error)
	//返回该collect对应的数据库大小、索引大小、文档个数、索引个数
	CollectInfor(collection string)(interface{})
	//创建collection 返回创建失败的错误信息；成功则返回nil
	CreateCollection(collection string) error
	//创建索引，返回创建后的索引名字
	CreateIndex(indexName,column string)(string,error)

	//删除索引
	DropIndex(indexName string)error
	CreateAccount() string
	//关闭连接
	Close()error
}
type Mgo struct {
	database *mongo.Database
	collect *mongo.Collection
}
type Response struct {
	Content []string
}
func  InitMgo(url string,database string,collection string) Mgo{
	mgo:=Mgo{}
	clientOptions:=options.Client().ApplyURI(url)
	client,error:=mongo.Connect(context.TODO(),clientOptions)
	if error!=nil{
		log.Fatal(error)
	}
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	if database==""{
		mgo.database=nil
	}else {
		mgo.database = client.Database(database)
	}

	if collection==""{
		mgo.collect=nil
	}else{
		mgo.collect=mgo.database.Collection(collection)
	}
	return mgo
}
//插入一个文档
func (mc *Mgo)Insert(val bson.D) (string, error){
	if mc.collect==nil{
		return "操作失败：",errors.New("缺少要操作的collection")
	}
	id, err :=mc.collect.InsertOne(context.TODO(),val)
	if err!=nil{
		log.Fatal(err)
	}
	return id.InsertedID.(primitive.ObjectID).Hex(),err
}
//根据key value删除集合下所有符合条件的文档
func (mc *Mgo)Delete(hash string)(int64,error){
	if mc.collect==nil{
		return 0,errors.New("缺少要操作的collection")
	}
	id,_ := primitive.ObjectIDFromHex(hash)
	filter:=bson.M{"_id":id}
	count, err := mc.collect.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		log.Fatal(err)
	}
	return count.DeletedCount,err
}
//根据fileter查询文档
func (mc *Mgo)Select(filter bson.D,sort bson.D,limit int64,skip int64)(Response,error){
	if mc.collect==nil{
		return Response{nil},errors.New("缺少要操作的collection")
	}
	result,error:= mc.collect.Find(context.TODO(), filter,options.Find().SetSort(sort).SetLimit(limit).SetSkip(skip))
	if error!=nil{
		log.Fatal(error)
	}
	var response Response
	for result.Next(context.TODO()) {
		response.Content= append(response.Content,result.Current.String())
	}
	return response,error
}
//根据主键查数据
func (mc *Mgo)SelectById(hash string)(Response,error){
	if mc.collect==nil{
		return Response{nil},errors.New("缺少要操作的collection")
	}
	id,_ := primitive.ObjectIDFromHex(hash)
	filter:=bson.M{"_id":id}
	result,error:= mc.collect.Find(context.TODO(), filter)
	if error!=nil{
		log.Fatal(error)
	}
	var response Response
	for result.Next(context.TODO()) {
		response.Content= append(response.Content,result.Current.String())
	}

	return response,error
}
//TODO根据filter更新所有符合条件的文档
func (mc *Mgo)Update(filter, update bson.D,operation string)(int64,error){
	if mc.collect==nil{
		return 0,errors.New("缺少要操作的collection")
	}
	var result *mongo.UpdateResult
	var err error
	switch operation {
	case "set":
		update1:= bson.M{"$set":update}
		result, err = mc.collect.UpdateMany(context.TODO(), filter, update1)
		if err != nil {
			log.Fatal(err)
		}
	case "unset":
		update1:= bson.M{"$unset":update}
		result, err = mc.collect.UpdateMany(context.TODO(), filter, update1)
		if err != nil {
			log.Fatal(err)
		}
	}
	return result.UpsertedCount,err
}
func (mc *Mgo)CreateCollection(collection string) error{
	if mc.database==nil{
		return errors.New("操作失败：没有指定数据库")
	}
	res:=mc.database.RunCommand(context.TODO(),bson.M{"create":collection})
	if res.Err()!=nil {
		log.Fatal(res.Err())
	}
	return res.Err()
}
//创建单个索引
func (mc *Mgo)CreateIndex(indexName,column string)(string,error){
	if mc.collect==nil{
		return "操作失败：",errors.New("缺少要操作的collection")
	}
	Doc:=bsonx.Doc{{column ,bsonx.Int32(1)}}
	idx:=mongo.IndexModel{
		Keys: Doc,
		Options: options.Index().SetUnique(false).SetName(indexName),
	}
	name,err:=mc.collect.Indexes().CreateOne(context.TODO(),idx)
	if err!=nil{
		log.Fatal(err)
	}
	return name,err
}
//index名字
func (mc *Mgo) DropIndex(indexName string)error{
	if mc.collect==nil{
		return errors.New("缺少要操作的collection")
	}
	_,err:=mc.collect.Indexes().DropOne(context.TODO(),indexName)
	if err != nil {
		log.Fatal(err)
	}
	return err

}
//返回数据库大小、索引大小、文档个数、索引个数
func (mc *Mgo)CollectInfor(collection string)(interface{}){
	if mc.database==nil{
		return errors.New("缺少要操作的数据库")
	}
	if mc.collect==nil{
		return errors.New("缺少要操作的collection")
	}
	res:=mc.database.RunCommand(context.TODO(),bson.M{"collStats":collection})
	var document bson.M
	err := res.Decode(&document)
	if err!=nil{
		log.Fatal(err)
	}
	response:= struct {
		storageSize interface{}
		totalIndexSize interface{}
		count interface{}
		indexes interface{}
	}{
		storageSize: document["storageSize"],
		totalIndexSize:document["totalIndexSize"],
		count:document["count"],
		indexes:document["nindexes"],
	}
	return response
}
//func (m *Mgo)CreateAccount()error
func (mc *Mgo) Close()error{
	if mc.database==nil{
		return errors.New("缺少要操作的数据库")
	}
	err:= mc.database.Client().Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	return err
}
