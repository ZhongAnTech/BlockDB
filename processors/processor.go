package processors

import "net"

type Processor interface {
	// handleConnection reads the connection and extract the incoming message
	ProcessConnection(conn net.Conn) error
	// Do not return until ready
	Start()
	Stop()
}
