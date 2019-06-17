package processors

import "net"

type Processor interface {
	// handleConnection reads the connection and extract the incoming message
	// note that this may be a long connection so take care of the connection reuse.
	ProcessConnection(conn net.Conn) error
	// Do not return until ready
	Start()
	Stop()
}
