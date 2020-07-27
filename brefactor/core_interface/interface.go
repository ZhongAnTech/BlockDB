package core_interface

type BlockDBCommand interface {
}

type BlockDBCommandProcessor interface {
	Process(command BlockDBCommand) (CommandProcessResult, error) // better to be implemented in async way.

}

type JsonCommandParser interface {
	FromJson(json string) (BlockDBCommand, error)
}
