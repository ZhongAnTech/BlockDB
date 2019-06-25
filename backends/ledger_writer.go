package backends

type LedgerWriter interface {
	EnqueueSendToLedger(data interface{})
}
