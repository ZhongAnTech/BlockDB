package web

import (
	"strings"

	"github.com/annchain/BlockDB/ogws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditDataQueryRequest struct {
	Type string `json:"type"`

	Ip string `json:"ip"`

	PrimaryKey string `json:"primary_key"`

	Timestamp      string `json:"timestamp"`
	Identity       string `json:"identity"`
	OtherCondition bson.M `json:"other_condition"`
	PageNum        int64  `json:"page_num"`
	PageSize       int64  `json:"page_size"`
}

type AuditDataQueryResponse struct {
	Total int64          `json:"total"`
	Data  []ogws.RawData `json:"data"`
}

type AuditDataGrammarRequest struct {
	Filter   bson.M `json:"filter"`
	PageNum  int64  `json:"page_num"`
	PageSize int64  `json:"page_size"`
}

type OriginalDataQueryResponse struct {
	Total int64               `json:"total"`
	Data  []ogws.OriginalData `json:"data"`
}

type OriginalDataRequest struct {
	Filter   bson.M `json:"filter"`
	PageNum  int64  `json:"page_num"`
	PageSize int64  `json:"page_size"`
}

type CommonResponse struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func (request *AuditDataQueryRequest) ToFilter() bson.M {
	userId := request.Identity
	filter := bson.M{}
	if request.Ip != "" {
		filter["data.ip"] = bson.M{"$regex": primitive.Regex{
			Pattern: request.Ip,
			Options: "i",
		}}
	}

	if request.PrimaryKey != "" {
		filter["data.primarykey"] = bson.M{"$regex": primitive.Regex{
			Pattern: request.PrimaryKey,
			Options: "i",
		}}
	}

	if str := strings.Split(request.Timestamp, ";"); len(str) == 2 {
		filter["data.timestamp"] = bson.M{
			"$gte": str[0],
			"$lt":  str[1],
		}
	} else if request.Timestamp != "" {
		filter["data.timestamp"] = bson.M{"$regex": primitive.Regex{
			Pattern: request.Timestamp,
			Options: "i",
		}}
	}
	if request.Type != "" {
		filter["data.type"] = bson.M{"$regex": primitive.Regex{
			Pattern: request.Type,
			Options: "i",
		}}
	}
	if userId != "" {
		filter["data.identity"] = userId
	}
	if len(request.OtherCondition) > 0 {
		if len(filter) > 0 {
			var filters []bson.M
			filters = append(filters, filter, request.OtherCondition)
			filter = bson.M{"$or": filters}
		}
	}
	return filter
}
