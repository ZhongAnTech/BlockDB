package httpserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ShutdownTimeoutSeconds = 5

type HttpListener struct {
	router *gin.Engine
	server *http.Server
	Port   string
	C      *RpcController
}

func (srv *HttpListener) InitDefault() {
	router := NewRouter()
	router = srv.C.addRouter(router)
	srv.router = router
	srv.server = &http.Server{
		Addr:    ":" + srv.Port,
		Handler: srv.router,
	}
}

func (srv *HttpListener) Start() {
	logrus.Infof("listening Http on %s", srv.Port)
	go func() {
		// service connections
		if err := srv.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatalf("error in Http server")
		}
	}()
}

func (srv *HttpListener) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeoutSeconds*time.Second)
	defer cancel()
	if err := srv.server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("error while shutting down the Http server")
	}
	logrus.Infof("http server Stopped")
}

func (srv *HttpListener) Name() string {
	return fmt.Sprintf("HttpServer at Port %s", srv.Port)
}
