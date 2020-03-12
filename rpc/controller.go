package rpc

import (
	"github.com/annchain/BlockDB/performance"
	"github.com/gin-gonic/gin"
)

type RpcController struct {
	Monitor performance.PerformanceMonitor
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
}

func (rc *RpcController) addRouter(router *gin.Engine) *gin.Engine {
	router.GET("/status", rc.Status)
}

func Response(c *gin.Context, status int, err error, data interface{}) {
	var msg string
	if err != nil {
		msg = err.Error()
	}
	c.JSON(status, gin.H{
		"err":  msg,
		"data": data,
	})
}

func (rc *RpcController) Status(c *gin.Context) {
	cors(c)
	reports := rc.Monitor.CollectData()
	Response(c, 200, nil, reports)
}
