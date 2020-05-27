package web

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/annchain/BlockDB/ogws"
	bson2 "github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RawData struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ogws.AuditEvent
}

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
	Total int64     `json:"total"`
	Data  []RawData `json:"data"`
}

type AuditDataGrammarRequest struct {
	Filter   bson.M `json:"filter"`
	PageNum  int64  `json:"page_num"`
	PageSize int64  `json:"page_size"`
}

func (request*AuditDataQueryRequest)ToFilter()bson.M{
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

func (l *HttpListener) Query(rw http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil || len(data) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}
	var request AuditDataQueryRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	filter:=request.ToFilter()
	if request.PageNum < 1 {
		request.PageNum = 1
	}
	if request.PageSize < 1 {
		request.PageNum = 10
	}
	skip := (request.PageNum - 1) * request.PageSize
	ctx, _ := context.WithTimeout(context.Background(), 8*time.Second)
	resp,err := l.queryData(ctx,filter,request.PageSize,skip)
	if err != nil {
		logrus.WithError(err).Error("read failed")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	RespOk(rw, resp)
	return

}


func (l *HttpListener) QueryGrammar(rw http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil || len(data) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}
	var request AuditDataGrammarRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	filter := request.Filter
	if request.PageNum < 1 {
		request.PageNum = 1
	}
	if request.PageSize < 1 {
		request.PageNum = 10
	}
	skip := (request.PageNum - 1) * request.PageSize

	ctx, _ := context.WithTimeout(context.Background(), 8*time.Second)

	resp,err := l.queryData(ctx,filter,request.PageSize,skip)
	if err != nil {
		logrus.WithError(err).Error("read failed")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	RespOk(rw, resp)
	return
}

func (l*HttpListener) queryData(ctx context.Context, filter bson.M ,limit,skip int64 ) (*AuditDataQueryResponse ,error) {
	if logrus.GetLevel() > logrus.DebugLevel {
		logData, _ := json.Marshal(&filter)
		logrus.WithField("filter", string(logData)).Trace("query filter")
	}
	count, err := l.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil,err
	}
	cur, err := l.coll.Find(ctx, filter, &options.FindOptions{Limit: &limit, Skip: &skip, Sort: bson.M{"_id": -1}})
	if err != nil {
		return nil,err
	}
	defer cur.Close(ctx)
	resp := &AuditDataQueryResponse{
		Total:count,
	}
	for cur.Next(ctx) {
		var o RawData
		var event ogws.AuditEvent
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
		resp.Data = append(resp.Data, o)
	}
	err = cur.Err()
	if err!=nil {
		logrus.WithError(err).Error("read  err")
		return nil,err
	}
	return resp,nil
}



func RespOk(rw http.ResponseWriter, result interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	data, err := json.Marshal(result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Write(data)
}
