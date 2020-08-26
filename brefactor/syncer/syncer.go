package syncer

// ChainOperator is the pulling way to query the latest sequencer and txs.
type ChainOperator interface {
	QueryHeight() (int64, error)
	// TODO: Qingyun enrich interface and implement plugins/clients/og/OgChainOperator
}

// ChainEventReceiver is the pushing way to receive the latest sequencer
type ChainEventReceiver interface {
	Connect()
	EventChannel() chan int64 // Maybe you will pass a more complicate object in the channel
	// TODO: Qingyun enrich interface and implement plugins/clients/og/OgChainEventReceiver

}

type OgChainSyncer struct {
	// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
	// TODO: receive websocket push (receive latest height) and sync (realtime)

	// table: op
	MaxSyncedHeight int64
	ChainOperator   ChainOperator
	InfoReceiver    ChainEventReceiver

	quit chan bool
}

func (o *OgChainSyncer) Start() {
	// load max height from ledger
	go o.loop()
}

func (o *OgChainSyncer) Stop() {
	panic("implement me")
}

func (o *OgChainSyncer) Name() string {
	return "OgChainSyncer"
}

func (o *OgChainSyncer) QueryHeight() (int64, error) {
	panic("implement me")
}

func (o *OgChainSyncer) loop() {
	for {
		select {
		case <-o.quit:
			return
		case newHeight := <-o.InfoReceiver.EventChannel():
			// TODO: compare local max height and sync if behind.
			if newHeight > o.MaxSyncedHeight {
				// TODO: sync.
			}

		//case <- timeout (be very careful when you handle the timer reset to prevent blocking.)
		default:

			// TODO: (priority) pull latest height and sync: startup, every 5 min (in case websocket is down)
			// TODO: receive websocket push (receive latest height) and sync (realtime)
		}
	}
}
