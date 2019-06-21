package multiplexer

import "net"

type DialogContext struct {
	Source net.Conn
	Target net.Conn
}
