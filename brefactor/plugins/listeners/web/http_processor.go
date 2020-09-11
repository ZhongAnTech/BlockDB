package web

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ZhongAnTech/BlockDB/brefactor/core"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type HttpListenerConfig struct {
	Port              int
	MaxContentLength  int64
	DBActionTimeoutMs int
}

type HttpListener struct {
	JsonCommandParser       core_interface.JsonCommandParser
	BlockDBCommandProcessor core_interface.BlockDBCommandProcessor
	Config                  HttpListenerConfig
	BusinessReader          *core.BusinessReader

	wg      sync.WaitGroup
	stopped bool
	router  *mux.Router
}

type Message struct {
	OpStr     json.RawMessage `json:"op_str"`
	PublicKey string          `json:"public_key"`
	OpHash    string          `json:"op_hash"`
	Signature string          `json:"signature"`
}

func (l *HttpListener) Name() string {
	return "HttpListener"
}

func (l *HttpListener) Setup() {
	if l.Config.MaxContentLength == 0 {
		l.Config.MaxContentLength = 1e7
	}

	l.router = mux.NewRouter()
	l.router.Methods("POST").Path("/audit").HandlerFunc(l.Handle)
	l.router.Methods("POST").Path("/info").HandlerFunc(l.Info)
	l.router.Methods("POST").Path("/actions").HandlerFunc(l.Actions)
	l.router.Methods("POST").Path("/action").HandlerFunc(l.Action)
	l.router.Methods("POST").Path("/values").HandlerFunc(l.Values)
	l.router.Methods("POST").Path("/value").HandlerFunc(l.Value)
	l.router.Methods("POST").Path("/currentValue").HandlerFunc(l.CurrentValue)
	//l.router.Methods("GET", "POST").Path("/query").HandlerFunc(l.Query)
	//l.router.Methods("GET", "POST").Path("/queryGrammar").HandlerFunc(l.QueryGrammar)
	l.router.Methods("GET", "POST").Path("/health").HandlerFunc(l.Health)
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
	if req.ContentLength > l.Config.MaxContentLength {
		http.Error(rw, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}

	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	var message Message
	err = json.Unmarshal(msg, &message)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	pubKeyBytes, err := hex.DecodeString(message.PublicKey)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	pubKey, err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	opHash, err := hex.DecodeString(message.OpHash)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	signatureBytes, err := hex.DecodeString(message.Signature)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	data := Normalize(string(message.OpStr))
	hash := sha256.Sum256([]byte(data))

	if !bytes.Equal(opHash, hash[:]) {
		http.Error(rw, "invalid op_hash", http.StatusBadRequest)
	}

	isSuccess, err := pubKey.Verify(hash[:], signatureBytes)
	if err != nil || !isSuccess {
		http.Error(rw, "invalid signature", http.StatusBadRequest)
		return
	}

	//logrus.Tracef("get audit request data: %s", string(data))
	//command, err := l.JsonCommandParser.FromJson(string(data))
	//
	//if err != nil {
	//	http.Error(rw, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//result, err := l.BlockDBCommandProcessor.Process(command)
	//if err != nil {
	//	logrus.WithError(err).Warn("failed to process command")
	//	http.Error(rw, err.Error(), http.StatusInternalServerError)
	//}
	//
	//logrus.WithField("result", result).Info("process result")
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("{}")) // TODO: write result of BlockDBCommandProcessor.Process

}

func (l *HttpListener) Info(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash string `json:"op_hash"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Info(ctx, input.OpHash)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Actions(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash string `json:"op_hash"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Actions(ctx, input.OpHash)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Action(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash  string `json:"op_hash"`
		Version int    `json:"version"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Action(ctx, input.OpHash, input.Version)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Values(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash string `json:"op_hash"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Values(ctx, input.OpHash)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Value(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash  string `json:"op_hash"`
		Version int    `json:"version"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Value(ctx, input.OpHash, input.Version)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) CurrentValue(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		OpHash string `json:"op_hash"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.CurrentValue(ctx, input.OpHash)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Query(rw http.ResponseWriter, req *http.Request) {
	msg, err := ioutil.ReadAll(req.Body)
	if err != nil || len(msg) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}

	input := struct {
		Filter map[string]interface{} `json:"filter"`
	}{}

	err = json.Unmarshal(msg, &input)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, _ := context.WithTimeout(req.Context(), time.Millisecond*time.Duration(l.Config.DBActionTimeoutMs))

	data, err := l.BusinessReader.Query(ctx, input.Filter)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

func (l *HttpListener) Health(rw http.ResponseWriter, req *http.Request) {
	// TODO: do real health check
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("ok"))
}

func (l *HttpListener) doListen() {
	logrus.WithField("port", l.Config.Port).Info("RPC server listening")
	logrus.Fatal(http.ListenAndServe(":"+fmt.Sprintf("%d", l.Config.Port), l.router))
}
