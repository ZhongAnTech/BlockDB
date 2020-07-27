package core_interface

// BlockDBCommand is the raw data operation applied on ledger. no additional info
type BlockDBCommand interface {
}

// BLockDBMessage is the enriched message including BlockDBCommand.
type BlockDBMessage interface {
}

type BlockDBCommandProcessor interface {
	Process(command BlockDBCommand) (CommandProcessResult, error) // better to be implemented in async way.

}

type JsonCommandParser interface {
	FromJson(json string) (BlockDBCommand, error)
}

type BlockchainOperator interface {
	EnqueueSendToLedger(command BlockDBMessage) error
}

type CommandExecutor interface{}

type StorageExecutor interface{}

type LedgerSyncer interface{}

type BlockchainListener interface{}
