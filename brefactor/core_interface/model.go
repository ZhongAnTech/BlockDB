package core_interface

import "go.mongodb.org/mongo-driver/bson"

type CommandProcessResult struct {
	Hash string
	OK   bool
}

type DefaultBlockDBCommand struct {
}

type SelectResponse struct {
	Content []bson.M
}

type CollectionInfoResponse struct {
	StorageSize    int32
	TotalIndexSize int32
	Count          int32
	NIndexes       int32
}
