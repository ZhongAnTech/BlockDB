package processors

import "net"

type ConnectionProcessor interface {
	// handleConnection reads the connection and extract the incoming message
	// note that this may be a long connection so take care of the connection reuse.
	ProcessConnection(conn net.Conn) error
	// Do not return until ready
	Start()
	Stop()
}

type DataProcessor interface {
	ParseCommand(bytes []byte) ([]*LogEvent, error)
}
