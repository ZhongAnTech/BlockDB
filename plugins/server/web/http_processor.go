package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/annchain/BlockDB/backends"
	"github.com/annchain/BlockDB/ogws"
	"github.com/annchain/BlockDB/processors"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HttpListenerConfig struct {
	Port             int
	EnableAudit      bool
	EnableHealth     bool
	MaxContentLength int64
}

type HttpListener struct {
	config        HttpListenerConfig
	ledgerWriter  backends.LedgerWriter
	dataProcessor processors.DataProcessor

	wg          sync.WaitGroup
	stopped     bool
	router      *mux.Router
	auditWriter ogws.AuditWriter
}

func (l *HttpListener) Name() string {
	return "HttpListener"
}

func NewHttpListener(config HttpListenerConfig, dataProcessor processors.DataProcessor, ledgerWriter backends.LedgerWriter, auditWriter ogws.AuditWriter) *HttpListener {
	if config.MaxContentLength == 0 {
		config.MaxContentLength = 1e7
	}
	l := &HttpListener{
		config:        config,
		ledgerWriter:  ledgerWriter,
		dataProcessor: dataProcessor,
		router:        mux.NewRouter(),
		auditWriter:   auditWriter,
	}
	if l.config.EnableAudit {
		l.router.Methods("POST").Path("/audit").HandlerFunc(l.Handle)
		l.router.Methods("GET", "POST").Path("/query").HandlerFunc(l.Query)
		l.router.Methods("GET", "POST").Path("/queryGrammar").HandlerFunc(l.QueryGrammar)
	}
	l.router.Methods("GET", "POST").Path("/health").HandlerFunc(l.Health)

	return l
}

func (l *HttpListener) Start() {
	go l.doListen()
	logrus.Info("HttpListener started")
}

func (l *HttpListener) Stop() {
	l.stopped = true
	logrus.Info("HttpListener stopped")
}

func (l *HttpListener) Handle(rw http.ResponseWriter, req *http.Request) {
	if req.ContentLength > l.config.MaxContentLength {
		http.Error(rw, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil || len(data) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}
	logrus.Tracef("get audit request data: %s", string(data))

	events, err := l.dataProcessor.ParseCommand(data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	for _, event := range events {
		logrus.Tracef("write event to ledger: %s", event.String())
		err = l.ledgerWriter.EnqueueSendToLedger(event)
		if err != nil {
			logrus.WithError(err).Warn("send to ledger err")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}

	logrus.Tracef("write to ledger ends, data: %s", events[0].PrimaryKey)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("{}"))

}

func (l *HttpListener) Health(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("ok"))
}

func (l *HttpListener) doListen() {
	logrus.Fatal(http.ListenAndServe(":"+fmt.Sprintf("%d", l.config.Port), l.router))
}
