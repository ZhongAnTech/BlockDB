package multiplexer

import "net"

type DialogContext struct {
	User   string
	Source net.Conn
	Target net.Conn
}
