package syncer

import "testing"

func TestWs(t *testing.T)  {
	ws := WebsocketInfoReceiver{
		WebsocketUrl: "ws://nbstock.top:30012/ws",
		HeightChan: make(chan int64,30),
	}

	ws.Connect()
}
