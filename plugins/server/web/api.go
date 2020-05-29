package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

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
	filter := request.ToFilter()
	if request.PageNum < 1 {
		request.PageNum = 1
	}
	if request.PageSize < 1 {
		request.PageNum = 10
	}
	skip := (request.PageNum - 1) * request.PageSize
	respData, total, err := l.auditWriter.Query(filter, request.PageSize, skip)
	if err != nil {
		logrus.WithError(err).Error("read failed")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp AuditDataQueryResponse
	resp.Total = total
	resp.Data = respData
	RespOk(rw, resp)
	return

}

func (l *HttpListener) QueryOriginal(rw http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil || len(data) == 0 {
		http.Error(rw, "miss content", http.StatusBadRequest)
		return
	}
	var request OriginalDataRequest
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
	originalData, total, err := l.auditWriter.GetOriginalDataProcessor().Query(filter, request.PageSize, skip)
	if err != nil {
		logrus.WithError(err).Error("read failed")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp OriginalDataQueryResponse
	resp.Total = total
	resp.Data = originalData
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
	respData, total, err := l.auditWriter.Query(filter, request.PageSize, skip)
	if err != nil {
		logrus.WithError(err).Error("read failed")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp AuditDataQueryResponse
	resp.Total = total
	resp.Data = respData
	RespOk(rw, resp)
	return
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
