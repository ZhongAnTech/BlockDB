package processors

import "net"

type Processor interface {
	// handleConnection reads the connection and extract the incoming message
	ProcessConnection(conn net.Conn)
	// Do not return until ready
	Start()
	Stop()
}
