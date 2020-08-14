package common

const (
	// URL Annchain.OG RPC URL
	URL = "http://47.100.122.212:30022"
	// DefaultValue 缺省交易额
	DefaultValue = "0"
	// DefaultData 缺省交易数据
	DefaultData = "="
	// DefaultCryptoType 缺省加密类型
	DefaultCryptoType = "secp256k1"
	// DefaultTokenID 缺省安全令牌
	DefaultTokenID = 0

	// NewAccountRPCMethod 新建账户RPC方法
	NewAccountRPCMethod = "new_account"
	// NewTransactionRPCMethod 新建交易RPC方法
	NewTransactionRPCMethod = "new_transaction"
	// QueryNonceRPCMethod 查询nonceRPC方法
	QueryNonceRPCMethod = "query_nonce"
	// QueryTransactionRPCMethod 查询单笔交易RPC方法
	QueryTransactionRPCMethod = "transaction"
	// QueryTransactionsRPCMethod 查询指定高度区块上交易哈希RPC方法
	QueryTransactionsRPCMethod = "transaction_hashes"
	// QuerySequencerRPCMethod 查询区块信息RPC方法
	QuerySequencerRPCMethod = "sequencer"
)
