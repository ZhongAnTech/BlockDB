package web

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpListener_Handle(t *testing.T) {
	httpListener := HttpListener{
		Config: HttpListenerConfig{MaxContentLength: 1e7},
	}
	data := `{
		"op": "insert",
		"collection": "sample_collection",
		"data": {
			"name": "fudan",
			"address": {
				"city": "Shanghai",
				"road": "xxx"
			},
			"logo": {
				"url": "http://a.png"
			},
			"teachers": [
				"T1", "T2", "T3"
			]
		}
    }`
	pri, pub, _ := crypto.GenerateSecp256k1Key(rand.Reader)
	pubBytes, _ := pub.Raw()
	dataBytes := []byte(Normalize(data))
	hash := sha256.Sum256(dataBytes)
	signature, _ := pri.Sign(hash[:])
	message := &Message{
		Data:      dataBytes,
		PublicKey: hex.EncodeToString(pubBytes),
		Signature: hex.EncodeToString(signature),
	}
	msg, _ := json.Marshal(message)
	fmt.Println(string(msg))
	req := httptest.NewRequest(http.MethodPost, "http://url.com", bytes.NewBuffer(msg))
	w := httptest.NewRecorder()
	httpListener.Handle(w, req)
	resp := w.Result()
	c, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(c))
}
