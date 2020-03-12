package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/annchain/BlockDB/ogws"
	"github.com/annchain/BlockDB/plugins/client/og"
	"github.com/annchain/BlockDB/processors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type RpcController struct {
	Processor             *og.OgProcessor
	AuditConnectionString string
	AuditDatabase         string
	AuditCollection       string
}

type HashResponse struct {
	Hash string `json:"hash"`
}

func (rc *RpcController) addRouter(router *gin.Engine) *gin.Engine {
	router.GET("/docs/:hash", rc.GetDoc)
	router.GET("/query", rc.QueryDoc)
	router.POST("/doc", rc.PostDoc)
	return router
}

func Response(c *gin.Context, status int, err error, data interface{}) {
	var msg interface{}
	if err != nil {
		msg = err.Error()
	}
	c.JSON(status, gin.H{
		"err":  msg,
		"data": data,
	})
}

func (rc *RpcController) GetDoc(ct *gin.Context) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(rc.AuditConnectionString))
	if err != nil {
		Response(ct, 500, err, nil)
		return
	}
	collection := client.Database(rc.AuditDatabase).Collection(rc.AuditCollection)

	logrus.Info(ct.Param("hash"))
	filter := bson.M{"hash": ct.Param("hash")}

	var result ogws.AuditEvent
	mongoResult := collection.FindOne(ctx, filter)
	err = mongoResult.Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			Response(ct, 404, errors.New("not found"), nil)
			return
		}

		logrus.WithError(err).Warn("failed to decode")
		Response(ct, 500, err, nil)
		return
	}

	Response(ct, 200, nil, result)
	return

}

func (rc *RpcController) PostDoc(ct *gin.Context) {
	body := ct.Request.Body
	userReq, err := ioutil.ReadAll(body)
	if err != nil {
		Response(ct, 400, err, nil)
		return
	}
	var c processors.LogEvent
	if err := json.Unmarshal(userReq, &c); err != nil {
		Response(ct, 400, err, nil)
		return
	}
	resp, err := rc.Processor.SendToLedger(c)
	if err != nil {
		Response(ct, 400, err, nil)
		return
	}

	if ss, ok := resp.(string); ok {
		Response(ct, 200, nil, HashResponse{Hash: ss})
	}

}

func (rc *RpcController) QueryDoc(ct *gin.Context) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(rc.AuditConnectionString))
	if err != nil {
		logrus.WithError(err).Error("error on connecting server")
		Response(ct, 500, errors.New("server error"), nil)
		return
	}
	collection := client.Database(rc.AuditDatabase).Collection(rc.AuditCollection)

	filter := bson.M{}

	for key, value := range ct.Request.URL.Query() {
		keys := strings.Split(key, "-")
		if len(keys) > 2 {
			Response(ct, 400, errors.New("key should be string or int-key format"), nil)
			return
		}
		if len(keys) == 1 {
			filter[key] = value[0]
		} else {
			if keys[0] == "int" {
				v, err := strconv.ParseInt(value[0], 0, 0)
				if err != nil {
					Response(ct, 400, errors.New("bad int format"), nil)
				}
				filter[keys[1]] = v
			} else if keys[0] == "float" {
				v, err := strconv.ParseFloat(value[0], 0)
				if err != nil {
					Response(ct, 400, errors.New("bad float format"), nil)
				}
				filter[keys[1]] = v
			} else if keys[0] == "bool" {
				v, err := strconv.ParseBool(value[0])
				if err != nil {
					Response(ct, 400, errors.New("bad bool format"), nil)
				}
				filter[keys[1]] = v
			} else if keys[0] == "string" {
				filter[keys[1]] = value[0]
			} else {
				Response(ct, 400, errors.New("unknown query type. supports int, float, bool, string"), nil)
			}
		}

	}
	logrus.WithField("dic", filter).Info("query")

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		logrus.WithError(err).Warn("failed to decode")
		Response(ct, 500, err, nil)
		return
	}

	docs := []ogws.AuditEvent{}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result ogws.AuditEvent
		err := cursor.Decode(&result)
		if err != nil {
			logrus.WithError(err).Warn("failed to decode")
			Response(ct, 500, err, nil)
			return
		}
		docs = append(docs, result)
	}

	Response(ct, 200, nil, docs)
	return
}
