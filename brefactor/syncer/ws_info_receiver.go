package syncer

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"strings"
)



type WebsocketInfoReceiver struct {
	WebsocketUrl string
	HeightChan   chan int64
	quit         chan bool
}

func (w WebsocketInfoReceiver) Start() {
	fmt.Println("ws start...")
	w.Connect()
}

func (w WebsocketInfoReceiver) Stop() {
	w.quit <- true
}

func (w WebsocketInfoReceiver) Name() string {
	return "WebsocketInfoReceiver"
}

func (w WebsocketInfoReceiver) Connect() {
	dialer := websocket.Dialer{}
	connect, _, err := dialer.Dial(w.WebsocketUrl, nil)
	if nil != err {
		log.Print(err)
	}
	for{
		_, messageData, err := connect.ReadMessage()
		if nil != err {
			log.Print(err)
		}
		str := string(messageData)
		fmt.Println(str)
		s1 := strings.Split(str, ",")
		s2 := strings.Split(s1[5], ":")
		height, err := strconv.Atoi(s2[1])
		fmt.Println(height)
		w.EventChannel() <- int64(height)
	}

}

func (w WebsocketInfoReceiver) EventChannel() chan int64 {
	return w.HeightChan
}
