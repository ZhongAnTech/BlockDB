package multiplexer

import "net"

type ConnectionBuilder interface {
	BuildConnection() (net.Conn, error)
}
type DefaultTCPConnectionBuilder struct {
	Url string
}

func NewDefaultTCPConnectionBuilder(url string) *DefaultTCPConnectionBuilder {
	return &DefaultTCPConnectionBuilder{
		Url: url,
	}
}

func (b *DefaultTCPConnectionBuilder) BuildConnection() (net.Conn, error) {
	c, err := net.Dial("tcp", b.Url)
	return c, err
}
