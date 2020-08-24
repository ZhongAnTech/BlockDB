package core_interface

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

// BlockDBCommand is the raw data operation applied on ledger. no additional info
type BlockDBCommand interface {
}

// BLockDBMessage is the enriched message including BlockDBCommand.
type BlockDBMessage interface {
}

type BlockDBCommandProcessor interface {
	Process(command BlockDBCommand) (CommandProcessResult, error) // better to be implemented in async way.

}

type JsonCommandParser interface {
	FromJson(json string) (BlockDBCommand, error)
}

type BlockchainOperator interface {
	EnqueueSendToLedger(command BlockDBMessage) error
}

type CommandExecutor interface{}

type StorageExecutor interface {
	//插入的json对应对bson形式；返回插入成功后的主键id
	Insert(ctx context.Context, collectionName string, val bson.D) (string, error)
	//在collect中删除主键id为hash
	Delete(ctx context.Context, collectionName string, id string) (int64, error)
	/**
	filter:筛选条件 为空：则表示全取
	sort:排序条件 为空：则表示不排序
	limit:查找出来的数据量;为0： 则表示全部
	skip:跳过skip条文档 为0：则表示逐条取
	skip+limit：跳过skip个文档后，取limit个文档
	*/
	Select(ctx context.Context, collectionName string,
		filter bson.D, sort bson.D, limit int64, skip int64) (response SelectResponse, err error)
	//在collect中查找主键id为hash的文档
	SelectById(ctx context.Context, collectionName string, id string) (response SelectResponse, err error)
	//将filter更新为update
	Update(ctx context.Context, collectionName string, filter, update bson.D, operation string) (count int64, err error)
	//创建collection 返回创建失败的错误信息；成功则返回nil
	CreateCollection(ctx context.Context, collectionName string) (err error)
	//创建索引，返回创建后的索引名字
	CreateIndex(ctx context.Context, collectionName string, indexName, column string) (createdIndexName string, err error)
	//删除索引
	DropIndex(ctx context.Context, collectionName string, indexName string) (err error)
	//返回该collect对应的数据库大小、索引大小、文档个数、索引个数
	CollectionInfo(ctx context.Context, collection string) (resp CollectionInfoResponse, err error)
	//关闭连接
	Close(ctx context.Context) error
}

type LedgerSyncer interface{}

type BlockchainListener interface{}
