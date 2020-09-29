package syncer

import (
	"context"
	"github.com/ZhongAnTech/BlockDB/brefactor/storage"
	"testing"
)

func TestSyncer(t *testing.T)  {
	ogChaninSyncerConfig := OgChainSyncerConfig{
		LatestHeightUrl: "http://nbstock.top:30010",
		WebsocketUrl: "ws://nbstock.top:30012/ws",
	}

	storageExecutor, err := storage.Connect(context.Background(),"mongodb://localhost:27017", "test", "", "", "" )
	if err != nil {
		t.Error(err.Error())
	}

	ws := WebsocketInfoReceiver{
		WebsocketUrl: ogChaninSyncerConfig.WebsocketUrl,
		HeightChan: make(chan int64,10),
	}

	go ws.Start()
	ogChainSyncer := OgChainSyncer{
		MaxSyncedHeight: 122759,
		SyncerConfig:    ogChaninSyncerConfig,
		StorageExecutor: storageExecutor,
		InfoReceiver:    ws,
		Quit:            nil,
	}



	//type OgChainSyncer struct {
	//	// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
	//	// TODO: receive websocket push (receive latest height) and sync (realtime)
	//
	//	SyncerConfig OgChainSyncerConfig
	//	// table: op
	//	MaxSyncedHeight int64
	//	ChainOperator   ChainOperator
	//	InfoReceiver    ChainEventReceiver
	//	storageExecutor core_interface.StorageExecutor
	//	quit            chan bool
	//}


	ogChainSyncer.Start()
	defer ogChainSyncer.Stop()
}
