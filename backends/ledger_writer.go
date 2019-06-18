package backends

type LedgerWriter interface {
	SendToLedger(data string)
}
