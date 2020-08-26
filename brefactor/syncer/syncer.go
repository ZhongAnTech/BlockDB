package syncer

type ChainSyncer interface {
	QueryHeight() (int64, error)
}

type OgChainSyncer struct {
}
